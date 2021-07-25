package kernel

import (
	"unsafe"

	"github.com/icexin/eggos/kernel/isyscall"
	"github.com/icexin/eggos/kernel/trap"
	"github.com/icexin/eggos/pic"
	"github.com/icexin/eggos/sys"
)

type trapFrame struct {
	AX, BX, CX, DX    uintptr
	BP, SI, DI, R8    uintptr
	R9, R10, R11, R12 uintptr
	R13, R14, R15     uintptr

	Trapno, Err uintptr

	// pushed by hardware
	IP, CS, FLAGS, SP, SS uintptr
}

func (t *trapFrame) SyscallRequest() isyscall.Request {
	return isyscall.Request{
		NO: t.AX,
		Args: [6]uintptr{
			t.DI, t.SI, t.DX, t.R10, t.R8, t.R9,
		},
	}
}

//go:nosplit
func trapret()

//go:nosplit
func trapPanic() {
	for {
	}
	panic("trap panic")
}

//go:nosplit
func ignoreHandler() {
}

//go:nosplit
func pageFaultPanic() {
	panic("nil pointer or invalid memory access")
}

//go:nosplit
func preparePanic(tf *trapFrame) {
	changeReturnPC(tf, sys.FuncPC(trapPanic))
}

// ChangeReturnPC change the return pc of a trap
// must be called in trap handler
//go:nosplit
func changeReturnPC(tf *trapFrame, pc uintptr) {
	// tf.Err, tf.IP, tf.CS, tf.FLAGS = pc, tf.CS, tf.FLAGS, tf.IP
	sp := tf.SP
	sp -= sys.PtrSize
	*(*uintptr)(unsafe.Pointer(sp)) = tf.IP
	tf.SP = sp
	tf.IP = pc
}

//go:nosplit
func dotrap(tf *trapFrame) {
	if sys.Flags()&_FLAGS_IF != 0 {
		throw("IF should clear")
	}
	Mythread().tf = tf
	// debug.PrintHex(tf.Trapno)
	// uart.WriteString("trap\n")
	handler := trap.Handler(int(tf.Trapno))
	if handler == nil {
		throw("kernel panic")
		preparePanic(tf)
		return
	}
	// timer and syscall interrupts are processed synchronously
	if tf.Trapno > 32 && tf.Trapno != 0x80 {
		// pci using level trigger irq, cause dead lock on trap handler
		// FIXME: hard code network irq line
		if tf.Trapno == 43 {
			pic.DisableIRQ(43 - pic.IRQ_BASE)
		}
		wakeIRQ(tf.Trapno)
		return
	}
	handler()
}

//go:nosplit
func trapInit() {
	trap.Register(39, ignoreHandler)
	trap.Register(47, ignoreHandler)
}
