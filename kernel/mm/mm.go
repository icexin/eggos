package mm

import (
	"unsafe"

	"github.com/icexin/eggos/drivers/multiboot"
	"github.com/icexin/eggos/kernel/sys"
)

const (
	PGSIZE = 4 << 10
	// 1-100 Mb memory reverse for kernel image
	MEMSTART = 100 << 20
	// 默认可以使用的物理内存终止地址，如果能从grub那里获取就用grub的
	DEFAULT_MEMTOP = 256 << 20
	// 虚拟内存起始地址
	VMSTART = 1 << 30

	PTE_P = 0x001
	PTE_W = 0x002
	PTE_U = 0x004

	_ENTRY_NUMBER = PGSIZE / sys.PtrSize
)

var (
	memtop uintptr

	kmm = kmmt{voffset: VMSTART}
	vmm vmmt
)

//go:nosplit
func pageEnable()

//go:nosplit
func lcr3(topPage *entryPage)

//go:linkname throw github.com/icexin/eggos/kernel.throw
func throw(msg string)

//go:nosplit
func pageRoundUp(size uintptr) uintptr {
	return (size + PGSIZE - 1) &^ (PGSIZE - 1)
}

//go:nosplit
func pageRoundDown(v uintptr) uintptr {
	return v &^ (PGSIZE - 1)
}

//go:nosplit
func pageEntryIdx(v uintptr, lvl int) uintptr {
	return (v >> (12 + (lvl-1)*9)) & (_ENTRY_NUMBER - 1)
}

//go:notinheap
type page struct {
	next *page
}

type kmmstat struct {
	alloc int
}

type kmmt struct {
	freelist *page
	voffset  uintptr
	stat     kmmstat
}

//go:nosplit
func (k *kmmt) sbrk(n uintptr) uintptr {
	p := k.voffset
	k.voffset = pageRoundUp(k.voffset + n)
	if k.voffset < p {
		throw("virtual memory address all used")
	}
	return p
}

//go:nosplit
func (k *kmmt) alloc() uintptr {
	r := k.freelist
	if r == nil {
		throw("kmemt.alloc")
	}
	k.stat.alloc++
	k.freelist = r.next
	return uintptr(unsafe.Pointer(r))
}

//go:nosplit
func (k *kmmt) freeRange(start, end uintptr) {
	p := pageRoundUp(start)
	for ; p+PGSIZE <= end; p += PGSIZE {
		k.free(p)
	}
}

//go:nosplit
func (k *kmmt) free(p uintptr) {
	if p%PGSIZE != 0 || p >= memtop {
		throw("kmemt.free")
	}
	r := (*page)(unsafe.Pointer(p))
	r.next = k.freelist
	k.freelist = r
}

//go:notinheap
type entryPage [_ENTRY_NUMBER]entry

type entry uintptr

//go:nosplit
func (p entry) present() bool {
	return p&PTE_P != 0
}

//go:nosplit
func (p entry) addr() uintptr {
	return uintptr(p) &^ 0xfff
}

//go:nosplit
func (p entry) entryPage() *entryPage {
	return (*entryPage)(unsafe.Pointer(p.addr()))
}

type vmmt struct {
	topPage *entryPage
}

//go:nosplit
func (v *vmmt) munmap(va, size uintptr) bool {
	// println("mumap va=", va, " size=", size)
	p := pageRoundDown(va)
	last := pageRoundDown(va + size - 1)
	for ; p != last; p += PGSIZE {
		pte := v.walkpgdir(p, false)
		if pte == nil {
			return false
		}
		if !pte.present() {
			return false
		}
		kmm.free(pte.addr())
		*pte = 0
	}
	return true
}

//go:nosplit
func (v *vmmt) mmap(va, size, perm uintptr) bool {
	// println("mmap va=", unsafe.Pointer(va), " size=", size>>10)
	var pa uintptr
	p := pageRoundDown(va)
	last := pageRoundDown(va + size - 1)
	for {
		pte := v.walkpgdir(p, true)
		if pte == nil {
			return false
		}
		if pte.present() {
			throw("mmap remap")
		} else {
			pa = kmm.alloc()
			sys.Memclr(pa, PGSIZE)
			*pte = entry(pa | perm)
		}
		if p == last {
			break
		}
		p += PGSIZE
	}
	return true
}

