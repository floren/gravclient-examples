package main

import (
	"errors"
	"encoding/csv"
	"os"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gravwell/gravwell/v3/client"
	"github.com/gravwell/gravwell/v3/client/objlog"
	"github.com/gravwell/gravwell/v3/client/types"
)

var (
	server   = flag.String("s", "", "Address and port of Gravwell webserver")
	username = flag.String("u", "admin", `Username (default "admin")`)
	useHttps = flag.Bool("https", false, "Set true to use https")
	password = flag.String("p", "changeme", `Password (default "changeme")`)
	query    = flag.String("q", "", "Query string")
	duration = flag.String("d", "-1h", "Search duration")
	count = flag.Int("c", 10, "Count of results to fetch")

	useTable bool
)

func main() {
	flag.Parse()

	// First things first: make sure they gave us a valid duration
	start, end, err := getTimeDuration(*duration)
	if err != nil {
		log.Fatal(err)
	}

	// For simplicity, we ignore cert validity (self-signed ok)
	c, err := client.NewClient(*server, false, *useHttps, &objlog.NilObjLogger{})
	if err != nil {
		log.Fatalf("NewClient failed: %v", err)
	}
	defer c.Close()

	// Log in
	if err = c.Login(*username, *password); err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	// Call Sync to update client internal data
	if err = c.Sync(); err != nil {
		log.Fatalf("Sync failed: %v", err)
	}

	// Parse the search and make sure they're using one of the basic renderers
	psr, err := c.ParseSearchWithResponse(*query, []types.FilterRequest{})
	if err != nil {
		log.Fatalf("Parse request failed: %v", err)
	}
	if psr.RenderModule != "text" && psr.RenderModule != "raw" && psr.RenderModule != "hex" && psr.RenderModule != "table" {
		log.Fatalf("Sorry, you can only use the text, raw, table, or hex renderers; %v isn't supported.", psr.RenderModule)
	}
	if psr.RenderModule == "table" {
		useTable = true
	}

	// Now start the search
	s, err := c.StartSearch(*query, start, end, false)
	if err != nil {
		log.Fatalf("Failed to start search: %v", err)
	}

	// Wait for the search to be completed
	if err = c.WaitForSearch(s); err != nil {
		log.Fatalf("Failed to wait for search: %v", err)
	}

	if !useTable {
		results, err := c.GetTextResults(s, 0, uint64(*count))
		if err != nil {
			log.Fatalf("Failed to get results: %v", err)
		}
		for _, r := range results.Entries {
			fmt.Println(string(r.Data))
		}
	} else {
		results, err := c.GetTableResults(s, 0, uint64(*count))
		if err != nil {
			log.Fatalf("Failed to get results: %v", err)
		}
		wtr := csv.NewWriter(os.Stdout)
		wtr.Write(results.Entries.Columns)
		for _, r := range results.Entries.Rows {
			wtr.Write(r.Row)
		}
		wtr.Flush()
	}
}

func getTimeDuration(trs string) (start time.Time, end time.Time, err error) {
	trs = strings.TrimRight(strings.TrimLeft(trs, " \t"), " \t\n")
	if len(trs) == 0 {
		err = errors.New("Invalid duration")
		return
	}
	if trs[0] != '-' {
		err = errors.New("Durations must be negative")
		return
	}
	var dur time.Duration
	dur, err = time.ParseDuration(trs)
	if err != nil {
		return
	}
	//we have a legit duration, lets make some start and end times
	end = time.Now()
	start = end.Add(dur)
	return
}
