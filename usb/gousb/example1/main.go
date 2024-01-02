package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/gousb"
)

// Bus: USB设备所在的总线号。
// Address: USB设备的地址。
// Speed: USB设备的速度。数字2通常代表全速（Full Speed，12 Mbps）。
// Port: USB设备所连接的端口号。
// Path: USB设备在USB端口上的路径，作为一个整数数组表示。
// Spec: USB规格版本，0x110代表USB 1.1。
// Device: 设备的版本号，0x100可能表示1.0.0。
// Vendor: 制造商的ID，0x8087是一个特定的供应商代码。
// Product: 产品ID，0x1024是一个特定的产品代码。
// Class: 设备类代码，0x0表示由接口描述的设备类。
// SubClass: 设备子类代码。
// Protocol: 设备协议代码。
// MaxControlPacketSize: 控制传输的最大包大小。
// Configs: 设备的配置描述符，包含有关设备配置的信息。
// 在Configs映射中的gousb.ConfigDesc结构体中，我们有：

// Number: 配置编号。
// SelfPowered: 表明设备是否自供电。
// RemoteWakeup: 表明设备是否支持远程唤醒。
// MaxPower: 设备最大功率消耗，0x64是十六进制表示，等于100毫安培。
// Interfaces数组包含了一个gousb.InterfaceDesc结构体，它描述了USB接口的参数：

// Number: 接口编号。
// AltSettings: 接口备用设置。
// 在AltSettings数组中的gousb.InterfaceSetting结构体中，我们有：

// Number: 接口设置编号。
// Alternate: 备用接口编号。
// Class: 接口类代码。
// SubClass: 接口子类代码。
// Protocol: 接口协议代码。
// Endpoints: 端点描述符的映射。
// 在Endpoints映射中gousb.EndpointDesc结构体描述了端点的参数：

// Address: 端点的地址。
// Number: 端点编号。
// Direction: 端点方向，false可能代表OUT（主机到设备），true代表IN（设备到主机）。
// MaxPacketSize: 端点的最大包大小。
// TransferType: 传输类型，0x2可能代表批量传输（Bulk Transfer）。
// PollInterval: 轮询间隔，对于中断和等时传输很重要。
// IsoSyncType: 等时同步类型。
// UsageType: 使用类型。
// iManufacturer, iProduct, iSerialNumber是字符串描述符的索引，
// 分别对应于制造商名称、产品名称和序列号。

func main() {
	// Initialize a new Context.
	ctx := gousb.NewContext()
	defer ctx.Close()

	fmt.Println("Initialized context")

	count := 0

	ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		fmt.Printf("%#v\n", desc)
		fmt.Println("-------------------------------------")
		count++
		return false
	})

	fmt.Println("Found", count, "devices")

	// Open any device with a given VID/PID using a convenience function.
	//  Vendor:0x3554, Product:0xf58a
	dev, err := ctx.OpenDeviceWithVIDPID(0x3554, 0xf58a)
	if err != nil {
		log.Fatalf("Could not open a device: %v", err)
	}
	defer dev.Close()

	// Claim the default interface using a convenience function.
	// The default interface is always #0 alt #0 in the currently active
	// config.
	intf, done, err := dev.DefaultInterface()
	if err != nil {
		log.Fatalf("%s.DefaultInterface(): %v", dev, err)
	}
	defer done()
	fmt.Printf("Interface: %#v\n", intf)

	dev.SetAutoDetach(true) //设置自动分离内核驱动(已知windows下的大容量存储设备无效)
	fmt.Println("Successfully claimed interface")

	time.Sleep(1000 * time.Second)

	// Open an OUT endpoint.
	ep, err := intf.OutEndpoint(7)
	if err != nil {
		log.Fatalf("%s.OutEndpoint(7): %v", intf, err)
	}

	// Generate some data to write.
	data := make([]byte, 5)
	for i := range data {
		data[i] = byte(i)
	}

	// Write data to the USB device.
	numBytes, err := ep.Write(data)
	if numBytes != 5 {
		log.Fatalf("%s.Write([5]): only %d bytes written, returned error is %v", ep, numBytes, err)
	}
	fmt.Println("5 bytes successfully sent to the endpoint")
}
