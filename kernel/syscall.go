package kernel

import (
	"syscall"
	"unsafe"

	"github.com/icexin/eggos/debug"
	"github.com/icexin/eggos/drivers/uart"
	"github.com/icexin/eggos/kernel/isyscall"
	"github.com/icexin/eggos/kernel/mm"
	"github.com/icexin/eggos/kernel/sys"
	"github.com/icexin/eggos/kernel/trap"
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
	SYS_EPOLL_NOTIFY = 503
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
		syscall.SYS_MADVISE,
		syscall.SYS_EXIT_GROUP,

		// TODO: real random
		unix.SYS_GETRANDOM,

		// may removed in the future
		syscall.SYS_EPOLL_CREATE1,
		syscall.SYS_EPOLL_CTL,
		syscall.SYS_EPOLL_WAIT,
		syscall.SYS_EPOLL_PWAIT,
		syscall.SYS_PIPE2,

		SYS_WAIT_IRQ,
		SYS_WAIT_SYSCALL,
		SYS_FIXED_MMAP,
		SYS_EPOLL_NOTIFY,
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
	doInKernel := !(bootstrapDone && canForward(&req))
	if doInKernel {
		doSyscall(&req)
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
}

//go:nosplit
func canForward(req *isyscall.Request) bool {
	no := req.NO()
	my := Mythread()
	// syscall thread can't call self
	if syscalltask != 0 && my == syscalltask.ptr() {
		return false
	}

	switch no {
	case syscall.SYS_WRITE:
		// handle panic write
		if req.Arg(0) == 2 {
			return false
		}
		// handle pipe write
		if req.Arg(0) == pipeWriteFd {
			return false
		}
	case syscall.SYS_READ:
		// handle pipe read
		if req.Arg(0) == pipeReadFd {
			return false
		}
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
	c := tf.SyscallRequest()
	ret, _, errno := syscall.Syscall6(c.NO(), c.Arg(0), c.Arg(1), c.Arg(2),
		c.Arg(3), c.Arg(4), c.Arg(5))
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
	debug.PrintStr(syscallName(int(req.NO())))
	debug.PrintStr("\n")
	throw("")
}

//go:nosplit
func doSyscall(req *isyscall.Request) {
	no := req.NO()
	req.SetRet(0)

	// if no != syscall.SYS_SCHED_YIELD {
	// 	debug.PrintStr("call ")
	// 	if int(no) < len(sysnum) {
	// 		debug.PrintStr(sysnum[no])
	// 	} else {
	// 		debug.PrintHex(no)
	// 	}
	// 	debug.PrintStr("\n")
	// }
	switch no {
	case syscall.SYS_ARCH_PRCTL:
		sysArchPrctl(req)
	case syscall.SYS_SCHED_GETAFFINITY:
		req.SetRet(0)
	case syscall.SYS_OPENAT:
		req.SetRet(isyscall.Errno(errno.ENOSYS))
	case syscall.SYS_MMAP:
		sysMmap(req)
	case syscall.SYS_MUNMAP:
		sysMunmap(req)
	case syscall.SYS_MADVISE:
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
		req.SetRet(uintptr(Mythread().id))
	case syscall.SYS_CLONE:
		sysClone(req)
	case syscall.SYS_FUTEX:
		sysFutex(req)
	case syscall.SYS_NANOSLEEP:
		sysNanosleep(req)
	case syscall.SYS_SCHED_YIELD:
		Yield()
	case syscall.SYS_EXIT_GROUP:
		sysExitGroup(req)

	case unix.SYS_GETRANDOM:
		req.SetRet(req.Arg(1))

	case syscall.SYS_EPOLL_CREATE1:
		sysEpollCreate(req)
	case syscall.SYS_EPOLL_CTL:
		sysEpollCtl(req)
	case syscall.SYS_EPOLL_WAIT, syscall.SYS_EPOLL_PWAIT:
		sysEpollWait(req)
	case syscall.SYS_PIPE2:
		sysPipe2(req)

	case SYS_WAIT_IRQ:
		sysWaitIRQ(req)
	case SYS_WAIT_SYSCALL:
		sysWaitSyscall(req)
	case SYS_FIXED_MMAP:
		sysFixedMmap(req)
	case SYS_EPOLL_NOTIFY:
		sysEpollNotify(req)

	default:
		req.SetRet(isyscall.Errno(errno.ENOSYS))
		if no == syscall.SYS_PIPE2 {
			changeReturnPC(Mythread().tf, sys.FuncPC(panicNosys))
		}
	}
}

//go:nosplit
func sysArchPrctl(req *isyscall.Request) {
	switch req.Arg(0) {
	case linux.ARCH_SET_FS:
		wrmsr(_MSR_FS_BASE, req.Arg(1))
		Mythread().fsBase = req.Arg(1)
	default:
		preparePanic(Mythread().tf)
		req.SetRet(errno.EINVAL)
	}
}

//go:nosplit
func sysMmap(req *isyscall.Request) {
	addr := req.Arg(0)
	n := req.Arg(1)
	prot := req.Arg(2)
	// called on sysReserve
	if prot == syscall.PROT_NONE {
		if addr == 0 {
			req.SetRet(mm.Sbrk(n))
		}
		return
	}

	// called on sysMap and sysAlloc
	req.SetRet(mm.Mmap(addr, n))
	return
}

//go:nosplit
func sysMunmap(req *isyscall.Request) {
	addr := req.Arg(0)
	n := req.Arg(1)
	mm.Munmap(addr, n)
}

//go:nosplit
func sysRead(req *isyscall.Request) {
	fd := req.Arg(0)
	if fd == pipeReadFd {
		sysPipeRead(req)
		return
	}
	req.SetRet(isyscall.Errno(errno.EINVAL))
	return
}

//go:nosplit
func sysWrite(req *isyscall.Request) {
	fd := req.Arg(0)
	buf := req.Arg(1)
	len := req.Arg(2)
	switch fd {
	case 2:
		buffer := sys.UnsafeBuffer(buf, int(len))
		uart.Write(buffer)
		req.SetRet(len)
		return
	case pipeWriteFd:
		sysPipeWrite(req)
		return
	default:
		req.SetErrorNO(syscall.EINVAL)
	}
}

//go:nosplit
func sysClockGetTime(req *isyscall.Request) {
	ts := (*linux.Timespec)(unsafe.Pointer(req.Arg(1)))
	*ts = clocktime()
}

//go:nosplit
func sysClone(req *isyscall.Request) {
	pc := Mythread().tf.IP
	flags := req.Arg(0)
	stack := req.Arg(1)
	tls := req.Arg(4)
	tid := clone(pc, stack, flags, tls)
	req.SetRet(uintptr(tid))
}

//go:nosplit
func sysFutex(req *isyscall.Request) {
	addr := (*uintptr)(unsafe.Pointer(req.Arg(0)))
	op := req.Arg(1)
	val := req.Arg(2)
	ts := (*linux.Timespec)(unsafe.Pointer(req.Arg(3)))
	futex(addr, op, val, ts)
}

//go:nosplit
func sysNanosleep(req *isyscall.Request) {
	tc := (*linux.Timespec)(unsafe.Pointer(req.Arg(0)))
	nanosleep(tc)
}

//go:nosplit
func sysEpollCreate(req *isyscall.Request) {
	req.SetRet(epollFd)
}

//go:nosplit
func sysEpollCtl(req *isyscall.Request) {
	efd := req.Arg(0)
	op := req.Arg(1)
	fd := req.Arg(2)
	desc := req.Arg(3)
	req.SetRet(epollCtl(efd, op, fd, desc))
}

//go:nosplit
func sysEpollWait(req *isyscall.Request) {
	efd := req.Arg(0)
	evs := req.Arg(1)
	len := req.Arg(2)
	_ms := req.Arg(3)
	req.SetRet(epollWait(efd, evs, len, _ms))
}

//go:nosplit
func sysWaitIRQ(req *isyscall.Request) {
	req.SetRet(waitIRQ())
}

//go:nosplit
func sysWaitSyscall(req *isyscall.Request) {
	req.SetRet(fetchPendingCall())
}

//go:nosplit
func sysFixedMmap(req *isyscall.Request) {
	vaddr := req.Arg(0)
	paddr := req.Arg(1)
	len := req.Arg(2)
	mm.Fixmap(vaddr, paddr, len)
}

//go:nosplit
func sysExitGroup(req *isyscall.Request) {
	debug.QemuExit(int(req.Arg(0)))
}

//go:nosplit
func sysEpollNotify(req *isyscall.Request) {
	epollNotify(req.Arg(0), req.Arg(1))
}

const vdsoGettimeofdaySym = 0xffffffffff600000

//go:nosplit
func vdsoGettimeofday()

//go:nosplit
func vdsoInit() {
	dst := sys.UnsafeBuffer(mm.Mmap(vdsoGettimeofdaySym, 0x100), 0x100)
	src := sys.UnsafeBuffer(sys.FuncPC(vdsoGettimeofday), 0x100)
	copy(dst, src)
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
	vdsoInit()
}
