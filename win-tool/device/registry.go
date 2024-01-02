//go:build windows

package device

import (
	"golang.org/x/sys/windows/registry"
)

// 需要管理员权限， 可能需要重启
func RegistryDisableAllUsb() error {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Services\USBSTOR`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	err = key.SetDWordValue("Start", 4)
	if err != nil {
		return err
	}
	return nil
}

// 需要管理员权限， 可能需要重启
func RegistryEnableAllUsb() error {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Services\USBSTOR`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	err = key.SetDWordValue("Start", 3)
	if err != nil {
		return err
	}
	return nil
}

// 需要管理员权限， 可能需要重启
//
// 0x00000004(4) 禁用
// 0x00000003(3) 启用
// 0x00000002(2) 自动
// 0x00000001(1) 系统
func RegistryGetAllUsbStatus() (uint, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Services\USBSTOR`, registry.QUERY_VALUE)
	if err != nil {
		return 0, err
	}
	defer key.Close()

	status, _, err := key.GetIntegerValue("Start")
	if err != nil {
		return 0, err
	}
	return uint(status), nil
}

// 设置全部usb设备为只读, 需要管理员权限
func RegistrySetAllUsbReadOnly() error {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\StorageDevicePolicies`, registry.SET_VALUE)
	if err != nil {
		if err != registry.ErrNotExist {
			return err
		}
		// 如果键不存在，则创建它
		key, _, err = registry.CreateKey(registry.LOCAL_MACHINE,
			`SYSTEM\CurrentControlSet\Control\StorageDevicePolicies`,
			registry.SET_VALUE)
		if err != nil {
			return err
		}
	}
	defer key.Close()

	err = key.SetDWordValue("WriteProtect", 1)
	if err != nil {
		return err
	}
	return nil
}

// 设置全部usb设备为可读写, 需要管理员权限
func RegistrySetAllUsbReadWrite() error {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\StorageDevicePolicies`, registry.SET_VALUE)
	if err != nil {
		if err != registry.ErrNotExist {
			return err
		}
		// 如果键不存在，则创建它
		key, _, err = registry.CreateKey(registry.LOCAL_MACHINE,
			`SYSTEM\CurrentControlSet\Control\StorageDevicePolicies`,
			registry.SET_VALUE)
		if err != nil {
			return err
		}
	}
	defer key.Close()

	err = key.SetDWordValue("WriteProtect", 0)
	if err != nil {
		return err
	}
	return nil
}

// 返回状态(只读/可读写), key是否存在, error。
// 需要管理员权限
func RegistryGetAllUsbReadOnly() (bool, bool, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\StorageDevicePolicies`, registry.QUERY_VALUE)
	if err != nil {
		if err != registry.ErrNotExist {
			return false, false, err
		}
		return false, false, nil
	}
	defer key.Close()

	status, _, err := key.GetIntegerValue("WriteProtect")
	if err != nil {
		return false, true, err
	}
	if status == 1 {
		return true, true, nil
	}
	return false, true, nil
}

// 检查在 Windows 系统上曾经挂载过的 USB 设备的数量
func GetUSBHistoryQuantity() (num int, err error) {
	Opened_Key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\ControlSet001\Enum\USBSTOR`, registry.QUERY_VALUE)
	if err != nil {
		return 0, err
	}
	defer Opened_Key.Close()

	keyInfo, err := Opened_Key.Stat()
	if err == nil {
		return int(keyInfo.SubKeyCount), nil
	}
	return 0, err
}

// USB物理存储设备的个数(不包括分区)
func GetUsbPhysicalDiskCount() (int, error) {
	//查询注册表，判断是否插入U盘
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services\USBSTOR\Enum`, registry.QUERY_VALUE)
	if err != nil {
		return 0, err
	}
	defer k.Close()
	// 获取注册表中值，得到插入了几个U盘
	count, _, err := k.GetIntegerValue("Count")
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
