package kernel

import (
	"runtime"
	"syscall"
	"unsafe"

	"github.com/jspc/eggos/kernel/isyscall"
	"github.com/jspc/eggos/kernel/sys"
	"github.com/jspc/eggos/log"
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

	// for debug purpose
	if -call.Ret() == uintptr(isyscall.EPANIC) {
		preparePanic(Mythread().tf)
	}
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
	log.Infof("[syscall] tid:%d", my.id)
	for {
		callptr, _, err := syscall.Syscall(SYS_WAIT_SYSCALL, 0, 0, 0)
		if err != 0 {
			throw("bad SYS_WAIT_SYSCALL return")
		}
		call := (*isyscall.Request)(unsafe.Pointer(callptr))

		no := call.NO()
		handler := isyscall.GetHandler(no)
		if handler == nil {
			log.Errorf("[syscall] unhandled %s(%d)(0x%x, 0x%x, 0x%x, 0x%x, 0x%x, 0x%x)",
				syscallName(int(no)), no,
				call.Arg(0), call.Arg(1), call.Arg(2), call.Arg(3),
				call.Arg(4), call.Arg(5))
			call.SetErrorNO(syscall.ENOSYS)
			call.Done()
			continue
		}
		go func() {
			handler(call)
			var iret interface{}
			ret := call.Ret()
			if hasErrno(ret) {
				iret = syscall.Errno(-ret)
			} else {
				iret = ret
			}
			log.Debugf("[syscall] %s(%d)(0x%x, 0x%x, 0x%x, 0x%x, 0x%x, 0x%x) = %v",
				syscallName(int(no)), no,
				call.Arg(0), call.Arg(1), call.Arg(2), call.Arg(3),
				call.Arg(4), call.Arg(5), iret,
			)
			call.Done()
		}()
	}
}

func hasErrno(n uintptr) bool {
	return 1<<(sys.PtrSize*8-1)&n != 0
}
