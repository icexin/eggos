package kernel

import (
	"unsafe"

	"github.com/icexin/eggos/mm"
	"github.com/icexin/eggos/multiboot"
	"github.com/icexin/eggos/pic"
	"github.com/icexin/eggos/sys"
	"github.com/icexin/eggos/uart"
)

const (
	STA_X    = 0x8
	STA_W    = 0x2
	STA_R    = 0x2
	DPL_USER = 0x60

	_AT_PAGESZ = 6
	_AT_NULL   = 0
)

var (
	gdt    [8]segDesc
	gdtptr [6]byte
)

type segDesc struct {
	limitLow        uint16
	baseLow         uint16
	baseMid         uint8
	attr1           uint8
	limitHightAttr2 uint8
	baseHigh        uint8
}

//go:nosplit
func sse_init()

//go:nosplit
func gdt_init()

//go:nosplit
func rt0()

//go:nosplit
func setSegDesc(seg *segDesc, base, limit uint32, flag uint16) {
	seg.limitLow = uint16((limit >> 12) & 0xffff)
	seg.baseLow = uint16(base & 0xffff)
	seg.baseMid = uint8((base >> 16) & 0xff)
	seg.attr1 = uint8(flag | 0x90)
	seg.limitHightAttr2 = uint8(0xc0 | (limit>>28)&0x0f)
	seg.baseHigh = uint8((base >> 24) & 0xff)
}

//go:nosplit
func setTssSegDesc(seg *segDesc, base, limit uint32) {
	seg.limitLow = uint16(limit & 0xffff)
	seg.baseLow = uint16(base & 0xffff)
	seg.baseMid = uint8((base >> 16) & 0xff)
	seg.attr1 = uint8(0x89)                              // P=1 DPL=0 TYPE=1001
	seg.limitHightAttr2 = uint8(0x00 | (limit>>16)&0x0f) // G=0
	seg.baseHigh = uint8((base >> 24) & 0xff)
}

//go:nosplit
func fillgdt() {
	gdt[0] = segDesc{}
	// kernel code
	setSegDesc(&gdt[1], 0, 0xffffffff, STA_X|STA_R)
	// kernel data
	setSegDesc(&gdt[2], 0, 0xffffffff, STA_W)
	// user code
	setSegDesc(&gdt[3], 0, 0xffffffff, STA_X|STA_R|DPL_USER)
	// user data
	setSegDesc(&gdt[4], 0, 0xffffffff, STA_W|DPL_USER)
	// task state(TSS)
	tssBase := uint32(uintptr(unsafe.Pointer(&taskstate)))
	tssLimit := uint32(unsafe.Sizeof(taskstate)) - 1
	setTssSegDesc(&gdt[5], tssBase, tssLimit)
	// go runtime tls
	setSegDesc(&gdt[6], tssBase, tssLimit, STA_W)
	// kernel tls
	setSegDesc(&gdt[7], tssBase, tssLimit, STA_W)

	limit := (*uint16)(unsafe.Pointer(&gdtptr[0]))
	base := (*uint32)(unsafe.Pointer(&gdtptr[2]))
	*limit = uint16(unsafe.Sizeof(gdt) - 1)
	*base = uint32(uintptr(unsafe.Pointer(&gdt[0])))
}

//go:nosplit
func settls(idx int, address, limit uint32) {
	setSegDesc(&gdt[idx], address, limit, STA_W|DPL_USER)
}

var argbuf [128]byte

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
	// no env, len(env) == 0
	// envput(&buf, 1)
	envTerm := (*uintptr)(unsafe.Pointer(envput(&buf, 0)))
	envGoDebug := (*uintptr)(unsafe.Pointer(envput(&buf, 0)))
	envput(&buf, 0)

	// put auxillary vector
	envput(&buf, _AT_PAGESZ)
	envput(&buf, sys.PageSize)
	envput(&buf, _AT_NULL)
	envput(&buf, 0)

	// alloc memory for argv[0]
	*argv0 = envdup(&buf, "eggos\x00")

	*envTerm = envdup(&buf, "TERM=xterm\x00")
	*envGoDebug = envdup(&buf, "GODEBUG=asyncpreemptoff=1\x00")
}

//go:nosplit
func preinit(magic uint32, mbi uintptr) {
	gdt_init()
	sse_init()
	multiboot.Init(magic, mbi)
	uart.PreInit()
	trap_init()
	mm.Init()
	thread_init()
	syscal_init()
	pic.Init()
	timer_init()
	schedule()
}
