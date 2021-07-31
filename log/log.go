package log

import (
	"bytes"
	"fmt"
	"os"
	"sync"

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

	loglvlDebug = iota
	loglvlInfo
	loglvlWarn
	loglvlError
	loglvlNone
)

var (
	loglvl     int
	loglvlonce sync.Once
)

func setLoglvl() {
	lvl := os.Getenv("EGGOS_LOGLVL")
	switch lvl {
	case loglvlEnvDebug:
		loglvl = loglvlDebug
	case loglvlEnvInfo:
		loglvl = loglvlInfo
	case loglvlEnvWarn:
		loglvl = loglvlWarn
	case loglvlEnvError:
		loglvl = loglvlError
	default:
		loglvl = loglvlNone
	}
}

func logf(lvl int, fmtstr string, args ...interface{}) {
	loglvlonce.Do(setLoglvl)
	if lvl < loglvl {
		return
	}

	buf := bytes.Buffer{}
	fmt.Fprintf(&buf, fmtstr, args...)
	buf.WriteByte('\n')
	buf.WriteTo(console.Console())
}

func Debugf(fmtstr string, args ...interface{}) {
	logf(loglvlDebug, fmtstr, args...)
}

func Infof(fmtstr string, args ...interface{}) {
	logf(loglvlInfo, fmtstr, args...)
}

func Warnf(fmtstr string, args ...interface{}) {
	logf(loglvlWarn, fmtstr, args...)
}

func Errorf(fmtstr string, args ...interface{}) {
	logf(loglvlError, fmtstr, args...)
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
