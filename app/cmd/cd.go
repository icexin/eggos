package cmd

import (
	"errors"

	"github.com/jspc/eggos/app"
)

func cdmain(ctx *app.Context) error {
	err := ctx.ParseFlags()
	if err != nil {
		return err
	}
	if ctx.Flag().NArg() == 0 {
		return errors.New("usage: cd $dir")
	}
	name := ctx.Flag().Arg(0)
	return ctx.Chdir(name)
}

func init() {
	app.Register("cd", cdmain)
}
