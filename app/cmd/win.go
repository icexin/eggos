package cmd

import (
	"fmt"
	"image"
	"image/color"
	"unsafe"

	"github.com/fogleman/fauxgl"
	"github.com/icexin/eggos/app"
	"github.com/icexin/eggos/kbd"
	"github.com/icexin/eggos/ps2/mouse"
	"github.com/icexin/eggos/vbe"
	"github.com/icexin/nk"
	"golang.org/x/image/draw"
)

type vertex struct {
	pos   [2]float32
	uv    [2]float32
	color [4]byte
}

type window struct {
	ctx    *nk.Nk_context
	ctxptr uintptr

	atlas    *nk.Nk_font_atlas
	atlasptr uintptr

	fctx   *fauxgl.Context
	tex    fauxgl.Texture
	shader *glshader
}

func winMain(ctx *app.Context) error {
	w := newWindow()
	for !kbd.Pressed('q') {
		w.newFrame()
		w.logic()
		img := w.drawImage()
		screen := vbe.DefaultView
		draw.Draw(screen.Canvas(), img.Bounds(), img, image.ZP, draw.Src)
		screen.Commit()
	}
	return nil
}

func newWindow() *window {
	fctx := fauxgl.NewContext(600, 480)
	fctx.Cull = fauxgl.CullNone
	fctx.ReadDepth = false
	fctx.WriteDepth = false

	null := nk.Nk_draw_null_texture{}

	var ctx = new(nk.Nk_context)
	var ctxptr = uintptr(unsafe.Pointer(ctx))
	nk.Xnk_init_default(ctxptr, 0)

	atlas := new(nk.Nk_font_atlas)
	var atlasptr = uintptr(unsafe.Pointer(atlas))
	nk.Xnk_font_atlas_init_default(atlasptr)
	wptr, hptr := 0, 0
	w, h := uintptr(unsafe.Pointer(&wptr)), uintptr(unsafe.Pointer(&hptr))
	piximg := nk.Xnk_font_atlas_bake(atlasptr, w, h, nk.NK_FONT_ATLAS_RGBA32)
	pix := make([]byte, wptr*hptr*4)
	copy(pix, (*[1 << 20]byte)(unsafe.Pointer(piximg))[:wptr*hptr*4])
	nk.Xnk_font_atlas_end(atlasptr, nk.Xnk_handle_id(1), uintptr(unsafe.Pointer(&null)))

	img := &image.RGBA{
		Pix:    pix,
		Rect:   image.Rect(0, 0, wptr, hptr),
		Stride: wptr * 4,
	}
	tex := fauxgl.NewImageTexture(img)
	// m := fauxgl.Scale(fauxgl.V(1, -1, 0)).Translate(fauxgl.V(0, 480, 0)).Orthographic(0, 600, 0, 480, -1, 1)
	// m := fauxgl.Orthographic(0, 600, 0, 480, -1, 1)
	m := fauxgl.Orthographic(0, 600, 480, 0, -1, 1)
	shader := &glshader{
		Matrix: m,
	}
	fctx.Shader = shader

	nk.Xnk_style_load_all_cursors(ctxptr, uintptr(unsafe.Pointer(&atlas.Cursors)))
	if atlas.Default_font != 0 {
		default_font := (*nk.Nk_font)(unsafe.Pointer(atlas.Default_font))
		nk.Xnk_style_set_font(ctxptr, uintptr(unsafe.Pointer(&default_font.Handle)))
	}
	return &window{
		ctx:    ctx,
		ctxptr: ctxptr,

		atlas:    atlas,
		atlasptr: atlasptr,

		fctx:   fctx,
		tex:    tex,
		shader: shader,
	}
}

func (w *window) newFrame() {
	ctx := w.ctxptr
	x, y := mouse.Cursor()
	left := int32(0)
	if mouse.LeftClick() {
		left = 1
	}
	right := int32(0)
	if mouse.RightClick() {
		right = 1
	}
	nk.Xnk_input_begin(ctx)
	nk.Xnk_input_motion(ctx, int32(x), int32(y))
	nk.Xnk_input_button(ctx, nk.NK_BUTTON_LEFT, int32(x), int32(y), left)
	nk.Xnk_input_button(ctx, nk.NK_BUTTON_RIGHT, int32(x), int32(y), right)
	nk.Xnk_input_end(ctx)
}

var title = []byte("hello")

