//go:build darwin || freebsd || linux

package libc

import "github.com/ebitengine/purego"

func OpenLibrary(name string) (uintptr, error) {
	return purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
}

func CloseLibrary(handle uintptr) error {
	// Dlclose减少动态库句柄上的引用计数。如果引用计数降为零，
	// 并且没有其他加载的库使用其中的符号，则卸载动态库。
	// 动态库里运行中的goroutine不主动退出就不会被影响
	return purego.Dlclose(handle)
}
