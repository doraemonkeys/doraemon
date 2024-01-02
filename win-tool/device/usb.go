//go:build windows

package device

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	// https://docs.microsoft.com/en-us/windows/win32/devio/wm-devicechange
	// https://godoc.org/github.com/AllenDang/w32#WM_DEVICECHANGE
	WM_DEVICECHANGE = 537

	// https://godoc.org/github.com/AllenDang/w32#HWND_MESSAGE
	HWND_MESSAGE = ^uintptr(2)

	// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-registerdevicenotificationa#device_notify_all_interface_classes
	DEVICE_NOTIFY_ALL_INTERFACE_CLASSES = 4

	// https://docs.microsoft.com/en-us/windows/win32/api/dbt/ns-dbt-_dev_broadcast_hdr#DBT_DEVTYP_DEVICEINTERFACE
	DBT_DEVTYP_DEVICEINTERFACE = 5

	// https://docs.microsoft.com/en-us/windows/win32/devio/wm-devicechange#DBT_DEVICEARRIVAL
	DBT_DEVICEARRIVAL = 0x8000

	// https://docs.microsoft.com/en-us/windows/win32/devio/wm-devicechange#DBT_DEVICEREMOVECOMPLETE
	DBT_DEVICEREMOVECOMPLETE = 0x8004
)

var (
	user32                      = syscall.NewLazyDLL("user32.dll")
	kernel32                    = syscall.NewLazyDLL("kernel32.dll")
	pDefWindowProc              = user32.NewProc("DefWindowProcW")
	pCreateWindowEx             = user32.NewProc("CreateWindowExW")
	pGetModuleHandle            = kernel32.NewProc("GetModuleHandleW")
	pRegisterClassEx            = user32.NewProc("RegisterClassExW")
	pGetMessage                 = user32.NewProc("GetMessageW")
	pDispatchMessage            = user32.NewProc("DispatchMessageW")
	pRegisterDeviceNotification = user32.NewProc("RegisterDeviceNotificationW")
)

// https://www.lifewire.com/device-class-guids-for-most-common-types-of-hardware-2619208
// 745A17A0-74D3-11D0-B6FE-00A0C90F57DA
var HID_DEVICE_CLASS = windows.GUID{
	Data1: 0x745a17a0,
	Data2: 0x74d3,
	Data3: 0x11d0,
	Data4: [8]byte{0xb6, 0xfe, 0x00, 0xa0, 0xc9, 0x0f, 0x57, 0xda},
}

// https://docs.microsoft.com/en-us/windows-hardware/drivers/install/guid-devinterface-usb-device
// A5DCBF10-6530-11D2-901F-00C04FB951ED
var GUID_DEVINTERFACE_USB_DEVICE = windows.GUID{
	Data1: 0xa5dcbf10,
	Data2: 0x6530,
	Data3: 0x11d2,
	Data4: [8]byte{0x90, 0x1f, 0x00, 0xc0, 0x4f, 0xb9, 0x51, 0xed},
}

// https://docs.microsoft.com/en-us/previous-versions//dd162805(v=vs.85)
type POINT struct {
	x uintptr
	y uintptr
}

// https://docs.microsoft.com/en-us/windows/win32/api/winuser/ns-winuser-tagmsg
type MSG struct {
	hWnd    syscall.Handle
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      POINT
}

// https://docs.microsoft.com/en-us/previous-versions/aa373931(v=vs.80)
// type GUID struct {
// 	Data1 uint32
// 	Data2 uint16
// 	Data3 uint16
// 	Data4 [8]byte
// }

// https://docs.microsoft.com/en-us/windows/win32/api/dbt/ns-dbt-_dev_broadcast_deviceinterface_a
type DevBroadcastDevinterface struct {
	dwSize       uint32
	dwDeviceType uint32
	dwReserved   uint32
	classGuid    windows.GUID
	szName       uint16
}

// https://docs.microsoft.com/en-us/windows/win32/api/winuser/ns-winuser-tagwndclassexa
// https://golang.org/src/runtime/syscall_windows_test.go
type Wndclassex struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   syscall.Handle
	Icon       syscall.Handle
	Cursor     syscall.Handle
	Background syscall.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     syscall.Handle
}

// https://docs.microsoft.com/en-us/previous-versions/windows/desktop/legacy/ms633573(v=vs.85)
func WndProc(hWnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	// TODO handle WM_DESTROY and deregister the hwnd class
	case WM_DEVICECHANGE:
		switch wParam {
		case uintptr(DBT_DEVICEARRIVAL):
			innerCallback(AddDevice, lParam)
		case uintptr(DBT_DEVICEREMOVECOMPLETE):
			innerCallback(RemoveDevice, lParam)
		}
		return 0
	default:
		ret, _, _ := pDefWindowProc.Call(uintptr(hWnd), uintptr(msg), uintptr(wParam), uintptr(lParam))
		innerCallback(Unknown, lParam)
		return ret
	}
}

