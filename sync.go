package doraemon

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

type WaitGroup struct {
	wg sync.WaitGroup
}

func (w *WaitGroup) Start(fn func()) {
	w.wg.Add(1)
	go func() {
		fn()
		w.wg.Done()
	}()
}

func (w *WaitGroup) Wait() {
	w.wg.Wait()
}

// 全局默认的Panic处理
var PanicHandlers = []func(any){
	func(recoveredErr any) {
		var buf [4096]byte
		n := runtime.Stack(buf[:], false)
		msg := fmt.Sprintf("%v\n%s\n", recoveredErr, buf[:n])
		_ = os.WriteFile(fmt.Sprintf("panic-%s.log",
			time.Now().Format("2006.01.02_15.04.05")),
			[]byte(msg), 0644)
		_, _ = os.Stderr.WriteString("\x1b[31mpanic\x1b[0m: " + msg + "\n")
	},
}

func handleCrash(additionalHandlers ...func(any)) {
	if r := recover(); r != nil {
		for _, fn := range PanicHandlers {
			fn(r)
		}
		for _, fn := range additionalHandlers {
			fn(r)
		}
	}
}

func GoWithRecover(fn func(), additionalHandlers ...func(recoveredErr any)) {
	go func() {
		defer handleCrash(additionalHandlers...)
		fn()
	}()
}
