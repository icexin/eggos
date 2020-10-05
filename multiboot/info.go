package multiboot

import "unsafe"

const (
	bootloaderMagic = 0x2BADB002
)

const (
	MemoryAvailable = 1 << iota
	MemoryReserved
	MemoryACPIReclaimable
	MemoryNVS
	MemoryBadRAM
)

type Flag uint32

const (
	FlagInfoMemory Flag = 1 << iota
	FlagInfoBootDev
	FlagInfoCmdline
	FlagInfoMods
	FlagInfoAoutSyms
	FlagInfoElfSHDR
	FlagInfoMemMap
	FlagInfoDriveInfo
	FlagInfoConfigTable
	FlagInfoBootLoaderName
	FlagInfoAPMTable
	FlagInfoVideoInfo
	FlagInfoFrameBuffer
)

// Info represents the Multiboot v1 info passed to the loaded kernel.
type Info struct {
	Flags    Flag
	MemLower uint32
	MemUpper uint32

	BootDevice uint32

	Cmdline uint32

	ModsCount uint32
	ModsAddr  uint32

	Syms [4]uint32

	MmapLength uint32
	MmapAddr   uint32

	DriversLength uint32
	DrivesrAddr   uint32

	ConfigTable uint32

	BootLoaderName uint32

	APMTable uint32

	VBEControlInfo  uint32
	VBEModeInfo     uint32
	VBEMode         uint16
	VBEInterfaceSeg uint16
	VBEInterfaceOff uint16
	VBEInterfaceLen uint16

	FramebufferAddr   uint64
	FramebufferPitch  uint32
	FramebufferWidth  uint32
	FramebufferHeight uint32
	FramebufferBPP    byte
	FramebufferType   byte
	ColorInfo         [6]byte
}

func (i *Info) MmapEntries() []MmapEntry {
	n := i.MmapLength / uint32(unsafe.Sizeof(MmapEntry{}))
	return (*[128]MmapEntry)(unsafe.Pointer(uintptr(i.MmapAddr)))[:n]
}

type MmapEntry struct {
	Size uint32
	Addr uint64
	Len  uint64
	Type uint32
}
