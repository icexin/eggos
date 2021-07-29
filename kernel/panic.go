package kernel

import (
	"unsafe"

	"github.com/icexin/eggos/debug"
	"github.com/icexin/eggos/drivers/uart"
	"github.com/icexin/eggos/sys"
)

var (
	panicPcs [32]uintptr
)

//go:nosplit
func throw(msg string) {
	sys.Cli()
	tf := Mythread().tf
	throwtf(tf, msg)
}

//go:nosplit
func throwtf(tf *trapFrame, msg string) {
	sys.Cli()
	n := callers(tf, panicPcs[:])
	uart.WriteString(msg)
	uart.WriteByte('\n')

	debug.PrintStr("0x")
	debug.PrintHex(tf.IP)
	debug.PrintStr("\n")
	for i := 0; i < n; i++ {
		debug.PrintStr("0x")
		debug.PrintHex(panicPcs[i])
		debug.PrintStr("\n")
	}

	for {
	}
}

//go:nosplit
func callers(tf *trapFrame, pcs []uintptr) int {
	fp := tf.BP
	var i int
	for i = 0; i < len(pcs); i++ {
		pc := deref(fp + 8)
		pcs[i] = pc
		fp = deref(fp)
		if fp == 0 {
			break
		}
	}
	return i
}

//go:nosplit
func deref(addr uintptr) uintptr {
	return *(*uintptr)(unsafe.Pointer(addr))
}
