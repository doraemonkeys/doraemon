package doraemon

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

// 选择文件夹(仅限windows)
func SelectFolderOnWindows() {
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
		fmt.Println(syscall.UTF16ToString(path))
	}
}

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
				if multiString[i] == 0 {
					break
				}
				str = append(str, multiString[i])
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
