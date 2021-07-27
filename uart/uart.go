package uart

import (
	"github.com/icexin/eggos/kernel/trap"
	"github.com/icexin/eggos/pic"
	"github.com/icexin/eggos/sys"
)

const (
	com1      = uint16(0x3f8)
	_IRQ_COM1 = pic.IRQ_BASE + pic.LINE_COM1
)

var (
	inputCallback func(byte)
)

//go:nosplit
func ReadByte() int {
	if sys.Inb(com1+5)&0x01 == 0 {
		return -1
	}
	return int(sys.Inb(com1 + 0))
}

//go:nosplit
func WriteByte(ch byte) {
	const lstatus = uint16(5)
	for {
		ret := sys.Inb(com1 + lstatus)
		if ret&0x20 != 0 {
			break
		}
	}
	sys.Outb(com1, uint8(ch))
}

//go:nosplit
func Write(s []byte) (int, error) {
	for _, c := range s {
		WriteByte(c)
	}
	return len(s), nil
}

//go:nosplit
func WriteString(s string) (int, error) {
	for i := 0; i < len(s); i++ {
		WriteByte(s[i])
	}
	return len(s), nil
}

//go:nosplit
func intr() {
	if inputCallback == nil {
		return
	}
	for {
		ch := ReadByte()
		if ch == -1 {
			break
		}
		inputCallback(byte(ch))
	}
	pic.EOI(_IRQ_COM1)
}

//go:nosplit
func PreInit() {
	sys.Outb(com1+3, 0x80) // unlock divisor
	sys.Outb(com1+0, 115200/9600)
	sys.Outb(com1+1, 0)

	sys.Outb(com1+3, 0x03) // lock divisor
	// disable fifo
	sys.Outb(com1+2, 0)

	// enable receive interrupt
	sys.Outb(com1+4, 0x00)
	sys.Outb(com1+1, 0x01)
}

func OnInput(callback func(byte)) {
	inputCallback = callback
}

func Init() {
	trap.Register(_IRQ_COM1, intr)
	pic.EnableIRQ(pic.LINE_COM1)
}
