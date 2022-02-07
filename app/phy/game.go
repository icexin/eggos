//go:build phy
// +build phy

package phy

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"
	"time"

	"github.com/icexin/eggos/app"
	"github.com/icexin/eggos/drivers/kbd"
	"github.com/icexin/eggos/drivers/vbe"
	"github.com/jakecoffman/cp"
)

const gravityStrength = 5.0e6

var planetBody *cp.Body

func planetGravityVelocity(body *cp.Body, gravity cp.Vector, damping, dt float64) {
	p := body.Position()
	sqdist := p.LengthSq()
	g := p.Mult(-gravityStrength / (sqdist * math.Sqrt(sqdist)))
	body.UpdateVelocity(g, damping, dt)
}

func randPos(radius float64) cp.Vector {
	var v cp.Vector
	for {
		v = cp.Vector{rand.Float64()*(640-2*radius) - (320 - radius), rand.Float64()*(480-2*radius) - (240 - radius)}
		if v.Length() >= 85 {
			return v
		}
	}
}

func addBox(space *cp.Space) {
	size := 10.0
	mass := 1.0

	verts := []cp.Vector{
		{-size, -size},
		{-size, size},
		{size, size},
		{size, -size},
	}

	radius := cp.Vector{size, size}.Length()
	pos := randPos(radius)

	body := space.AddBody(cp.NewBody(mass, cp.MomentForPoly(mass, len(verts), verts, cp.Vector{}, 0)))
	body.SetVelocityUpdateFunc(planetGravityVelocity)
	body.SetPosition(pos)

	r := pos.Length()
	v := math.Sqrt(gravityStrength/r) / r
	body.SetVelocityVector(pos.Perp().Mult(v))

	body.SetAngularVelocity(v)
	body.SetAngle(math.Atan2(pos.Y, pos.X))

	shape := space.AddShape(cp.NewPolyShape(body, 4, verts, cp.NewTransformIdentity(), 0))
	// shape := space.AddShape(cp.NewCircle(body, float64(size), cp.Vector{0, 0}))
	shape.SetElasticity(0)
	shape.SetFriction(0.7)
}

var GRABBABLE_MASK_BIT uint = 1 << 31

type Game struct {
	space  *cp.Space
	drawer *Drawer
}

func NewGame(w, h int) *Game {
	g := &Game{
		space: cp.NewSpace(),
		drawer: NewDrawer(DrawOption{
			Width:          w,
			Height:         h,
			Flags:          cp.DRAW_SHAPES | cp.DRAW_CONSTRAINTS | cp.DRAW_COLLISION_POINTS,
			Outline:        cp.FColor{200.0 / 255.0, 210.0 / 255.0, 230.0 / 255.0, 1},
			Constraint:     cp.FColor{0, 0.75, 0, 1},
			CollisionPoint: cp.FColor{1, 0, 0, 1},
		}),
	}
	g.Init()
	return g
}

func (g *Game) Init() {
	planetBody = g.space.AddBody(cp.NewKinematicBody())
	planetBody.SetAngularVelocity(0.2)

	N := 50
	for i := 0; i < N; i++ {
		addBox(g.space)
	}

	shape := g.space.AddShape(cp.NewCircle(planetBody, 70, cp.Vector{}))
	shape.SetElasticity(1)
	shape.SetFriction(1)
	shape.SetFilter(cp.ShapeFilter{
		cp.NO_GROUP, ^GRABBABLE_MASK_BIT, ^GRABBABLE_MASK_BIT,
	})
}

func (g *Game) Update(dt float64) error {
	// g.space.Step(dt)
	g.space.Step(1 / (120.0))
	return nil
}

func (g *Game) Draw(screen *vbe.View) {
	g.drawer.NewFrame()
	cp.DrawSpace(g.space, g.drawer)
	draw.Draw(screen.Canvas(), g.drawer.Image().Bounds(), image.NewUniform(color.RGBA{38, 47, 57, 255}), image.ZP, draw.Src)
	draw.Draw(screen.Canvas(), g.drawer.Image().Bounds(), g.drawer.Image(), image.ZP, draw.Src)

	screen.Commit()
}

func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return outsideWidth, outsideHeight
}

func main(ctx *app.Context) error {
	game := NewGame(600, 480)
	game.Init()
	last := time.Now()
	for !kbd.Pressed('q') {
		now := time.Now()
		game.Update(now.Sub(last).Seconds())
		game.Draw(vbe.DefaultView)
		last = now
	}
	return nil
}

func init() {
	app.Register("phy", main)
}
