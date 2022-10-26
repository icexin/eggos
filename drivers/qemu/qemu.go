package qemu

import "github.com/jspc/eggos/kernel/sys"

const (
	qemuExitPort = 0x501
)

func Exit(code int) {
	sys.Outb(qemuExitPort, byte(code))
}
