//go:build windows

package main

import (
	"fmt"
	"strings"

	winDevice "github.com/doraemonkeys/doraemon/win-tool/device"
)

func main() {
	winDevice.WatchUsbDevice(func(behavior winDevice.UsbDeviceBehavior, lParam uintptr) {
		fmt.Printf("behavior: %s lParam: %x\n", behavior, lParam)
		if behavior == winDevice.Unknown {
			return
		}
		device, err := winDevice.ReadDeviceInfo(lParam)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("device: %+v\n", device)
		if behavior != winDevice.AddDevice {
			return
		}
		if !strings.Contains(device.Name, "Volume") {
			return
		}
		names := strings.Split(device.Name, "#")
		volGUID := ""
		for _, name := range names {
			if strings.HasPrefix(name, "{") {
				volGUID = name
				break
			}
		}
		fmt.Println("volume GUID:", volGUID)
		ltr, err := winDevice.GetVolumeLetterByVolumeGUID(volGUID)
		if err != nil {
			fmt.Println("GetDiskByVolumeGUID error:", err)
			return
		}
		fmt.Println("volume LTR:", ltr)
		fmt.Println("----------------------------------")
	})
	select {}
}
