package isyscall

import (
	"fmt"
	"syscall"
	_ "unsafe"
)

const (
	_SYS_futex  = 240
	_FUTEX_WAKE = 1
)

var (
	handlers [512]Handler
)

//go:linkname wakeup github.com/icexin/eggos/kernel.wakeup
func wakeup(lock *uintptr, n int)

type Handler func(req *Request)

type Request struct {
	NO   uintptr
	Args [6]uintptr
	Ret  uintptr

	Lock uintptr
}

func (r *Request) Done() {
	wakeup(&r.Lock, 1)
	// syscall.Syscall6(_SYS_futex, uintptr(unsafe.Pointer(&r.Lock)), _FUTEX_WAKE, 1, 0, 0, 0)
}

func GetHandler(no uintptr) Handler {
	return handlers[no]
}

func Register(no uintptr, handler Handler) {
	handlers[no] = handler
}

func Errno(code syscall.Errno) uintptr {
	return uintptr(-code)
}

func Error(err error) uintptr {
	if err == nil {
		return 0
	}
	if code, ok := err.(syscall.Errno); ok {
		return Errno(code)
	}
	fmt.Printf("syscall error: %s\n", err)
	return ^uintptr(0)
}
