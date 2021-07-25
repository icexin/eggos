package mm

import "syscall"

const (
	// sync with kernel
	_SYS_FIXED_MMAP = 502
)

// SysMmap like Mmap but can run in user mode
// wraper of syscall.Mmap
func SysMmap(vaddr, size uintptr) uintptr {
	mem, _, err := syscall.Syscall6(syscall.SYS_MMAP, uintptr(vaddr), size, syscall.PROT_READ|syscall.PROT_WRITE, 0, 0, 0)
	if err != 0 {
		panic(err.Error())
	}
	return mem
}

// SysFixedMmap map the same physical address to the virtual address
// run in user mode
func SysFixedMmap(addr, size uintptr) {
	_, _, err := syscall.Syscall(_SYS_FIXED_MMAP, addr, size, 0)
	if err != 0 {
		panic(err.Error())
	}
}
