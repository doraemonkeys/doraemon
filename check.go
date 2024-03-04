package doraemon

import (
	"errors"
	"fmt"
	"sync/atomic"
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
