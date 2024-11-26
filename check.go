package doraemon

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"
)

// https://mp.weixin.qq.com/s/c9tDVoNt3dzI-FY1c6QsfA
// noCopy可以用于标记一个结构体不允许被复制。
// 在结构体中嵌入noCopy结构体，go vet工具会检查是否有复制操作。
// copyChecker可以用于检测对象是否被复制，如果对象被复制，调用check方法会返回错误。

// CopyChecker holds back pointer to itself to detect object copying.
type CopyChecker uintptr

// Check checks if the CopyChecker object has been copied.
// The Check method records its address the first time it is called and
// checks whether its address has changed the next time it is called.
// The Check method is thread-safe and can be called from multiple goroutines.
func (c *CopyChecker) Check() error {
	if uintptr(*c) != uintptr(unsafe.Pointer(c)) &&
		!atomic.CompareAndSwapUintptr((*uintptr)(c), 0, uintptr(unsafe.Pointer(c))) &&
		uintptr(*c) != uintptr(unsafe.Pointer(c)) {
		return errors.New("copyChecker: object is copied")
	}
	return nil
}

// NoCopy may be added to structs which must not be copied
// after the first use.
//
// See https://golang.org/issues/8005#issuecomment-190753527
// for details.
//
// Note that it must not be embedded, due to the Lock and Unlock methods.
type NoCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
//
// Don't remove it, as `go vet` won't recognize the type as
func (*NoCopy) Lock()   {}
func (*NoCopy) Unlock() {}

// Incomparable is a zero-width, non-comparable type. Adding it to a struct
// makes that struct also non-comparable, and generally doesn't add
// any size (as long as it's first).
type Incomparable [0]func()

func Assert(ok bool) {
	if !ok {
		panic("assertion failed")
	}
}

func AssertError(err error) {
	if err != nil {
		panic(err)
	}
}

func AssertEqual[T comparable](t1, t2 T, tn ...T) {
	if t1 != t2 {
		panic(fmt.Sprintf("%v != %v", t1, t2))
	}
	for i := 0; i < len(tn); i++ {
		if t1 != tn[i] {
			panic(fmt.Sprintf("%v != %v", t1, tn[i]))
		}
	}
}

// DebugInfo collects and returns a string containing various debug information
// about the current program execution environment.
func DebugInfo() string {
	var buf bytes.Buffer
	startTime := time.Now()

	// Helper function to write section headers
	writeSection := func(name string) {
		buf.WriteString("\n")
		buf.WriteString(strings.Repeat("=", 20))
		buf.WriteString("\n")
		buf.WriteString(name)
		buf.WriteString("\n")
		buf.WriteString(strings.Repeat("=", 20))
		buf.WriteString("\n")
	}

	// Program Arguments
	writeSection("Program Arguments")
	for i, arg := range os.Args {
		fmt.Fprintf(&buf, "[%d] %s\n", i, arg)
	}

	// Admin Information
	writeSection("Admin Information")
	fmt.Fprintf(&buf, "Is Admin: %v\n", IsAdmin())
	fmt.Fprintf(&buf, "uid: %d\n", os.Getuid())

	// System Information
	writeSection("System Information")
	fmt.Fprintf(&buf, "OS: %s\n", runtime.GOOS)
	fmt.Fprintf(&buf, "Architecture: %s\n", runtime.GOARCH)
	fmt.Fprintf(&buf, "CPUs: %d\n", runtime.NumCPU())
	fmt.Fprintf(&buf, "Go Version: %s\n", runtime.Version())
	fmt.Fprintf(&buf, "GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))

	// Directory Information
	writeSection("Directory Information")
	if dir, err := os.Getwd(); err == nil {
		fmt.Fprintf(&buf, "Current Directory: %s\n", dir)
	} else {
		fmt.Fprintf(&buf, "Current Directory Error: %v\n", err)
	}
	if exe, err := os.Executable(); err == nil {
		fmt.Fprintf(&buf, "Executable Path: %s\n", exe)
	} else {
		fmt.Fprintf(&buf, "Executable Path Error: %v\n", err)
	}
	if home, err := os.UserHomeDir(); err == nil {
		fmt.Fprintf(&buf, "User Home Directory: %s\n", home)
	} else {
		fmt.Fprintf(&buf, "User Home Directory Error: %v\n", err)
	}
	if config, err := os.UserConfigDir(); err == nil {
		fmt.Fprintf(&buf, "User Config Directory: %s\n", config)
	} else {
		fmt.Fprintf(&buf, "User Config Directory Error: %v\n", err)
	}
	if cache, err := os.UserCacheDir(); err == nil {
		fmt.Fprintf(&buf, "User Cache Directory: %s\n", cache)
	} else {
		fmt.Fprintf(&buf, "User Cache Directory Error: %v\n", err)
	}

	// Memory Usage
	writeSection("Memory Usage")
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Fprintf(&buf, "Alloc = %v\n", HumanReadableByteSize(m.Alloc))
	fmt.Fprintf(&buf, "TotalAlloc = %v\n", HumanReadableByteSize(m.TotalAlloc))
	fmt.Fprintf(&buf, "Sys = %v\n", HumanReadableByteSize(m.Sys))
	fmt.Fprintf(&buf, "NumGC = %v\n", m.NumGC)

	// Time Information
	writeSection("Time Information")
	zone, offset := time.Now().Zone()
	fmt.Fprintf(&buf, "Time Zone: %v, Offset: %v\n", zone, offset)
	fmt.Fprintf(&buf, "Current Time: %v\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(&buf, "Execution Time: %v\n", time.Since(startTime))

	// Environment Variables
	writeSection("Environment Variables")
	for _, env := range os.Environ() {
		if !strings.HasPrefix(strings.ToLower(env), "path=") {
			buf.WriteString(env)
			buf.WriteString("\n")
		}
	}

	// PATH Environment Variable
	writeSection("PATH Environment Variable")
	pathEnv := os.Getenv("PATH")
	paths := strings.Split(pathEnv, string(os.PathListSeparator))
	for _, path := range paths {
		if path != "" {
			buf.WriteString(path)
			buf.WriteString("\n")
		}
	}

	return buf.String()
}

func HumanReadableByteSize[T uint64 | int64 | uint32 | int32 | uint | int](bytes T) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
