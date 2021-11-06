package inet

import (
	"bytes"
	"errors"
	"net"
	"syscall"
	"time"
	"unsafe"

	"github.com/icexin/eggos/fs"
	"github.com/icexin/eggos/log"

	"gvisor.dev/gvisor/pkg/abi/linux"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/waiter"
)

//go:linkname evnotify github.com/icexin/eggos/kernel.epollNotify
func evnotify(fd, events uintptr)

type sockFile struct {
	fd int
	ep tcpip.Endpoint
	wq *waiter.Queue
}

func allocSockFile(ep tcpip.Endpoint, wq *waiter.Queue) *sockFile {
	fd, ni := fs.AllocInode()

	sfile := &sockFile{
		fd: fd,
		ep: ep,
		wq: wq,
	}
	sfile.setupEvent()

	ni.File = sfile
	return sfile
}

func findSockFile(fd uintptr) (*sockFile, error) {
	ni, err := fs.GetInode(int(fd))
	if err != nil {
		return nil, err
	}
	sf, ok := ni.File.(*sockFile)
	if !ok {
		return nil, syscall.EBADF
	}
	return sf, nil
}

func (s *sockFile) Read(p []byte) (int, error) {
	var terr tcpip.Error
	var result tcpip.ReadResult

	w := tcpip.SliceWriter(p)
	result, terr = s.ep.Read(&w, tcpip.ReadOptions{})

	switch terr.(type) {
	case nil:
	case *tcpip.ErrWouldBlock:
		return 0, syscall.EAGAIN
	case *tcpip.ErrClosedForReceive:
		return 0, nil
	default:
		log.Infof("[socket] read error:%s", terr)
		return 0, e(terr)
	}
	if result.Count < result.Total {
		// make next epoll_wait success
		s.evcallback(nil, waiter.EventIn)
	}
	return result.Count, nil
}

func (s *sockFile) Write(p []byte) (int, error) {
	n, terr := s.ep.Write(bytes.NewBuffer(p), tcpip.WriteOptions{})
	if n != 0 {
		return int(n), nil
	}

	switch terr.(type) {
	case *tcpip.ErrWouldBlock:
		return 0, syscall.EAGAIN
	case *tcpip.ErrClosedForSend:
		return 0, syscall.EPIPE
	default:
		log.Infof("[socket] write error:%s", terr)
		return 0, e(terr)
	}
}

func (s *sockFile) Close() error {
	s.ep.Close()
	return nil
}

type evcallback func(*waiter.Entry, waiter.EventMask)

func (e evcallback) Callback(entry *waiter.Entry, mask waiter.EventMask) {
	e(entry, mask)
}

func (s *sockFile) setupEvent() {
	s.wq.EventRegister(&waiter.Entry{
		Callback: evcallback(s.evcallback),
	}, waiter.EventIn|waiter.EventOut|waiter.EventErr|waiter.EventHUp)
}

func (s *sockFile) stopEvent() {
	s.wq.EventUnregister(nil)
}

func (s *sockFile) evcallback(e *waiter.Entry, mask waiter.EventMask) {
	// log.Infof("ev:%x fd:%d", mask, s.fd)
	// syscall.Syscall(kernel.SYS_EPOLL_NOTIFY, uintptr(s.fd), uintptr(mask.ToLinux()), 0)
	evnotify(uintptr(s.fd), uintptr(mask.ToLinux()))
}

func (s *sockFile) Bind(uaddr, uaddrlen uintptr) error {
	var saddr *linux.SockAddrInet
	if uaddrlen < unsafe.Sizeof(*saddr) {
		return errors.New("bad bind address")
	}
	saddr = (*linux.SockAddrInet)(unsafe.Pointer(uaddr))
	ip := net.IPv4(saddr.Addr[0], saddr.Addr[1], saddr.Addr[2], saddr.Addr[3])
	addr := tcpip.FullAddress{
		// NIC:  defaultNIC,
		Addr: tcpip.Address(ip),
		Port: ntohs(saddr.Port),
	}
	err := s.ep.Bind(addr)
	if err != nil {
		log.Infof("[socket] bind error:%s", err)
		return e(err)
	}
	return nil
}

func (s *sockFile) Connect(uaddr, uaddrlen uintptr) error {
	var saddr *linux.SockAddrInet
	if uaddrlen < unsafe.Sizeof(*saddr) {
		return syscall.EINVAL
	}
	saddr = (*linux.SockAddrInet)(unsafe.Pointer(uaddr))
	addr := tcpip.FullAddress{
		Addr: tcpip.Address(saddr.Addr[:]),
		Port: ntohs(saddr.Port),
	}
	err := s.ep.Connect(addr)
	if _, ok := err.(*tcpip.ErrConnectStarted); ok {
		return syscall.EINPROGRESS
	}
	if err != nil {
		log.Infof("[socket] connect error:%s", err)
		return e(err)
	}
	return nil
}

