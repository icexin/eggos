package mouse

import (
	"github.com/icexin/eggos/kernel/trap"
	"github.com/icexin/eggos/pic"
	"github.com/icexin/eggos/ps2"
)

const (
	_IRQ_MOUSE = pic.IRQ_BASE + pic.LINE_MOUSE
)

var (
	mouseCnt int

	packet     [3]byte
	status     byte
	xpos, ypos int
)

func Cursor() (int, int) {
	return xpos, ypos
}

func LeftClick() bool {
	return status&0x01 != 0
}

func RightClick() bool {
	return status&0x02 != 0
}

func intr() {
	pic.EOI(_IRQ_MOUSE)

	x := ps2.ReadDataNoWait()
	if x < 0 {
		return
	}

	switch mouseCnt {
	case 0:
		packet[0] = byte(x)
		mouseCnt++
	case 1:
		packet[1] = byte(x)
		mouseCnt++
	case 2:
		packet[2] = byte(x)
		// x overflow or y overflow, discard packet
		if packet[0]&0xC0 != 0 {
			return
		}
		status = packet[0]
		xpos += xrel(status, int(packet[1]))
		ypos += yrel(status, int(packet[2]))
		mouseCnt = 0
	}
	// debug.Logf("x:%d y:%d status:%08b", xpos, ypos, status)
}

func xrel(status byte, value int) int {
	if status&0x10 != 0 {
		return value - 0x100
	}
	return value
}

func yrel(status byte, value int) int {
	if status&0x20 != 0 {
		return value - 0x100
	}
	return value
}

func Init() {
	// enable ps2 mouse port
	ps2.WriteCmd(0xA8)
	// enable mouse device and IRQ of mouse(12)
	status := ps2.ReadCmd()
	status |= 0x22
	ps2.WriteCmd(0x60)
	ps2.WriteData(status, false)
	// enable mouse send packet
	ps2.WriteMouseData(0xF4)

	trap.Register(_IRQ_MOUSE, intr)
	pic.EnableIRQ(pic.LINE_MOUSE)
}
