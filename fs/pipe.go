package fs

import (
	"errors"
	"syscall"
	"unsafe"

	"github.com/icexin/eggos/kernel/isyscall"
)

const (
	pipeFileBuffer = 256
)

var (
	ErrClosed = errors.New("file closed")
)

//go:linkname evnotify github.com/icexin/eggos/kernel.epollNotify
func evnotify(fd, events uintptr)

func sysPipe2(call *isyscall.Request) {
	fds := (*[2]int32)(unsafe.Pointer(call.Args[0]))
	flags := call.Args[1]
	_ = flags
	w := newPipeFile()
	r := w.Share()
	wfd, _ := AllocFileNode(w)
	rfd, _ := AllocFileNode(r)
	fds[0] = int32(rfd)
	fds[1] = int32(wfd)
	w.fd = wfd
	r.fd = rfd
	call.Done()
}

type pipeFile struct {
	fd     int
	closed bool
	ch     chan byte
}

func newPipeFile() *pipeFile {
	return &pipeFile{
		ch: make(chan byte, pipeFileBuffer),
	}
}

func (f *pipeFile) Share() *pipeFile {
	return &pipeFile{
		ch: f.ch,
	}
}

func (f *pipeFile) Read(p []byte) (n int, err error) {
	if f.closed {
		return 0, ErrClosed
	}
	var idx int
	for idx < len(p) {
		select {
		case b := <-f.ch:
			p[idx] = b
			idx++
		default:
			break
		}
	}
	if idx == 0 {
		return 0, syscall.EAGAIN
	}
	evnotify(uintptr(f.fd), syscall.EPOLLOUT)
	return idx, nil
}

func (f *pipeFile) Write(p []byte) (n int, err error) {
	if f.closed {
		return 0, ErrClosed
	}
	var idx int
	for idx < len(p) {
		select {
		case f.ch <- p[idx]:
			idx++
		default:
			break
		}
	}
	if idx == 0 {
		return 0, syscall.EAGAIN
	}
	evnotify(uintptr(f.fd), syscall.EPOLLIN)
	return idx, nil
}

func (f *pipeFile) Close() error {
	if f.closed {
		return nil
	}
	f.closed = true
	close(f.ch)
	return nil
}

func init() {
	isyscall.Register(syscall.SYS_PIPE2, sysPipe2)
}
