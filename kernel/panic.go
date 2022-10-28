package kernel

import (
	"unsafe"

	"github.com/jspc/eggos/drivers/qemu"
	"github.com/jspc/eggos/drivers/uart"
	"github.com/jspc/eggos/kernel/sys"
	"github.com/jspc/eggos/log"
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

	log.PrintStr("0x")
	log.PrintHex(tf.IP)
	log.PrintStr("\n")
	for i := 0; i < n; i++ {
		log.PrintStr("0x")
		log.PrintHex(panicPcs[i])
		log.PrintStr("\n")
	}

	qemu.Exit(0xff)
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
