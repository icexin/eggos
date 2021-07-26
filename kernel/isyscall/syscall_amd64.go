package isyscall

import "unsafe"

// must sync with kernel.trapFrame
type trapFrame struct {
	AX, BX, CX, DX    uintptr
	BP, SI, DI, R8    uintptr
	R9, R10, R11, R12 uintptr
	R13, R14, R15     uintptr

	Trapno, Err uintptr

	// pushed by hardware
	IP, CS, FLAGS, SP, SS uintptr
}

func NewRequest(tf uintptr) Request {
	return Request{
		tf: (*trapFrame)(unsafe.Pointer(tf)),
	}
}

//go:nosplit
func (t *trapFrame) NO() uintptr {
	return t.AX
}

//go:nosplit
func (t *trapFrame) Arg(n int) uintptr {
	switch n {
	case 0:
		return t.DI
	case 1:
		return t.SI
	case 2:
		return t.DX
	case 3:
		return t.R10
	case 4:
		return t.R8
	case 5:
		return t.R9
	default:
		return 0
	}
}

//go:nosplit
func (t *trapFrame) SetRet(v uintptr) {
	t.AX = v
}
