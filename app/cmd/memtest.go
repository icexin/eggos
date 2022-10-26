package cmd

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/jspc/eggos/app"
)

func memtestmain(ctx *app.Context) error {
	if len(ctx.Args) < 2 {
		fmt.Fprintln(ctx.Stderr, "usage: memtest $duration")
		return nil
	}

	dura, err := time.ParseDuration(ctx.Args[1])
	if err != nil {
		return err
	}
	deadline := time.Now().Add(dura)
	for {
		if time.Now().After(deadline) {
			return nil
		}

		buf := make([]byte, 1024)
		rand.Read(buf)
		fmt.Fprintf(ioutil.Discard, "%v", buf)
	}
}

func init() {
	app.Register("memtest", memtestmain)
}