//go:nosplit
func Sbrk(n uintptr) uintptr {
	return kmm.sbrk(n)
}

//go:nosplit
func Mmap(va, size uintptr) uintptr {
	if va == 0 {
		va = kmm.sbrk(size)
	}
	vmm.mmap(va, size, PTE_P|PTE_W|PTE_U)
	// flush page table cache
	lcr3(vmm.topPage)
	return va
}

//go:nosplit
func Munmap(va, size uintptr) bool {
	ok := vmm.munmap(va, size)
	lcr3(vmm.topPage)
	return ok
}

//go:nosplit
func Fixmap(va, pa, size uintptr) {
	vmm.fixmap(va, pa, size, PTE_P|PTE_W|PTE_U)
	// flush page table cache
	lcr3(vmm.topPage)
}

//go:nosplit
func Alloc() uintptr {
	ptr := kmm.alloc()
	buf := sys.UnsafeBuffer(ptr, PGSIZE)
	for i := range buf {
		buf[i] = 0
	}
	return ptr
}

//go:nosplit
func (v *vmmt) fixmap(va, pa, size, perm uintptr) bool {
	p := pageRoundDown(va)
	last := pageRoundDown(va + size - 1)
	for {
		pte := v.walkpgdir(p, true)
		if pte == nil {
			return false
		}
		if pte.present() {
			throw("fixmap remap")
		}
		*pte = entry(pa | perm)
		if p == last {
			break
		}
		p += PGSIZE
		pa += PGSIZE
	}
	return true
}

//go:nosplit
func (v *vmmt) walkpglvl(pg *entryPage, va uintptr, lvl int, alloc bool) *entry {
	idx := pageEntryIdx(va, lvl)
	if int(idx) >= len(pg) {
		throw("bad page index")
	}
	pe := &pg[idx]
	if lvl == 1 {
		return pe
	}

	// find entry
	if pe.present() {
		return pe
	}
	// not found and no alloc
	if !alloc {
		return nil
	}
	// alloc a page to map entry page
	addr := kmm.alloc()
	if addr == 0 {
		return nil
	}
	sys.Memclr(addr, PGSIZE)
	// map new page to entry
	*pe = entry(addr | PTE_P | PTE_W | PTE_U)
	// debug.PrintHex(uintptr(unsafe.Pointer(pg)))
	// debug.PrintStr("@")
	// debug.PrintHex(idx)
	// debug.PrintStr("=")
	// debug.PrintHex(addr)
	// debug.PrintStr("\n")
	return pe
}

//go:nosplit
func (v *vmmt) walkpgdir(va uintptr, alloc bool) *entry {
	epg := v.topPage
	var pe *entry
	for i := 4; i >= 1; i-- {
		pe = v.walkpglvl(epg, va, i, alloc)
		if pe == nil {
			return nil
		}
		epg = pe.entryPage()
	}
	return pe
}

//go:nosplit
func findMemTop() uintptr {
	if !multiboot.Enabled() {
		return DEFAULT_MEMTOP
	}
	var top uintptr
	for _, e := range multiboot.BootInfo.MmapEntries() {
		if e.Type != multiboot.MemoryAvailable {
			continue
		}
		ptop := e.Addr + e.Len
		if ptop > VMSTART {
			ptop = VMSTART
		}
		if top < uintptr(ptop) {
			top = uintptr(ptop)
		}
	}
	if top == 0 {
		return DEFAULT_MEMTOP
	}
	return top
}

//go:nosplit
func Init() {
	memtop = findMemTop()
	kmm.voffset = VMSTART
	kmm.freeRange(MEMSTART, memtop)

	vmm.topPage = (*entryPage)(unsafe.Pointer(kmm.alloc()))
	sys.Memclr(uintptr(unsafe.Pointer(vmm.topPage)), PGSIZE)
	// 4096-MEMTOP 用来让微内核访问到所有的地址空间
	// identity map all phy memory
	vmm.fixmap(4096, 4096, memtop-4096, PTE_P|PTE_W|PTE_U)

	lcr3(vmm.topPage)
	pageEnable()
}
