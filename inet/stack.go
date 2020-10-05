package inet

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/icexin/eggos/debug"
	"github.com/icexin/eggos/inet/dhcp"

	"github.com/google/netstack/tcpip"
	"github.com/google/netstack/tcpip/adapters/gonet"
	"github.com/google/netstack/tcpip/header"
	"github.com/google/netstack/tcpip/network/arp"
	"github.com/google/netstack/tcpip/network/ipv4"
	"github.com/google/netstack/tcpip/stack"
	"github.com/google/netstack/tcpip/transport/tcp"
	"github.com/google/netstack/tcpip/transport/udp"
)

const (
	defaultNIC = 1
)

var (
	nstack *stack.Stack
)

func Listen(proto, addr string) (net.Listener, error) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	return gonet.NewListener(nstack, tcpip.FullAddress{
		NIC:  defaultNIC,
		Addr: tcpip.Address(tcpaddr.IP),
		Port: uint16(tcpaddr.Port),
	}, ipv4.ProtocolNumber)
}

func DialTCP(proto, addr string) (net.Conn, error) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	return gonet.DialTCP(nstack, tcpip.FullAddress{
		NIC:  defaultNIC,
		Addr: tcpip.Address(tcpaddr.IP),
		Port: uint16(tcpaddr.Port),
	}, ipv4.ProtocolNumber)
}

func e(err *tcpip.Error) error {
	if err == nil {
		return nil
	}
	return errors.New(err.String())
}

func Init() error {
	nstack = stack.New(stack.Options{
		NetworkProtocols:   []stack.NetworkProtocol{arp.NewProtocol(), ipv4.NewProtocol()},
		TransportProtocols: []stack.TransportProtocol{tcp.NewProtocol(), udp.NewProtocol()},
	})
	endpoint := New(&Options{})
	err := nstack.CreateNIC(defaultNIC, endpoint)
	if err != nil {
		return e(err)
	}
	err = nstack.AddAddress(defaultNIC, arp.ProtocolNumber, arp.ProtocolAddress)
	if err != nil {
		return e(err)
	}

	err1 := dodhcp(endpoint.LinkAddress())
	if err1 != nil {
		return err1
	}
	return nil

	// nstack.AddAddress(defaultNIC, ipv4.ProtocolNumber, tcpip.Address([]byte{10, 0, 2, 15}))
	nstack.AddAddress(defaultNIC, ipv4.ProtocolNumber, tcpip.Address([]byte{10, 0, 0, 7}))
	setroute(nstack, dhcp.Config{
		SubnetMask: tcpip.AddressMask([]byte{255, 255, 255, 0}),
		// Gateway:    tcpip.Address([]byte{10, 0, 2, 2}),
		Gateway: tcpip.Address([]byte{10, 0, 0, 1}),
	})

	return nil
}

func dodhcp(linkaddr tcpip.LinkAddress) error {
	dhcpclient := dhcp.NewClient(nstack, defaultNIC, linkaddr)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	debug.Logf("[inet] begin dhcp")
	err1 := dhcpclient.Request(ctx, "")
	cancel()
	if err1 != nil {
		return err1
	}
	debug.Logf("[inet] dhcp done")
	cfg := dhcpclient.Config()
	debug.Logf("[inet] addr:%v", dhcpclient.Address())
	debug.Logf("[inet] gateway:%v", cfg.Gateway)
	debug.Logf("[inet] mask:%v", cfg.SubnetMask)
	debug.Logf("[inet] dns:%v", cfg.DomainNameServer)

	setroute(nstack, dhcpclient.Config())
	return nil
}

func setroute(nstack *stack.Stack, cfg dhcp.Config) {
	nstack.SetRouteTable([]tcpip.Route{
		{
			Destination: header.IPv4EmptySubnet,
			Gateway:     cfg.Gateway,
			NIC:         defaultNIC,
		}},
	)
}
