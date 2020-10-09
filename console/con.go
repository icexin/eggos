package console

import (
	"io"
	"sync"
	"syscall"
	"unsafe"

	"github.com/icexin/eggos/cga"
	"github.com/icexin/eggos/kbd"
	"github.com/icexin/eggos/uart"
)

const (
	RAW_BUFLEN = 128
	CON_BUFLEN = 128
	BACKSPACE  = 0x100
)

type console struct {
	rawch chan byte

	buf     [CON_BUFLEN]byte
	r, w, e uint

	tios syscall.Termios

	mutex  sync.Mutex
	notify *sync.Cond
}

var (
	con *console
)

func newConsole() *console {
	c := &console{
		rawch: make(chan byte, RAW_BUFLEN),
		tios: syscall.Termios{
			Lflag: syscall.ICANON | syscall.ECHO,
		},
	}
	c.notify = sync.NewCond(&c.mutex)
	return c
}

func ctrl(c byte) byte {
	return c - '@'
}

//go:nosplit
func (c *console) intr(ch byte) {
	c.handleInput(ch)
}

func (c *console) rawmode() bool {
	return c.tios.Lflag&syscall.ICANON == 0
}

func (c *console) handleRaw(ch byte) {
	if c.e-c.r >= CON_BUFLEN {
		return
	}
	idx := c.e % CON_BUFLEN
	c.e++
	c.buf[idx] = byte(ch)
	c.w = c.e
	c.notify.Broadcast()
}

func (c *console) handleInput(ch byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.rawmode() {
		c.handleRaw(ch)
		return
	}

	switch ch {
	case 0x7f, ctrl('H'):
		if c.e > c.w {
			c.e--
			c.putc(0x7f)
		}
		return
	}

	if c.e-c.r >= CON_BUFLEN {
		return
	}
	if ch == '\r' {
		ch = '\n'
	}
	idx := c.e % CON_BUFLEN
	c.e++
	c.buf[idx] = byte(ch)
	c.putc(ch)
	if ch == '\n' || c.e == c.r+CON_BUFLEN {
		c.w = c.e
		c.notify.Broadcast()
	}
}

func (c *console) loop() {
	for ch := range c.rawch {
		c.handleInput(ch)
	}
}

func (c *console) putc(ch byte) {
	uart.WriteByte(ch)
	cga.WriteByte(ch)
}

func (c *console) read(p []byte) int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	i := 0
	for i < len(p) {
		for c.r == c.w {
			c.notify.Wait()
		}
		idx := c.r
		c.r++
		ch := c.buf[idx%CON_BUFLEN]
		p[i] = byte(ch)
		i++
		if ch == '\n' || c.rawmode() {
			break
		}
	}
	return i
}

func (c *console) Read(p []byte) (int, error) {
	return c.read(p), nil
}

func (c *console) Write(p []byte) (int, error) {
	for _, ch := range p {
		c.putc(ch)
	}
	return len(p), nil
}

func (c *console) Ioctl(op, arg uintptr) error {
	switch op {
	case syscall.TIOCGWINSZ:
		w := (*winSize)(unsafe.Pointer(arg))
		w.row = 25
		w.col = 80
		return nil
	case syscall.TCGETS:
		tios := (*syscall.Termios)(unsafe.Pointer(arg))
		*tios = c.tios
		return nil
	case syscall.TCSETS:
		tios := (*syscall.Termios)(unsafe.Pointer(arg))
		c.tios = *tios
		return nil

	default:
		return syscall.EINVAL
	}
}

type winSize struct {
	row, col       uint16
	xpixel, ypixel uint16
}

func Console() io.ReadWriter {
	return con
}

func Init() {
	con = newConsole()
	// go con.loop()
	uart.OnInput(con.intr)
	kbd.OnInput(con.intr)
}
