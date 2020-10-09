package cga

import (
	"unsafe"

	"github.com/icexin/eggos/cga/fbcga"
	"github.com/icexin/eggos/sys"
	"github.com/icexin/eggos/vbe"
)

type Backend interface {
	GetPos() int
	SetPos(int)
	WritePos(int, byte)
	WriteByte(ch byte)
}

const (
	CRTPORT   = 0x3d4
	BACKSPACE = 0x7f
)

var (
	crt = (*[25 * 80]uint16)(unsafe.Pointer(uintptr(0xb8000)))
)

type cgabackend struct {
}

func (c *cgabackend) SetPos(pos int) {
	sys.Outb(CRTPORT, 14)
	sys.Outb(CRTPORT+1, byte(pos>>8))
	sys.Outb(CRTPORT, 15)
	sys.Outb(CRTPORT+1, byte(pos))
}

func (c *cgabackend) GetPos() int {
	var pos int

	// Cursor position: col + 80*row.
	sys.Outb(CRTPORT, 14)
	pos = int(sys.Inb(CRTPORT+1)) << 8
	sys.Outb(CRTPORT, 15)
	pos |= int(sys.Inb(CRTPORT + 1))
	return pos
}

func (c *cgabackend) WritePos(pos int, ch byte) {
	crt[pos] = uint16(ch) | 0x0700
}

func (c *cgabackend) WriteByte(ch byte) {
	var pos int

	// Cursor position: col + 80*row.
	sys.Outb(CRTPORT, 14)
	pos = int(sys.Inb(CRTPORT+1)) << 8
	sys.Outb(CRTPORT, 15)
	pos |= int(sys.Inb(CRTPORT + 1))

	switch ch {
	case '\n':
		pos += 80 - pos%80
	case BACKSPACE:
		if pos > 0 {
			pos--
		}
	default:
		// black on white
		crt[pos] = uint16(ch&0xff) | 0x0700
		pos++
	}

	// Scroll up
	if pos/80 >= 25 {
		copy(crt[:], crt[80:25*80])
		pos -= 80
		s := crt[pos : 25*80]
		for i := range s {
			// 在mac下开启qemu的-M accel=hvf,在滚屏的时候会出现`failed to decode instruction f 7f`
			// 猜测是memclrNoHeapPointer造成的，具体原因未知
			if false {
			}
			s[i] = 0
		}
	}
	c.SetPos(pos)
	crt[pos] = ' ' | 0x0700
}

func getbackend() Backend {
	if vbe.IsEnable() {
		return &fbcga.Backend
	}
	return &backend
}
