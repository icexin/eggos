package inet

import "github.com/google/netstack/tcpip"

var DefaultDevice Device

type Device interface {
	Mac() [6]byte
	Transmit(pkt tcpip.PacketBuffer) error
	SetReceiveCallback(func(b []byte))
}

func RegisterDevice(d Device) {
	DefaultDevice = d
}
