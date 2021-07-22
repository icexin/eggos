package kernel64

import "github.com/icexin/eggos/uart"

//go:nosplit
func rt0()

//go:nosplit
func go_entry()

//go:nosplit
func preinit() {
	uart.WriteString("kernel64\n")
	go_entry()
	for {
	}
}
