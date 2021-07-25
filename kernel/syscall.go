package kernel

import (
	"syscall"
	"unsafe"

	"github.com/icexin/eggos/debug"
	"github.com/icexin/eggos/kernel/isyscall"
	"github.com/icexin/eggos/kernel/mm"
	"github.com/icexin/eggos/kernel/trap"
	"github.com/icexin/eggos/sys"
	"github.com/icexin/eggos/uart"
	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/abi/linux"
	"gvisor.dev/gvisor/pkg/abi/linux/errno"
)

const (
	_MSR_LSTAR = 0xc0000082
	_MSR_STAR  = 0xc0000081
	_MSR_FSTAR = 0xc0000084

	_MSR_IA32_EFER = 0xc0000080
	_MSR_FS_BASE   = 0xc0000100
	_MSR_GS_BASE   = 0xc0000101

	_EFER_SCE = 1 << 0 // Enable SYSCALL.
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

	kernelCalls = [...]uintptr{
		syscall.SYS_ARCH_PRCTL,
		syscall.SYS_MMAP,
		syscall.SYS_MUNMAP,
		syscall.SYS_CLOCK_GETTIME,
		syscall.SYS_RT_SIGPROCMASK,
		syscall.SYS_SIGALTSTACK,
		syscall.SYS_RT_SIGACTION,
		syscall.SYS_GETTID,
		syscall.SYS_CLONE,
		syscall.SYS_FUTEX,
		syscall.SYS_NANOSLEEP,
		syscall.SYS_SCHED_YIELD,

		// may removed in the future
		syscall.SYS_EPOLL_CREATE1,
		syscall.SYS_EPOLL_CTL,
		syscall.SYS_EPOLL_WAIT,
		syscall.SYS_EPOLL_PWAIT,

		SYS_WAIT_IRQ,
		SYS_WAIT_SYSCALL,
		SYS_FIXED_MMAP,
	}
)

//go:nosplit
func syscallEntry()

//go:nosplit
func getg() uintptr

//go:linkname readgstatus runtime.readgstatus
func readgstatus(uintptr) uint32

//go:nosplit
func syscallIntr() {
	my := Mythread()
	tf := my.tf
	my.systf = *tf

	req := tf.SyscallRequest()
	req.Ret = 0
	doInKernel := !(bootstrapDone && canForward(tf))
	if doInKernel {
		doSyscall(&req)
		tf.AX = req.Ret
		return
	}

	// use tricks to get whether the current g has p
	status := readgstatus(getg())
	if status != _Grunning {
		forwardCall(&req)
	} else {
		// tf.AX = doForwardSyscall(tf.AX, tf.BX, tf.CX, tf.DX, tf.SI, tf.DI, tf.BP)
		// making all forwarded syscall as blocked syscall, so the syscall task can acquire a P
		// make the caller call blocksyscall, it will call syscallIntr again with syscall status.
		my.systf = *tf
		changeReturnPC(tf, sys.FuncPC(blocksyscall))
		return
	}

	tf.AX = req.Ret
}

