package debug

import (
	"fmt"
	"time"

	"github.com/icexin/eggos/console"
	"github.com/icexin/eggos/sys"
	"github.com/icexin/eggos/uart"
)

var (
	bootTime = time.Now()
	lastTime = time.Now()
)

func Logf(fmtstr string, args ...interface{}) {
	fmt.Fprintf(console.Console(), fmtstr+"\n", args...)
}

//go:nosplit
func PrintStr(s string) {
	uart.WriteString(s)
}

const hextab = "0123456789abcdef"

//go:nosplit
func PrintHex(n uintptr) {
	shift := sys.PtrSize*8 - 4
	for ; shift > 0; shift = shift - 4 {
		v := (n >> shift) & 0x0F
		ch := hextab[v]
		uart.WriteByte(ch)
	}
	uart.WriteByte(hextab[n&0x0F])
}
