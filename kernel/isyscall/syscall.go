package isyscall

import (
	"syscall"
	_ "unsafe"
)

const (
	EPANIC syscall.Errno = 0xfffff
)

var (
	handlers [512]Handler
)

//go:linkname wakeup github.com/jspc/eggos/kernel.wakeup
func wakeup(lock *uintptr, n int)

type Handler func(req *Request)

type Request struct {
	tf *trapFrame

	Lock uintptr
}

//go:nosplit
func (r *Request) NO() uintptr {
	return r.tf.NO()
}

//go:nosplit
func (r *Request) Arg(n int) uintptr {
	return r.tf.Arg(n)
}

//go:nosplit
func (r *Request) SetRet(v uintptr) {
	r.tf.SetRet(v)
}

//go:nosplit
func (r *Request) Ret() uintptr {
	return r.tf.Ret()
}

//go:nosplit
func (r *Request) SetErrorNO(errno syscall.Errno) {
	r.SetRet(Errno(errno))
}

//go:nosplit
func (r *Request) SetError(err error) {
	if err == nil {
		r.SetRet(0)
		return
	}
	r.SetRet(Error(err))
}

func (r *Request) Done() {
	wakeup(&r.Lock, 1)
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
	ret := uintptr(syscall.EINVAL)
	return -ret
}
