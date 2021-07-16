package kernel

import (
	"runtime"
	"syscall"
	"unsafe"

	"github.com/icexin/eggos/debug"
	"github.com/icexin/eggos/kernel/isyscall"
	"github.com/icexin/eggos/kernel/trap"
	"github.com/icexin/eggos/mm"
	"github.com/icexin/eggos/sys"
	"github.com/icexin/eggos/uart"
)

const (
	SYS_read              = 3
	SYS_write             = 4
	SYS_open              = 5
	SYS_close             = 6
	SYS_brk               = 45
	SYS_munmap            = 91
	SYS_clone             = 120
	SYS_uname             = 122
	SYS_sched_yield       = 158
	SYS_nanosleep         = 162
	SYS_rt_sigaction      = 174
	SYS_rt_sigprocmask    = 175
	SYS_sigaltstack       = 186
	SYS_mmap2             = 192
	SYS_madvise           = 219
	SYS_gettid            = 224
	SYS_futex             = 240
	SYS_sched_getaffinity = 242
	SYS_set_thread_area   = 243
	SYS_exit_group        = 252
	SYS_clock_gettime     = 265

	SYS_EXIT       = 1
	SYS_FCNTL      = 55
	SYS_FCNTL64    = 221
	SYS_READLINKAT = 305
	SYS_RANDOM     = 355

	SYS_WAIT_IRQ     = 500
	SYS_WAIT_SYSCALL = 501
	SYS_FIXED_MMAP   = 502
)

const (
	_PROT_NONE  = 0x0
	_PROT_READ  = 0x1
	_PROT_WRITE = 0x2
	_PROT_EXEC  = 0x4

	_MAP_ANON    = 0x20
	_MAP_PRIVATE = 0x2
	_MAP_FIXED   = 0x10
)

const (
	// copy from runtime, need by readgstatus
	_Gsyscall = 3
)

var (
	bootstrapDone = false

	// kernelCalls is the syscalls must be implement in kernel
	kernelCalls = [...]uintptr{
		SYS_EXIT, SYS_set_thread_area, SYS_sched_yield, SYS_nanosleep, SYS_brk,
		SYS_munmap, SYS_mmap2, SYS_madvise, SYS_clone, SYS_gettid,
		SYS_futex, SYS_rt_sigaction, SYS_rt_sigprocmask, SYS_sigaltstack,
		SYS_clock_gettime, SYS_exit_group,
		syscall.SYS_EPOLL_CREATE1, syscall.SYS_EPOLL_CTL, syscall.SYS_EPOLL_WAIT,
		SYS_WAIT_IRQ, SYS_WAIT_SYSCALL, SYS_FIXED_MMAP,
	}

	syscalltask threadptr

	// pendingCall is the address of pending forward syscall
	pendingCall uintptr
)

type userDesc struct {
	entryNumber int
	baseAddr    uintptr
	limit       int
	flags       int
}

type timespec struct {
	tv_sec  int32
	tv_nsec int32
}

//go:nosplit
func getg() uintptr

//go:linkname readgstatus runtime.readgstatus
func readgstatus(uintptr) uint32

//go:nosplit
func blocksyscall()

//go:nosplit
func syscallIntr() {
	tf := Mythread().tf
	dokernel := !(bootstrapDone && canForward(tf))
	if dokernel {
		tf.AX = doBootSyscall(tf.AX, tf.BX, tf.CX, tf.DX, tf.SI, tf.DI, tf.BP)
		return
	}

	// use tricks to get whether the current g has p
	status := readgstatus(getg())
	if status == _Gsyscall {
		tf.AX = doForwardSyscall(tf.AX, tf.BX, tf.CX, tf.DX, tf.SI, tf.DI, tf.BP)
	} else {
		// making all forwarded syscall as blocked syscall, so the syscall task can acquire a P
		// make the caller call blocksyscall, it will call syscallIntr again with syscall status.
		ChangeReturnPC(tf, sys.FuncPC(blocksyscall))
		return
	}
	if tf.AX == ^uintptr(0) {
		// Signal(uintptr(syscall.SIGSEGV), 10, 0)
	}
	return
}

//go:nosplit
func canForward(tf *TrapFrame) bool {
	no := tf.AX
	my := Mythread()
	// syscall thread can't call self
	if syscalltask != 0 && my == syscalltask.ptr() {
		return false
	}
	// handle panic write
	if no == SYS_write && tf.BX == 2 {
		return false
	}
	for i := 0; i < len(kernelCalls); i++ {
		if no == kernelCalls[i] {
			return false
		}
	}
	return true
}

//go:nosplit
func doForwardSyscall(no, a0, a1, a2, a3, a4, a5 uintptr) uintptr {
	call := isyscall.Request{
		NO: no,
		Args: [...]uintptr{
			a0, a1, a2, a3, a4, a5,
		},
	}
	forwardCall(&call)
	return call.Ret
}

