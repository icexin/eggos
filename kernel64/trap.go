package kernel64

import "github.com/icexin/eggos/uart"

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
func dotrap(tf *trapFrame) {
	uart.WriteString("trap\n")
}
