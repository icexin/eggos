package shiny

import (
	"image"

	"golang.org/x/exp/shiny/screen"
)

var (
	_ screen.Buffer = (*bufferImpl)(nil)
	_ screen.Window = (*windowImpl)(nil)

	defaultScreen screenImpl
)

type screenImpl struct {
}

// NewBuffer returns a new Buffer for this screen.
func (s *screenImpl) NewBuffer(size image.Point) (screen.Buffer, error) {
	m := image.NewRGBA(image.Rectangle{Max: size})
	return &bufferImpl{
		buf:  m.Pix,
		rgba: *m,
		size: size,
	}, nil
}

// NewTexture returns a new Texture for this screen.
func (s *screenImpl) NewTexture(size image.Point) (screen.Texture, error) {
	panic("not implemented") // TODO: Implement
}

// NewWindow returns a new Window for this screen.
//
// A nil opts is valid and means to use the default option values.
func (s *screenImpl) NewWindow(opts *screen.NewWindowOptions) (screen.Window, error) {
	return newWindow(), nil
}
