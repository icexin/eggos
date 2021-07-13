package kbd

import (
	"github.com/icexin/eggos/kernel/trap"
	"github.com/icexin/eggos/pic"
	"github.com/icexin/eggos/sys"
)

const (
	_IRQ_KBD = pic.IRQ_BASE + pic.LINE_KBD
)

const (
	KBSTATP = 0x64 // kbd controller status port(I)
	KBS_DIB = 0x01 // kbd data in buffer
	KBDATAP = 0x60 // kbd data port(I)

	NO = 0

	SHIFT = (1 << 0)
	CTL   = (1 << 1)
	ALT   = (1 << 2)

	CAPSLOCK   = (1 << 3)
	NUMLOCK    = (1 << 4)
	SCROLLLOCK = (1 << 5)

	E0ESC = (1 << 6)

	// Special keycodes
	KEY_HOME = 0xE0
	KEY_END  = 0xE1
	KEY_UP   = 0xE2
	KEY_DN   = 0xE3
	KEY_LF   = 0xE4
	KEY_RT   = 0xE5
	KEY_PGUP = 0xE6
	KEY_PGDN = 0xE7
	KEY_INS  = 0xE8
	KEY_DEL  = 0xE9
)

var (
	inputCallback func(byte)
	keyPressed    [255]bool
)

func ctrl(c byte) byte {
	return c - '@'
}

var shiftcode = [256]byte{
	0x1d: CTL,
	0x2A: SHIFT,
	0x36: SHIFT,
	0x38: ALT,
	0x9d: CTL,
	0xB8: ALT,
}

var togglecode = [256]byte{
	0x3a: CAPSLOCK,
	0x45: NUMLOCK,
	0x46: SCROLLLOCK,
}

var normalmap = [256]byte{
	NO, 0x1B, '1', '2', '3', '4', '5', '6', // 0x00
	'7', '8', '9', '0', '-', '=', '\b', '\t',
	'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', // 0x10
	'o', 'p', '[', ']', '\n', NO, 'a', 's',
	'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', // 0x20
	'\'', '`', NO, '\\', 'z', 'x', 'c', 'v',
	'b', 'n', 'm', ',', '.', '/', NO, '*', // 0x30
	NO, ' ', NO, NO, NO, NO, NO, NO,
	NO, NO, NO, NO, NO, NO, NO, '7', // 0x40
	'8', '9', '-', '4', '5', '6', '+', '1',
	'2', '3', '0', '.', NO, NO, NO, NO, // 0x50,
	0x9c: '\n', // KP_Enter
	0xB5: '/',  // KP_Div
	0xc8: KEY_UP, 0xd0: KEY_DN,
	0xC9: KEY_PGUP, 0xD1: KEY_PGDN,
	0xCB: KEY_LF, 0xCD: KEY_RT,
	0x97: KEY_HOME, 0xCF: KEY_END,
	0xD2: KEY_INS, 0xD3: KEY_DEL,
}

var shiftmap = [256]byte{
	NO, 033, '!', '@', '#', '$', '%', '^', // 0x00
	'&', '*', '(', ')', '_', '+', '\b', '\t',
	'Q', 'W', 'E', 'R', 'T', 'Y', 'U', 'I', // 0x10
	'O', 'P', '{', '}', '\n', NO, 'A', 'S',
	'D', 'F', 'G', 'H', 'J', 'K', 'L', ':', // 0x20
	'"', '~', NO, '|', 'Z', 'X', 'C', 'V',
	'B', 'N', 'M', '<', '>', '?', NO, '*', // 0x30
	NO, ' ', NO, NO, NO, NO, NO, NO,
	NO, NO, NO, NO, NO, NO, NO, '7', // 0x40
	'8', '9', '-', '4', '5', '6', '+', '1',
	'2', '3', '0', '.', NO, NO, NO, NO, // 0x50
	0x9C: '\n', // KP_Enter
	0xB5: '/',  // KP_Div
	0xC8: KEY_UP, 0xD0: KEY_DN,
	0xC9: KEY_PGUP, 0xD1: KEY_PGDN,
	0xCB: KEY_LF, 0xCD: KEY_RT,
	0x97: KEY_HOME, 0xCF: KEY_END,
	0xD2: KEY_INS, 0xD3: KEY_DEL,
}

