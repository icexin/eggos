package kernel

import (
	"github.com/jspc/eggos/drivers/multiboot"
	"github.com/jspc/eggos/drivers/pic"
	"github.com/jspc/eggos/drivers/uart"
	"github.com/jspc/eggos/kernel/mm"
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
