package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/icexin/eggos/app"
	"github.com/icexin/eggos/drivers/kbd"
	"github.com/icexin/eggos/kernel"
)

func printstat(ctx *app.Context) {
	var stat1, stat2 [20]int64
	kernel.ThreadStat(&stat1)
	time.Sleep(time.Second)
	kernel.ThreadStat(&stat2)

	var sum int64
	for i := range stat1 {
		sum += stat2[i] - stat1[i]
	}
	var tids []string
	var percents []string
	for i := range stat1 {
		if stat1[i] == 0 {
			continue
		}
		tids = append(tids, fmt.Sprintf("%3d", i))
		percent := int(float32(stat2[i]-stat1[i]) / float32(sum) * 100)
		percents = append(percents, fmt.Sprintf("%3d", percent))
	}
	fmt.Fprintf(ctx.Stdout, "%s\n", strings.Join(tids, " "))
	fmt.Fprintf(ctx.Stdout, "%s\n\n", strings.Join(percents, " "))
}

func topmain(ctx *app.Context) error {
	for !kbd.Pressed('q') {
		printstat(ctx)
	}
	return nil
}

func init() {
	app.Register("top", topmain)
}
