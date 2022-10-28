package cga

import (
	"unsafe"

	"github.com/jspc/eggos/drivers/cga/fbcga"
	"github.com/jspc/eggos/drivers/vbe"
	"github.com/jspc/eggos/kernel/sys"
)

type Backend interface {
	GetPos() int
	SetPos(pos int)
	// WritePos write char at given pos but not update pos
	WritePos(pos int, char byte)
	// WriteByte write char and advance pos
	WriteByte(ch byte)
}

type EraseMethod uint8

const (
	EraseMethod_Unknown EraseMethod = iota
	EraseMethod_Line
	EraseMethod_All
)

const (
	CRTPORT = 0x3d4
	bs      = '\b'
	del     = 0x7f
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
	case bs, del:
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
	//crt[pos] = ' ' | 0x0700
}

func getbackend() Backend {
	if vbe.IsEnable() {
		return &fbcga.Backend
	}
	return &backend
}
