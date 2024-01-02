//go:build windows

package main

import (
	"fmt"
	"strings"

	win_tool "github.com/doraemonkeys/doraemon/win-tool/device"
)

func main() {
	win_tool.WatchUsbDevice(func(behavior win_tool.UsbDeviceBehavior, lParam uintptr) {
		fmt.Printf("behavior: %s lParam: %x\n", behavior, lParam)
		if behavior == win_tool.Unknown {
			return
		}
		device, err := win_tool.ReadDeviceInfo(lParam)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("device: %+v\n", device)
		if behavior != win_tool.AddDevice {
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
		ltr, err := win_tool.GetVolumeLetterByVolumeGUID(volGUID)
		if err != nil {
			fmt.Println("GetDiskByVolumeGUID error:", err)
			return
		}
		fmt.Println("volume LTR:", ltr)
		fmt.Println("----------------------------------")
	})
	select {}
}
