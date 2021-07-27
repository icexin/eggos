package kernel

import (
	"syscall"
	"unsafe"

	"github.com/icexin/eggos/kernel/isyscall"
	"github.com/icexin/eggos/kernel/mm"
	"gvisor.dev/gvisor/pkg/abi/linux"
	"gvisor.dev/gvisor/pkg/abi/linux/errno"
)

const (
	epollFd = 3
)

var (
	// to manage epoll event
	eventpool mm.Pool

	// header of registered epoll events
	epollEvents epollEvent

	// notify of epoll events
	epollNote note
)

//go:notinheap
type epollEvent struct {
	events uintptr
	mask   uintptr

	fd   uintptr
	data [8]byte

	pre, next *epollEvent
}

type userEpollEvent struct {
	events uint32
	data   [8]byte // to match amd64
}

//go:nosplit
func newEpollEvent() *epollEvent {
	ptr := eventpool.Alloc()
	e := (*epollEvent)(unsafe.Pointer(ptr))
	e.pre = &epollEvents
	e.next = epollEvents.next
	if epollEvents.next != nil {
		epollEvents.next.pre = e
	}
	epollEvents.next = e
	return e
}

//go:nosplit
func freeEpollEvent(e *epollEvent) {
	e.pre.next = e.next
	if e.next != nil {
		e.next.pre = e.pre
	}
	eventpool.Free(uintptr(unsafe.Pointer(e)))
}

//go:nosplit
func findEpollEvent(fd uintptr) *epollEvent {
	for e := epollEvents.next; e != nil; e = e.next {
		if e.fd == fd {
			return e
		}
	}
	return nil
}

//go:nosplit
func epollCtl(epfd, op, fd, desc uintptr) uintptr {
	euser := (*userEpollEvent)(unsafe.Pointer(desc))
	var e *epollEvent
	switch op {
	case syscall.EPOLL_CTL_ADD:
		e = newEpollEvent()
		e.fd = fd
		e.data = euser.data
		e.mask = uintptr(euser.events) | syscall.EPOLLHUP
		return 0
	case syscall.EPOLL_CTL_MOD:
		e = findEpollEvent(fd)
		if e == nil {
			return isyscall.Errno(errno.EINVAL)
		}
		e.mask = uintptr(euser.events) | syscall.EPOLLHUP
		e.data = euser.data
		return 0
	case syscall.EPOLL_CTL_DEL:
		e = findEpollEvent(fd)
		if e == nil {
			return isyscall.Errno(errno.EINVAL)
		}
		freeEpollEvent(e)
		return 0
	default:
		return isyscall.Errno(errno.EINVAL)
	}
}

//go:nosplit
func epollWait(epfd, eventptr, len, _ms uintptr) uintptr {
	if _ms != 0 {
		ts := linux.Timespec{
			Sec:  int64(_ms / 1000),
			Nsec: int64(_ms%1000) * ms,
		}
		// wait fd event
		epollNote.sleep(&ts)
		epollNote.clear()
	}

	events := (*[256]userEpollEvent)(unsafe.Pointer(eventptr))[:len]
	var cnt uintptr = 0
	for e := epollEvents.next; e != nil && cnt < len; e = e.next {
		if e.events == 0 {
			continue
		}
		ue := &events[cnt]
		ue.data = e.data
		ue.events = uint32(e.events)
		// clear events
		e.events = 0
		cnt++
	}
	return cnt
}

//go:nosplit
func epollNotify(fd, events uintptr) {
	e := findEpollEvent(fd)
	if e == nil {
		return
	}
	if e.mask&events != 0 {
		e.events |= e.mask & events
	}
	epollNote.wakeup()
}

//go:nosplit
func epollInit() {
	mm.PoolInit(&eventpool, unsafe.Sizeof(epollEvent{}))
}
