package debug

import "github.com/icexin/eggos/sys"

const (
	qemuExitPort = 0x501
)

func QemuExit(code int) {
	sys.Outb(qemuExitPort, byte(code))
}
