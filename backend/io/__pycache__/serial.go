package main

import (
	"fmt"
	"log"
	"time"

	"://github.com"
)

func main() {
	portName := "COM3"
	baudRate := 115200

	options := serial.OpenOptions{
		PortName:        portName,
		BaudRate:        uint(baudRate),
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 1,
	}


	port, err := serial.OpenPort(options)
	if err != nil {
		log.Fatalf("connection error: %v", err)
	}
	defer port.Close()

	time.Sleep(2 * time.Second)
	fmt.Printf("Connected: %s\n", portName)

	for {
		bytesToRead, err := port.BytesToRead()
		if err != nil {
			log.Printf("error: %v", err)
			break
		}


		if bytesToRead > 0 {
			

			buf := make([]byte, bytesToRead)

	
			n, err := port.Read(buf)
			if err != nil {
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
						fmt.Printf("Spos: %d", i)
					}
				}

				fmt.Println("                            	  ")
			}
		}

		time.Sleep(10 * time.Millisecond)
	}
}
