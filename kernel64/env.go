package kernel64

import (
	"unsafe"

	"github.com/icexin/eggos/sys"
	"gvisor.dev/gvisor/pkg/abi/linux"
)

//go:nosplit
func envput(pbuf *[]byte, v uintptr) uintptr {
	buf := *pbuf
	*(*uintptr)(unsafe.Pointer(&buf[0])) = v
	*pbuf = buf[unsafe.Sizeof(v):]
	return uintptr(unsafe.Pointer(&buf[0]))
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
	const argc = 1
	buf := sys.UnsafeBuffer(sp, 256)

	// put args
	envput(&buf, argc)
	argv0 := (*uintptr)(unsafe.Pointer(envput(&buf, 0)))
	// end of args
	envput(&buf, 0)
	envTerm := (*uintptr)(unsafe.Pointer(envput(&buf, 0)))
	envGoDebug := (*uintptr)(unsafe.Pointer(envput(&buf, 0)))
	// end of env
	envput(&buf, 0)

	// put auxillary vector
	envput(&buf, linux.AT_PAGESZ)
	envput(&buf, sys.PageSize)
	envput(&buf, linux.AT_NULL)
	envput(&buf, 0)

	// alloc memory for argv[0]
	*argv0 = envdup(&buf, "eggos\x00")

	*envTerm = envdup(&buf, "TERM=xterm\x00")
	*envGoDebug = envdup(&buf, "GODEBUG=asyncpreemptoff=1\x00")
}
