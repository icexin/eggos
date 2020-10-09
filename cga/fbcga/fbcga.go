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
	del = 0x7f
	bs  = '\b'

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
	cursor int
	buffer [TERM_WIDTH * TERM_HEIGHT]byte
}

func (f *fbbackend) xy(pos int) (int, int) {
	x, y := pos%TERM_WIDTH*face.Advance, (pos/TERM_WIDTH+1)*face.Height
	if face.Left < 0 {
		x += -face.Left
	}
	return x, y
}

func (f *fbbackend) glyphPos(pos int) fixed.Point26_6 {
	x, y := f.xy(pos)
	return fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}
}

func (f *fbbackend) drawByte(pos int, c byte, overide bool) {
	if c == 0 {
		return
	}
	fpos := f.glyphPos(pos)
	rect, _, _, _, _ := face.Glyph(fpos, rune(c))
	if overide {
		// clear old font
		draw.Draw(drawer.Dst, rect, image.NewUniform(color.Black), image.Point{}, draw.Src)
	}
	drawer.Dot = fpos
	drawer.DrawBytes([]byte{c})
	view.CommitRect(rect)
}

func (f *fbbackend) setByte(pos int, c byte) {
	f.drawByte(pos, c, true)
	f.buffer[pos] = c
}

func (f *fbbackend) scrollup(pos int) int {
	copy(f.buffer[:], f.buffer[TERM_WIDTH:TERM_HEIGHT*TERM_WIDTH])
	pos -= TERM_WIDTH
	s := f.buffer[pos : TERM_HEIGHT*TERM_WIDTH]
	for i := range s {
		s[i] = 0
	}
	f.refresh()
	return pos
}

func (f *fbbackend) refresh() {
	rect := image.Rect(0, 0, TERM_WIDTH*face.Advance+face.Left, TERM_HEIGHT*face.Height+face.Descent)
	draw.Draw(drawer.Dst, rect, image.NewUniform(color.Black), image.Point{}, draw.Src)

	for i := 0; i < TERM_HEIGHT; i++ {
		y := (i + 1) * face.Height
		drawer.Dot = fixed.Point26_6{X: fixed.I(0), Y: fixed.I(y)}
		drawer.DrawBytes(f.buffer[i*TERM_WIDTH : (i+1)*TERM_WIDTH])
	}
	view.CommitRect(rect)
}

func (f *fbbackend) updateCursor(n int) {
	old := f.cursor
	ch := f.buffer[old]
	f.drawByte(old, ch, true)

	f.drawByte(n, '_', false)
	f.cursor = n
}

func (f *fbbackend) SetPos(n int) {
	f.pos = n
	f.updateCursor(n)
}

func (f *fbbackend) GetPos() int {
	return f.pos
}

func (f *fbbackend) WritePos(n int, ch byte) {
	f.setByte(n, ch)
}

func (f *fbbackend) WriteByte(c byte) {
	pos := f.GetPos()
	switch c {
	case '\n', '\r':
		pos += TERM_WIDTH - pos%TERM_WIDTH
	case bs, del:
		if pos > 0 {
			f.setByte(pos, ' ')
			pos--
			f.setByte(pos, ' ')
		}
	default:
		f.setByte(pos, c)
		pos++
	}

	// Scroll up
	if pos/TERM_WIDTH >= TERM_HEIGHT {
		pos = f.scrollup(pos)
	}
	f.setByte(pos, ' ')
	f.SetPos(pos)
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
