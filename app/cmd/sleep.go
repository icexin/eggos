package cmd

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/icexin/eggos/app"
)

func sleepmain(ctx *app.Context) error {
	var (
		flagset = flag.NewFlagSet(ctx.Args[0], flag.ContinueOnError)
		istick  = flagset.Bool("t", false, "use ticker")
	)
	err := flagset.Parse(ctx.Args[1:])
	if err != nil {
		return err
	}
	if len(flagset.Args()) == 0 {
		return errors.New("usage: sleep $duration")
	}
	dura, err := time.ParseDuration(flagset.Arg(0))
	if err != nil {
		return err
	}

	if !*istick {
		time.Sleep(dura)
		return nil
	}

	begin := time.Now()
	ticker := time.NewTicker(dura)
	for t := range ticker.C {
		fmt.Fprintln(ctx.Stdout, t.Sub(begin).Milliseconds())
	}
	return nil
}

func init() {
	app.Register("sleep", sleepmain)
}
