package shiny

import (
	"github.com/icexin/eggos/console"
	imouse "github.com/icexin/eggos/drivers/ps2/mouse"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/sys/unix"
)

func (w *windowImpl) listenKeyboardEvent() {
	termios := unix.Termios{}
	termios.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	unix.IoctlSetTermios(0, unix.TCSETS, &termios)

	buf := make([]byte, 16)
	for {
		n, _ := console.Console().Read(buf)
		content := buf[:n]
		for _, ch := range content {
			var code key.Code
			var char rune
			if ch == '\b' {
				code = key.CodeDeleteBackspace
			} else {
				char = rune(ch)
			}
			event := key.Event{
				Code:      code,
				Rune:      char,
				Direction: key.DirPress,
			}
			w.eventch <- event
		}
	}
}

func (w *windowImpl) listenMouseEvent() {
	for e := range imouse.EventQueue() {
		w.updateCursor()
		w.sendMouseEvent(e)
		w.cursor = e
	}
}

func (w *windowImpl) sendMouseEvent(e imouse.Packet) {
	var btn mouse.Button
	var dir mouse.Direction

	if e.Left {
		if !w.cursor.Left {
			btn = mouse.ButtonLeft
			dir = mouse.DirPress
		}
	} else {
		if w.cursor.Left {
			btn = mouse.ButtonLeft
			dir = mouse.DirRelease
		}
	}
	if e.Right {
		if !w.cursor.Right {
			btn = mouse.ButtonRight
			dir = mouse.DirPress
		}
	} else {
		if w.cursor.Right {
			btn = mouse.ButtonRight
			dir = mouse.DirRelease
		}
	}

	event := mouse.Event{
		X:         float32(e.X),
		Y:         float32(e.Y),
		Button:    btn,
		Direction: dir,
	}
	w.eventch <- event
}
