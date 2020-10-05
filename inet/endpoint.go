package inet

import (
	"sync"

	"github.com/google/netstack/tcpip"
	"github.com/google/netstack/tcpip/buffer"
	"github.com/google/netstack/tcpip/header"
	"github.com/google/netstack/tcpip/stack"
)

type endpoint struct {
	hdrsize    int
	cap        stack.LinkEndpointCapabilities
	addr       tcpip.LinkAddress
	dispatcher stack.NetworkDispatcher

	// protect the following fields.
	mutex  sync.Mutex
	device Device
}

type Options struct {
	// MTU is the mtu to use for this endpoint.
	MTU uint32

	// Address is the link address for this endpoint. Only used if
	// EthernetHeader is true.
	Address tcpip.LinkAddress
}

func New(opt *Options) stack.LinkEndpoint {
	mac := DefaultDevice.Mac()
	caps := stack.LinkEndpointCapabilities(0)
	caps |= stack.CapabilityResolutionRequired
	e := &endpoint{
		hdrsize: header.EthernetMinimumSize,
		cap:     caps,
		addr:    tcpip.LinkAddress(mac[:]),
		device:  DefaultDevice,
	}

	e.device.SetReceiveCallback(e.onrx)
	return e
}

// MTU is the maximum transmission unit for this endpoint. This is
// usually dictated by the backing physical network; when such a
// physical network doesn't exist, the limit is generally 64k, which
// includes the maximum size of an IP packet.
func (e *endpoint) MTU() uint32 {
	return 1500
}

// Capabilities returns the set of capabilities supported by the
// endpoint.
func (e *endpoint) Capabilities() stack.LinkEndpointCapabilities {
	return e.cap
}

// MaxHeaderLength returns the maximum size the data link (and
// lower level layers combined) headers can have. Higher levels use this
// information to reserve space in the front of the packets they're
// building.
func (e *endpoint) MaxHeaderLength() uint16 {
	return uint16(e.hdrsize)
}

// LinkAddress returns the link address (typically a MAC) of the
// link endpoint.
func (e *endpoint) LinkAddress() tcpip.LinkAddress {
	return e.addr
}

// WritePacket writes a packet with the given protocol through the
// given route. It sets pkt.LinkHeader if a link layer header exists.
// pkt.NetworkHeader and pkt.TransportHeader must have already been
// set.
//
// To participate in transparent bridging, a LinkEndpoint implementation
// should call eth.Encode with header.EthernetFields.SrcAddr set to
// r.LocalLinkAddress if it is provided.
func (e *endpoint) WritePacket(r *stack.Route, gso *stack.GSO, protocol tcpip.NetworkProtocolNumber, pkt tcpip.PacketBuffer) *tcpip.Error {
	// Add ethernet header if needed.
	eth := header.Ethernet(pkt.Header.Prepend(header.EthernetMinimumSize))
	pkt.LinkHeader = buffer.View(eth)
	ethHdr := &header.EthernetFields{
		DstAddr: r.RemoteLinkAddress,
		Type:    protocol,
	}

	// Preserve the src address if it's set in the route.
	if r.LocalLinkAddress != "" {
		ethHdr.SrcAddr = r.LocalLinkAddress
	} else {
		ethHdr.SrcAddr = e.addr
	}
	eth.Encode(ethHdr)
	e.mutex.Lock()
	err := e.device.Transmit(pkt)
	if err != nil {
		e.mutex.Unlock()
		return tcpip.ErrAlreadyBound
	}
	e.mutex.Unlock()
	return nil
}

// WritePackets writes packets with the given protocol through the
// given route.
//
// Right now, WritePackets is used only when the software segmentation
// offload is enabled. If it will be used for something else, it may
// require to change syscall filters.
func (e *endpoint) WritePackets(r *stack.Route, gso *stack.GSO, hdrs []stack.PacketDescriptor, payload buffer.VectorisedView, protocol tcpip.NetworkProtocolNumber) (int, *tcpip.Error) {
	panic("not implemented") // TODO: Implement
}

// WriteRawPacket writes a packet directly to the link. The packet
// should already have an ethernet header.
func (e *endpoint) WriteRawPacket(vv buffer.VectorisedView) *tcpip.Error {
	panic("not implemented") // TODO: Implement
}

// Attach attaches the data link layer endpoint to the network-layer
// dispatcher of the stack.
func (e *endpoint) Attach(dispatcher stack.NetworkDispatcher) {
	e.dispatcher = dispatcher
}

// IsAttached returns whether a NetworkDispatcher is attached to the
// endpoint.
func (e *endpoint) IsAttached() bool {
	return e.dispatcher != nil
}

// Wait waits for any worker goroutines owned by the endpoint to stop.
//
// For now, requesting that an endpoint's worker goroutine(s) stop is
// implementation specific.
//
// Wait will not block if the endpoint hasn't started any goroutines
// yet, even if it might later.
func (e *endpoint) Wait() {
}

func (e *endpoint) onrx(buf []byte) {
	view := buffer.View(buf)
	eth := header.Ethernet(view[:header.EthernetMinimumSize])
	p := eth.Type()
	remote := eth.SourceAddress()
	local := eth.DestinationAddress()

	//fmt.Printf("type:%v, remote:%s, local:%s\n", p, remote, local)

	pkt := tcpip.PacketBuffer{
		Data:       buffer.NewVectorisedView(len(buf), []buffer.View{view}),
		LinkHeader: buffer.View(eth),
	}
	pkt.Data.TrimFront(e.hdrsize)
	e.dispatcher.DeliverNetworkPacket(e, remote, local, p, pkt)
}
