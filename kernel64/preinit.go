package kernel64

import "github.com/icexin/eggos/uart"

//go:nosplit
func rt0()

//go:nosplit
func go_entry()

//go:nosplit
func sseInit()

//go:nosplit
func wrmsr(reg uint32, value uintptr)

//go:nosplit
func rdmsr(reg uint32) (value uintptr)

//go:nosplit
func preinit() {
	sseInit()
	gdtInit()
	idtInit()
	syscallInit()
	uart.PreInit()
	uart.WriteString("kernel64\n")
	go_entry()
	for {
	}
}
