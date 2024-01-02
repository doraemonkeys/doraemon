package main

import (
	"fmt"

	"github.com/younglifestyle/gousb"
)

// 热拔插 但unsupported(windows)
func main() {

	ctx := gousb.NewContext()
	defer ctx.Close()

	h, err := ctx.RegisterHotplug(func(evt gousb.HotplugEvent) {
		if evt.Type() == gousb.HotplugEventDeviceArrived {
			fmt.Println("arrived device", evt.Type())
		} else {
			fmt.Println("un-arrived device", evt.Type())
		}
	})
	if err != nil {
		// panic(err)
		fmt.Println(err)
	}
	h()
	fmt.Println("Initialized context")
}
