//+build eggos
package eggos

import (
	"runtime"

	"github.com/jspc/eggos/console"
	"github.com/jspc/eggos/drivers/cga/fbcga"
	_ "github.com/jspc/eggos/drivers/e1000"
	"github.com/jspc/eggos/drivers/kbd"
	"github.com/jspc/eggos/drivers/pci"
	"github.com/jspc/eggos/drivers/ps2/mouse"
	"github.com/jspc/eggos/drivers/uart"
	"github.com/jspc/eggos/drivers/vbe"
	"github.com/jspc/eggos/fs"
	"github.com/jspc/eggos/inet"
	"github.com/jspc/eggos/kernel"
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
	inet.Init()
}

func init() {
	kernelInit()
}
