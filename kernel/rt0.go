package kernel

import (
	"github.com/icexin/eggos/drivers/multiboot"
	"github.com/icexin/eggos/drivers/pic"
	"github.com/icexin/eggos/drivers/uart"
	"github.com/icexin/eggos/kernel/mm"
)

//go:nosplit
func rt0()

//go:nosplit
func go_entry()

//go:nosplit
func wrmsr(reg uint32, value uintptr)

//go:nosplit
func rdmsr(reg uint32) (value uintptr)

//go:nosplit
func preinit(magic, mbiptr uintptr) {
	simdInit()
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
	schedule()
}
