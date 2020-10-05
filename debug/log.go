package debug

import (
	"fmt"
	"time"

	"github.com/icexin/eggos/console"
)

var (
	bootTime = time.Now()
	lastTime = time.Now()
)

func Logf(fmtstr string, args ...interface{}) {
	fmt.Fprintf(console.Console(), fmtstr+"\n", args...)
}
