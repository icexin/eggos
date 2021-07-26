package inet

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/icexin/eggos/debug"
	"github.com/icexin/eggos/kernel/isyscall"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"
)

const (
	// see linux/net.h
	_SOCKET      = 1
	_BIND        = 2
	_CONNECT     = 3
	_LISTEN      = 4
	_ACCEPT      = 5
	_GETSOCKNAME = 6
	_GETPEERNAME = 7
	_SOCKETPAIR  = 8
	_SEND        = 9
	_RECV        = 10
	_SENDTO      = 11
	_RECVFROM    = 12
	_SHUTDOWN    = 13
	_SETSOCKOPT  = 14
	_GETSOCKOPT  = 15
	_SENDMSG     = 16
	_RECVMSG     = 17
	_ACCEPT4     = 18
	_RECVMMSG    = 19
	_SENDMMSG    = 20
)

func socketcall(c *isyscall.Request) {
	// Kernel interface gets call sub-number and pointer to a0.
	// see syscall/asm_linux_386.s rawsocketcall
	fn := c.Arg(0)
	args := (*[5]uintptr)(unsafe.Pointer(c.Arg(1)))

	if fn == _SOCKET {
		c.SetRet(sysSocket(uintptr(args[0]), uintptr(args[1]), uintptr(args[2])))
		c.Done()
		return
	}

	c.SetRet(0)

	sf, err := findSockFile(args[0])
	if err != nil {
		c.SetRet(isyscall.Error(err))
		c.Done()
		return
	}
	switch fn {
	case _LISTEN:
		err = sf.Listen(args[1])
	case _ACCEPT4:
		var fd int
		fd, err = sf.Accept4(args[1], args[2], args[3])
		c.SetRet(uintptr(fd))
	case _BIND:
		err = sf.Bind(args[1], args[2])
	case _CONNECT:
		err = sf.Connect(args[1], args[2])
	case _SETSOCKOPT:
		err = sf.Setsockopt(args[1], args[2], args[3], args[4])
	case _GETSOCKOPT:
		err = sf.Getsockopt(args[1], args[2], args[3], args[4])
	case _GETPEERNAME:
		err = sf.Getpeername(args[1], args[2])
	case _GETSOCKNAME:
		err = sf.Getsockname(args[1], args[2])

	default:
		debug.Logf("[socket] usupport fn:%d", fn)
		err = fmt.Errorf("unsupported socket fn:%d", fn)
	}

	if err != nil {
		c.SetRet(isyscall.Error(err))
	}
	c.Done()
}

func sysSocket(domain, typ, proto uintptr) uintptr {
	if domain != syscall.AF_INET {
		return isyscall.Errno(syscall.EINVAL)
	}
	if typ&syscall.SOCK_STREAM == 0 && typ&syscall.SOCK_DGRAM == 0 {
		return isyscall.Errno(syscall.EINVAL)
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
		return isyscall.Error(e(err))
	}

	sfile := allocSockFile(ep, wq)
	return uintptr(sfile.fd)
}

func ntohs(n uint16) uint16 {
	return (n >> 8 & 0xff) | (n&0xff)<<8
}

func htons(n uint16) uint16 {
	return ntohs(n)
}

func init() {
	// isyscall.Register(syscall.SYS_SOCKETCALL, socketcall)
}
