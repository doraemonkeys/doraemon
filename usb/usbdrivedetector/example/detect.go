package main

import (
	"fmt"

	"github.com/doraemonkeys/doraemon/usb/usbdrivedetector"
)

func main() {
	if drives, err := usbdrivedetector.Detect(); err == nil {
		fmt.Printf("%d USB Devices Found\n", len(drives))
		for _, d := range drives {
			fmt.Println(d)
		}
	} else {
		fmt.Println(err)
	}
}
