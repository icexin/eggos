package kernel

import (
	"syscall"
	"unsafe"

	"github.com/jspc/eggos/kernel/isyscall"
)

// Timer depends on epoll and pipe.
// When timer is frequently used, it will be more efficient
// to implement it in the kernel

const (
	pipeReadFd  = epollFd + 1
	pipeWriteFd = epollFd + 2
)

var (
	// FIXME:
	// avoid dup create pipe
	epollPipeCreated bool
	// the bytes number in pipe
	pipeBufferBytes int
)

//go:nosplit
func sysPipe2(req *isyscall.Request) {
	if epollPipeCreated {
		req.SetErrorNO(syscall.EINVAL)
		return
	}
	epollPipeCreated = true
	fds := (*[2]int32)(unsafe.Pointer(req.Arg(0)))
	fds[0] = pipeReadFd
	fds[1] = pipeWriteFd
	req.SetRet(0)
}

//go:nosplit
func sysPipeRead(req *isyscall.Request) {
	fd := req.Arg(0)
	buffer := req.Arg(1)
	len := req.Arg(2)
	_ = fd
	_ = buffer

	var n int
	if int(len) < pipeBufferBytes {
		n = int(len)
	} else {
		n = pipeBufferBytes
	}

	pipeBufferBytes -= n
	epollNotify(pipeWriteFd, syscall.EPOLLOUT)
	req.SetRet(uintptr(n))
}

//go:nosplit
func sysPipeWrite(req *isyscall.Request) {
	fd := req.Arg(0)
	buffer := req.Arg(1)
	len := req.Arg(2)
	_ = fd
	_ = buffer

	pipeBufferBytes += int(len)
	epollNotify(pipeReadFd, syscall.EPOLLIN)
	req.SetRet(len)
}
