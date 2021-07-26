package kernel

import (
	"runtime"
	"syscall"
	"unsafe"

	"github.com/icexin/eggos/debug"
	"github.com/icexin/eggos/kernel/isyscall"
)

var (
	syscalltask threadptr

	// pendingCall is the address of pending forward syscall
	pendingCall uintptr
)

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

// runSyscallThread run in normal go code space
func runSyscallThread() {
	runtime.LockOSThread()
	my := Mythread()
	syscalltask = (threadptr)(unsafe.Pointer(my))
	debug.Logf("[syscall] tid:%d", my.id)
	for {
		callptr, _, err := syscall.Syscall(SYS_WAIT_SYSCALL, 0, 0, 0)
		if err != 0 {
			throw("bad SYS_WAIT_SYSCALL return")
		}
		call := (*isyscall.Request)(unsafe.Pointer(callptr))
		handler := isyscall.GetHandler(call.NO())
		if handler == nil {
			debug.Logf("[syscall] unhandled syscall %s(%d)", syscallName(int(call.NO())), call.NO())
			// call.SetRet(isyscall.Errno(syscall.EINVAL))
			// call.SetRet(isyscall.Errno(syscall.EPERM))
			call.Done()
			continue
		}
		go handler(call)
	}
}
