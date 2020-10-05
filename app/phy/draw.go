package phy

import (
	"image"
	"math"

	"github.com/fogleman/gg"
	"github.com/jakecoffman/cp"
)

const (
	LineWidth = 2
)

type DrawOption struct {
	Width, Height  int
	Flags          uint
	Outline        cp.FColor
	Constraint     cp.FColor
	CollisionPoint cp.FColor
	Data           interface{}
}

type Drawer struct {
	ctx    *gg.Context
	option DrawOption
}

func NewDrawer(opt DrawOption) *Drawer {
	ctx := gg.NewContext(opt.Width, opt.Height)
	ctx.Translate(float64(opt.Width)/2, float64(opt.Height)/2)
	ctx.Scale(1, -1)
	return &Drawer{
		ctx:    ctx,
		option: opt,
	}
}

func (d *Drawer) NewFrame() {
	d.ctx.SetRGBA(0, 0, 0, 0)
	d.ctx.Clear()
}

func (d *Drawer) Image() image.Image {
	return d.ctx.Image()
}

func setColor(ctx *gg.Context, c cp.FColor) {
	ctx.SetRGBA(float64(c.R), float64(c.G), float64(c.B), float64(c.A))
}

func (d *Drawer) DrawCircle(pos cp.Vector, angle float64, radius float64, outline cp.FColor, fill cp.FColor, data interface{}) {
	d.ctx.DrawCircle(pos.X, pos.Y, radius)
	d.ctx.SetLineWidth(LineWidth)
	setColor(d.ctx, outline)
	d.ctx.StrokePreserve()
	setColor(d.ctx, fill)
	d.ctx.Fill()
	d.DrawFatSegment(pos, pos.Add(cp.ForAngle(angle).Mult(radius-LineWidth*0.5)), 0, outline, fill, nil)
}

func (d *Drawer) DrawSegment(a cp.Vector, b cp.Vector, fill cp.FColor, data interface{}) {
	d.ctx.DrawLine(a.X, a.Y, b.X, b.Y)
	d.ctx.SetLineWidth(LineWidth)
	setColor(d.ctx, fill)
	d.ctx.Stroke()
}

func (d *Drawer) DrawFatSegment(a cp.Vector, b cp.Vector, radius float64, outline cp.FColor, fill cp.FColor, data interface{}) {
	d.ctx.DrawLine(a.X, a.Y, b.X, b.Y)
	width := radius * 2
	color := fill
	if width == 0 {
		width = 1
		color = outline
	}
	d.ctx.SetLineWidth(width)
	d.ctx.SetLineCapRound()
	setColor(d.ctx, color)
	d.ctx.Stroke()
}

func (d *Drawer) DrawPolygon(count int, verts []cp.Vector, radius float64, outline cp.FColor, fill cp.FColor, data interface{}) {
	for _, v := range verts {
		d.ctx.LineTo(v.X, v.Y)
	}
	d.ctx.LineTo(verts[0].X, verts[0].Y)
	d.ctx.SetLineWidth(LineWidth)
	setColor(d.ctx, outline)
	d.ctx.StrokePreserve()
	setColor(d.ctx, fill)
	d.ctx.Fill()
}

func (d *Drawer) DrawDot(size float64, pos cp.Vector, fill cp.FColor, data interface{}) {
	radius := size/2 + LineWidth
	d.ctx.DrawCircle(pos.X, pos.Y, radius)
	setColor(d.ctx, fill)
	d.ctx.Fill()
}

func (d *Drawer) Flags() uint {
	return d.option.Flags
}

func (d *Drawer) OutlineColor() cp.FColor {
	return d.option.Outline
}

func (d *Drawer) ShapeColor(shape *cp.Shape, data interface{}) cp.FColor {
	return ColorForShape(shape, data)
}

func (d *Drawer) ConstraintColor() cp.FColor {
	return d.option.Constraint
}

func (d *Drawer) CollisionPointColor() cp.FColor {
	return d.option.CollisionPoint
}

func (d *Drawer) Data() interface{} {
	return d.option.Data
}

func ColorForShape(shape *cp.Shape, data interface{}) cp.FColor {
	if shape.Sensor() {
		return cp.FColor{R: 1, G: 1, B: 1, A: .1}
	}

	body := shape.Body()

	if body.IsSleeping() {
		return cp.FColor{R: .2, G: .2, B: .2, A: 1}
	}

	if body.IdleTime() > shape.Space().SleepTimeThreshold {
		return cp.FColor{R: .66, G: .66, B: .66, A: 1}
	}

	val := shape.HashId()

	// scramble the bits up using Robert Jenkins' 32 bit integer hash function
	val = (val + 0x7ed55d16) + (val << 12)
	val = (val ^ 0xc761c23c) ^ (val >> 19)
	val = (val + 0x165667b1) + (val << 5)
	val = (val + 0xd3a2646c) ^ (val << 9)
	val = (val + 0xfd7046c5) + (val << 3)
	val = (val ^ 0xb55a4f09) ^ (val >> 16)

	r := float32((val >> 0) & 0xFF)
	g := float32((val >> 8) & 0xFF)
	b := float32((val >> 16) & 0xFF)

	max := float32(math.Max(math.Max(float64(r), float64(g)), float64(b)))
	min := float32(math.Min(math.Min(float64(r), float64(g)), float64(b)))
	var intensity float32
	if body.GetType() == cp.BODY_STATIC {
		intensity = 0.15
	} else {
		intensity = 0.75
	}

	if min == max {
		return cp.FColor{R: intensity, A: 1}
	}

	var coef float32 = intensity / (max - min)
	return cp.FColor{
		R: (r - min) * coef,
		G: (g - min) * coef,
		B: (b - min) * coef,
		A: 1,
	}
}
