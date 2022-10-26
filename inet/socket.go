package inet

import (
	"syscall"

	"github.com/jspc/eggos/kernel/isyscall"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"
)

func sysSocket(c *isyscall.Request) {
	domain := c.Arg(0)
	typ := c.Arg(1)
	// proto := c.Arg(2)
	if domain != syscall.AF_INET {
		c.SetErrorNO(syscall.EINVAL)
		return
	}
	if typ&syscall.SOCK_STREAM == 0 && typ&syscall.SOCK_DGRAM == 0 {
		c.SetErrorNO(syscall.EINVAL)
		return
	}

	var protoNum tcpip.TransportProtocolNumber
	switch {
	case typ&syscall.SOCK_STREAM != 0:
		protoNum = tcp.ProtocolNumber
	case typ&syscall.SOCK_DGRAM != 0:
		protoNum = udp.ProtocolNumber
	default:
		panic(typ)
	}

	wq := new(waiter.Queue)
	ep, err := nstack.NewEndpoint(protoNum, ipv4.ProtocolNumber, wq)
	if err != nil {
		c.SetError(e(err))
		return
	}

	sfile := allocSockFile(ep, wq)
	c.SetRet(uintptr(sfile.fd))

}

func sysListen(c *isyscall.Request) {
	sf, err := findSockFile(c.Arg(0))
	if err != nil {
		c.SetError(err)
		return
	}
	err = sf.Listen(c.Arg(1))
	if err != nil {
		c.SetError(err)
		return
	}
	c.SetRet(0)
}

func sysBind(c *isyscall.Request) {
	sf, err := findSockFile(c.Arg(0))
	if err != nil {
		c.SetError(err)
		return
	}
	err = sf.Bind(c.Arg(1), c.Arg(2))
	if err != nil {
		c.SetError(err)
		return
	}
	c.SetRet(0)

}

func sysAccept4(c *isyscall.Request) {
	sf, err := findSockFile(c.Arg(0))
	if err != nil {
		c.SetError(err)
		return
	}
	fd, err := sf.Accept4(c.Arg(1), c.Arg(2), c.Arg(3))
	if err != nil {
		c.SetError(err)
		return
	}
	c.SetRet(uintptr(fd))

}

func sysConnect(c *isyscall.Request) {
	sf, err := findSockFile(c.Arg(0))
	if err != nil {
		c.SetError(err)
		return
	}
	uaddr := c.Arg(1)
	uaddrlen := c.Arg(2)
	err = sf.Connect(uaddr, uaddrlen)
	c.SetError(err)
}

func sysSetsockopt(c *isyscall.Request) {
	sf, err := findSockFile(c.Arg(0))
	if err != nil {
		c.SetError(err)
		return
	}
	err = sf.Setsockopt(c.Arg(1), c.Arg(2), c.Arg(3), c.Arg(4))
	// if err != nil {
	// 	err = isyscall.EPANIC
	// }
	c.SetError(err)
}

func sysGetsockopt(c *isyscall.Request) {
	sf, err := findSockFile(c.Arg(0))
	if err != nil {
		c.SetError(err)
		return
	}
	err = sf.Getsockopt(c.Arg(1), c.Arg(2), c.Arg(3), c.Arg(4))
	if err != nil {
		c.SetError(err)
		return
	}
	c.SetRet(0)
}

func sysGetsockname(c *isyscall.Request) {
	sf, err := findSockFile(c.Arg(0))
	if err != nil {
		c.SetError(err)
		return
	}
	err = sf.Getsockname(c.Arg(1), c.Arg(2))
	// if err != nil {
	// 	err = isyscall.EPANIC
	// }
	c.SetError(err)
}

func sysGetpeername(c *isyscall.Request) {
	sf, err := findSockFile(c.Arg(0))
	if err != nil {
		c.SetError(err)
		return
	}
	err = sf.Getpeername(c.Arg(1), c.Arg(2))
	// if err != nil {
	// 	err = isyscall.EPANIC
	// }
	c.SetError(err)
}

func ntohs(n uint16) uint16 {
	return (n >> 8 & 0xff) | (n&0xff)<<8
}

func htons(n uint16) uint16 {
	return ntohs(n)
}

func init() {
	isyscall.Register(syscall.SYS_SOCKET, sysSocket)
	isyscall.Register(syscall.SYS_BIND, sysBind)
	isyscall.Register(syscall.SYS_LISTEN, sysListen)
	isyscall.Register(syscall.SYS_ACCEPT4, sysAccept4)
	isyscall.Register(syscall.SYS_CONNECT, sysConnect)
	isyscall.Register(syscall.SYS_SETSOCKOPT, sysSetsockopt)
	isyscall.Register(syscall.SYS_GETSOCKOPT, sysGetsockopt)
	isyscall.Register(syscall.SYS_GETSOCKNAME, sysGetsockname)
	isyscall.Register(syscall.SYS_GETPEERNAME, sysGetpeername)
}
