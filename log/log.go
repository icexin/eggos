package log

import (
	"bytes"
	"fmt"
	"os"

	"github.com/icexin/eggos/console"
	"github.com/icexin/eggos/drivers/uart"
	"github.com/icexin/eggos/kernel/sys"
)

const (
	loglvlEnv      = "EGGOS_LOGLVL"
	loglvlEnvDebug = "debug"
	loglvlEnvInfo  = "info"
	loglvlEnvWarn  = "warn"
	loglvlEnvError = "error"
	loglvlEnvNone  = "none"

	LoglvlDebug = iota
	LoglvlInfo
	LoglvlWarn
	LoglvlError
	LoglvlNone

	defaultLoglvl = LoglvlError
)

var (
	Level int
)

func init() {
	lvl := os.Getenv("EGGOS_LOGLVL")
	switch lvl {
	case loglvlEnvDebug:
		Level = LoglvlDebug
	case loglvlEnvInfo:
		Level = LoglvlInfo
	case loglvlEnvWarn:
		Level = LoglvlWarn
	case loglvlEnvError:
		Level = LoglvlError
	default:
		Level = defaultLoglvl
	}
}

func logf(lvl int, fmtstr string, args ...interface{}) {
	if lvl < Level {
		return
	}

	buf := bytes.Buffer{}
	fmt.Fprintf(&buf, fmtstr, args...)
	buf.WriteByte('\n')
	buf.WriteTo(console.Console())
}

func Debugf(fmtstr string, args ...interface{}) {
	logf(LoglvlDebug, fmtstr, args...)
}

func Infof(fmtstr string, args ...interface{}) {
	logf(LoglvlInfo, fmtstr, args...)
}

func Warnf(fmtstr string, args ...interface{}) {
	logf(LoglvlWarn, fmtstr, args...)
}

func Errorf(fmtstr string, args ...interface{}) {
	logf(LoglvlError, fmtstr, args...)
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