func (s *sockFile) Listen(n uintptr) error {
	err := s.ep.Listen(int(n))
	if err != nil {
		log.Infof("[socket] listen error:%s", err)
	}
	return e(err)
}

func (s *sockFile) Accept4(uaddr, uaddrlen, flag uintptr) (int, error) {
	var saddr *linux.SockAddrInet
	if uaddrlen < unsafe.Sizeof(*saddr) {
		return 0, syscall.EINVAL
	}
	saddr = (*linux.SockAddrInet)(unsafe.Pointer(uaddr))
	newep, wq, err := s.ep.Accept(nil)
	switch err.(type) {
	case nil:
	case *tcpip.ErrWouldBlock:
		return 0, syscall.EAGAIN
	default:
		log.Infof("[socket] accept error:%s", err)
		return 0, e(err)
	}

	newaddr, err := newep.GetRemoteAddress()
	if err != nil {
		log.Infof("[socket] accept getRemoteAddress error:%s", err)
		return 0, e(err)
	}
	saddr.Family = syscall.AF_INET
	saddr.Port = htons(newaddr.Port)
	copy(saddr.Addr[:], newaddr.Addr)
	sfile := allocSockFile(newep, wq)
	return sfile.fd, nil
}

func (s *sockFile) Setsockopt(level, opt, vptr, vlen uintptr) error {
	switch level {
	case syscall.SOL_SOCKET, syscall.IPPROTO_TCP:
	default:
		log.Infof("[socket] setsockopt:unsupport socket opt level:%d", level)
		return syscall.EINVAL
	}

	if vlen != 4 {
		log.Infof("[socket] setsockopt:bad opt value length:%d", vlen)
		return syscall.EINVAL
	}

	var terr tcpip.Error
	value := *(*uint32)(unsafe.Pointer(vptr))
	sockopt := s.ep.SocketOptions()

	switch opt {
	case syscall.SO_REUSEADDR:
		sockopt.SetReuseAddress(value != 0)
	case syscall.SO_BROADCAST:
		sockopt.SetBroadcast(value != 0)
	case syscall.TCP_NODELAY:
		sockopt.SetDelayOption(value != 0)
	case syscall.SO_KEEPALIVE:
		sockopt.SetKeepAlive(value != 0)
	case syscall.TCP_KEEPINTVL:
		v := tcpip.KeepaliveIntervalOption(time.Duration(value) * time.Second)
		terr = s.ep.SetSockOpt(&v)
	case syscall.TCP_KEEPIDLE:
		v := tcpip.KeepaliveIdleOption(time.Duration(value) * time.Second)
		terr = s.ep.SetSockOpt(&v)
	default:
		log.Infof("[socket] setsockopt:unknow socket option:%d", opt)
		return syscall.EINVAL
	}

	if terr != nil {
		return e(terr)
	}
	return nil
}

func (s *sockFile) Getsockopt(level, opt, vptr, vlenptr uintptr) error {
	if level != syscall.SOL_SOCKET {
		log.Infof("[socket] getsockopt:unsupport socket opt level:%d", level)
		return syscall.EINVAL
	}
	vlen := (*int)(unsafe.Pointer(vlenptr))
	if *vlen != 4 {
		log.Infof("[socket] getsockopt:bad opt value length:%d", vlen)
		return syscall.EINVAL
	}
	value := (*uint32)(unsafe.Pointer(vptr))

	switch opt {
	case syscall.SO_ERROR:
		terr := s.ep.SocketOptions().GetLastError()
		switch terr.(type) {
		case nil:
		case *tcpip.ErrConnectionRefused:
			*value = uint32(syscall.ECONNREFUSED)
		case *tcpip.ErrNoRoute:
			*value = uint32(syscall.EHOSTUNREACH)
		default:
			log.Infof("[socket] getsockopt:unknow socket error:%s", terr)
			return e(terr)
		}
	default:
		log.Infof("[socket] getsockopt:unknow socket option:%d", opt)
		return syscall.EINVAL
	}
	return nil
}

func (s *sockFile) Getpeername(uaddr, uaddrlen uintptr) error {
	saddr := (*linux.SockAddrInet)(unsafe.Pointer(uaddr))
	addr, err := s.ep.GetRemoteAddress()
	if err != nil {
		log.Infof("[socket] getpeername error:%s", err)
		return e(err)
	}
	saddr.Family = syscall.AF_INET
	copy(saddr.Addr[:], addr.Addr)
	saddr.Port = ntohs(addr.Port)
	return nil
}

func (s *sockFile) Getsockname(uaddr, uaddrlen uintptr) error {
	saddr := (*linux.SockAddrInet)(unsafe.Pointer(uaddr))
	addr, err := s.ep.GetLocalAddress()
	if err != nil {
		log.Infof("[socket] getsockname error:%s", err)
		return e(err)
	}
	saddr.Family = syscall.AF_INET
	copy(saddr.Addr[:], addr.Addr)
	saddr.Port = htons(addr.Port)
	return nil
}
