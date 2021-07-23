package kernel64

import (
	"unsafe"

	"github.com/icexin/eggos/debug"
	"github.com/icexin/eggos/sys"
	"github.com/icexin/eggos/uart"
)

type trapFrame struct {
	AX, BX, CX, DX    uintptr
	BP, SI, DI, R8    uintptr
	R9, R10, R11, R12 uintptr
	R13, R14, R15     uintptr

	Trapno, Err uintptr

	// pushed by hardward
	IP, CS, FLAGS, SP, SS uintptr
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
	debug.PrintHex(tf.Trapno)
	uart.WriteString("trap\n")
	preparePanic(tf)
}
