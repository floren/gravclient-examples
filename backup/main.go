package main

import (
	"flag"
	"log"
	"os"

	"github.com/gravwell/gravwell/v3/client"
	"github.com/gravwell/gravwell/v3/client/objlog"
)

var (
	server   = flag.String("s", "", "Address and port of Gravwell webserver")
	username = flag.String("u", "admin", `Username (default "admin")`)
	useHttps = flag.Bool("https", false, "Set true to use https")
	password = flag.String("p", "changeme", `Password (default "changeme")`)
	output   = flag.String("f", "gravwell.bak", "Output file path")
)

func main() {
	flag.Parse()

	f, err := os.Create(*output)
	if err != nil {
		log.Fatalf("Couldn't create output file: %v", err)
	}
	defer f.Close()

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

	// Now fire the backup
	if err = c.Backup(f, true); err != nil {
		log.Fatalf("Backup failed: %v", err)
	}
}
