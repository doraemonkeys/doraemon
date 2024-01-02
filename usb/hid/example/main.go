package main

import (
	"fmt"

	"github.com/doraemonkeys/doraemon/usb/hid"
)

func main() {

	allDevices := hid.Devices()
	for dev := range allDevices {
		fmt.Printf("%+v\n", dev)
	}

	fmt.Println("-------------------------------")

	infoCh := hid.FindDevices(0x0781, 0x558b)
	dev := <-infoCh
	if dev == nil {
		fmt.Println("Can't find HID device")
		return
	}

	fmt.Println(dev)

	hdev, err := dev.Open()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(hdev)
	defer hdev.Close()
}
