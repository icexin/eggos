package fbcga

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/icexin/eggos/vbe"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

const (
	BACKSPACE = 0x7f

	TERM_WIDTH  = 80
	TERM_HEIGHT = 25
)

var (
	view   *vbe.View
	face   *basicfont.Face
	drawer *font.Drawer

	Backend = fbbackend{}
)

type fbbackend struct {
	pos    int
	buffer [TERM_WIDTH * TERM_HEIGHT]byte
}

func (f *fbbackend) xy(pos int) (int, int) {
	return pos%TERM_WIDTH*face.Advance + face.Left, (pos/TERM_WIDTH + 1) * face.Height
}

func (f *fbbackend) glyphPos(pos int) fixed.Point26_6 {
	x, y := f.xy(pos)
	return fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}
}

func (f *fbbackend) clear(pos int) image.Rectangle {
	ch := f.buffer[pos]
	if ch == 0 {
		return image.ZR
	}
	f.buffer[pos] = 0
	rect, _, _, _, _ := face.Glyph(f.glyphPos(pos), rune(ch))
	draw.Draw(drawer.Dst, rect, image.NewUniform(color.Black), image.Point{}, draw.Src)
	return rect
}

func (f *fbbackend) drawByte(pos int, c byte) image.Rectangle {
	rect := f.clear(pos)
	if c != 0 {
		fpos := f.glyphPos(pos)
		rect, _, _, _, _ = face.Glyph(fpos, rune(c))
		drawer.Dot = fpos
		drawer.DrawBytes([]byte{c})
	}
	f.buffer[pos] = c
	return rect
}

func (f *fbbackend) scrollup() image.Rectangle {
	copy(f.buffer[:], f.buffer[TERM_WIDTH:TERM_HEIGHT*TERM_WIDTH])
	f.pos -= TERM_WIDTH
	s := f.buffer[f.pos : TERM_HEIGHT*TERM_WIDTH]
	for i := range s {
		s[i] = 0
	}
	return f.refresh()
}

func (f *fbbackend) refresh() image.Rectangle {
	rect := image.Rect(0, 0, TERM_WIDTH*face.Advance+face.Left, TERM_HEIGHT*face.Height+face.Descent)
	draw.Draw(drawer.Dst, rect, image.NewUniform(color.Black), image.Point{}, draw.Src)

	for i := 0; i < TERM_HEIGHT; i++ {
		y := (i + 1) * face.Height
		drawer.Dot = fixed.Point26_6{X: fixed.I(0), Y: fixed.I(y)}
		drawer.DrawBytes(f.buffer[i*TERM_WIDTH : (i+1)*TERM_WIDTH])
	}
	return rect
}

func (f *fbbackend) SetPos(n int) {
	f.pos = n
}

func (f *fbbackend) GetPos() int {
	return f.pos
}

func (f *fbbackend) WritePos(n int, ch byte) {
	var rect image.Rectangle
	if ch == 0 {
		rect = f.clear(n)
	} else {
		rect = f.drawByte(n, ch)
	}
	view.CommitRect(rect)
}

func (f *fbbackend) WriteByte(c byte) {
	var rect image.Rectangle

	switch c {
	case '\n':
		f.pos += TERM_WIDTH - f.pos%TERM_WIDTH
	case '\b', BACKSPACE:
		if f.pos > 0 {
			f.pos--
			rect = f.clear(f.pos)
		}
	default:
		rect = f.drawByte(f.pos, c)
		f.pos++
	}

	// Scroll up
	if f.pos/TERM_WIDTH >= TERM_HEIGHT {
		rect = f.scrollup()
	}
	if rect != image.ZR {
		// this line does not work, don't know why
		// draw.Draw(vbe.Screen(), rect, buffer, image.ZP, draw.Src)
		view.CommitRect(rect)
	}
}

func Init() {
	if !vbe.IsEnable() {
		return
	}
	face = inconsolata.Regular8x16
	// face = inconsolata.Bold8x16
	view = vbe.DefaultView
	drawer = &font.Drawer{
		Dst:  view.Canvas(),
		Src:  image.NewUniform(color.RGBA{199, 199, 199, 0}),
		Face: face,
	}
}
