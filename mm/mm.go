package mm

import (
	"unsafe"

	"github.com/icexin/eggos/multiboot"
	"github.com/icexin/eggos/sys"
)

const (
	PGSIZE = 4 << 10
	// 1-20 Mb memory reverse for kernel image
	MEMSTART = 20 << 20
	// 默认可以使用的物理内存终止地址，如果能从grub那里获取就用grub的
	DEFAULT_MEMTOP = 256 << 20
	// 虚拟内存起始地址
	VMSTART = 1 << 30

	PTE_P = 0x001
	PTE_W = 0x002
	PTE_U = 0x004
)

var (
	memtop uintptr

	kmm = kmmt{voffset: VMSTART}
	vmm vmmt
)

//go:nosplit
func get_cr2() uintptr

//go:nosplit
func page_enable()

//go:nosplit
func lcr3(pgdir *pgdir)

func pageRoundUp(size uintptr) uintptr {
	return (size + PGSIZE - 1) &^ (PGSIZE - 1)
}

func pageRoundDown(v uintptr) uintptr {
	return v &^ (PGSIZE - 1)
}

func pageDirIdx(v uintptr) uintptr {
	return v >> 22 & 0x3ff
}

func pageTabIdx(v uintptr) uintptr {
	return v >> 12 & 0x3ff
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
		panic("virtual memory address all used")
	}
	return p
}

//go:nosplit
func (k *kmmt) alloc() uintptr {
	r := k.freelist
	if r == nil {
		panic("kmemt.alloc")
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
		panic("kmemt.free")
		return
	}
	r := (*page)(unsafe.Pointer(p))
	r.next = k.freelist
	k.freelist = r
}

//go:notinheap
type pgtab [1024]pdet

//go:notinheap
type pgdir [1024]pdet

//go:notinheap
type pdet uintptr

func (p pdet) present() bool {
	return p&PTE_P != 0
}

func (p pdet) addr() uintptr {
	return uintptr(p) &^ 0xfff
}

func (p pdet) pgtab() *pgtab {
	return (*pgtab)(unsafe.Pointer(p.addr()))
}

type vmmt struct {
	pgdir *pgdir
}

//go:nosplit
func (v *vmmt) munmap(va, size uintptr) bool {
	// println("mumap va=", va, " size=", size)
	p := pageRoundDown(va)
	last := pageRoundDown(va + size - 1)
	for {
		pte := v.walkpgdir(p, false)
		if pte == nil {
			return false
		}
		if !pte.present() {
			return false
		}
		kmm.free(pte.addr())
		*pte = pdet(pte.addr())
		if p == last {
			break
		}
		p += PGSIZE
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
			println("remap va=", unsafe.Pointer(p), " pa=", pte.addr(), " pte", pte)
			panic("remap")
		} else {
			pa = kmm.alloc()
			sys.Memclr(pa, PGSIZE)
			*pte = pdet(pa | perm)
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
	return va
}

//go:nosplit
func Fixmap(va, pa, size uintptr) {
	vmm.fixmap(va, pa, size, PTE_P|PTE_W|PTE_U)
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
			panic("remap")
		}
		*pte = pdet(pa | perm)
		if p == last {
			break
		}
		p += PGSIZE
		pa += PGSIZE
	}
	return true
}

//go:nosplit
func (v *vmmt) walkpgdir(va uintptr, alloc bool) *pdet {
	pdx := pageDirIdx(va)
	pde := &v.pgdir[pdx]
	if pde.present() {
		pgtab := pde.pgtab()
		return &pgtab[pageTabIdx(va)]
	}
	if !alloc {
		return nil
	}
	addr := kmm.alloc()
	if addr == 0 {
		return nil
	}
	sys.Memclr(addr, PGSIZE)
	*pde = pdet(addr | PTE_P | PTE_W | PTE_U)
	pgtab := pde.pgtab()
	return &pgtab[pageTabIdx(va)]
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

	vmm.pgdir = (*pgdir)(unsafe.Pointer(kmm.alloc()))
	sys.Memclr(uintptr(unsafe.Pointer(vmm.pgdir)), PGSIZE)
	// 4096-MEMTOP 用来让微内核访问到所有的地址空间
	// identity map all phy memory
	vmm.fixmap(4096, 4096, memtop-4096, PTE_P|PTE_W|PTE_U)

	lcr3(vmm.pgdir)
	page_enable()
}
