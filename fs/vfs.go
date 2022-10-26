package fs

import (
	"io"
	"math/rand"
	"os"
	"sync"
	"syscall"
	"unsafe"

	"github.com/jspc/eggos/console"
	"github.com/jspc/eggos/fs/mount"
	"github.com/jspc/eggos/kernel/isyscall"
	"github.com/jspc/eggos/kernel/sys"

	"github.com/spf13/afero"
)

var (
	inodeLock sync.Mutex
	inodes    []*Inode

	Root = mount.NewMountableFs(afero.NewMemMapFs())
)

type Ioctler interface {
	Ioctl(op, arg uintptr) error
}

type Inode struct {
	File  io.ReadWriteCloser
	Fd    int
	inuse bool
}

func (i *Inode) Release() {
	inodeLock.Lock()
	defer inodeLock.Unlock()

	i.inuse = false
	i.File = nil
	i.Fd = -1
}

func AllocInode() (int, *Inode) {
	inodeLock.Lock()
	defer inodeLock.Unlock()

	var fd int
	var ni *Inode
	for i := range inodes {
		entry := inodes[i]
		if !entry.inuse {
			fd = i
			ni = entry
			break
		}
	}
	if fd == 0 {
		ni = new(Inode)
		fd = len(inodes)
		inodes = append(inodes, ni)
	}
	ni.inuse = true
	ni.Fd = fd
	return fd, ni
}

func AllocFileNode(r io.ReadWriteCloser) (int, *Inode) {
	fd, ni := AllocInode()
	ni.File = r
	return fd, ni
}

func GetInode(fd int) (*Inode, error) {
	inodeLock.Lock()
	defer inodeLock.Unlock()

	if fd >= len(inodes) || fd < 0 {
		return nil, syscall.EBADF
	}
	ni := inodes[fd]
	if !ni.inuse {
		return nil, syscall.EBADF
	}
	return ni, nil
}

func fscall(fn int) isyscall.Handler {
	return func(c *isyscall.Request) {
		var err error
		if fn == syscall.SYS_OPENAT {
			var fd int
			fd, err = sysOpen(c.Arg(0), c.Arg(1), c.Arg(2), c.Arg(3))
			if err != nil {
				c.SetRet(isyscall.Error(err))
			} else {
				c.SetRet(uintptr(fd))
			}

			return
		}

		var ni *Inode

		ni, err = GetInode(int(c.Arg(0)))
		if err != nil {
			c.SetRet(isyscall.Error(err))

			return
		}

		switch fn {
		case syscall.SYS_READ:
			var n int
			n, err = sysRead(ni, c.Arg(1), c.Arg(2))
			c.SetRet(uintptr(n))
		case syscall.SYS_WRITE:
			var n int
			n, err = sysWrite(ni, c.Arg(1), c.Arg(2))
			c.SetRet(uintptr(n))
		case syscall.SYS_CLOSE:
			err = sysClose(ni)
		case syscall.SYS_FSTAT:
			err = sysStat(ni, c.Arg(1))
		case syscall.SYS_IOCTL:
			err = sysIoctl(ni, c.Arg(1), c.Arg(2))
		}

		if err != nil {
			c.SetError(err)
		}

	}
}

func sysOpen(dirfd, name, flags, perm uintptr) (int, error) {
	path := cstring(name)
	f, err := Root.OpenFile(path, int(flags), os.FileMode(perm))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, syscall.ENOENT
		}
		return 0, err
	}
	fd, ni := AllocInode()
	ni.File = f
	return fd, nil
}

func sysClose(ni *Inode) error {
	err := ni.File.Close()
	ni.Release()
	return err
}

func sysRead(ni *Inode, p, n uintptr) (int, error) {
	buf := sys.UnsafeBuffer(p, int(n))
	ret, err := ni.File.Read(buf)

	switch {
	case ret != 0:
		return ret, nil
	case err == io.EOF:
		return 0, nil
	case err != nil:
		return 0, err
	default:
		return ret, err
	}
}

func sysWrite(ni *Inode, p, n uintptr) (int, error) {
	buf := sys.UnsafeBuffer(p, int(n))
	_n, err := ni.File.Write(buf)
	if _n != 0 {
		return _n, nil
	}
	return 0, err
}

func sysStat(ni *Inode, statptr uintptr) error {
	file, ok := ni.File.(afero.File)
	if !ok {
		return syscall.EINVAL
	}
	stat := (*syscall.Stat_t)(unsafe.Pointer(statptr))
	info, err := file.Stat()
	if err != nil {
		return err
	}
	stat.Mode = uint32(info.Mode())
	stat.Mtim.Sec = int64(info.ModTime().Unix())
	stat.Size = info.Size()

	return nil
}

func sysIoctl(ni *Inode, op, arg uintptr) error {
	ctl, ok := ni.File.(Ioctler)
	if !ok {
		return syscall.EINVAL
	}
	return ctl.Ioctl(op, arg)
}

