package e1000

import (
	"errors"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/icexin/eggos/drivers/pci"
	"github.com/icexin/eggos/drivers/pic"
	"github.com/icexin/eggos/inet"
	"github.com/icexin/eggos/kernel/mm"
	"github.com/icexin/eggos/kernel/sys"
	"github.com/icexin/eggos/log"

	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

var _ pci.Driver = (*driver)(nil)

type driver struct {
	mac   [6]byte
	ident []pci.Identity

	rxdescs *[NUM_RX_DESCS]rxdesc
	txdescs *[NUM_TX_DESCS]txdesc

	txidx, rxidx uint32

	rxfunc func([]byte)

	bar uintptr
	dev *pci.Device
}

func newDriver() *driver {
	return &driver{}
}

func (d *driver) Name() string {
	return "e1000"
}

func (d *driver) Idents() []pci.Identity {
	return []pci.Identity{
		{0x8086, 0x100e},
		{0x8086, 0x153a},
		{0x8086, 0x10ea},
		{0x8086, 0x10d3},
		{0x8086, 0x15b8},
	}
}

func (d *driver) Mac() [6]byte {
	return d.mac
}

func (d *driver) SetReceiveCallback(cb func([]byte)) {
	d.rxfunc = cb
}

func (d *driver) writecmd(reg uint16, val uint32) {
	addr := (*uint32)(unsafe.Pointer(d.bar + uintptr(reg)))
	atomic.StoreUint32(addr, val)
}

func (d *driver) readcmd(reg uint16) uint32 {
	addr := (*uint32)(unsafe.Pointer(d.bar + uintptr(reg)))
	return atomic.LoadUint32(addr)
}

func (d *driver) detecteeprom() bool {
	haseeprom := false
	d.writecmd(REG_EEPROM, 1)
	for i := 0; i < 1000 && !haseeprom; i++ {
		val := d.readcmd(REG_EEPROM)
		if val&0x10 != 0 {
			haseeprom = true
		}
	}
	return haseeprom
}

func (d *driver) readeeprom(addr uint8) uint16 {
	var temp uint32
	d.writecmd(REG_EEPROM, 1|(uint32(addr)<<8))
	for {
		temp = d.readcmd(REG_EEPROM)
		if temp&(1<<4) != 0 {
			break
		}
	}
	return uint16((temp >> 16) & 0xFFFF)
}

func (d *driver) readmac() {
	haseeprom := d.detecteeprom()
	if !haseeprom {
		temp := d.readcmd(REG_RXADDR)
		d.mac[0] = byte(temp & 0xff)
		d.mac[1] = byte(temp >> 8 & 0xff)
		d.mac[2] = byte(temp >> 16 & 0xff)
		d.mac[3] = byte(temp >> 24 & 0xff)
		temp = d.readcmd(REG_RXADDR + 4)
		d.mac[4] = byte(temp & 0xff)
		d.mac[5] = byte(temp >> 8 & 0xff)
		return
	}

	var temp uint16
	temp = d.readeeprom(0)
	d.mac[0] = byte(temp & 0xff)
	d.mac[1] = byte(temp >> 8 & 0xff)
	temp = d.readeeprom(1)
	d.mac[2] = byte(temp & 0xff)
	d.mac[3] = byte(temp >> 8 & 0xff)
	temp = d.readeeprom(2)
	d.mac[4] = byte(temp & 0xff)
	d.mac[5] = byte(temp >> 8 & 0xff)
}

func (d *driver) Init(dev *pci.Device) error {
	if unsafe.Sizeof(rxdesc{}) != 16 {
		panic("bad rxdesc size")
	}

	if unsafe.Sizeof(txdesc{}) != 16 {
		panic("bad txdesc size")
	}

	d.dev = dev

	log.Infof("[e1000] enable bus master")
	dev.Addr.EnableBusMaster()

	// mmap bar address
	baddr, blen, _, ismem := dev.Addr.ReadBAR(0)
	if !ismem {
		panic("not memory bar")
	}
	mm.SysFixedMmap(uintptr(baddr), uintptr(baddr), uintptr(blen))
	d.bar = uintptr(baddr)
	log.Infof("[e1000] mmap for bar0 0x%x", d.bar)

	// alloc desc
	descptr := mm.Alloc()
	d.rxdescs = (*[NUM_RX_DESCS]rxdesc)(unsafe.Pointer(descptr))
	descptr += unsafe.Sizeof([NUM_RX_DESCS + 1]rxdesc{})
	d.txdescs = (*[NUM_TX_DESCS]txdesc)(unsafe.Pointer(descptr))

	// disable all intrs
	d.writecmd(REG_IMC, 0xffffffff)

	log.Infof("[e1000] begin reset")
	// Reset the device.
	d.writecmd(REG_CTRL, d.readcmd(REG_CTRL)|CTRL_RST)

	// Wait until the device gets reset.
	// for (d.readcmd(REG_CTRL) & CTRL_RST) != 0 {
	// }
	time.Sleep(time.Microsecond)
	log.Infof("[e1000] reset done")
	// again disable all intrs after reset
	d.writecmd(REG_IMC, 0xffffffff)

	// Link up!
	d.writecmd(REG_CTRL, d.readcmd(REG_CTRL)|CTRL_SLU|CTRL_ASDE)
	log.Infof("[e1000] link up")

	// Fill Multicast Table Array with zeros.
	for i := uint16(0); i < 0x80; i++ {
		d.writecmd(REG_MTA_BASE+i*4, 0)
	}

	// Initialize RX queue.
	for i := 0; i < NUM_RX_DESCS; i++ {
		desc := &d.rxdescs[i]
		desc.paddr = uint64(mm.Alloc())
	}

	rxDescAddr := uintptr(unsafe.Pointer(&d.rxdescs[0]))
	if rxDescAddr&0xfffffff0 != rxDescAddr {
		panic("addr of rx desc must be 16 byte align")
	}
	d.writecmd(REG_RDBAL, uint32(rxDescAddr&0xfffffff0))
	d.writecmd(REG_RDBAH, 0)
	d.writecmd(REG_RDLEN, uint32(unsafe.Sizeof(*d.rxdescs)))
	d.writecmd(REG_RDH, 0)
	d.writecmd(REG_RDT, NUM_RX_DESCS-1)
	d.writecmd(REG_RCTL, RCTL_EN|RCTL_SECRC|RCTL_BSIZE|RCTL_BAM)

	// Initialize TX queue.
	for i := 0; i < NUM_TX_DESCS; i++ {
		desc := &d.txdescs[i]
		desc.paddr = uint64(mm.Alloc())
		desc.status = 1
	}
	txDescAddr := uintptr(unsafe.Pointer(&d.txdescs[0]))
	if txDescAddr&0xfffffff0 != txDescAddr {
		panic("addr of tx desc must be 16 byte align")
	}
	d.writecmd(REG_TDBAL, uint32(txDescAddr&0xffffffff))
	d.writecmd(REG_TDBAH, 0)
	d.writecmd(REG_TDLEN, uint32(unsafe.Sizeof(*d.txdescs)))
	d.writecmd(REG_TDH, 0)
	d.writecmd(REG_TDT, 0)
	d.writecmd(REG_TCTL, TCTL_EN|TCTL_PSP)

	// Enable interrupts.
	d.writecmd(REG_IMS, IMS_RXT0)
	// clear pending intrs
	d.readcmd(REG_ICR)
	d.writecmd(REG_ICR, ^uint32(0))

	log.Infof("[e1000] begin read mac")
	d.readmac()
	log.Infof("[e1000] mac:%x", d.mac)
	// go d.recvloop()
	return nil
}

func (d *driver) Transmit(pkt *stack.PacketBuffer) error {
	desc := &d.txdescs[d.txidx]
	if desc.status == 0 {
		return errors.New("tx queue full")
	}

	txbuf := sys.UnsafeBuffer(uintptr(desc.paddr), mm.PGSIZE)

	r := buffer.NewVectorisedView(pkt.Size(), pkt.Views())
	pktlen, _ := r.Read(txbuf)

	desc.cmd = TX_DESC_IFCS | TX_DESC_EOP | TX_DESC_RS
	desc.len = uint16(pktlen)
	desc.cso = 0
	desc.status = 0
	desc.css = 0
	desc.special = 0

	// Notify the device.
	d.txidx = (d.txidx + 1) % NUM_TX_DESCS
	d.writecmd(REG_TDT, uint32(d.txidx))

	// for atomic.LoadInt32((*int32)(unsafe.Pointer(&desc.status))) == 0 {
	// }
	// log.Infof("[e1000] send %d bytes", pktlen)
	return nil
}

func (d *driver) Intr() {
	defer pic.EnableIRQ(uint16(d.dev.IRQLine))
	defer pic.EOI(uintptr(d.dev.IRQNO))
	cause := d.readcmd(REG_ICR)
	// log.Infof("[e1000] cause %x", cause)
	// clear ICR register
	// e1000e may not clear upon read
	d.writecmd(REG_ICR, ^uint32(0))
	if cause&ICR_RXT0 == 0 {
		return
	}

	for d.readpkt() {
	}
}

func (d *driver) recvloop() {
	for {
		// syscall.Nanosleep(&syscall.Timespec{
		// 	Nsec: 1000000 * 10,
		// }, nil)
		time.Sleep(time.Millisecond * 10)
		for d.readpkt() {
		}
	}
}

func (d *driver) readpkt() bool {
	d.rxidx = d.readcmd(REG_RDT)
	d.rxidx = (d.rxidx + 1) % NUM_RX_DESCS
	// log.Infof("[e1000] head:%d", d.readcmd(0x02810))
	// log.Infof("[e1000] tail:%d", d.readcmd(0x02818))
	// log.Infof("[e1000] rxidx:%d", d.rxidx)
	desc := &d.rxdescs[d.rxidx]

	// fmt.Printf("status:%x\n", atomic.LoadUint32((*uint32)(unsafe.Pointer(&desc.status))))
	// We don't support a large packet which spans multiple descriptors.
	const bits = RX_DESC_DD | RX_DESC_EOP
	if (desc.status & bits) != bits {
		return false
	}

	buf := sys.UnsafeBuffer(uintptr(desc.paddr), int(desc.len))
	// log.Infof("[e1000] read %d bytes", desc.len)
	if d.rxfunc != nil {
		d.rxfunc(buf)
	}

	// Tell the device that we've tasked a received packet.
	desc.status = 0
	d.writecmd(REG_RDT, d.rxidx)

	return true
}

func init() {
	d := newDriver()
	inet.RegisterDevice(d)
	pci.Register(d)
}
