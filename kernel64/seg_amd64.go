package kernel64

import (
	"unsafe"

	"github.com/icexin/eggos/sys"
)

const (
	segDplKernel = 0x00
	segDplUser   = 0x60
)

var (
	gdt    [5]gdtSegDesc
	gdtptr [10]byte

	idt    [256]idtSetDesc
	idtptr [10]byte
)

type gdtSegDesc [8]byte

type idtSetDesc struct {
	Addr1    uint16
	Selector uint16
	Attr     uint16
	Addr2    uint16
	Addr3    uint32
	Reserved uint32
}

//go:nosplit
func lgdt(gdtptr uintptr)

//go:nosplit
func lidt(idtptr uintptr)

//go:nosplit
func reloadCS()

//go:nosplit
func setGdtCodeDesc(desc *gdtSegDesc, dpl uint8) {
	desc[5] |= dpl | 0x98 // P=1 C=0
	desc[6] |= 0x20       // D=0 L=1
}

//go:nosplit
func setGdtDataDesc(desc *gdtSegDesc, dpl uint8) {
	desc[5] |= dpl | 0x92 // P=1 W=1
}

//go:nosplit
func gdtInit() {
	// leave gdt[0] untouch
	setGdtCodeDesc(&gdt[1], segDplKernel)
	setGdtDataDesc(&gdt[2], segDplKernel)
	setGdtCodeDesc(&gdt[3], segDplUser)
	setGdtDataDesc(&gdt[4], segDplUser)

	limit := (*uint16)(unsafe.Pointer(&gdtptr[0]))
	base := (*uint64)(unsafe.Pointer(&gdtptr[2]))
	*limit = uint16(unsafe.Sizeof(gdt) - 1)
	*base = uint64(uintptr(unsafe.Pointer(&gdt[0])))
	lgdt(uintptr(unsafe.Pointer(&gdtptr[0])))
	reloadCS()
}

//go:nosplit
func setIdtDesc(desc *idtSetDesc, addr uintptr) {
	desc.Addr1 = uint16(addr & 0xffff)
	desc.Selector = 8
	desc.Attr = 0x8e00 // P=1 TYPE=e interrupt gate
	desc.Addr2 = uint16(addr >> 16 & 0xffff)
	desc.Addr3 = uint32(addr>>32) & 0xffffffff
}

//go:nosplit
func idtInit() {
	for i := 0; i < 256; i++ {
		setIdtDesc(&idt[i], sys.FuncPC(vectors[i]))
	}

	limit := (*uint16)(unsafe.Pointer(&idtptr[0]))
	base := (*uint64)(unsafe.Pointer(&idtptr[2]))
	*limit = uint16(unsafe.Sizeof(idt) - 1)
	*base = uint64(uintptr(unsafe.Pointer(&idt[0])))
	lidt(uintptr(unsafe.Pointer(&idtptr)))
}