func (w *window) logic() {
	ctx := w.ctxptr
	ok := nk.Xnk_begin(ctx, uintptr(unsafe.Pointer(&title[0])), nk.Xnk_rect(50, 50, 220, 220), nk.NK_WINDOW_BORDER|nk.NK_WINDOW_MOVABLE|nk.NK_WINDOW_CLOSABLE|nk.NK_WINDOW_SCALABLE)
	if ok != 0 {
		x, y := mouse.Cursor()
		left, right := mouse.LeftClick(), mouse.RightClick()
		nk.Xnk_layout_row_static(ctx, 30, 80, 1)
		str := []byte(fmt.Sprintf("l:%v r:%v", left, right))
		nk.Xnk_label(ctx, uintptr(unsafe.Pointer(&str[0])), nk.NK_TEXT_LEFT)
		nk.Xnk_layout_row_static(ctx, 30, 80, 1)
		str1 := []byte(fmt.Sprintf("x:%d y:%d", x, y))
		nk.Xnk_button_label(ctx, uintptr(unsafe.Pointer(&str1[0])))
	}
	nk.Xnk_end(ctx)

}
func (w *window) drawImage() image.Image {
	ctx := w.ctxptr
	fctx := w.fctx
	tex := w.tex
	shader := w.shader

	fctx.ClearColorBuffer()
	fctx.ClearDepthBuffer()

	var vertex_layout = [...]nk.Nk_draw_vertex_layout_element{
		{nk.NK_VERTEX_POSITION, nk.NK_FORMAT_FLOAT, nk.Nk_size(unsafe.Offsetof(vertex{}.pos))},
		{nk.NK_VERTEX_TEXCOORD, nk.NK_FORMAT_FLOAT, nk.Nk_size(unsafe.Offsetof(vertex{}.uv))},
		{nk.NK_VERTEX_COLOR, nk.NK_FORMAT_R8G8B8A8, nk.Nk_size(unsafe.Offsetof(vertex{}.color))},
		{nk.NK_VERTEX_ATTRIBUTE_COUNT, nk.NK_FORMAT_COUNT, 0},
	}
	var cfg nk.Nk_convert_config
	cfg.Shape_AA = nk.NK_ANTI_ALIASING_ON
	cfg.Line_AA = nk.NK_ANTI_ALIASING_ON
	cfg.Vertex_layout = uintptr(unsafe.Pointer(&vertex_layout[0]))
	cfg.Vertex_size = nk.Nk_size(unsafe.Sizeof(vertex{}))
	cfg.Vertex_alignment = nk.Nk_size(unsafe.Alignof(vertex{}))
	cfg.Circle_segment_count = 22
	cfg.Curve_segment_count = 22
	cfg.Arc_segment_count = 22
	cfg.Global_alpha = 1.0
	null := nk.Nk_draw_null_texture{}
	cfg.Null = null

	cmdbuf := uintptr(unsafe.Pointer(new(nk.Nk_buffer)))
	vertbuf := uintptr(unsafe.Pointer(new(nk.Nk_buffer)))
	idxbuf := uintptr(unsafe.Pointer(new(nk.Nk_buffer)))

	nk.Xnk_buffer_init_default(cmdbuf)
	nk.Xnk_buffer_init_default(vertbuf)
	nk.Xnk_buffer_init_default(idxbuf)

	nk.Xnk_convert(ctx, cmdbuf, vertbuf, idxbuf, uintptr(unsafe.Pointer(&cfg)))
	vertptr := nk.Xnk_buffer_memory(vertbuf)
	vertsize := nk.Xnk_buffer_total(vertbuf)
	verts := ((*[1 << 20]vertex)(unsafe.Pointer(vertptr)))[:vertsize]

	idxptr := nk.Xnk_buffer_memory(idxbuf)
	idxsize := nk.Xnk_buffer_total(idxbuf)
	idxs := ((*[1 << 20]uint16)(unsafe.Pointer(idxptr)))[:idxsize]

	cmd := nk.Xnk__draw_begin(ctx, cmdbuf)
	for ; cmd != 0; cmd = nk.Xnk__draw_next(cmd, cmdbuf, ctx) {
		cmdp := (*nk.Nk_draw_command)(unsafe.Pointer(cmd))
		if cmdp.Elem_count == 0 {
			continue
		}
		// fmt.Printf("verts:%d, tex:%d\n", cmdp.Elem_count, cmdp.Texture.Ptr)
		if cmdp.Texture.Ptr != 0 {
			shader.Texture = tex
		} else {
			shader.Texture = nil
		}

		idx := idxs[:cmdp.Elem_count]
		vs := make([]fauxgl.Vertex, 0, cmdp.Elem_count)
		for _, i := range idx {
			v := verts[i]
			vs = append(vs, fauxgl.Vertex{
				Position: fauxgl.V(float64(v.pos[0]), float64(v.pos[1]), 0),
				Texture:  fauxgl.V(float64(v.uv[0]), float64(v.uv[1]), 0),
				Color:    fauxgl.MakeColor(color.RGBA{v.color[0], v.color[1], v.color[2], v.color[3]}),
			})
		}

		for i := 0; i < len(vs); i += 3 {
			t := &fauxgl.Triangle{
				V1: vs[i],
				V2: vs[i+1],
				V3: vs[i+2],
			}
			// t := (*fauxgl.Triangle)(unsafe.Pointer(&vs[i]))
			fctx.DrawTriangle(t)
		}
		idxs = idxs[cmdp.Elem_count:]
	}
	nk.Xnk_clear(ctx)

	nk.Xnk_buffer_free(cmdbuf)
	nk.Xnk_buffer_free(vertbuf)
	nk.Xnk_buffer_free(idxbuf)

	return fctx.Image()
}

type glshader struct {
	Matrix  fauxgl.Matrix
	Texture fauxgl.Texture
}

func (s *glshader) Vertex(v fauxgl.Vertex) fauxgl.Vertex {
	v.Output = s.Matrix.MulPositionW(v.Position)
	return v
}

func (s *glshader) Fragment(v fauxgl.Vertex) fauxgl.Color {
	if s.Texture != nil {
		return s.Texture.Sample(v.Texture.X, 1-v.Texture.Y).Mul(v.Color)
	}
	return v.Color
}

func init() {
	app.Register("win", winMain)
}
