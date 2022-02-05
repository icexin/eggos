package generate

import (
	"text/template"
)

var (
	eggosImportTpl = template.Must(template.New("eggos").Parse(`//+build eggos
package {{.name}}

import (
	"runtime"

	"github.com/icexin/eggos/console"
	"github.com/icexin/eggos/drivers/cga/fbcga"
	_ "github.com/icexin/eggos/drivers/e1000"
	"github.com/icexin/eggos/drivers/kbd"
	"github.com/icexin/eggos/drivers/pci"
	"github.com/icexin/eggos/drivers/ps2/mouse"
	"github.com/icexin/eggos/drivers/uart"
	"github.com/icexin/eggos/drivers/vbe"
	"github.com/icexin/eggos/fs"
	"github.com/icexin/eggos/inet"
	"github.com/icexin/eggos/kernel"
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
`))
)
