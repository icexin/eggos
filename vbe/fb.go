package vbe

import (
	"image"
	"unsafe"

	"github.com/icexin/eggos/kernel/mm"
	"github.com/icexin/eggos/multiboot"
	"github.com/icexin/eggos/uart"
)

const bootloaderMagic = 0x2BADB002

type framebufferInfo struct {
	Addr   uint64
	Pitch  uint32
	Width  uint32
	Height uint32
}

var (
	info   framebufferInfo
	buffer []uint8
	fbbuf  []uint8

	DefaultView *View
	currentView *View
)

func bufcopy(dst, src []uint8, stride int, rect image.Rectangle, op func([]uint8, []uint8)) {
	miny := rect.Min.Y
	maxy := rect.Max.Y
	minx := rect.Min.X * 4
	maxx := rect.Max.X * 4
	for j := miny; j < maxy; j++ {
		srcline := src[j*stride : (j+1)*stride]
		dstline := dst[j*stride : (j+1)*stride]
		op(dstline[minx:maxx], srcline[minx:maxx])
	}
}

func SaveCurrView() *View {
	return currentView
}

func SetCurrView(v *View) {
	currentView = v
	v.Commit()
}

func IsEnable() bool {
	return fbbuf != nil
}

func Init() {
	bootInfo := &multiboot.BootInfo
	if bootInfo.Flags&multiboot.FlagInfoVideoInfo == 0 {
		uart.WriteString("[video] can't found video info from bootloader, video disabled\n")
		return
	}
	if bootInfo.FramebufferType != 1 {
		uart.WriteString("[video] framebuffer only support RGB color\n")
		return
	}
	if bootInfo.FramebufferBPP != 32 {
		uart.WriteString("[video] framebuffer only support 32 bit color\n")
		return
	}

	info = framebufferInfo{
		Addr:   bootInfo.FramebufferAddr,
		Width:  bootInfo.FramebufferWidth,
		Height: bootInfo.FramebufferHeight,
		Pitch:  bootInfo.FramebufferPitch,
	}

	mm.SysFixedMmap(uintptr(info.Addr), uintptr(info.Width*info.Height*4))
	fbbuf = (*[10 << 20]uint8)(unsafe.Pointer(uintptr(info.Addr)))[:info.Width*info.Height*4]
	buffer = make([]uint8, len(fbbuf))
	DefaultView = NewView()
	currentView = DefaultView
}
