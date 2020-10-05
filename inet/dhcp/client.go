// Copyright 2016 The Netstack Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/icexin/eggos/debug"

	"github.com/google/netstack/tcpip"
	"github.com/google/netstack/tcpip/adapters/gonet"
	nheader "github.com/google/netstack/tcpip/header"
	"github.com/google/netstack/tcpip/network/ipv4"
	"github.com/google/netstack/tcpip/stack"
)

// Client is a DHCP client.
type Client struct {
	stack    *stack.Stack
	nicid    tcpip.NICID
	linkAddr tcpip.LinkAddress

	mu          sync.Mutex
	addr        tcpip.Address
	cfg         Config
	lease       time.Duration
	cancelRenew func()
}

// NewClient creates a DHCP client.
//
// TODO(crawshaw): add s.LinkAddr(nicid) to *stack.Stack.
func NewClient(s *stack.Stack, nicid tcpip.NICID, linkAddr tcpip.LinkAddress) *Client {
	return &Client{
		stack:    s,
		nicid:    nicid,
		linkAddr: linkAddr,
	}
}

// Start starts the DHCP client.
// It will periodically search for an IP address using the Request method.
func (c *Client) Start() {
	go func() {
		for {
			log.Print("DHCP request")
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			err := c.Request(ctx, "")
			cancel()
			if err == nil {
				break
			}
		}
		log.Printf("DHCP acquired IP %s for %s", c.Address(), c.Config().LeaseLength)
	}()
}

// Address reports the IP address acquired by the DHCP client.
func (c *Client) Address() tcpip.Address {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.addr
}

// Config reports the DHCP configuration acquired with the IP address lease.
func (c *Client) Config() Config {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cfg
}

// Shutdown relinquishes any lease and ends any outstanding renewal timers.
func (c *Client) Shutdown() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.addr != "" {
		c.stack.RemoveAddress(c.nicid, c.addr)
	}
	if c.cancelRenew != nil {
		c.cancelRenew()
	}
}

func e(err *tcpip.Error) error {
	return errors.New(err.String())
}

// Request executes a DHCP request session.
//
// On success, it adds a new address to this client's TCPIP stack.
// If the server sets a lease limit a timer is set to automatically
// renew it.
func (c *Client) Request(ctx context.Context, requestedAddr tcpip.Address) error {
	tcperr := c.stack.AddAddress(c.nicid, ipv4.ProtocolNumber, nheader.IPv4Any)
	if tcperr != nil {
		return e(tcperr)
	}
	defer c.stack.RemoveAddress(c.nicid, nheader.IPv4Any)

	clientAddr := tcpip.FullAddress{
		Addr: nheader.IPv4Broadcast,
		Port: clientPort,
		NIC:  c.nicid,
	}

	serverAddr := &net.UDPAddr{
		IP:   net.IPv4(255, 255, 255, 255),
		Port: serverPort,
	}

	conn, err := gonet.DialUDP(c.stack, &clientAddr, nil, ipv4.ProtocolNumber)
	if err != nil {
		return err
	}
	defer conn.Close()

	var xid [4]byte
	rand.Read(xid[:])

	// DHCPDISCOVERY
	options := options{
		{optDHCPMsgType, []byte{byte(dhcpDISCOVER)}},
		{optParamReq, []byte{
			1,  // request subnet mask
			3,  // request router
			15, // domain name
			6,  // domain name server
		}},
	}
	if requestedAddr != "" {
		options = append(options, option{optReqIPAddr, []byte(requestedAddr)})
	}
	h := make(header, headerBaseSize+options.len())
	h.init()
	h.setOp(opRequest)
	copy(h.xidbytes(), xid[:])
	h.setBroadcast()
	copy(h.chaddr(), c.linkAddr)
	h.setOptions(options)

	_, err = conn.WriteTo(h, serverAddr)
	if err != nil {
		return err
	}

	v := make([]byte, 1024)
	// DHCPOFFER
	for {
		n, err := conn.Read(v)
		if err != nil {
			return err
		}
		h = header(v[:n])
		if h.isValid() && h.op() == opReply && bytes.Equal(h.xidbytes(), xid[:]) {
			break
		}
	}
	opts, err := h.options()
	if err != nil {
		return fmt.Errorf("dhcp offer: %v", err)
	}
	debug.Logf("[dhcp] offer done")

	var ack bool
	var cfg Config

	err = cfg.decode(opts)
	if err != nil {
		return err
	}

	// DHCPREQUEST
	addr := tcpip.Address(h.yiaddr())
	if err := c.stack.AddAddress(c.nicid, ipv4.ProtocolNumber, addr); err != nil {
		if err != tcpip.ErrDuplicateAddress {
			return e(err)
		}
	}
	defer func() {
		if ack {
			c.mu.Lock()
			c.addr = addr
			c.cfg = cfg
			c.mu.Unlock()
		} else {
			c.stack.RemoveAddress(c.nicid, addr)
		}
	}()
	h.setOp(opRequest)
	for i, b := 0, h.yiaddr(); i < len(b); i++ {
		b[i] = 0
	}
	h.setOptions([]option{
		{optDHCPMsgType, []byte{byte(dhcpREQUEST)}},
		{optReqIPAddr, []byte(addr)},
		{optDHCPServer, []byte(cfg.ServerAddress)},
	})
	debug.Logf("[dhcp] offer ip:%s server:%s", addr, cfg.ServerAddress)
	_, err = conn.WriteTo(h, serverAddr)
	if err != nil {
		return err
	}

	// DHCPACK
	for {
		n, err := conn.Read(v)
		if err != nil {
			return err
		}
		h = header(v[:n])
		if h.isValid() && h.op() == opReply && bytes.Equal(h.xidbytes(), xid[:]) {
			break
		}
	}
	opts, err = h.options()
	if err != nil {
		return fmt.Errorf("dhcp ack: %v", err)
	}
	if err := cfg.decode(opts); err != nil {
		return fmt.Errorf("dhcp ack bad options: %v", err)
	}
	msgtype, err := opts.dhcpMsgType()
	if err != nil {
		return fmt.Errorf("dhcp ack: %v", err)
	}
	ack = msgtype == dhcpACK
	if !ack {
		return fmt.Errorf("dhcp: request not acknowledged")
	}
	debug.Logf("[dhcp] lease:%s", cfg.LeaseLength)
	if cfg.LeaseLength != 0 {
		go c.renewAfter(cfg.LeaseLength)
	}
	return nil
}

func (c *Client) renewAfter(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancelRenew != nil {
		c.cancelRenew()
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.cancelRenew = cancel
	go func() {
		timer := time.NewTimer(d)
		defer timer.Stop()
		select {
		case <-ctx.Done():
		case <-timer.C:
			if err := c.Request(ctx, c.addr); err != nil {
				log.Printf("address renewal failed: %v", err)
				go c.renewAfter(1 * time.Minute)
			}
		}
	}()
}