//go:nosplit
func doBootSyscall(no, a0, a1, a2, a3, a4, a5 uintptr) uintptr {
	my := Mythread()

	switch no {
	case SYS_EXIT:
		exit()
		return 0
	case SYS_set_thread_area:
		desc := (*userDesc)(unsafe.Pointer(a0))
		settls(_GO_TLS_IDX, uint32(desc.baseAddr), uint32(desc.limit))
		desc.entryNumber = _GO_TLS_IDX
		my.tls = *desc
		return 0
	case SYS_read:
		return a2
	case SYS_write:
		fd, p, n := a0, a1, a2
		if fd == 1 || fd == 2 {
			buf := sys.UnsafeBuffer(p, int(n))
			uart.Write(buf)
		}
		return n
	case SYS_open:
		return errno(-1)
	case SYS_close:
		return 0
	case SYS_uname:
		return 0
	case SYS_sched_yield:
		Yield()
		return 0
	case SYS_nanosleep:
		tc := (*timespec)(unsafe.Pointer(a0))
		nanosleep(tc)
		return 0
	case SYS_brk:
		return mm.Sbrk(0)
	case SYS_munmap:
		return 0
	case SYS_mmap2:
		return mmap(unsafe.Pointer(a0), a1, int32(a2), int32(a3), int32(a4), uint32(a5))
	case SYS_madvise:
		return 0
	case SYS_clone:
		return uintptr(clone(my.tf.IP, a1))
	case SYS_gettid:
		return uintptr(my.id)
	case SYS_futex:
		futex((*uintptr)(unsafe.Pointer(a0)), a1, a2, (*timespec)(unsafe.Pointer(a3)))
		return 0
	case SYS_sched_getaffinity:
		return ^uintptr(0)
	case SYS_rt_sigaction:
		_new := (*sigactiont)(unsafe.Pointer(a1))
		old := (*sigactiont)(unsafe.Pointer(a2))
		return uintptr(rt_sigaction(a0, _new, old, a3))
	case SYS_rt_sigprocmask:
		_new := (*sigset)(unsafe.Pointer(a1))
		old := (*sigset)(unsafe.Pointer(a2))
		rtsigprocmask(int32(a0), _new, old, int32(a3))
		return 0
	case SYS_sigaltstack:
		_new := (*stackt)(unsafe.Pointer(a0))
		old := (*stackt)(unsafe.Pointer(a1))
		sigaltstack(_new, old)
		return 0
	case SYS_clock_gettime:
		tspec := (*timespec)(unsafe.Pointer(a1))
		*tspec = clocktime()
		// tspec.tv_sec = int32(counter) / _HZ
		// tspec.tv_nsec = int32(counter) % _HZ * (second / _HZ)
		return 0
	case SYS_exit_group:
		for {
			sys.Hlt()
		}
	case SYS_FCNTL, SYS_FCNTL64:
		return errno(-1)
	case SYS_READLINKAT:
		return errno(-1)
	case syscall.SYS_EPOLL_CREATE1:
		return 3
	case syscall.SYS_EPOLL_CTL:
		return epollCtl(a0, a1, a2, a3)
	case syscall.SYS_EPOLL_WAIT:
		return epollWait(a0, a1, a2, a3)
	case SYS_RANDOM:
		// the lenth arg
		return a1
	case SYS_WAIT_IRQ:
		return waitIRQ()
	case SYS_WAIT_SYSCALL:
		return fetchPendingCall()
	case SYS_FIXED_MMAP:
		return fixedmmap(a0, a1)

	default:
		uart.WriteString("unknown syscall\n")
		// PreparePanic(my.tf)
		Signal(uintptr(syscall.SIGABRT), no, 0)
		return 0
	}
}

//go:nosplit
func mmap(addr unsafe.Pointer, n uintptr, prot, flags, fd int32, off uint32) uintptr {
	// called on sysReserve
	if prot == syscall.PROT_NONE {
		return mm.Sbrk(n)
	}

	// called on sysMap and sysAlloc
	return mm.Mmap(uintptr(addr), n)
}

//go:nopslit
func fixedmmap(addr uintptr, size uintptr) uintptr {
	mm.Fixmap(addr, addr, size)
	return addr
}

//go:nosplit
func syscal_init() {
	epollInit()
	trap.Register(0x80, syscallIntr)
}

//go:nosplit
func errno(n int) uintptr {
	return uintptr(n)
}

//go:nosplit
func forwardCall(call *isyscall.Request) {
	// wait syscall task fetch pendingCall
	for pendingCall != 0 {
		sleepon(&pendingCall)
	}
	pendingCall = uintptr(unsafe.Pointer(call))
	// tell syscall task pendingCall is avaiable
	// we can't only wakeup only one thread here
	wakeup(&pendingCall, -1)

	// wait on syscall task handle request
	sleepon(&call.Lock)
}

//go:nosplit
func fetchPendingCall() uintptr {
	// waiting someone call forward syscall
	for pendingCall == 0 {
		sleepon(&pendingCall)
	}
	ret := pendingCall
	pendingCall = 0
	// wakeup one thread, pendingCall is avaiable
	wakeup(&pendingCall, 1)
	return ret
}

// handleForward run in normal go code space
func handleForward() {
	runtime.LockOSThread()
	my := Mythread()
	syscalltask = (threadptr)(unsafe.Pointer(my))
	debug.Logf("[syscall] tid:%d", my.id)
	for {
		callptr, _, _ := syscall.Syscall(SYS_WAIT_SYSCALL, 0, 0, 0)
		call := (*isyscall.Request)(unsafe.Pointer(callptr))
		handler := isyscall.GetHandler(call.NO)
		if handler == nil {
			debug.Logf("[syscall] unhandled syscall %d", call.NO)
			call.Ret = isyscall.Errno(syscall.EINVAL)
			// call.Ret = isyscall.Errno(syscall.EPERM)
			call.Done()
			continue
		}
		go handler(call)
	}
}
