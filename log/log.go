package log

import (
	"bytes"
	"fmt"
	"os"

	"github.com/jspc/eggos/console"
	"github.com/jspc/eggos/drivers/uart"
	"github.com/jspc/eggos/kernel/sys"
)

type LogLevel int8

const (
	LoglvlDebug LogLevel = iota
	LoglvlInfo
	LoglvlWarn
	LoglvlError
	LoglvlNone
)

const (
	loglvlEnv      = "EGGOS_LOGLVL"
	loglvlEnvDebug = "debug"
	loglvlEnvInfo  = "info"
	loglvlEnvWarn  = "warn"
	loglvlEnvError = "error"
	loglvlEnvNone  = "none"

	defaultLoglvl = LoglvlError
)

var (
	Level LogLevel

	ErrInvalidLogLevel = fmt.Errorf("invalid log level")
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

func SetLevel(l LogLevel) error {
	if l < LoglvlDebug || l > LoglvlNone {
		return ErrInvalidLogLevel
	}

	Level = l

	return nil
}

func logf(lvl LogLevel, fmtstr string, args ...interface{}) {
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
