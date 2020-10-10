package cmd

import (
	"errors"
	"fmt"
	"io"

	"github.com/icexin/eggos/app"
	"github.com/spf13/afero"
)

func catmain(ctx *app.Context) error {
	err := ctx.ParseFlags()
	if err != nil {
		return err
	}

	if ctx.Flag().NArg() == 0 {
		return errors.New("usage: cat $filename")
	}
	name := ctx.Flag().Arg(0)
	if isdir, _ := afero.IsDir(ctx, name); isdir {
		return errors.New("can't cat dir")
	}

	f, err := ctx.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	io.Copy(ctx.Stdout, f)
	fmt.Fprintln(ctx.Stdout)
	return nil
}

func init() {
	app.Register("cat", catmain)
}