var ctlmap = [256]byte{
	NO, NO, NO, NO, NO, NO, NO, NO,
	NO, NO, NO, NO, NO, NO, NO, NO,
	ctrl('Q'), ctrl('W'), ctrl('E'), ctrl('R'), ctrl('T'), ctrl('Y'), ctrl('U'), ctrl('I'),
	ctrl('O'), ctrl('P'), NO, NO, '\r', NO, ctrl('A'), ctrl('S'),
	ctrl('D'), ctrl('F'), ctrl('G'), ctrl('H'), ctrl('J'), ctrl('K'), ctrl('L'), NO,
	NO, NO, NO, ctrl('\\'), ctrl('Z'), ctrl('X'), ctrl('C'), ctrl('V'),
	ctrl('B'), ctrl('N'), ctrl('M'), NO, NO, ctrl('/'), NO, NO,
	0x9C: '\r',      // KP_Enter
	0xB5: ctrl('/'), // KP_Div
	0xC8: KEY_UP, 0xD0: KEY_DN,
	0xC9: KEY_PGUP, 0xD1: KEY_PGDN,
	0xCB: KEY_LF, 0xCD: KEY_RT,
	0x97: KEY_HOME, 0xCF: KEY_END,
	0xD2: KEY_INS, 0xD3: KEY_DEL,
}

var (
	shift    byte
	charcode = [...][256]byte{
		normalmap, shiftmap, ctlmap, ctlmap,
	}
)

//go:nosplit
func ReadByte() int {
	var (
		st      byte
		data, c byte
	)
	st = sys.Inb(KBSTATP)
	if st&KBS_DIB == 0 {
		return -1
	}
	// from mouse?
	if st&0x20 != 0 {
		return -1
	}
	data = byte(sys.Inb(KBDATAP))

	switch {
	case data == 0xE0:
		shift |= E0ESC
		return 0
	case data&0x80 != 0:
		// Key released
		if shift&E0ESC == 0 {
			data &= 0x7f
		}
		shift &= ^(shiftcode[data] | E0ESC)
		c = charcode[shift&(CTL|SHIFT)][data]
		keyPressed[c] = false
		return 0
	case shift&E0ESC != 0:
		data |= 0x80
		shift &= ^byte(E0ESC)
	}

	shift |= shiftcode[data]
	shift ^= togglecode[data]
	c = charcode[shift&(CTL|SHIFT)][data]
	if shift&CAPSLOCK != 0 {
		if 'a' <= c && c <= 'z' {
			c -= 'a' - 'A'
		} else if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
	}
	keyPressed[c] = true
	return int(c)
}

const (
	csiUp    = "\x1b[A"
	csiDown  = "\x1b[B"
	csiLeft  = "\x1b[D"
	csiRight = "\x1b[C"
)

func csiEscape(ch byte) string {
	switch ch {
	case KEY_UP:
		return csiUp
	case KEY_DN:
		return csiDown
	case KEY_LF:
		return csiLeft
	case KEY_RT:
		return csiRight
	default:
		return ""
	}
}

//go:nosplit
func intr() {
	if inputCallback == nil {
		return
	}
	for {
		ch := ReadByte()
		if ch <= 0 {
			break
		}
		esc := csiEscape(byte(ch))
		if esc == "" {
			inputCallback(byte(ch))
		} else {
			for i := range esc {
				inputCallback(esc[i])
			}
		}
	}
	pic.EOI(_IRQ_KBD)
}

func OnInput(callback func(byte)) {
	inputCallback = callback
}

func Pressed(key byte) bool {
	return keyPressed[key]
}

func Init() {
	trap.Register(_IRQ_KBD, intr)
	pic.EnableIRQ(pic.LINE_KBD)
}