type UsbDeviceBehavior int

const (
	AddDevice UsbDeviceBehavior = iota
	RemoveDevice
	Unknown
)

func (behavior UsbDeviceBehavior) String() string {
	switch behavior {
	case AddDevice:
		return "AddDevice"
	case RemoveDevice:
		return "RemoveDevice"
	case Unknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}

var innerCallback func(UsbDeviceBehavior, uintptr)

// 监控usb设备插拔。
// 代码修改自：https://github.com/jake-dog/opensimdash
func WatchUsbDevice(callback func(UsbDeviceBehavior, uintptr)) error {
	// TODO clean this up a bit
	// The whole thing needs to be run in a single scope/closure otherwise golang
	// will GC all the structs and the message window will not work.

	if innerCallback != nil {
		return fmt.Errorf("callback already set")
	}
	innerCallback = callback
	// Create callback
	cb := syscall.NewCallback(WndProc)
	mh, _, _ := pGetModuleHandle.Call(0)

	// Create a class and window name
	lpClassName, _ := syscall.UTF16PtrFromString("opensimdash")
	lpWindowName, _ := syscall.UTF16PtrFromString("opensimdash")

	// Register our invisible window class
	// Code from: https://golang.org/src/runtime/syscall_windows_test.go
	wc := Wndclassex{
		WndProc:   cb,
		Instance:  syscall.Handle(mh),
		ClassName: lpClassName,
	}
	wc.Size = uint32(unsafe.Sizeof(wc))
	a, _, err := pRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if a == 0 {
		return fmt.Errorf("RegisterClassEx failed: %v", err)
	}

	// Create a message only window
	// https://docs.microsoft.com/en-us/windows/win32/winmsg/window-features#message-only-windows
	// https://stackoverflow.com/a/4081383
	ret, _, err := pCreateWindowEx.Call(
		uintptr(0),                            //dwExStyle
		uintptr(unsafe.Pointer(lpClassName)),  //lpClassName
		uintptr(unsafe.Pointer(lpWindowName)), //lpWindowName
		uintptr(0),                            //dwStyle
		uintptr(0),                            //X
		uintptr(0),                            //Y
		uintptr(0),                            //nWidth
		uintptr(0),                            //nHeight
		HWND_MESSAGE,                          //hWndParent
		uintptr(0),                            //hMenu
		uintptr(0),                            //hInstance
		uintptr(0))                            //lpParam

	if ret == 0 {
		return fmt.Errorf("CreateWindowEx failed: %v", err)
	}
	hWnd := syscall.Handle(ret)

	// Register for device notifications
	// https://github.com/google/cloud-print-connector/blob/master/winspool/win32.go
	// https://www.lifewire.com/device-class-guids-for-most-common-types-of-hardware-2619208
	var notificationFilter DevBroadcastDevinterface
	notificationFilter.dwSize = uint32(unsafe.Sizeof(notificationFilter))
	notificationFilter.dwDeviceType = DBT_DEVTYP_DEVICEINTERFACE
	notificationFilter.dwReserved = 0
	notificationFilter.classGuid = HID_DEVICE_CLASS
	notificationFilter.szName = 0
	ret, _, err = pRegisterDeviceNotification.Call(uintptr(hWnd), uintptr(unsafe.Pointer(&notificationFilter)), DEVICE_NOTIFY_ALL_INTERFACE_CLASSES)
	if ret == 0 {
		return fmt.Errorf("RegisterDeviceNotification failed: %v", err)
	}

	// If we made it here, start the main message loop
	var msg MSG
	for {
		if ret, _, _ := pGetMessage.Call(uintptr(unsafe.Pointer(&msg)), uintptr(0), uintptr(0), uintptr(0)); ret == 0 {
			break
		}
		pDispatchMessage.Call((uintptr(unsafe.Pointer(&msg))))
	}
	return nil
}

const (
	DBT_DEVICEQUERYREMOVE       = 0x8001 // wants to remove, may fail
	DBT_DEVICEQUERYREMOVEFAILED = 0x8002 // removal aborted
	DBT_DEVICEREMOVEPENDING     = 0x8003 // about to remove, still avail.
	DBT_DEVICETYPESPECIFIC      = 0x8005 // type specific event

	DEVICE_NOTIFY_WINDOW_HANDLE = 0x0
)

var (
	// user32                      = windows.MustLoadDLL("user32.dll")
	// registerDeviceNotificationProc = user32.MustFindProc("RegisterDeviceNotificationW")

	SMARTCARD_DEVICE_CLASS = windows.GUID{
		Data1: 0xDEEBE6AD,
		Data2: 0x9E01,
		Data3: 0x47E2,
		Data4: [8]byte{0xA3, 0xB2, 0xA6, 0x6A, 0xA2, 0xC0, 0x36, 0xC9},
	}
)

// 代码修改自：https://github.com/unreality/nCryptAgent/blob/master/deviceevents/events.go
func RegisterDeviceNotification(hwnd windows.HWND) error {

	var notificationFilter struct {
		dwSize       uint32
		dwDeviceType uint32
		dwReserved   uint32
		classGuid    windows.GUID
		szName       uint16
	}
	notificationFilter.dwSize = uint32(unsafe.Sizeof(notificationFilter))
	notificationFilter.dwDeviceType = DBT_DEVTYP_DEVICEINTERFACE
	notificationFilter.dwReserved = 0
	//notificationFilter.classGuid = SMARTCARD_DEVICE_CLASS // seems to be ignored
	notificationFilter.szName = 0

	r1, _, err := pRegisterDeviceNotification.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&notificationFilter)),
		DEVICE_NOTIFY_WINDOW_HANDLE|DEVICE_NOTIFY_ALL_INTERFACE_CLASSES,
	)
	if r1 == 0 {
		return err
	}
	return nil
}

