//go:build windows

package doraemon

import "golang.org/x/sys/windows"

func OpenLibrary(name string) (uintptr, error) {
	handle, err := windows.LoadLibrary(name)
	return uintptr(handle), err
}

func CloseLibrary(handle uintptr) error {
	return nil
}
