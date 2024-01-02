package main

import (
	"fmt"

	winDevice "github.com/doraemonkeys/doraemon/win-tool/device"
)

func main() {
	devices, err := winDevice.DevconListAllUsbDeviceNames("./devcon.exe")
	if err != nil {
		panic(err)
	}
	for _, device := range devices {
		fmt.Println(device)
	}
	fmt.Printf("total: %d\n", len(devices))

	fmt.Println("----------------------------------")

	fmt.Println(winDevice.PnputilEnumConnectedDevices())
	fmt.Println("----------------------------------")
	dpdi, err := winDevice.DiskPartListDisk()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", dpdi)
	dpdd, err := winDevice.DiskPartDetailDisk(0)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", dpdd)
	fmt.Println("----------------------------------")
	fmt.Println(winDevice.DiskPartIsVolumeReadOnly('F'))
	fmt.Println(winDevice.DiskPartIsVolumeReadOnly('G'))
	fmt.Println(winDevice.DiskPartIsDiskReadOnly(2))
	fmt.Println(winDevice.DiskPartSwitchDiskReadOnly(20))
	fmt.Println(winDevice.DiskPartIsDiskReadOnly(2))
}
