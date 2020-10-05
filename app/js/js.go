package js

import (
	"github.com/icexin/eggos/app"
	"github.com/icexin/eggos/fs"
	"github.com/robertkrimen/otto"
	"github.com/spf13/afero"
)

func runFile(vm *otto.Otto, fname string) error {
	buf, err := afero.ReadFile(fs.Root, fname)
	if err != nil {
		return err
	}
	_, err = vm.Run(buf)
	return err
}

func jsmain(ctx *app.Context) error {
	flag := ctx.Flag()
	ctx.ParseFlags()

	vm := NewVM()
	if flag.NArg() == 0 {
		return repl(ctx, vm)
	}

	return runFile(vm, flag.Arg(0))
}

func init() {
	app.Register("js", jsmain)
}
