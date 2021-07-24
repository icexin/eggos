package kernel64

import (
	"syscall"

	"github.com/icexin/eggos/debug"
	"github.com/icexin/eggos/kernel/isyscall"
	"github.com/icexin/eggos/kernel/trap"
	"github.com/icexin/eggos/kernel64/mm"
	"github.com/icexin/eggos/sys"
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

//go:nosplit
func syscallEntry()

//go:nosplit
func syscallIntr() {
	my := Mythread()
	tf := my.tf
	my.systf = *tf
	// SYSCALL store IP in RCX and FLAGS in R11
	// tf.IP = tf.CX
	// tf.FLAGS = tf.R11
	// after we enter syscallIntr from SYSCALL,
	// the CS changed to 8 and SS changed to 16,
	// both need be changed to USER
	// tf.CS = _UCODE_IDX<<3 | _RPL_USER
	// tf.SS = _UDATA_IDX<<3 | _RPL_USER

	req := tf.SyscallRequest()
	doSyscall(&req)
	tf.AX = req.Ret
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
	switch req.NO {
	case syscall.SYS_ARCH_PRCTL:
		sysArchPrctl(req)
	case syscall.SYS_SCHED_GETAFFINITY:
		req.Ret = 0
	case syscall.SYS_OPENAT:
		req.Ret = isyscall.Errno(errno.ENOSYS)
	case syscall.SYS_MMAP:
		sysMmap(req)
	case syscall.SYS_READ:
		sysRead(req)
	case syscall.SYS_CLOSE:

	default:
		changeReturnPC(Mythread().tf, sys.FuncPC(panicNosys))
	}
}

//go:nosplit
func sysArchPrctl(req *isyscall.Request) {
	switch req.Args[0] {
	case linux.ARCH_SET_FS:
		wrmsr(_MSR_FS_BASE, req.Args[1])
		Mythread().fsBase = req.Args[1]
	default:
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
		req.Ret = mm.Sbrk(n)
		return
	}

	// called on sysMap and sysAlloc
	req.Ret = mm.Mmap(addr, n)
	return
}

//go:nosplit
func sysRead(req *isyscall.Request) {
	req.Ret = isyscall.Errno(errno.EINVAL)
	return
}

//go:nosplit
func syscallInit() {
	// write syscall selector
	wrmsr(_MSR_STAR, 8<<32)
	// clear IF when enter syscall
	// wrmsr(_MSR_FSTAR, 0x200)
	// set syscall entry
	wrmsr(_MSR_LSTAR, sys.FuncPC(syscallEntry))

	// Enable SYSCALL instruction.
	efer := rdmsr(_MSR_IA32_EFER)
	wrmsr(_MSR_IA32_EFER, efer|_EFER_SCE)

	trap.Register(0x80, syscallIntr)
}
