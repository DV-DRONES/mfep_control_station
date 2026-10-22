package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tarm/serial"
)

func main() {
	portName := "COM3"
	baudRate := 115200

	c := &serial.Config{
		Name:        portName,
		Baud:        baudRate,
		ReadTimeout: time.Millisecond * 100,
	}

	port, err := serial.OpenPort(c)
	if err != nil {
		log.Fatalf("connection error: %v", err)
	}
	defer port.Close()

	time.Sleep(2 * time.Second)
	fmt.Printf("Connected: %s\n", portName)

	buf := make([]byte, 128)

	for {
		n, err := port.Read(buf)
		if err != nil {

			if n == 0 {
				continue
			}
			log.Printf("reading error: %v", err)
			break
		}

		if n > 0 {
			rawPayload := buf[:n]

			fmt.Printf("received: %d Bytes\n", len(rawPayload))

			fmt.Print(" [")
			for i, b := range rawPayload {
				fmt.Printf("0x%02X", b)
				if i < len(rawPayload)-1 {
					fmt.Print(", ")
				}
			}
			fmt.Println("]")

			for i, byteValue := range rawPayload {
				if byteValue == 0xAA {
					fmt.Printf("Spos: %d ", i)
				}
			}
			fmt.Println()
		}
	}
}
