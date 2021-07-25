package main

import (
	"fmt"
	"io"
	"runtime"

	_ "github.com/icexin/eggos/app"
	"github.com/icexin/eggos/app/sh"
	"github.com/icexin/eggos/console"
	"github.com/icexin/eggos/fs"
	"github.com/icexin/eggos/kbd"
	kernel "github.com/icexin/eggos/kernel"
	"github.com/icexin/eggos/ps2/mouse"
	"github.com/icexin/eggos/uart"
)

func write(fmtstr string, args ...interface{}) {
	uart.WriteString(fmt.Sprintf(fmtstr, args...))
}

func main() {
	// trap and syscall threads use two Ps,
	// and the remainings are for other goroutines
	runtime.GOMAXPROCS(6)

	uart.Init()
	kbd.Init()
	mouse.Init()
	console.Init()
	kernel.Init()

	fs.Init()

	// fmt.Println("hello world")
	// fmt.Fprintf(os.Stderr, "%s\n", time.Now())
	// time.Sleep(time.Second)
	// panic("main")
	// for {
	// }
	w := console.Console()
	io.WriteString(w, "\nwelcome to eggos\n")
	sh.Bootstrap()
}
