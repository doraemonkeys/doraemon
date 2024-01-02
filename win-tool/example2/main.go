package main

import (
	"fmt"

	win_tool "github.com/doraemonkeys/doraemon/win-tool/device"
)

func main() {
	devices, err := win_tool.DevconListAllUsbDeviceNames("./devcon.exe")
	if err != nil {
		panic(err)
	}
	for _, device := range devices {
		fmt.Println(device)
	}
	fmt.Printf("total: %d\n", len(devices))

	fmt.Println("----------------------------------")

	fmt.Println(win_tool.PnputilEnumConnectedDevices())
	fmt.Println("----------------------------------")
	dpdi, err := win_tool.DiskPartListDisk()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", dpdi)
	dpdd, err := win_tool.DiskPartDetailDisk(0)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", dpdd)
	fmt.Println("----------------------------------")
	fmt.Println(win_tool.DiskPartIsVolumeReadOnly('F'))
	fmt.Println(win_tool.DiskPartIsVolumeReadOnly('G'))
	fmt.Println(win_tool.DiskPartIsDiskReadOnly(2))
	fmt.Println(win_tool.DiskPartSwitchDiskReadOnly(20))
	fmt.Println(win_tool.DiskPartIsDiskReadOnly(2))
}
