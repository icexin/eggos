// Package vbe provides access to the VESA VBE framebuffer.
package vbe

import (
	"image"
	"image/draw"
)

// View provides a canvas for offscreen rendering which may be commited to the
// framebuffer.
type View struct {
	// underlying offscreen rendering buffer.
	buffer *image.RGBA
}

// NewView returns a new view for offscreen rendering which may be commited to
// the framebuffer.
func NewView() *View {
	return &View{
		buffer: image.NewRGBA(image.Rect(0, 0, int(info.Width), int(info.Height))),
	}
}

// Canvas returns a canvas for rendering into the given view.
func (v *View) Canvas() draw.Image {
	return v.buffer
}

// Clear clears the given view, setting every pixel to black colour.
func (v *View) Clear() {
	for i := range v.buffer.Pix {
		v.buffer.Pix[i] = 0
	}
}

// CommitRect commits the given rectangle of the view to the framebuffer,
// copying the pixels of the view within rect to the framebuffer.
func (v *View) CommitRect(rect image.Rectangle) {
	if fbbuf == nil {
		return
	}
	if v != currentView {
		return
	}

	// let rect in the range of view rect
	rect = v.buffer.Rect.Intersect(rect).Canon()

	bufcopy(buffer, v.buffer.Pix, v.buffer.Stride, rect, func(dst, src []uint8) {
		for i := 0; i < len(dst); i += 4 {
			_ = dst[i+3]
			_ = src[i+3]
			dst[i] = src[i+2]
			dst[i+1] = src[i+1]
			dst[i+2] = src[i]
			dst[i+3] = src[i+3]
		}
	})

	bufcopy(fbbuf, buffer, v.buffer.Stride, rect, func(dst, src []uint8) {
		copy(dst, src)
	})
}

// Commit commits the view to the framebuffer, copying every pixel of the view
// to the framebuffer.
func (v *View) Commit() {
	if fbbuf == nil {
		return
	}
	if v != currentView {
		return
	}

	pix := v.buffer.Pix
	for i, j := 0, 0; i < len(pix); i, j = i+4, j+4 {
		_ = buffer[j+3]
		_ = pix[i+3]
		buffer[j] = pix[i+2]
		buffer[j+1] = pix[i+1]
		buffer[j+2] = pix[i]
		buffer[j+3] = pix[i+3]
	}
	copy(fbbuf, buffer)
}
