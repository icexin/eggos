package inet

import (
	"context"
	"errors"
	"time"

	"github.com/icexin/eggos/inet/dhcp"
	"github.com/icexin/eggos/log"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/loopback"
	"gvisor.dev/gvisor/pkg/tcpip/network/arp"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
)

const (
	defaultNIC  = 1
	loopbackNIC = 2
)

var (
	nstack *stack.Stack
)

func e(err tcpip.Error) error {
	if err == nil {
		return nil
	}
	return errors.New(err.String())
}

func Init() {
	nstack = stack.New(stack.Options{
		NetworkProtocols:   []stack.NetworkProtocolFactory{arp.NewProtocol, ipv4.NewProtocol},
		TransportProtocols: []stack.TransportProtocolFactory{tcp.NewProtocol, udp.NewProtocol},
		HandleLocal:        true,
	})

	// add net card interface
	endpoint := New(&Options{})
	err := nstack.CreateNIC(defaultNIC, endpoint)
	if err != nil {
		panic(err)
	}
	err1 := dodhcp(endpoint.LinkAddress())
	if err1 != nil {
		panic(err)
	}

	// add loopback interface
	err = nstack.CreateNIC(loopbackNIC, loopback.New())
	if err != nil {
		panic(err)
	}
	addInterfaceAddr(nstack, loopbackNIC, tcpip.Address([]byte{127, 0, 0, 1}))
	return
}

func addInterfaceAddr(s *stack.Stack, nic tcpip.NICID, addr tcpip.Address) {
	s.AddAddress(nic, ipv4.ProtocolNumber, addr)
	// Add route for local network if it doesn't exist already.
	localRoute := tcpip.Route{
		Destination: addr.WithPrefix().Subnet(),
		Gateway:     "", // No gateway for local network.
		NIC:         nic,
	}

	for _, rt := range s.GetRouteTable() {
		if rt.Equal(localRoute) {
			return
		}
	}

	// Local route does not exist yet. Add it.
	s.AddRoute(localRoute)
}

func dodhcp(linkaddr tcpip.LinkAddress) error {
	dhcpclient := dhcp.NewClient(nstack, defaultNIC, linkaddr)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	log.Infof("[inet] begin dhcp")
	err1 := dhcpclient.Request(ctx, "")
	cancel()
	if err1 != nil {
		return err1
	}
	log.Infof("[inet] dhcp done")
	cfg := dhcpclient.Config()
	log.Infof("[inet] addr:%v", dhcpclient.Address())
	log.Infof("[inet] gateway:%v", cfg.Gateway)
	log.Infof("[inet] mask:%v", cfg.SubnetMask)
	log.Infof("[inet] dns:%v", cfg.DomainNameServer)

	addInterfaceAddr(nstack, defaultNIC, dhcpclient.Address())
	nstack.AddRoute(tcpip.Route{
		Destination: header.IPv4EmptySubnet,
		Gateway:     cfg.Gateway,
		NIC:         defaultNIC,
	})
	return nil
}
