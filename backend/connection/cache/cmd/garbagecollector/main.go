package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"mfep_control_station/backend/connection/cache"
)

func main() {
	root := flag.String("dir", "", "dfiles")
	address := flag.String("listen", "127.0.0.1:8765", "command")
	flag.Parse()

	if *root == "" {
		log.Fatal("the -dir flag is required")
	}
	if err := validateLoopbackAddress(*address); err != nil {
		log.Fatal(err)
	}
	handler, err := cache.NewCommandHandler(*root)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("garbage collector listening on http://%s/collect for %s", *address, *root)
	if err := http.ListenAndServe(*address, handler); err != nil {
		log.Fatal(err)
	}
}

func validateLoopbackAddress(address string) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("invalid listen address %q: %w", address, err)
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("listen address must use a loopback IP")
	}
	return nil
}
