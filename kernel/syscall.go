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
	"golang.org/x/sys/unix"
)

const (
	SYS_WAIT_IRQ     = 500
	SYS_WAIT_SYSCALL = 501
	SYS_FIXED_MMAP   = 502
)

const (
	// copy from runtime, need by readgstatus
	_Grunning = 2
	_Gsyscall = 3
)

var (
	bootstrapDone = false

	// kernelCalls is the syscalls must be implement in kernel
	kernelCalls = [...]uintptr{
		syscall.SYS_EXIT,
		syscall.SYS_SET_THREAD_AREA,
		syscall.SYS_SCHED_YIELD,
		syscall.SYS_NANOSLEEP,
		syscall.SYS_BRK,
		syscall.SYS_MUNMAP,
		syscall.SYS_MMAP2,
		syscall.SYS_MADVISE,
		syscall.SYS_CLONE,
		syscall.SYS_GETTID,
		syscall.SYS_FUTEX,
		syscall.SYS_RT_SIGACTION,
		syscall.SYS_RT_SIGPROCMASK,
		syscall.SYS_SIGALTSTACK,
		syscall.SYS_CLOCK_GETTIME,
		syscall.SYS_EXIT_GROUP,
		syscall.SYS_EPOLL_CREATE1,
		syscall.SYS_EPOLL_CTL,
		syscall.SYS_EPOLL_WAIT,

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
func syscallIntr() {
	my := Mythread()
	tf := my.tf
	dokernel := !(bootstrapDone && canForward(tf))
	if dokernel {
		tf.AX = doBootSyscall(tf.AX, tf.BX, tf.CX, tf.DX, tf.SI, tf.DI, tf.BP)
		return
	}

	// use tricks to get whether the current g has p
	status := readgstatus(getg())
	if status != _Grunning {
		tf.AX = doForwardSyscall(tf.AX, tf.BX, tf.CX, tf.DX, tf.SI, tf.DI, tf.BP)
	} else {
		// tf.AX = doForwardSyscall(tf.AX, tf.BX, tf.CX, tf.DX, tf.SI, tf.DI, tf.BP)
		// making all forwarded syscall as blocked syscall, so the syscall task can acquire a P
		// make the caller call blocksyscall, it will call syscallIntr again with syscall status.
		my.systf = *tf
		ChangeReturnPC(tf, sys.FuncPC(blocksyscall))
		return
	}
	if tf.AX == ^uintptr(0) {
		// Signal(uintptr(syscall.SIGSEGV), 10, 0)
	}
	return
}

//go:nosplit
func blocksyscall() {
	tf := *&Mythread().systf
	ret, _, errno := syscall.Syscall6(tf.AX, tf.BX, tf.CX, tf.DX, tf.SI, tf.DI, tf.BP)
	if errno != 0 {
		sys.SetAX(-uintptr(errno))
	} else {
		sys.SetAX(ret)
	}
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
	if no == syscall.SYS_WRITE && tf.BX == 2 {
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
	// if no != syscall.SYS_WRITE && no != syscall.SYS_SCHED_YIELD && no < 500 {
	// 	debug.PrintStr(sysnum[no])
	// 	debug.PrintStr("\n")
	// }

	switch no {
	case syscall.SYS_EXIT:
		exit()
		return 0
	case syscall.SYS_SET_THREAD_AREA:
		desc := (*userDesc)(unsafe.Pointer(a0))
		settls(_GO_TLS_IDX, uint32(desc.baseAddr), uint32(desc.limit))
		desc.entryNumber = _GO_TLS_IDX
		my.tls = *desc
		return 0
	case syscall.SYS_READ:
		return a2
	case syscall.SYS_WRITE:
		fd, p, n := a0, a1, a2
		if fd == 1 || fd == 2 {
			buf := sys.UnsafeBuffer(p, int(n))
			uart.Write(buf)
		}
		return n
	case syscall.SYS_OPEN:
		return errno(-1)
	case syscall.SYS_CLOSE:
		return 0
	case syscall.SYS_UNAME:
		return 0
	case syscall.SYS_SCHED_YIELD:
		Yield()
		return 0
	case syscall.SYS_NANOSLEEP:
		tc := (*timespec)(unsafe.Pointer(a0))
		nanosleep(tc)
		return 0
	case syscall.SYS_BRK:
		return mm.Sbrk(0)
	case syscall.SYS_MUNMAP:
		return 0
	case syscall.SYS_MMAP2:
		return mmap(unsafe.Pointer(a0), a1, int32(a2), int32(a3), int32(a4), uint32(a5))
	case syscall.SYS_MADVISE:
		return 0
	case syscall.SYS_CLONE:
		return uintptr(clone(my.tf.IP, a1))
	case syscall.SYS_GETTID:
		return uintptr(my.id)
	case syscall.SYS_FUTEX:
		futex((*uintptr)(unsafe.Pointer(a0)), a1, a2, (*timespec)(unsafe.Pointer(a3)))
		return 0
	case syscall.SYS_SCHED_GETAFFINITY:
		return ^uintptr(0)
	case syscall.SYS_RT_SIGACTION:
		_new := (*sigactiont)(unsafe.Pointer(a1))
		old := (*sigactiont)(unsafe.Pointer(a2))
		return uintptr(rt_sigaction(a0, _new, old, a3))
	case syscall.SYS_RT_SIGPROCMASK:
		_new := (*sigset)(unsafe.Pointer(a1))
		old := (*sigset)(unsafe.Pointer(a2))
		rtsigprocmask(int32(a0), _new, old, int32(a3))
		return 0
	case syscall.SYS_SIGALTSTACK:
		_new := (*stackt)(unsafe.Pointer(a0))
		old := (*stackt)(unsafe.Pointer(a1))
		sigaltstack(_new, old)
		return 0
	case syscall.SYS_CLOCK_GETTIME:
		tspec := (*timespec)(unsafe.Pointer(a1))
		*tspec = clocktime()
		// tspec.tv_sec = int32(counter) / _HZ
		// tspec.tv_nsec = int32(counter) % _HZ * (second / _HZ)
		return 0
	case syscall.SYS_EXIT_GROUP:
		for {
			sys.Hlt()
		}
	case syscall.SYS_FCNTL, syscall.SYS_FCNTL64:
		return errno(-1)
	case syscall.SYS_READLINKAT:
		return errno(-1)
	case syscall.SYS_EPOLL_CREATE1:
		return 3
	case syscall.SYS_EPOLL_CTL:
		return epollCtl(a0, a1, a2, a3)
	case syscall.SYS_EPOLL_WAIT:
		return epollWait(a0, a1, a2, a3)
	case unix.SYS_GETRANDOM:
		// the lenth arg
		return a1
	case syscall.SYS_GETPID:
		return 0
	case SYS_WAIT_IRQ:
		return waitIRQ()
	case SYS_WAIT_SYSCALL:
		return fetchPendingCall()
	case SYS_FIXED_MMAP:
		return fixedmmap(a0, a1)

	default:
		debug.PrintStr("syscall ")
		debug.PrintStr(sysnum[no])
		debug.PrintStr(" not found\n")
		PreparePanic(my.tf)
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
			debug.Logf("[syscall] unhandled syscall %s(%d)", sysnum[call.NO], call.NO)
			call.Ret = isyscall.Errno(syscall.EINVAL)
			// call.Ret = isyscall.Errno(syscall.EPERM)
			call.Done()
			continue
		}
		go handler(call)
	}
}
