//go:build windows

package win_tool

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	win32 "github.com/arduino/go-win32-utils"
	"github.com/doraemonkeys/doraemon"
	"github.com/doraemonkeys/doraemon/win-tool/device"
	"github.com/leibnewton/winapi/kbcap"
	"github.com/lxn/win"
)

// 返回选择的文件路径(绝对路径)
func SelectMultiFilesOnWindows() ([]string, error) {
	var ofn win.OPENFILENAME
	fileNames := make([]uint16, 1024*1024)

	ofn.LStructSize = uint32(unsafe.Sizeof(ofn))
	ofn.Flags = win.OFN_ALLOWMULTISELECT | win.OFN_EXPLORER | win.OFN_LONGNAMES | win.OFN_FILEMUSTEXIST | win.OFN_PATHMUSTEXIST

	ofn.NMaxFile = uint32(len(fileNames))
	ofn.LpstrFile = &fileNames[0]

	ret := win.GetOpenFileName(&ofn)
	if ret {
		return parseMultiString(fileNames), nil
	}
	// 用户取消选择或者选择失败(比如选择了太多文件)
	return nil, fmt.Errorf("user cancel or select too many files")
}

// Helper function to convert the multistring returned by GetOpenFileName to a slice of strings
func parseMultiString(multiString []uint16) []string {
	var ret []string = make([]string, 0)
	for i := 0; i < len(multiString); i++ {
		if multiString[i] != 0 {
			var str []uint16
			for ; i < len(multiString); i++ {
				str = append(str, multiString[i])
				if multiString[i] == 0 {
					break
				}
			}
			ret = append(ret, win.UTF16PtrToString(&str[0]))
		}
	}
	if len(ret) <= 1 {
		return ret
	}
	var dir = ret[0]
	for i := 1; i < len(ret); i++ {
		ret[i] = filepath.Join(dir, ret[i])
	}
	return ret[1:]
}

// 选择文件夹(仅限windows)
func SelectFolderOnWindows() (string, error) {
	const BIF_RETURNONLYFSDIRS = 0x00000001
	const BIF_NEWDIALOGSTYLE = 0x00000040
	var bi win.BROWSEINFO
	bi.HwndOwner = win.GetDesktopWindow()
	bi.UlFlags = BIF_RETURNONLYFSDIRS | BIF_NEWDIALOGSTYLE
	bi.LpszTitle, _ = syscall.UTF16PtrFromString("Select a folder")

	id := win.SHBrowseForFolder(&bi)
	if id != 0 {
		path := make([]uint16, win.MAX_PATH)
		win.SHGetPathFromIDList(id, &path[0])
		return syscall.UTF16ToString(path), nil
	}
	return "", fmt.Errorf("user cancel")
}

// 获取系统默认桌面路径
func GetDesktopPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	desktopPath := filepath.Join(homeDir, "Desktop")
	if _, err := os.Stat(desktopPath); os.IsNotExist(err) {
		return "", fmt.Errorf("desktop path not exist")
	}
	return desktopPath, nil
}

// 获取系统默认文档路径(D:\xxx\computer\Documents)
func GetDocumentsFolder() (string, error) {
	return win32.GetDocumentsFolder()
}

// C:\Users\xxx\AppData\Local
func GetRoamingAppDataFolder() (string, error) {
	return win32.GetRoamingAppDataFolder()
}

// C:\Users\xxx\AppData\Local
func GetLocalAppDataFolder() (string, error) {
	return win32.GetLocalAppDataFolder()
}

// 监控键盘输入
func WatchKeyboardInput(callback func(string), codeCallback func(byte)) error {
	// kbcap.Debug = true
	return kbcap.MonitorKeyboard(callback, codeCallback)
}

// 获取系统中所有盘符
func GetSystemDiskLetters() []string {
	// 获取系统dll
	kernel32 := syscall.MustLoadDLL("kernel32.dll")
	// 获取dll中函数
	GetLogicalDrives := kernel32.MustFindProc("GetLogicalDrives")
	// 调用dll中函数
	n, _, _ := GetLogicalDrives.Call()
	s := strconv.FormatInt(int64(n), 2)
	var allDrives = []string{"A:", "B:", "C:", "D:", "E:", "F:", "G:", "H:",
		"I:", "J:", "K:", "L:", "M:", "N:", "O:", "P:", "Q:", "R:", "S:", "T:",
		"U:", "V:", "W:", "X:", "Y:", "Z:"}
	temp := allDrives[0:len(s)]
	var d []string
	for i, v := range s {
		if v == 49 {
			l := len(s) - i - 1
			d = append(d, temp[l])
		}
	}
	var drives []string
	for i, v := range d {
		drives = append(drives[i:], append([]string{v}, drives[:i]...)...)
	}
	return drives
}

// 获取系统中所有盘符
//
// PS C:\Users> fsutil fsinfo drives
//
// 驱动器: C:\ D:\ E:\ F:\
func GetSystemDiskLetters2() ([]string, error) {
	cmd := exec.Command("fsutil", "fsinfo", "drives")
	cmd_out := bytes.NewBuffer(nil)
	cmd.Stdout = cmd_out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	lines, err := doraemon.ReadLines(cmd_out)
	if err != nil {
		return nil, err
	}
	for strings.TrimSpace(string(lines[0])) == "" {
		if len(lines) == 1 {
			return nil, fmt.Errorf("no disk")
		}
		lines = lines[1:]
	}
	drives := strings.Split(string(lines[0]), `\`)
	drives[0] = drives[0][len(drives[0])-2:]
	drives = drives[:len(drives)-1]
	return drives, nil
}

// 获取插入的U盘盘符
func GetRemoveableVolumeLetters() ([]string, error) {
	n, err := device.GetUsbPhysicalDiskCount()
	if err != nil {
		return nil, err
	}
	// 获取全部盘符
	disks := GetSystemDiskLetters()

	return disks[len(disks)-int(n):], nil
}

// 获取系统硬盘的盘符
func GetHardVolumeLetters() ([]string, error) {
	n, err := device.GetUsbPhysicalDiskCount()
	if err != nil {
		return nil, err
	}
	// 获取全部盘符
	disks := GetSystemDiskLetters()

	return disks[:len(disks)-int(n)], nil
}