//go:nosplit
func canForward(tf *trapFrame) bool {
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
func panicNosys() {
	req := Mythread().systf.SyscallRequest()
	debug.PrintStr("syscall not found:")
	debug.PrintStr(sysnum[req.NO])
	debug.PrintStr("\n")
	for {
	}
}

//go:nosplit
func doSyscall(req *isyscall.Request) {
	// if req.NO != syscall.SYS_SCHED_YIELD {
	// 	debug.PrintStr("call ")
	// 	if int(req.NO) < len(sysnum) {
	// 		debug.PrintStr(sysnum[req.NO])
	// 	} else {
	// 		debug.PrintHex(req.NO)
	// 	}
	// 	debug.PrintStr("\n")
	// }
	switch req.NO {
	case syscall.SYS_ARCH_PRCTL:
		sysArchPrctl(req)
	case syscall.SYS_SCHED_GETAFFINITY:
		req.Ret = 0
	case syscall.SYS_OPENAT:
		req.Ret = isyscall.Errno(errno.ENOSYS)
	case syscall.SYS_MMAP:
		sysMmap(req)
	case syscall.SYS_MUNMAP:
		sysMunmap(req)
	case syscall.SYS_READ:
		sysRead(req)
	case syscall.SYS_WRITE:
		sysWrite(req)
	case syscall.SYS_CLOSE:
	case syscall.SYS_CLOCK_GETTIME:
		sysClockGetTime(req)
	case syscall.SYS_RT_SIGPROCMASK:
	case syscall.SYS_SIGALTSTACK:
	case syscall.SYS_RT_SIGACTION:
	case syscall.SYS_GETTID:
		req.Ret = uintptr(Mythread().id)
	case syscall.SYS_CLONE:
		sysClone(req)
	case syscall.SYS_FUTEX:
		sysFutex(req)
	case syscall.SYS_NANOSLEEP:
		sysNanosleep(req)
	case syscall.SYS_SCHED_YIELD:
		Yield()

	case unix.SYS_GETRANDOM:
		req.Ret = req.Args[1]

	case syscall.SYS_EPOLL_CREATE1:
		sysEpollCreate(req)
	case syscall.SYS_EPOLL_CTL:
		sysEpollCtl(req)
	case syscall.SYS_EPOLL_WAIT, syscall.SYS_EPOLL_PWAIT:
		sysEpollWait(req)

	case SYS_WAIT_IRQ:
		sysWaitIRQ(req)
	case SYS_WAIT_SYSCALL:
		sysWaitSyscall(req)
	case SYS_FIXED_MMAP:
		sysFixedMmap(req)

	default:
		req.Ret = isyscall.Errno(errno.ENOSYS)
		if req.NO == syscall.SYS_PIPE2 {
			changeReturnPC(Mythread().tf, sys.FuncPC(panicNosys))
		}
	}
}

//go:nosplit
func sysArchPrctl(req *isyscall.Request) {
	switch req.Args[0] {
	case linux.ARCH_SET_FS:
		wrmsr(_MSR_FS_BASE, req.Args[1])
		Mythread().fsBase = req.Args[1]
	default:
		preparePanic(Mythread().tf)
		req.Ret = errno.EINVAL
	}
}

//go:nosplit
func sysMmap(req *isyscall.Request) {
	addr := req.Args[0]
	n := req.Args[1]
	prot := req.Args[2]
	// called on sysReserve
	if prot == syscall.PROT_NONE {
		if addr == 0 {
			req.Ret = mm.Sbrk(n)
		}
		return
	}

	// called on sysMap and sysAlloc
	req.Ret = mm.Mmap(addr, n)
	return
}

//go:nosplit
func sysMunmap(req *isyscall.Request) {
	addr := req.Args[0]
	n := req.Args[1]
	mm.Munmap(addr, n)
}

//go:nosplit
func sysRead(req *isyscall.Request) {
	req.Ret = isyscall.Errno(errno.EINVAL)
	return
}

//go:nosplit
func sysWrite(req *isyscall.Request) {
	fd := req.Args[0]
	buf := req.Args[1]
	len := req.Args[2]
	if fd != 2 {
		req.Ret = isyscall.Errno(errno.EINVAL)
		return
	}
	buffer := sys.UnsafeBuffer(buf, int(len))
	uart.Write(buffer)
	req.Ret = len
	return
}

//go:nosplit
func sysClockGetTime(req *isyscall.Request) {
	ts := (*linux.Timespec)(unsafe.Pointer(req.Args[1]))
	*ts = clocktime()
}

//go:nosplit
func sysClone(req *isyscall.Request) {
	pc := Mythread().tf.IP
	stack := req.Args[1]
	tls := req.Args[4]
	tid := clone(pc, stack, tls)
	req.Ret = uintptr(tid)
}

//go:nosplit
func sysFutex(req *isyscall.Request) {
	addr := (*uintptr)(unsafe.Pointer(req.Args[0]))
	op := req.Args[1]
	val := req.Args[2]
	ts := (*linux.Timespec)(unsafe.Pointer(req.Args[3]))
	futex(addr, op, val, ts)
}

//go:nosplit
func sysNanosleep(req *isyscall.Request) {
	tc := (*linux.Timespec)(unsafe.Pointer(req.Args[0]))
	nanosleep(tc)
}

//go:nosplit
func sysEpollCreate(req *isyscall.Request) {
	req.Ret = 3
}

//go:nosplit
func sysEpollCtl(req *isyscall.Request) {
	efd := req.Args[0]
	op := req.Args[1]
	fd := req.Args[2]
	desc := req.Args[3]
	req.Ret = epollCtl(efd, op, fd, desc)
}

//go:nosplit
func sysEpollWait(req *isyscall.Request) {
	efd := req.Args[0]
	evs := req.Args[1]
	len := req.Args[2]
	_ms := req.Args[3]
	req.Ret = epollWait(efd, evs, len, _ms)
}

//go:nosplit
func sysWaitIRQ(req *isyscall.Request) {
	req.Ret = waitIRQ()
}

//go:nosplit
func sysWaitSyscall(req *isyscall.Request) {
	req.Ret = fetchPendingCall()
}

//go:nosplit
func sysFixedMmap(req *isyscall.Request) {
	addr := req.Args[0]
	len := req.Args[1]
	mm.Fixmap(addr, addr, len)
}

//go:nosplit
func syscallInit() {
	// write syscall selector
	wrmsr(_MSR_STAR, 8<<32)
	// clear IF when enter syscall
	wrmsr(_MSR_FSTAR, 0x200)
	// set syscall entry
	wrmsr(_MSR_LSTAR, sys.FuncPC(syscallEntry))

	// Enable SYSCALL instruction.
	efer := rdmsr(_MSR_IA32_EFER)
	wrmsr(_MSR_IA32_EFER, efer|_EFER_SCE)

	trap.Register(0x80, syscallIntr)
	epollInit()
}
