package shiny

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/fogleman/gg"
	imouse "github.com/icexin/eggos/ps2/mouse"
	"github.com/icexin/eggos/uart"
	"github.com/icexin/eggos/vbe"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/math/f64"
	"golang.org/x/mobile/event/mouse"
)

type windowImpl struct {
	view   *vbe.View
	img    image.Image
	cursor imouse.Packet
}

func newWindow() *windowImpl {
	ctx := gg.NewContext(100, 100)
	ctx.SetRGBA(1, 1, 1, 1)
	ctx.Clear()
	ctx.SetRGB(0, 0, 0)
	ctx.DrawCircle(5, 5, 5)
	ctx.Stroke()
	img := ctx.Image()
	return &windowImpl{
		view: vbe.DefaultView,
		img:  img,
	}
}

// Release closes the window.
//
// The behavior of the Window after Release, whether calling its methods or
// passing it as an argument, is undefined.
func (w *windowImpl) Release() {
	panic("not implemented") // TODO: Implement
}

// Send adds an event to the end of the deque. They are returned by
// NextEvent in FIFO order.
func (w *windowImpl) Send(event interface{}) {
}

// SendFirst adds an event to the start of the deque. They are returned by
// NextEvent in LIFO order, and have priority over events sent via Send.
func (w *windowImpl) SendFirst(event interface{}) {
}

// NextEvent returns the next event in the deque. It blocks until such an
// event has been sent.
//
// Typical event types include:
//	- lifecycle.Event
//	- size.Event
//	- paint.Event
//	- key.Event
//	- mouse.Event
//	- touch.Event
// from the golang.org/x/mobile/event/... packages. Other packages may send
// events, of those types above or of other types, via Send or SendFirst.
func (w *windowImpl) NextEvent() interface{} {
	p := imouse.EventQueue()
	e := <-p
	var btn mouse.Button
	if e.Left {
		btn = mouse.ButtonLeft
	}
	if e.Right {
		btn = mouse.ButtonRight
	}
	var dir mouse.Direction
	if btn != mouse.ButtonNone {
		dir = mouse.DirPress
	}
	pt := image.Pt(e.X-5, e.Y-5)
	draw.Draw(w.view.Canvas(), w.img.Bounds(), w.img, pt, draw.Src)
	w.view.Commit()
	return mouse.Event{
		X:         float32(e.X),
		Y:         float32(e.Y),
		Direction: dir,
	}
}

type writer struct {
}

func (w writer) Write(b []byte) (int, error) {
	return uart.Write(b)
}

// Upload uploads the sub-Buffer defined by src and sr to the destination
// (the method receiver), such that sr.Min in src-space aligns with dp in
// dst-space. The destination's contents are overwritten; the draw operator
// is implicitly draw.Src.
//
// It is valid to upload a Buffer while another upload of the same Buffer
// is in progress, but a Buffer's image.RGBA pixel contents should not be
// accessed while it is uploading. A Buffer is re-usable, in that its pixel
// contents can be further modified, once all outstanding calls to Upload
// have returned.
//
// TODO: make it optional that a Buffer's contents is preserved after
// Upload? Undoing a swizzle is a non-trivial amount of work, and can be
// redundant if the next paint cycle starts by clearing the buffer.
//
// When uploading to a Window, there will not be any visible effect until
// Publish is called.
func (w *windowImpl) Upload(dp image.Point, src screen.Buffer, sr image.Rectangle) {
	fmt.Fprintf(writer{}, "upload\n")
	draw.Draw(w.view.Canvas(), sr, src.RGBA(), dp, draw.Src)
}

// Fill fills that part of the destination (the method receiver) defined by
// dr with the given color.
//
// When filling a Window, there will not be any visible effect until
// Publish is called.
func (w *windowImpl) Fill(dr image.Rectangle, src color.Color, op draw.Op) {
	panic("not implemented") // TODO: Implement
}

// Draw draws the sub-Texture defined by src and sr to the destination (the
// method receiver). src2dst defines how to transform src coordinates to
// dst coordinates. For example, if src2dst is the matrix
//
// m00 m01 m02
// m10 m11 m12
//
// then the src-space point (sx, sy) maps to the dst-space point
// (m00*sx + m01*sy + m02, m10*sx + m11*sy + m12).
func (w *windowImpl) Draw(src2dst f64.Aff3, src screen.Texture, sr image.Rectangle, op draw.Op, opts *screen.DrawOptions) {
	panic("not implemented") // TODO: Implement
}

// DrawUniform is like Draw except that the src is a uniform color instead
// of a Texture.
func (w *windowImpl) DrawUniform(src2dst f64.Aff3, src color.Color, sr image.Rectangle, op draw.Op, opts *screen.DrawOptions) {
	panic("not implemented") // TODO: Implement
}

// Copy copies the sub-Texture defined by src and sr to the destination
// (the method receiver), such that sr.Min in src-space aligns with dp in
// dst-space.
func (w *windowImpl) Copy(dp image.Point, src screen.Texture, sr image.Rectangle, op draw.Op, opts *screen.DrawOptions) {
	panic("not implemented") // TODO: Implement
}

// Scale scales the sub-Texture defined by src and sr to the destination
// (the method receiver), such that sr in src-space is mapped to dr in
// dst-space.
func (w *windowImpl) Scale(dr image.Rectangle, src screen.Texture, sr image.Rectangle, op draw.Op, opts *screen.DrawOptions) {
	panic("not implemented") // TODO: Implement
}

// Publish flushes any pending Upload and Draw calls to the window, and
// swaps the back buffer to the front.
func (w *windowImpl) Publish() screen.PublishResult {
	fmt.Fprintf(writer{}, "publish\n")
	// p := image.Pt(w.cursor.X-5, w.cursor.Y-5)
	// draw.Draw(w.view.Canvas(), w.img.Bounds(), w.img, p, draw.Src)
	w.view.Commit()
	return screen.PublishResult{}
}
