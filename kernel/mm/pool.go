package mm

import (
	"unsafe"

	"github.com/jspc/eggos/kernel/sys"
)

//go:notinheap
type memblk struct {
	next uintptr
}

// Pool used to manage fixed size memory block
//go:notinheap
type Pool struct {
	size uintptr
	head uintptr
}

// size will align ptr size
//go:nosplit
func PoolInit(p *Pool, size uintptr) {
	const align = sys.PtrSize - 1
	size = (size + align) &^ align
	p.size = size
}

//go:nosplit
func (p *Pool) grow() {
	start := kmm.alloc()
	end := start + PGSIZE
	for v := start; v+p.size <= end; v += p.size {
		p.Free(v)
	}
}

//go:nosplit
func (p *Pool) Alloc() uintptr {
	if p.head == 0 {
		p.grow()
	}
	ret := p.head
	h := (*memblk)(unsafe.Pointer(p.head))
	p.head = h.next
	sys.Memclr(ret, int(p.size))
	return ret
}

//go:nosplit
func (p *Pool) Free(ptr uintptr) {
	v := (*memblk)(unsafe.Pointer(ptr))
	v.next = p.head
	p.head = ptr
}
