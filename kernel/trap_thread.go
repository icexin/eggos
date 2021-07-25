package kernel

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/icexin/eggos/debug"
	"github.com/icexin/eggos/kernel/trap"
	"github.com/icexin/eggos/pic"
)

var (
	// 因为中断处理是异步的，在获取一次中断期间可能发生了多次中断，
	// irqset按位保存发生的中断，对应的中断号为IRQ_BASE+1<<bit
	irqset uintptr

	traptask threadptr
)

func runTrapThread() {
	runtime.LockOSThread()
	var trapset uintptr
	var err syscall.Errno
	const setsize = unsafe.Sizeof(irqset) * 8

	my := Mythread()
	traptask = (threadptr)(unsafe.Pointer(my))
	debug.Logf("[trap] tid:%d", my.id)

	for {
		trapset, _, err = syscall.Syscall(SYS_WAIT_IRQ, 0, 0, 0)
		if err != 0 {
			throw("bad SYS_WAIT_IRQ return")
		}
		for i := uintptr(0); i < setsize; i++ {
			if trapset&(1<<i) == 0 {
				continue
			}
			trapno := uintptr(pic.IRQ_BASE + i)

			handler := trap.Handler(int(trapno))
			if handler == nil {
				fmt.Printf("trap handler for %d not found\n", trapno)
				pic.EOI(trapno)
				continue
			}
			handler()
		}
	}
}

//go:nosplit
func wakeIRQ(no uintptr) {
	irqset |= 1 << (no - pic.IRQ_BASE)
	wakeup(&irqset, 1)
	Yield()
}

//go:nosplit
func waitIRQ() uintptr {
	if irqset != 0 {
		ret := irqset
		irqset = 0
		return ret
	}
	sleepon(&irqset)
	ret := irqset
	irqset = 0
	return ret
}