func sysFcntl(call *isyscall.Request) {
	call.SetRet(0)
}

// func Uname(buf *Utsname)
func sysUname(c *isyscall.Request) {
	unsafebuf := func(b *[65]int8) []byte {
		return (*[65]byte)(unsafe.Pointer(b))[:]
	}
	buf := (*syscall.Utsname)(unsafe.Pointer(c.Arg(0)))
	copy(unsafebuf(&buf.Machine), "x86_32")
	copy(unsafebuf(&buf.Domainname), "jspc.com")
	copy(unsafebuf(&buf.Nodename), "jspc.local")
	copy(unsafebuf(&buf.Release), "0")
	copy(unsafebuf(&buf.Sysname), "eggos")
	copy(unsafebuf(&buf.Version), "0")
	c.SetRet(0)

}

// func fstatat(dirfd int, path string, stat *Stat_t, flags int)
func sysFstatat64(c *isyscall.Request) {
	name := cstring(c.Arg(1))
	stat := (*syscall.Stat_t)(unsafe.Pointer(c.Arg(2)))
	info, err := Root.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			c.SetRet(isyscall.Errno(syscall.ENOENT))
		} else {
			c.SetRet(isyscall.Error(err))
		}

		return
	}
	stat.Mode = uint32(info.Mode())
	stat.Mtim.Sec = int64(info.ModTime().Unix())
	stat.Size = info.Size()
	c.SetRet(0)

}

func sysLseek(c *isyscall.Request) {
	fd := c.Arg(0)
	offset := c.Arg(1)
	whence := c.Arg(2)

	_ = offset
	_ = whence

	_, err := GetInode(int(fd))
	if err != nil {
		c.SetRet(isyscall.Error(err))
		return
	}
	c.SetRet(0)
}

func sysRandom(call *isyscall.Request) {
	p, n := call.Arg(0), call.Arg(1)
	buf := sys.UnsafeBuffer(p, int(n))
	rand.Read(buf)
	call.SetRet(n)
}

func cstring(ptr uintptr) string {
	var n int
	for p := ptr; *(*byte)(unsafe.Pointer(p)) != 0; p++ {
		n++
	}
	return string(sys.UnsafeBuffer(ptr, n))
}

type fileHelper struct {
	r io.Reader
	w io.Writer
	c io.Closer
}

func NewFile(r io.Reader, w io.Writer, c io.Closer) io.ReadWriteCloser {
	return &fileHelper{
		r: r,
		w: w,
		c: c,
	}
}

func (r *fileHelper) Read(p []byte) (int, error) {
	if r.r != nil {
		return r.r.Read(p)
	}
	return 0, syscall.EINVAL
}
func (r *fileHelper) Write(p []byte) (int, error) {
	if r.w != nil {
		return r.w.Write(p)
	}
	return 0, syscall.EROFS
}

func (r *fileHelper) Ioctl(op, arg uintptr) error {
	var x interface{}
	if r.r != nil {
		x = r.r
	} else {
		x = r.w
	}
	ctl, ok := x.(Ioctler)
	if !ok {
		return syscall.EBADF
	}
	return ctl.Ioctl(op, arg)
}

func (r *fileHelper) Close() error {
	if r.c != nil {
		return r.c.Close()
	}
	return syscall.EINVAL
}

func Mount(target string, fs afero.Fs) error {
	return Root.Mount(target, fs)
}

func vfsInit() {
	c := console.Console()
	// stdin
	AllocFileNode(NewFile(c, nil, nil))
	// stdout
	AllocFileNode(NewFile(nil, c, nil))
	// stderr
	AllocFileNode(NewFile(nil, c, nil))
	// epoll fd
	AllocFileNode(NewFile(nil, nil, nil))
	// pipe read fd
	AllocFileNode(NewFile(nil, nil, nil))
	// pipe write fd
	AllocFileNode(NewFile(nil, nil, nil))

	etcInit()
}

func sysInit() {
	isyscall.Register(syscall.SYS_OPENAT, fscall(syscall.SYS_OPENAT))
	isyscall.Register(syscall.SYS_WRITE, fscall(syscall.SYS_WRITE))
	isyscall.Register(syscall.SYS_READ, fscall(syscall.SYS_READ))
	isyscall.Register(syscall.SYS_CLOSE, fscall(syscall.SYS_CLOSE))
	isyscall.Register(syscall.SYS_FSTAT, fscall(syscall.SYS_FSTAT))
	isyscall.Register(syscall.SYS_IOCTL, fscall(syscall.SYS_IOCTL))
	isyscall.Register(syscall.SYS_FCNTL, sysFcntl)
	isyscall.Register(syscall.SYS_NEWFSTATAT, sysFstatat64)
	isyscall.Register(syscall.SYS_LSEEK, sysLseek)
	isyscall.Register(syscall.SYS_UNAME, sysUname)
	isyscall.Register(355, sysRandom)
}

func Init() {
	vfsInit()
	sysInit()
}
