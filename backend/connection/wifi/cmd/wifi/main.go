package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"mfep_control_station/backend/connection/wifi"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usageError()
	}

	switch args[0] {
	case "receive":
		flags := flag.NewFlagSet("receive", flag.ContinueOnError)
		address := flags.String("listen", "127.0.0.1:50000", "TCP address to listen on")
		directory := flags.String("dir", "received", "directory where received files are saved")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 {
			return usageError()
		}
		return wifi.Serve(*address, *directory)
	case "send":
		flags := flag.NewFlagSet("send", flag.ContinueOnError)
		address := flags.String("address", "", "receiver address, for example 192.168.1.20:50000")
		filePath := flags.String("file", "", "JSON or YAML file to send")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 || *address == "" || *filePath == "" {
			return usageError()
		}
		if err := wifi.Send(*address, *filePath); err != nil {
			return err
		}
		fmt.Println("File sent.")
		return nil
	default:
		return usageError()
	}
}

func usageError() error {
	return fmt.Errorf("usage:\n  go run ./backend/connection/wifi/cmd/wifi receive [-listen 127.0.0.1:50000] [-dir received]\n  go run ./backend/connection/wifi/cmd/wifi send -address <host:port> -file <path>")
}
