//go:build windows

package device

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/doraemonkeys/doraemon"
)

type DeviceInfo struct {
	Name string
	Desc string
}

// 列出电脑上全部的设备名称和描述(包括曾经的)。
func DevconListAllDeviceNames(devconPath string) ([]DeviceInfo, error) {
	cmd := exec.Command(devconPath, "find", "*")
	cmd_out := bytes.NewBuffer(nil)
	cmd.Stdout = cmd_out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	// var buf = make([]byte, 1024)
	var deviceInfos []DeviceInfo
	for {
		line, err := cmd_out.ReadBytes('\n')
		if err != nil {
			// 最后一行：No matching devices found./272 matching device(s) found.
			break
		}
		nameDesc := strings.Split(string(line), ":")
		if len(nameDesc) < 2 {
			continue
		}
		deviceInfos = append(deviceInfos, DeviceInfo{
			Name: strings.TrimSpace(nameDesc[0]),
			Desc: strings.TrimSpace(strings.Join(nameDesc[1:], ":")),
		})
	}
	return deviceInfos, nil
}

// 列出所有蓝牙设备(包括曾经的)
func DevconListAllBluetoothDeviceNames(devconPath string) ([]DeviceInfo, error) {
	allDevices, err := DevconListAllDeviceNames(devconPath)
	if err != nil {
		return nil, err
	}
	var bluetoothDevices []DeviceInfo
	for _, device := range allDevices {
		name := strings.ToLower(device.Name)
		desc := strings.ToLower(device.Desc)
		if strings.Contains(name, "bluetooth") ||
			strings.Contains(desc, "bluetooth") {
			bluetoothDevices = append(bluetoothDevices, device)
			continue
		}
		if strings.HasPrefix(name, "bth") {
			bluetoothDevices = append(bluetoothDevices, device)
			continue
		}
	}
	return bluetoothDevices, nil
}

// 列出所有的USB设备名称
func DevconListAllUsbDeviceNames(devconPath string) ([]DeviceInfo, error) {
	allDevices, err := DevconListAllDeviceNames(devconPath)
	if err != nil {
		return nil, err
	}
	var usbDevices []DeviceInfo
	for _, device := range allDevices {
		if strings.HasPrefix(device.Name, "USB") {
			usbDevices = append(usbDevices, device)
			continue
		}
	}
	return usbDevices, nil
}

// disable后需要Enable才能恢复。
// deviceName: *VID_8087&PID_1024*
func DevconDisableDevice(devconPath string, deviceName string) (int, error) {
	cmd := exec.Command(devconPath, "disable", deviceName)
	cmd_out := bytes.NewBuffer(nil)
	cmd.Stdout = cmd_out
	err := cmd.Run()
	if err != nil {
		return 0, err
	}
	// 1 device(s) were disabled.
	lines, err := doraemon.ReadLines(cmd_out)
	if err != nil {
		return 0, err
	}
	if len(lines) == 0 {
		return 0, fmt.Errorf("unexpected cmd output: %s", lines)
	}
	line := strings.TrimSpace(string(lines[len(lines)-1]))
	if !strings.Contains(line, "disabled") {
		return 0, fmt.Errorf("unexpected cmd output: %s", lines)
	}
	return len(lines) - 1, nil
}

// deviceName: *VID_8087&PID_1024*
func DevconEnableDevice(devconPath string, deviceName string) (int, error) {
	cmd := exec.Command(devconPath, "enable", deviceName)
	cmd_out := bytes.NewBuffer(nil)
	cmd.Stdout = cmd_out
	err := cmd.Run()
	if err != nil {
		return 0, err
	}
	// 1 device(s) were enabled.
	lines, err := doraemon.ReadLines(cmd_out)
	if err != nil {
		return 0, err
	}
	if len(lines) == 0 {
		return 0, fmt.Errorf("unexpected cmd output: %s", lines)
	}
	line := strings.TrimSpace(string(lines[len(lines)-1]))
	if !strings.Contains(line, "enabled") {
		return 0, fmt.Errorf("unexpected cmd output: %s", lines)
	}
	return len(lines) - 1, nil
}

// 重新插拔后即可恢复，需要管理员。
// deviceName: *VID_8087&PID_1024*
func DevconRemoveDevice(devconPath string, deviceName string) (int, error) {
	cmd := exec.Command(devconPath, "remove", deviceName)
	cmd_out := bytes.NewBuffer(nil)
	cmd.Stdout = cmd_out
	err := cmd.Run()
	if err != nil {
		return 0, err
	}
	// 1 device(s) were removed.
	lines, err := doraemon.ReadLines(cmd_out)
	if err != nil {
		return 0, err
	}
	if len(lines) == 0 {
		return 0, fmt.Errorf("unexpected cmd output: %s", lines)
	}
	line := strings.TrimSpace(string(lines[len(lines)-1]))
	if !strings.Contains(line, "removed") {
		return 0, fmt.Errorf("unexpected cmd output: %s", lines)
	}
	return len(lines) - 1, nil
}
