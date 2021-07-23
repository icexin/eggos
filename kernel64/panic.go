package kernel64

import "github.com/icexin/eggos/uart"

//go:nosplit
func throw(msg string) {
	uart.WriteString(msg)
	uart.WriteByte('\n')
	for {
	}
}
