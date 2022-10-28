package kernel

import (
	"unsafe"

	"github.com/jspc/eggos/drivers/multiboot"
	"github.com/jspc/eggos/kernel/sys"
	"gvisor.dev/gvisor/pkg/abi/linux"
)

//go:nosplit
func envput(pbuf *[]byte, v uintptr) uintptr {
	buf := *pbuf
	p := unsafe.Pointer(&buf[0])
	// *p = v
	*(*uintptr)(p) = v
	// advance buffer
	*pbuf = buf[unsafe.Sizeof(v):]
	// return p
	return uintptr(unsafe.Pointer(&buf[0]))
}

// envptr used to alloc an *uintptr
//go:nosplit
func envptr(pbuf *[]byte) *uintptr {
	return (*uintptr)(unsafe.Pointer(envput(pbuf, 0)))
}

//go:nosplit
func envdup(pbuf *[]byte, s string) uintptr {
	buf := *pbuf
	copy(buf, s)
	*pbuf = buf[len(s):]
	return uintptr(unsafe.Pointer(&buf[0]))
}

//go:nosplit
func prepareArgs(sp uintptr) {
	buf := sys.UnsafeBuffer(sp, 256)

	var argc uintptr = 1
	// put argc slot
	envput(&buf, argc)
	arg0 := envptr(&buf)
	// end of args
	envput(&buf, 0)

	envTerm := envptr(&buf)
	envGoDebug := envptr(&buf)
	putKernelArgs(&buf)
	// end of env
	envput(&buf, 0)

	// put auxillary vector
	envput(&buf, linux.AT_PAGESZ)
	envput(&buf, sys.PageSize)
	envput(&buf, linux.AT_NULL)
	envput(&buf, 0)

	*arg0 = envdup(&buf, "eggos\x00")
	*envTerm = envdup(&buf, "TERM=xterm\x00")
	*envGoDebug = envdup(&buf, "GODEBUG=asyncpreemptoff=1\x00")
}

//go:nosplit
func putKernelArgs(pbuf *[]byte) uintptr {
	var cnt uintptr
	info := multiboot.BootInfo
	var flag = info.Flags
	if flag&multiboot.FlagInfoCmdline == 0 {
		return 0
	}
	var pos uintptr = uintptr(info.Cmdline)
	if pos == 0 {
		return cnt
	}

	var arg uintptr
	for {
		arg = strtok(&pos)
		if arg == 0 {
			break
		}
		envput(pbuf, arg)
		cnt++
	}
	return cnt
}

//go:nosplit
func strtok(pos *uintptr) uintptr {
	addr := *pos

	// skip spaces
	for {
		ch := bytedef(addr)
		if ch == 0 {
			return 0
		}
		if ch != ' ' {
			break
		}
		addr++
	}
	ret := addr
	// scan util read space and \0
	for {
		ch := bytedef(addr)
		if ch == ' ' {
			*(*byte)(unsafe.Pointer(addr)) = 0
			addr++
			break
		}
		if ch == 0 {
			break
		}
		addr++
	}
	*pos = addr
	return ret
}

//go:nosplit
func bytedef(addr uintptr) byte {
	return *(*byte)(unsafe.Pointer(addr))
}
