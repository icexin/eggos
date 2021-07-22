package kernel64

import (
	"github.com/icexin/eggos/sys"
	"github.com/icexin/eggos/uart"
)

const (
	_MSR_LSTAR = 0xc0000082
	_MSR_STAR  = 0xc0000081
	_MSR_FSTAR = 0xc0000084

	_MSR_IA32_EFER = 0xc0000080
	_MSR_FS_BASE   = 0xc0000100

	_EFER_SCE = 1 << 0 // Enable SYSCALL.
)

//go:nosplit
func syscallEntry() {
	uart.WriteString("syscall")
	for {
	}
}

//go:nosplit
func syscallInit() {
	// write syscall selector
	wrmsr(_MSR_STAR, 8<<32)
	// clear IF when enter syscall
	wrmsr(_MSR_FSTAR, 0x200)
	// set syscall entry
	wrmsr(_MSR_LSTAR, sys.FuncPC(syscallEntry))

	// Enable SYSCALL instruction.
	efer := rdmsr(_MSR_IA32_EFER)
	wrmsr(_MSR_IA32_EFER, efer|_EFER_SCE)
}
