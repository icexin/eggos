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
	pos    int
	view   *vbe.View
	face   *basicfont.Face
	drawer *font.Drawer

	Backend = fbbackend{}
)

func clear(pos int) image.Rectangle {
	x, y := pos%TERM_WIDTH*face.Width, (pos/TERM_WIDTH+1)*face.Height
	rect, _, _, _, _ := face.Glyph(fixed.Point26_6{fixed.I(x), fixed.I(y)}, 'a')
	draw.Draw(drawer.Dst, rect, image.NewUniform(color.Black), image.Point{}, draw.Src)
	return rect
}

func drawByte(pos int, c byte) image.Rectangle {
	x, y := pos%TERM_WIDTH*face.Width, (pos/TERM_WIDTH+1)*face.Height
	drawer.Dot = fixed.Point26_6{fixed.I(x), fixed.I(y)}
	rect, _, _, _, _ := face.Glyph(drawer.Dot, rune(c))
	drawer.DrawBytes([]byte{c})
	return rect
}

func scrollup() {
	start := face.Width * face.Height * TERM_WIDTH * 4
	pix := view.Canvas().(*image.RGBA).Pix
	copy(pix, pix[start:])
	clearStart := face.Width * face.Height * pos * 4
	for i := clearStart; i < len(pix); i++ {
		pix[i] = 0
	}
}

type fbbackend struct {
}

func (f *fbbackend) SetPos(n int) {
	pos = n
}

func (f *fbbackend) GetPos() int {
	return pos
}

func (f *fbbackend) WritePos(n int, ch byte) {
	var rect image.Rectangle
	if ch == 0 {
		rect = clear(n)
	} else {
		rect = drawByte(n, ch)
	}
	view.CommitRect(rect)
}

func (f *fbbackend) WriteByte(c byte) {
	var rect image.Rectangle

	switch c {
	case '\n':
		pos += TERM_WIDTH - pos%TERM_WIDTH
	case BACKSPACE:
		if pos > 0 {
			pos--
			rect = clear(pos)
		}
	default:
		rect = drawByte(pos, c)
		pos++
	}

	// Scroll up
	if pos/TERM_WIDTH >= TERM_HEIGHT-1 {
		pos -= TERM_WIDTH
		scrollup()
		rect = view.Canvas().Bounds()
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
	view = vbe.DefaultView
	drawer = &font.Drawer{
		Dst:  view.Canvas(),
		Src:  image.NewUniform(color.White),
		Face: face,
	}
}
