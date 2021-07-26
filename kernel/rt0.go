package kernel

import (
	"github.com/icexin/eggos/kernel/mm"
	"github.com/icexin/eggos/multiboot"
	"github.com/icexin/eggos/pic"
	"github.com/icexin/eggos/uart"
)

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
func preinit(magic, mbiptr uintptr) {
	sseInit()
	gdtInit()
	idtInit()
	multiboot.Init(magic, mbiptr)
	mm.Init()
	uart.PreInit()
	syscallInit()
	trapInit()
	threadInit()
	pic.Init()
	timerInit()
	uart.WriteString("kernel\n")
	schedule()
}
