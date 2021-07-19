package inet

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"syscall"
	"time"
	"unsafe"

	"github.com/icexin/eggos/fs"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/waiter"
)

type _sockaddr struct {
	family uint16
	port   uint16
	ip     [4]byte
}

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
	evnotify(uintptr(s.fd), uintptr(mask.ToLinux()))
}

func (s *sockFile) Bind(uaddr, uaddrlen uintptr) error {
	if uaddrlen < unsafe.Sizeof(_sockaddr{}) {
		return errors.New("bad bind address")
	}
	saddr := (*_sockaddr)(unsafe.Pointer(uaddr))
	ip := net.IPv4(saddr.ip[0], saddr.ip[1], saddr.ip[2], saddr.ip[3])
	addr := tcpip.FullAddress{
		NIC:  defaultNIC,
		Addr: tcpip.Address(ip),
		Port: ntohs(saddr.port),
	}
	err := s.ep.Bind(addr)
	if err != nil {
		return e(err)
	}
	return nil
}

func (s *sockFile) Connect(uaddr, uaddrlen uintptr) error {
	if uaddrlen < unsafe.Sizeof(_sockaddr{}) {
		return syscall.EINVAL
	}
	saddr := (*_sockaddr)(unsafe.Pointer(uaddr))
	addr := tcpip.FullAddress{
		Addr: tcpip.Address(saddr.ip[:]),
		Port: ntohs(saddr.port),
	}
	err := s.ep.Connect(addr)
	if _, ok := err.(*tcpip.ErrConnectStarted); ok {
		return syscall.EINPROGRESS
	}
	if err != nil {
		return e(err)
	}
	return nil
}

func (s *sockFile) Listen(n uintptr) error {
	err := s.ep.Listen(int(n))
	return e(err)
}

func (s *sockFile) Accept4(uaddr, uaddrlen, flag uintptr) (int, error) {
	if uaddrlen < unsafe.Sizeof(_sockaddr{}) {
		return 0, syscall.EINVAL
	}
	saddr := (*_sockaddr)(unsafe.Pointer(uaddr))
	newep, wq, err := s.ep.Accept(nil)
	switch err.(type) {
	case nil:
	case *tcpip.ErrWouldBlock:
		return 0, syscall.EAGAIN
	default:
		return 0, e(err)
	}

	newaddr, err := newep.GetRemoteAddress()
	if err != nil {
		return 0, e(err)
	}
	saddr.family = syscall.AF_INET
	saddr.port = htons(newaddr.Port)
	copy(saddr.ip[:], newaddr.Addr)
	sfile := allocSockFile(newep, wq)
	return sfile.fd, nil
}

func (s *sockFile) Setsockopt(level, opt, vptr, vlen uintptr) error {
	switch level {
	case syscall.SOL_SOCKET, syscall.IPPROTO_TCP:
	default:
		return fmt.Errorf("setsockopt:unsupport socket opt level:%d", level)
	}

	if vlen != 4 {
		return errors.New("setsockopt:bad opt value length")
	}

	var terr tcpip.Error
	value := *(*int32)(unsafe.Pointer(vptr))
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
		return fmt.Errorf("setsockopt:unsupport socket opt:%d", level)
	}

	if terr != nil {
		return e(terr)
	}
	return nil
}

func (s *sockFile) Getsockopt(level, opt, vptr, vlenptr uintptr) error {
	if level != syscall.SOL_SOCKET {
		return fmt.Errorf("unsupport opt level:%d", level)
	}
	vlen := (*int)(unsafe.Pointer(vlenptr))
	if *vlen != 4 {
		return errors.New("bad opt value length")
	}

	switch opt {
	case syscall.SO_ERROR:
		terr := s.ep.SocketOptions().GetLastError()
		switch terr.(type) {
		case *tcpip.ErrConnectionRefused:
			return syscall.ECONNREFUSED
		default:
			return e(terr)
		}
	default:
		return fmt.Errorf("unknown socket option:%d", opt)
	}
	return nil
}

func (s *sockFile) Getpeername(uaddr, uaddrlen uintptr) error {
	saddr := (*_sockaddr)(unsafe.Pointer(uaddr))
	addr, err := s.ep.GetRemoteAddress()
	if err != nil {
		return e(err)
	}
	saddr.family = syscall.AF_INET
	copy(saddr.ip[:], addr.Addr)
	return nil
}

func (s *sockFile) Getsockname(uaddr, uaddrlen uintptr) error {
	saddr := (*_sockaddr)(unsafe.Pointer(uaddr))
	addr, err := s.ep.GetLocalAddress()
	if err != nil {
		return e(err)
	}
	saddr.family = syscall.AF_INET
	copy(saddr.ip[:], addr.Addr)
	return nil
}
