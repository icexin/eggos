package main

import (
	"io"
	_ "net/http/pprof"
	"os"
	"runtime"

	_ "github.com/jspc/eggos"
	"github.com/jspc/eggos/app/sh"
	"github.com/jspc/eggos/console"
	"github.com/jspc/eggos/log"
)

func main() {
	log.Infof("[runtime] go version:%s", runtime.Version())
	log.Infof("[runtime] args:%v", os.Args)
	w := console.Console()
	io.WriteString(w, "\nwelcome to eggos\n")
	sh.Bootstrap()
}
