package cmd

import (
	"fmt"
	"strings"

	"github.com/icexin/eggos/app"
)

func echomain(ctx *app.Context) error {
	if len(ctx.Args) == 1 {
		fmt.Fprintf(ctx.Stdout, "\n")
		return nil
	}
	fmt.Fprintf(ctx.Stdout, "%s\n", strings.Join(ctx.Args[1:], " "))
	return nil
}

func init() {
	app.Register("echo", echomain)
}
