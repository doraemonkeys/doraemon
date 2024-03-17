//go:build windows

package device

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/doraemonkeys/doraemon"
	"github.com/doraemonkeys/doraemon/encode"
)

type WmicDeviceInfo struct {
	Name        string
	Caption     string
	DeviceID    string
	PNPDeviceID string
	Status      string
}

// 物理存储设备的个数(不包括分区)
func WmicGetPhysicalDiskCount() (int, error) {
	cmd := exec.Command("wmic", "diskdrive")
	cmd_out := bytes.NewBuffer(nil)
	cmd.Stdout = cmd_out
	err := cmd.Run()
	if err != nil {
		return 0, err
	}
	lines, err := doraemon.ReadLines(cmd_out)
	if err != nil {
		return 0, err
	}
	count := 0
	for i := 0; i < len(lines); i++ {
		if strings.TrimSpace(string(lines[i])) != "" {
			count++
		}
	}
	return count - 1, nil
}

// 枚举系统上已连接的设备
func PnputilEnumConnectedDevices() ([][]doraemon.Pair[string, string], error) {
	cmd := exec.Command("pnputil", "/enum-devices", "/connected")
	cmd_out := bytes.NewBuffer(nil)
	cmd.Stdout = cmd_out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	lines, err := doraemon.ReadTrimmedLines(cmd_out)
	if err != nil {
		return nil, err
	}
	if len(lines) < 2 {
		return nil, fmt.Errorf("unexpected cmd output: %s", lines)
	}
	lines = lines[2:]
	var devices [][]doraemon.Pair[string, string]
	var device []doraemon.Pair[string, string]
	for i := 0; i < len(lines); i++ {
		lines[i] = encode.GbkToUtf8(lines[i])
		line := strings.TrimSpace(string(lines[i]))
		if strings.TrimSpace(line) == "" && device != nil {
			devices = append(devices, device)
			device = nil
			continue
		}
		keyVal := strings.Split(line, ":")
		if len(keyVal) < 2 {
			return nil, fmt.Errorf("invalid line: %s", line)
		}
		device = append(device, doraemon.Pair[string, string]{
			First:  strings.TrimSpace(keyVal[0]),
			Second: strings.TrimSpace(strings.Join(keyVal[1:], ":")),
		})
	}
	if device != nil {
		devices = append(devices, device)
	}
	return devices, nil
}

func GetVolumeLetterByVolumeGUID(guid string) (string, error) {
	devices, err := PnputilEnumConnectedDevices()
	if err != nil {
		return "", err
	}
	// fmt.Println(devices)
	for _, device := range devices {
		if len(device) < 2 {
			continue
		}
		if strings.Contains(device[0].Second, guid) &&
			len(device[1].Second) == 3 &&
			device[1].Second[2] == '\\' {
			return device[1].Second[:2], nil
		}
	}
	return "", fmt.Errorf("not found volume: %s", guid)
}

// 删除不在系统中的、卷的装入点目录和注册表设置。
func MountvolRemoveUnusedMountPoint() error {
	cmd := exec.Command("mountvol", "/R")
	cmd_out := bytes.NewBuffer(nil)
	cmd.Stdout = cmd_out
	err := cmd.Run()
	if err != nil {
		return err
	}
	content := encode.GbkToUtf8(cmd_out.Bytes())
	if bytes.Contains(content, []byte("拒绝")) ||
		bytes.Contains(content, []byte("refused")) {
		return errors.New(strings.TrimSpace(string(content)))
	}
	return nil
}

// 删除指定的卷的装入点目录和注册表设置。
// 仅限于此U盘，重新插入也不会自动装载。
// 恢复使用diskpart assign
func MountvolRemoveMountPoint(volumeLetter byte) error {
	cmd := exec.Command("mountvol", fmt.Sprintf("%c: /D", volumeLetter))
	cmd_out := bytes.NewBuffer(nil)
	cmd.Stdout = cmd_out
	err := cmd.Run()
	if err != nil {
		return err
	}
	content := encode.GbkToUtf8(cmd_out.Bytes())
	if bytes.Contains(content, []byte("拒绝")) ||
		bytes.Contains(content, []byte("refused")) {
		return errors.New(strings.TrimSpace(string(content)))
	}
	return nil
}

// 禁用新卷的自动装入。
func MountvolDisableAutoMount() error {
	cmd := exec.Command("mountvol", "/N")
	cmd_out := bytes.NewBuffer(nil)
	cmd.Stdout = cmd_out
	err := cmd.Run()
	if err != nil {
		return err
	}
	content := encode.GbkToUtf8(cmd_out.Bytes())
	if bytes.Contains(content, []byte("拒绝")) ||
		bytes.Contains(content, []byte("refused")) {
		return errors.New(strings.TrimSpace(string(content)))
	}
	return nil
}

// 再次启用新卷的自动装入。
func MountvolEnableAutoMount() error {
	cmd := exec.Command("mountvol", "/E")
	cmd_out := bytes.NewBuffer(nil)
	cmd.Stdout = cmd_out
	err := cmd.Run()
	if err != nil {
		return err
	}
	content := encode.GbkToUtf8(cmd_out.Bytes())
	if bytes.Contains(content, []byte("拒绝")) ||
		bytes.Contains(content, []byte("refused")) {
		return errors.New(strings.TrimSpace(string(content)))
	}
	return nil
}
