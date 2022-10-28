package vbe

import (
	"image"
	"unsafe"

	"github.com/jspc/eggos/drivers/multiboot"
	"github.com/jspc/eggos/drivers/uart"
	"github.com/jspc/eggos/kernel/mm"
)

// framebufferInfo specifies the dimensions and DMA address of the VBE
// framebuffer.
type framebufferInfo struct {
	// DMA address of the VBE framebuffer.
	Addr uint64
	// Pitch is the stride (in bytes) between vertically adjacent pixels.
	Pitch uint32
	// Width of frame buffer in pixels.
	Width uint32
	// Height of frame buffer in pixels.
	Height uint32
}

var (
	// info specifies VBE framebuffer information.
	info framebufferInfo
	// buffer is an RGBA-buffer with the same pixel format and dimensions as
	// fbbuf, used for double-buffering during rendering.
	buffer []uint8
	// fbbuf provides direct memory access to the VBE framebuffer.
	fbbuf []uint8

	// DefaultView is the default view used for rendering.
	DefaultView *View
	// currentView is the current view used for rendering.
	currentView *View
)

// bufcopy copies pixel data within the given rectangle from src to dst, based
// on the given stride and copy operation.
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

// SaveCurrView returns the current view.
func SaveCurrView() *View {
	return currentView
}

// SetCurrView sets the current view, commiting pixel data to the framebuffer.
func SetCurrView(v *View) {
	currentView = v
	v.Commit()
}

// IsEnable reports whether the VBE framebuffer has been initialized.
func IsEnable() bool {
	return fbbuf != nil
}

// Init initializes the VBE framebuffer based on the info data provided by
// multiboot.
func Init() {
	bootInfo := &multiboot.BootInfo
	if bootInfo.Flags&multiboot.FlagInfoVideoInfo == 0 {
		uart.WriteString("[video] can't find video info from bootloader, video disabled\n")
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

	mm.SysFixedMmap(uintptr(info.Addr), uintptr(info.Addr), uintptr(info.Width*info.Height*4))
	fbbuf = (*[10 << 20]uint8)(unsafe.Pointer(uintptr(info.Addr)))[:info.Width*info.Height*4]
	buffer = make([]uint8, len(fbbuf))
	DefaultView = NewView()
	currentView = DefaultView
}
