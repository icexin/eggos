package eggos

import (
	"runtime"

	"github.com/icexin/eggos/cga/fbcga"
	"github.com/icexin/eggos/console"
	_ "github.com/icexin/eggos/e1000"
	"github.com/icexin/eggos/fs"
	"github.com/icexin/eggos/inet"
	"github.com/icexin/eggos/kbd"
	"github.com/icexin/eggos/kernel"
	"github.com/icexin/eggos/pci"
	"github.com/icexin/eggos/ps2/mouse"
	"github.com/icexin/eggos/uart"
	"github.com/icexin/eggos/vbe"
)

func kernelInit() {
	// trap and syscall threads use two Ps,
	// and the remainings are for other goroutines
	runtime.GOMAXPROCS(6)

	kernel.Init()
	uart.Init()
	kbd.Init()
	mouse.Init()
	console.Init()

	fs.Init()
	vbe.Init()
	fbcga.Init()
	pci.Init()

	err := inet.Init()
	if err != nil {
		panic(err)
	}
}

func init() {
	kernelInit()
}
