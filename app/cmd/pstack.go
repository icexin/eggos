package cmd

import (
	"runtime"

	"github.com/icexin/eggos/app"
)

func gostack() []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}

func pstackmain(ctx *app.Context) error {
	ctx.Stderr.Write(gostack())
	// tstack()
	return nil
}

func init() {
	app.Register("pstack", pstackmain)
}
