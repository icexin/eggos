package cmd

import (
	"fmt"
	"time"

	"github.com/jspc/eggos/app"
	"github.com/jspc/eggos/drivers/clock"
)

func datemain(ctx *app.Context) error {
	fmt.Fprintf(ctx.Stdout, "%s\n", time.Now())
	t := clock.ReadCmosTime()
	fmt.Fprintf(ctx.Stdout, "%#v\n", t)
	fmt.Fprintf(ctx.Stdout, "%s\n", t.Time())
	return nil
}

func init() {
	app.Register("date", datemain)
}
