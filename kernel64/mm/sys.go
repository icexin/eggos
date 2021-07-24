package mm

import "syscall"

// SysMmap like Mmap but can run in user mode
// wraper of syscall.Mmap
func SysMmap(vaddr, size uintptr) uintptr {
	mem, _, err := syscall.Syscall6(syscall.SYS_MMAP, uintptr(vaddr), size, syscall.PROT_READ|syscall.PROT_WRITE, 0, 0, 0)
	if err != 0 {
		panic(err.Error())
	}
	return mem
}
