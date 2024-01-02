package main

import (
	"fmt"
	"log"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

func main() {

	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		fmt.Println("No serial ports found!")
		return
	}

	for _, port := range ports {
		fmt.Printf("Found port: %s\n", port.Name)
		if port.IsUSB {
			fmt.Printf("   USB ID     %s:%s\n", port.VID, port.PID)
			fmt.Printf("   USB serial %s\n", port.SerialNumber)
		}
	}

	fmt.Println("------------------------------------")

	ports2, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	for _, port := range ports2 {
		fmt.Printf("Found port: %v\n", port)
	}

	// output:

	// Found port: COM4
	// Found port: COM3
	// Found port: COM5
	// Found port: COM6
	// ------------------------------------
	// Found port: COM3
	// Found port: COM4
	// Found port: COM5
	// Found port: COM6
}
