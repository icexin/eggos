package inet

import "gvisor.dev/gvisor/pkg/tcpip/stack"

var DefaultDevice Device

type Device interface {
	Mac() [6]byte
	Transmit(pkt *stack.PacketBuffer) error
	SetReceiveCallback(func(b []byte))
}

func RegisterDevice(d Device) {
	DefaultDevice = d
}
