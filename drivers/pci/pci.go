package pci

import (
	"github.com/icexin/eggos/debug"
	"github.com/icexin/eggos/drivers/pic"
	"github.com/icexin/eggos/kernel/trap"
)

type Identity struct {
	Vendor uint16
	Device uint16
}

type Device struct {
	Ident Identity
	Addr  Address

	Class, SubClass uint8

	IRQLine uint8
	IRQNO   uint8
}

var devices []*Device

func Scan() []*Device {
	var devices []*Device
	for bus := int(0); bus < 256; bus++ {
		for dev := uint8(0); dev < 32; dev++ {
			for f := uint8(0); f < 8; f++ {
				addr := Address{
					Bus:    uint8(bus),
					Device: dev,
					Func:   f,
				}
				vendor := addr.ReadVendorID()
				if vendor == 0xffff {
					continue
				}
				devid := addr.ReadDeviceID()
				class := addr.ReadPCIClass()
				irqline := addr.ReadIRQLine()
				device := &Device{
					Ident: Identity{
						Vendor: vendor,
						Device: devid,
					},
					Addr:     addr,
					Class:    uint8((class >> 8) & 0xff),
					SubClass: uint8(class & 0xff),
					IRQLine:  irqline,
					IRQNO:    pic.IRQ_BASE + irqline,
				}
				devices = append(devices, device)
			}
		}
	}
	return devices
}

func findDev(idents []Identity) *Device {
	for _, ident := range idents {
		for _, dev := range devices {
			if dev.Ident == ident {
				return dev
			}
		}
	}
	return nil
}

func Init() {
	devices = Scan()
	for _, driver := range drivers {
		dev := findDev(driver.Idents())
		if dev == nil {
			debug.Logf("[pci] no pci device found for %v\n", driver.Name())
			continue
		}
		debug.Logf("[pci] found %x:%x for %s, irq:%d\n", dev.Ident.Vendor, dev.Ident.Device, driver.Name(), dev.IRQNO)
		driver.Init(dev)
		pic.EnableIRQ(uint16(dev.IRQLine))
		trap.Register(int(dev.IRQNO), driver.Intr)
	}
}
