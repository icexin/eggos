package pic

import "github.com/jspc/eggos/kernel/sys"

const (
	PIC1_CMD  = 0x20
	PIC1_DATA = PIC1_CMD + 1
	PIC2_CMD  = 0xA0
	PIC2_DATA = PIC2_CMD + 1

	ICW4_8086 = 0x01 /* 8086/88 (MCS-80/85) mode */
	ICW4_AUTO = 0x02 /* Auto (normal) EOI */

	IRQ_BASE = 0x20

	LINE_TIMER = 0
	LINE_KBD   = 1
	LINE_COM1  = 4
	LINE_COM2  = 3
	LINE_MOUSE = 12
)

//go:nosplit
func Init() {
	// send init command
	sys.Outb(PIC1_CMD, 0x11)
	sys.Outb(PIC2_CMD, 0x11)

	// set offset
	sys.Outb(PIC1_DATA, IRQ_BASE)
	sys.Outb(PIC2_DATA, IRQ_BASE+8)

	// set master and slave chip
	sys.Outb(PIC1_DATA, 0x4)
	sys.Outb(PIC2_DATA, 0x2)

	sys.Outb(PIC1_DATA, ICW4_8086)
	sys.Outb(PIC2_DATA, ICW4_8086)
	// sys.Outb(PIC1_DATA, ICW4_8086|ICW4_AUTO)
	// sys.Outb(PIC2_DATA, ICW4_8086|ICW4_AUTO)

	// disable all ints
	sys.Outb(PIC1_DATA, 0xff)
	sys.Outb(PIC2_DATA, 0xff)

	// Enable slave intr pin on master
	EnableIRQ(0x02)
}

//go:nosplit
func EnableIRQ(line uint16) {
	var port uint16 = PIC1_DATA
	if line >= 8 {
		port = PIC2_DATA
		line -= 8
	}
	sys.Outb(port, byte(sys.Inb(port)&^(1<<line)))
}

//go:nosplit
func DisableIRQ(line uint16) {
	var port uint16 = PIC1_DATA
	if line >= 8 {
		port = PIC2_DATA
		line -= 8
	}
	sys.Outb(port, byte(sys.Inb(port)|(1<<line)))
}

//go:nosplit
func EOI(irq uintptr) {
	if irq >= 0x28 {
		sys.Outb(PIC2_CMD, 0x20)
	}
	sys.Outb(PIC1_CMD, 0x20)
}
