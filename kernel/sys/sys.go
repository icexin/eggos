package sys

import "unsafe"

const PtrSize = 4 << (^uintptr(0) >> 63) // unsafe.Sizeof(uintptr(0)) but an ideal const

const PageSize = 4 << 10

//go:nosplit
func Outb(port uint16, data byte)

//go:nosplit
func Inb(port uint16) byte

//go:nosplit
func Outl(port uint16, data uint32)

//go:nosplit
func Inl(port uint16) uint32

//go:nosplit
func Cli()

//go:nosplit
func Sti()

//go:nosplit
func Hlt()

//go:nosplit
func Cr2() uintptr

//go:nosplit
func Flags() uintptr

//go:nosplit
func UnsafeBuffer(p uintptr, n int) []byte {
	return (*[1 << 30]byte)(unsafe.Pointer(p))[:n]
}

//go:nosplit
func Memclr(p uintptr, n int) {
	s := (*(*[1 << 30]byte)(unsafe.Pointer(p)))[:n]
	// the compiler will emit runtime.memclrNoHeapPointers
	for i := range s {
		s[i] = 0
	}
}

// funcPC returns the entry PC of the function f.
// It assumes that f is a func value. Otherwise the behavior is undefined.
// CAREFUL: In programs with plugins, funcPC can return different values
// for the same function (because there are actually multiple copies of
// the same function in the address space). To be safe, don't use the
// results of this function in any == expression. It is only safe to
// use the result as an address at which to start executing code.
//go:nosplit
func FuncPC(f interface{}) uintptr {
	return **(**uintptr)(unsafe.Pointer((uintptr(unsafe.Pointer(&f)) + PtrSize)))
}

//go:nosplit
func Fxsave(addr uintptr)

//go:nosplit
func SetAX(val uintptr)

//go:nosplit
func CS() uintptr
