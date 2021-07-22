package main

import (
	_ "github.com/icexin/eggos/kernel64"
	"github.com/icexin/eggos/uart"
)

func main() {
	uart.WriteString("call main")
	for {
	}
}
