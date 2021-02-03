package main

import (
	"errors"
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
	title    = flag.String("title", "", "Chart title")
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
	if psr.RenderModule != "chart" {
		log.Fatalf("Sorry, you can only use the chart renderer; %v isn't supported.", psr.RenderModule)
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

	// Grab the results, requesting only 1 result so it gets condense to a single result
	results, err := c.GetChartTsRange(s, start, end, 0, 1)
	if err != nil {
		log.Fatalf("Failed to get results: %v", err)
	}

	// Print gnuplot preamble
	fmt.Printf("set terminal png\n")
	fmt.Printf("set boxwidth 0.5\n")
	fmt.Printf("set style fill solid\n")
	fmt.Printf("set title \"%v\"\n", *title)
	fmt.Printf("set nokey\n")

	// Generate a here document
	r := results.Entries
	if len(r.Values) != 1 {
		log.Fatalf("Got too many results: %v, expected 1", len(r.Values))
	}
	fmt.Printf("$data << EOD\n")
	for i, v := range r.Names {
		fmt.Printf("%v %v %v\n", i, v, r.Values[0].Data[i])
	}
	fmt.Printf("EOD\n")
	// And issue the plot command
	fmt.Printf("plot $data using 1:3:xtic(2) with boxes\n")
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
