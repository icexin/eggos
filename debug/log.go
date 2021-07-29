package debug

import (
	"bytes"
	"fmt"

	"github.com/icexin/eggos/console"
	"github.com/icexin/eggos/drivers/uart"
	"github.com/icexin/eggos/sys"
)

func Logf(fmtstr string, args ...interface{}) {
	buf := bytes.Buffer{}
	fmt.Fprintf(&buf, fmtstr+"\n", args...)
	buf.WriteTo(console.Console())
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
