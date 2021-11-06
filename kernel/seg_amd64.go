package kernel

import (
	"unsafe"

	"github.com/icexin/eggos/kernel/sys"
)

const (
	segDplKernel = 0x00
	segDplUser   = 0x60
)

const (
	_KCODE_IDX = 1
	_KDATA_IDX = 2
	_UCODE_IDX = 3
	_UDATA_IDX = 4
	_TSS_IDX   = 5
)

var (
	gdt    [7]gdtSegDesc
	gdtptr [10]byte

	idt    [256]idtSetDesc
	idtptr [10]byte

	tss [26]uint32
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
func ltr(sel uintptr)

//go:nosplit
func reloadCS()

//go:nosplit
func setGdtCodeDesc(desc *gdtSegDesc, dpl uint8) {
	desc[5] |= dpl | 0x99 // P=1 C=0
	desc[6] |= 0x20       // D=0 L=1
}

//go:nosplit
func setGdtDataDesc(desc *gdtSegDesc, dpl uint8) {
	desc[5] |= dpl | 0x92 // P=1 W=1
}

//go:nosplit
func setTssDesc(lo, hi *gdtSegDesc, addr, limit uintptr) {
	// tss limit 0-15
	lo[0] = byte(limit)
	lo[1] = byte(limit >> 8)
	// tss base 0-15
	lo[2] = byte(addr)
	lo[3] = byte(addr >> 8)
	// tss base 16-23
	lo[4] = byte(addr >> 16)
	// type 64 bit tss, P=1
	lo[5] = 0x89
	// limit 16-19 and AVL=1
	lo[6] = 0x80 | byte(limit>>16)&0x07
	// base addr 24-31 G=0
	lo[7] = byte(addr >> 24)

	// base addr 32-63
	*(*uint32)(unsafe.Pointer(hi)) = uint32(addr >> 32)
}

//go:nosplit
func gdtInit() {
	// leave gdt[0] untouched
	setGdtCodeDesc(&gdt[_KCODE_IDX], segDplKernel)
	setGdtDataDesc(&gdt[_KDATA_IDX], segDplKernel)
	setGdtCodeDesc(&gdt[_UCODE_IDX], segDplUser)
	setGdtDataDesc(&gdt[_UDATA_IDX], segDplUser)
	tssAddr := uintptr(unsafe.Pointer(&tss[0]))
	tssLimit := uintptr(unsafe.Sizeof(tss)) - 1
	setTssDesc(&gdt[_TSS_IDX], &gdt[_TSS_IDX+1], tssAddr, tssLimit)

	limit := (*uint16)(unsafe.Pointer(&gdtptr[0]))
	base := (*uint64)(unsafe.Pointer(&gdtptr[2]))
	*limit = uint16(unsafe.Sizeof(gdt) - 1)
	*base = uint64(uintptr(unsafe.Pointer(&gdt[0])))
	lgdt(uintptr(unsafe.Pointer(&gdtptr[0])))
	ltr(_TSS_IDX << 3)
	reloadCS()
}

//go:nosplit
func setIdtDesc(desc *idtSetDesc, addr uintptr, dpl byte) {
	desc.Addr1 = uint16(addr & 0xffff)
	desc.Selector = 8
	desc.Attr = 0x8e00 | uint16(dpl)<<8 // P=1 TYPE=e interrupt gate
	desc.Addr2 = uint16(addr >> 16 & 0xffff)
	desc.Addr3 = uint32(addr>>32) & 0xffffffff
}

//go:nosplit
func idtInit() {
	for i := 0; i < 256; i++ {
		setIdtDesc(&idt[i], sys.FuncPC(vectors[i]), segDplKernel)
	}
	setIdtDesc(&idt[0x80], sys.FuncPC(vectors[0x80]), segDplUser)

	limit := (*uint16)(unsafe.Pointer(&idtptr[0]))
	base := (*uint64)(unsafe.Pointer(&idtptr[2]))
	*limit = uint16(unsafe.Sizeof(idt) - 1)
	*base = uint64(uintptr(unsafe.Pointer(&idt[0])))
	lidt(uintptr(unsafe.Pointer(&idtptr)))
}

//go:nosplit
func setTssSP0(addr uintptr) {
	tss[1] = uint32(addr)
	tss[2] = uint32(addr >> 32)
}