type UsbDevInfo struct {
	DbccSize       uint32
	DbccDeviceType uint32
	DbccReserved   uint32
	Name           string
	GUID           windows.GUID
}

// 代码修改自：https://github.com/unreality/nCryptAgent/blob/master/deviceevents/events.go
func ReadDeviceInfo(devInfoPtr uintptr) (UsbDevInfo, error) {
	var devInfo struct {
		dbccSize       uint32
		dbccDeviceType uint32
		dbccReserved   uint32
		GUID           windows.GUID
	}

	var err error

	// var devInfoBytes []byte
	// slice := (*reflect.SliceHeader)(unsafe.Pointer(&devInfoBytes))
	// slice.Data = devInfoPtr
	// slice.Len = int(uint32(unsafe.Sizeof(devInfo)))
	// slice.Cap = int(uint32(unsafe.Sizeof(devInfo)))
	devInfoBytes := unsafe.Slice((*byte)(unsafe.Pointer(devInfoPtr)), unsafe.Sizeof(devInfo))
	reader := bytes.NewReader(devInfoBytes)

	// TODO: ARM might need to use different endianness
	err = binary.Read(reader, binary.LittleEndian, &devInfo.dbccSize)
	if err != nil {
		return UsbDevInfo{}, fmt.Errorf("read dbccSize failed: %w", err)
	}
	err = binary.Read(reader, binary.LittleEndian, &devInfo.dbccDeviceType)
	if err != nil {
		return UsbDevInfo{}, fmt.Errorf("read dbccDeviceType failed: %w", err)
	}
	err = binary.Read(reader, binary.LittleEndian, &devInfo.dbccReserved)
	if err != nil {
		return UsbDevInfo{}, fmt.Errorf("read dbccReserved failed: %w", err)
	}
	err = binary.Read(reader, binary.LittleEndian, &devInfo.GUID)
	if err != nil {
		return UsbDevInfo{}, fmt.Errorf("read GUID failed: %w", err)
	}

	// var devNameBytes []byte
	// devNameSlice := (*reflect.SliceHeader)(unsafe.Pointer(&devNameBytes))
	// devNameSlice.Data = devInfoPtr + unsafe.Sizeof(devInfo)
	// devNameSlice.Len = int(devInfo.dbccSize - uint32(unsafe.Sizeof(devInfo)))
	// devNameSlice.Cap = int(devInfo.dbccSize - uint32(unsafe.Sizeof(devInfo)))
	devNameBytes := unsafe.Slice((*byte)(unsafe.Pointer(devInfoPtr+unsafe.Sizeof(devInfo))),
		int(devInfo.dbccSize-uint32(unsafe.Sizeof(devInfo))))

	// return devInfo.dbccDeviceType, devInfo.GUID, string(devNameBytes), nil
	var devInfo2 UsbDevInfo
	devInfo2.DbccSize = devInfo.dbccSize
	devInfo2.DbccDeviceType = devInfo.dbccDeviceType
	devInfo2.DbccReserved = devInfo.dbccReserved
	devInfo2.GUID = devInfo.GUID
	devInfo2.Name = string(trimAllZerobyte(devNameBytes))
	return devInfo2, nil
}

func trimAllZerobyte(b []byte) []byte {
	if !bytes.Contains(b, []byte{0}) {
		return b
	}
	var ret []byte
	for _, c := range b {
		if c != 0 {
			ret = append(ret, c)
		}
	}
	return ret
}
