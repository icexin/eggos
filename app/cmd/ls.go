package cmd

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/icexin/eggos/app"
)

func printfiles(w io.Writer, files ...os.FileInfo) {
	tw := tabwriter.NewWriter(w, 0, 4, 1, ' ', 0)
	for _, file := range files {
		fmt.Fprintf(tw, "%s\t%d\t%s\n", file.Mode(), file.Size(), file.Name())
	}
	tw.Flush()
}

func lsmain(ctx *app.Context) error {
	err := ctx.ParseFlags()
	if err != nil {
		return err
	}

	var name string
	if ctx.Flag().NArg() > 0 {
		name = ctx.Flag().Arg(0)
	}
	f, err := ctx.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		printfiles(ctx.Stdout, stat)
		return nil
	}
	stats, err := f.Readdir(-1)
	if err != nil {
		return err
	}
	printfiles(ctx.Stdout, stats...)
	return nil
}

func init() {
	app.Register("ls", lsmain)
}
