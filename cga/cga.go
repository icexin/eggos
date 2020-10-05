package cga

import (
	"unsafe"

	"github.com/icexin/eggos/sys"
)

const (
	CRTPORT = 0x3d4

	BACKSPACE = 0x100
)

var (
	crt = (*[25 * 80]uint16)(unsafe.Pointer(uintptr(0xb8000)))
)

//go:nosplit
func WriteString(s string) {
	for _, c := range s {
		WriteByte(int(c))
	}
}

//go:nosplit
func WriteByte(c int) {
	var pos int

	// Cursor position: col + 80*row.
	sys.Outb(CRTPORT, 14)
	pos = int(sys.Inb(CRTPORT+1)) << 8
	sys.Outb(CRTPORT, 15)
	pos |= int(sys.Inb(CRTPORT + 1))

	switch c {
	case '\n':
		pos += 80 - pos%80
	case BACKSPACE:
		if pos > 0 {
			pos--
		}
	default:
		// black on white
		crt[pos] = uint16((c & 0xff) | 0x0700)
		pos++
	}

	// Scroll up
	if pos/80 >= 24 {
		copy(crt[:], crt[80:24*80])
		pos -= 80
		s := crt[pos : 24*80]
		for i := range s {
			// 在mac下开启qemu的-M accel=hvf,在滚屏的时候会出现`failed to decode instruction f 7f`
			// 猜测是memclrNoHeapPointer造成的，具体原因未知
			if false {
			}
			s[i] = 0
		}
	}
	sys.Outb(CRTPORT, 14)
	sys.Outb(CRTPORT+1, byte(pos>>8))
	sys.Outb(CRTPORT, 15)
	sys.Outb(CRTPORT+1, byte(pos))
	crt[pos] = ' ' | 0x0700
}
