package multiboot

import "unsafe"

var (
	enabled  bool
	BootInfo Info
)

func Enabled() bool {
	return enabled
}

func Init(magic uintptr, mbiptr uintptr) {
	if magic != bootloaderMagic {
		return
	}
	enabled = true
	mbi := (*Info)(unsafe.Pointer(mbiptr))
	BootInfo = *mbi
}
