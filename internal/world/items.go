package world

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"superdario/internal/core"
	"superdario/internal/gfx"
)

// ---------------------------------------------------------------------------
// Pickups
// ---------------------------------------------------------------------------

type ItemKind uint8

const (
	ItToken    ItemKind = iota // a context token — the game's coin
	ItCompute                  // Compute Cluster: grow
	ItScroll                   // Constitution Scroll: alignment bolts
	ItSnapshot                 // extra life
	ItHalo                     // Safety Halo: invincibility
	ItModel                    // a rescued Claude, purely cosmetic pickup
)

type Item struct {
	Body
	Kind     ItemKind
	Alive    bool
	animT    float64
	emerge   float64 // pixels still to rise out of the block that spawned it
	walks    bool
	modelIdx int
}

func newItem(kind ItemKind, x, y float64) *Item {
	it := &Item{Kind: kind, Alive: true, Body: Body{Facing: core.Right}}
	it.W, it.H = 12, 12
	switch kind {
	case ItToken:
		it.W, it.H = 10, 13
	case ItCompute, ItSnapshot:
		it.W, it.H = 13, 14
		it.walks = true
	case ItScroll:
		it.W, it.H = 13, 14
	case ItHalo:
		it.W, it.H = 13, 14
		it.walks = true
	}
	it.X = x + (TS-it.W)/2
	it.Y = y + TS - it.H
	// Stagger the animation by position. Without this every coin in a row
	// spins in lockstep and they all go edge-on at the same moment.
	it.animT = math.Mod(math.Abs(x)*0.37+math.Abs(y)*0.71, 6)
	return it
}

// spawnFromBlock makes the item rise slowly out of the block it came from.
func (it *Item) spawnFromBlock() {
	it.emerge = TS
	it.Y += TS
}

func (it *Item) Update(w *World) {
	if !it.Alive {
		return
	}
	l := w.Level
	it.animT += 0.14

	if it.emerge > 0 {
		step := math.Min(0.6, it.emerge)
		it.Y -= step
		it.emerge -= step
		return
	}

	switch it.Kind {
	case ItToken, ItScroll, ItModel:
		// stationary
	case ItCompute, ItSnapshot, ItHalo:
		if it.VX == 0 {
			it.VX = 0.8
		}
		it.VX = sign(it.VX) * 0.8
		if it.Kind == ItHalo {
			it.VX = sign(it.VX) * 1.5
		}
		it.ApplyGravity(l.Def, it.InLiquid(l))
		it.MoveX(l, it.VX)
		if it.HitLeft || it.HitRight {
			it.VX = -it.VX
		}
		wasGround := it.OnGround
		it.MoveY(l, it.VY, true)
		if it.Kind == ItHalo && it.OnGround && wasGround {
			it.VY = -3.6 // the Halo bounces along
		}
		if it.Y > l.PixelHeight()+64 {
			it.Alive = false
		}
	}
}

func (it *Item) Sprite() *ebiten.Image {
	switch it.Kind {
	case ItToken:
		return gfx.ImgToken[int(it.animT)%len(gfx.ImgToken)]
	case ItCompute:
		return gfx.ImgCompute
	case ItScroll:
		return gfx.ImgScroll
	case ItSnapshot:
		return gfx.ImgSnapshot
	case ItHalo:
		return gfx.ImgHalo[(int(it.animT)/2)%2]
	case ItModel:
		return gfx.ImgModel[it.modelIdx%4][(int(it.animT)/3)%2]
	}
	return nil
}

func (it *Item) Draw(dst *ebiten.Image, camX, camY float64) {
	if !it.Alive {
		return
	}
	img := it.Sprite()
	if img == nil {
		return
	}
	sw := float64(img.Bounds().Dx())
	sh := float64(img.Bounds().Dy())
	gfx.DrawSprite(dst, img, it.CX()-sw/2-camX, it.Bottom()-sh-camY, gfx.SpriteOpt{})
}

// ---------------------------------------------------------------------------
// Alignment bolts
// ---------------------------------------------------------------------------

type Bolt struct {
	Body
	Alive  bool
	life   int
	animT  float64
	bounce int
}

func newBolt(x, y float64, dir int) *Bolt {
	b := &Bolt{Alive: true, life: 200}
	b.W, b.H = 8, 8
	b.X, b.Y = x, y
	b.VX = float64(dir) * 4.2
	b.VY = 1.5
	b.Facing = dir
	return b
}

func (b *Bolt) Update(w *World) {
	if !b.Alive {
		return
	}
	l := w.Level
	b.animT += 0.5
	b.life--
	if b.life <= 0 {
		b.Alive = false
		return
	}
	b.VY += 0.30
	if b.VY > 5 {
		b.VY = 5
	}
	b.MoveX(l, b.VX)
	if b.HitLeft || b.HitRight {
		b.explode(w)
		return
	}
	b.MoveY(l, b.VY, false)
	if b.OnGround {
		b.VY = -3.2
		b.bounce++
		if b.bounce > 8 {
			b.explode(w)
			return
		}
	}
	if b.X < 0 || b.X > l.PixelWidth() || b.Y > l.PixelHeight()+32 {
		b.Alive = false
	}
}

func (b *Bolt) explode(w *World) {
	b.Alive = false
	w.Particles.Burst(b.CX(), b.CY(), 9, gfx.ColCream, 1.6)
	w.Particles.Ring(b.CX(), b.CY(), gfx.Alpha(gfx.ColCoral, 0.8), 4)
}

func (b *Bolt) Draw(dst *ebiten.Image, camX, camY float64) {
	if !b.Alive {
		return
	}
	img := gfx.ImgBolt[int(b.animT)%2]
	gfx.DrawSprite(dst, img, b.CX()-8-camX, b.CY()-8-camY, gfx.SpriteOpt{})
}

// ---------------------------------------------------------------------------
// Moving and falling platforms
// ---------------------------------------------------------------------------

type PlatKind uint8

const (
	PlatH PlatKind = iota
	PlatV
	PlatFall
)

type Platform struct {
	Body
	Kind     PlatKind
	Alive    bool
	origin   [2]float64
	phase    float64
	span     float64
	fallT    int
	Carrying bool
	DX, DY   float64 // movement applied this frame, so riders can follow
}

func newPlatform(kind PlatKind, x, y float64) *Platform {
	p := &Platform{Kind: kind, Alive: true}
	p.W, p.H = 3*TS, 8
	if kind == PlatFall {
		p.W = 2 * TS
	}
	p.X, p.Y = x, y
	p.origin = [2]float64{x, y}
	switch kind {
	case PlatH:
		p.span = 6 * TS
	case PlatV:
		p.span = 4 * TS
	}
	return p
}

func (p *Platform) Update(w *World) {
	if !p.Alive {
		return
	}
	px, py := p.X, p.Y
	switch p.Kind {
	case PlatH:
		p.phase += 0.014
		p.X = p.origin[0] + math.Sin(p.phase)*p.span
	case PlatV:
		p.phase += 0.016
		p.Y = p.origin[1] + math.Sin(p.phase)*p.span
	case PlatFall:
		if p.fallT > 0 {
			p.fallT--
			// shake before it goes
			p.X = p.origin[0] + math.Sin(float64(p.fallT)*0.9)*1.2
			if p.fallT == 0 {
				p.VY = 0.4
			}
		} else if p.VY > 0 {
			p.VY += 0.22
			p.Y += p.VY
			if p.Y > w.Level.PixelHeight()+64 {
				p.Alive = false
			}
		}
	}
	p.DX, p.DY = p.X-px, p.Y-py
}

// Trigger starts a falling platform's countdown when stepped on.
func (p *Platform) Trigger() {
	if p.Kind == PlatFall && p.fallT == 0 && p.VY == 0 {
		p.fallT = 28
	}
}

func (p *Platform) Draw(dst *ebiten.Image, ts *gfx.Tileset, camX, camY float64) {
	if !p.Alive {
		return
	}
	n := int(p.W) / TS
	for i := 0; i < n; i++ {
		gfx.DrawSprite(dst, ts.GirderBeam, p.X+float64(i*TS)-camX, p.Y-camY, gfx.SpriteOpt{})
	}
}

// ---------------------------------------------------------------------------
// Fire bars — a rotating arm of hot nodes, the classic castle hazard
// ---------------------------------------------------------------------------

type FireBar struct {
	X, Y   float64
	angle  float64
	speed  float64
	length int
	Alive  bool
}

func newFireBar(x, y float64) *FireBar {
	return &FireBar{X: x + TS/2, Y: y + TS/2, speed: 0.024, length: 5, Alive: true}
}

func (f *FireBar) Update() { f.angle += f.speed }

// Hits reports whether a body touches any node of the bar.
func (f *FireBar) Hits(b *Body) bool {
	for i := 1; i <= f.length; i++ {
		r := float64(i) * 11
		x := f.X + math.Cos(f.angle)*r
		y := f.Y + math.Sin(f.angle)*r
		if b.OverlapsRect(x-4, y-4, 8, 8) {
			return true
		}
	}
	return false
}

func (f *FireBar) Draw(dst *ebiten.Image, camX, camY float64, tick int) {
	for i := 0; i <= f.length; i++ {
		r := float64(i) * 11
		x := f.X + math.Cos(f.angle)*r - camX
		y := f.Y + math.Sin(f.angle)*r - camY
		sz := 7.0 - float64(i)*0.4
		flick := 0.75 + 0.25*math.Sin(float64(tick)*0.4+float64(i))
		gfx.FillRectF(dst, x-sz/2-1, y-sz/2-1, sz+2, sz+2, gfx.Alpha(gfx.ColRed, 0.55*flick))
		gfx.FillRectF(dst, x-sz/2, y-sz/2, sz, sz, gfx.Scale(gfx.ColGold, flick))
		gfx.FillRectF(dst, x-sz/4, y-sz/4, sz/2, sz/2, gfx.ColGoldLight)
	}
}

// ---------------------------------------------------------------------------
// Springs, checkpoints and the rescue pod
// ---------------------------------------------------------------------------

type Spring struct {
	Body
	compress int
}

func newSpring(x, y float64) *Spring {
	s := &Spring{}
	s.W, s.H = 14, 10
	s.X = x + 1
	s.Y = y + TS - s.H
	return s
}

func (s *Spring) Update() {
	if s.compress > 0 {
		s.compress--
	}
}

func (s *Spring) Draw(dst *ebiten.Image, ts *gfx.Tileset, camX, camY float64) {
	squish := float64(s.compress) * 0.35
	y := s.Y - camY + squish
	h := s.H - squish
	gfx.FillRectF(dst, s.X-camX, y+h-3, s.W, 3, gfx.Scale(gfx.ColSlate, 1.2))
	for i := 0.0; i < h-3; i += 3 {
		gfx.FillRectF(dst, s.X-camX+1, y+i, s.W-2, 2, gfx.ColGold)
	}
	gfx.FillRectF(dst, s.X-camX-1, y-2, s.W+2, 3, gfx.ColCoral)
	gfx.StrokeRectF(dst, s.X-camX-1, y-2, s.W+2, 3, gfx.ColCoralDark)
}

type Checkpoint struct {
	Body
	Active bool
	animT  float64
}

func newCheckpoint(x, y float64) *Checkpoint {
	c := &Checkpoint{}
	c.W, c.H = 10, TS*3
	c.X = x + 3
	c.Y = y + TS - c.H
	return c
}

func (c *Checkpoint) Draw(dst *ebiten.Image, camX, camY float64, tick int) {
	c.animT += 0.06
	x := c.X - camX
	y := c.Y - camY
	// post
	gfx.FillRectF(dst, x+3, y, 4, c.H, gfx.ColSlate)
	gfx.FillRectF(dst, x+3, y, 1, c.H, gfx.Scale(gfx.ColSlate, 1.6))
	// banner
	col := gfx.Alpha(gfx.ColCream, 0.65)
	if c.Active {
		col = gfx.ColCoral
	}
	wob := math.Sin(c.animT*2) * 2
	for i := 0.0; i < 12; i++ {
		wv := math.Sin(c.animT*2+i*0.35) * 1.6
		gfx.FillRectF(dst, x+7, y+2+i, 14+wob*0.2, 1, gfx.Scale(col, 1-0.02*i))
		_ = wv
	}
	if c.Active {
		gfx.FillRectF(dst, x+1, y-4, 8, 4, gfx.ColGoldLight)
	}
}

// Goal is the caged Claude model at the end of a stage.
type Goal struct {
	Body
	animT  float64
	Model  int
	Opened bool
}

func newGoal(x, y float64, model int) *Goal {
	g := &Goal{Model: model}
	g.W, g.H = TS, TS*2
	g.X, g.Y = x, y
	return g
}

func (g *Goal) Draw(dst *ebiten.Image, camX, camY float64, tick int) {
	g.animT += 0.05
	x := g.X - camX
	y := g.Y - camY
	// pedestal
	gfx.FillRectF(dst, x-6, y+TS*2-6, TS+12, 6, gfx.ColSlate)
	gfx.FillRectF(dst, x-6, y+TS*2-6, TS+12, 1, gfx.Scale(gfx.ColSlate, 1.5))
	// glow
	pulse := 0.5 + 0.5*math.Sin(g.animT*2)
	for r := 22.0; r > 6; r -= 4 {
		gfx.FillRectF(dst, x+8-r/2, y+16-r/2, r, r, gfx.Alpha(gfx.ModelColors[maxInt(g.Model, 0)%4], 0.06*pulse))
	}
	// the model itself
	idx := 0
	if g.Model >= 0 {
		idx = g.Model
	}
	bob := math.Sin(g.animT*3) * 2
	gfx.DrawSprite(dst, gfx.ImgModel[idx%4][(tick/12)%2], x, y+10+bob, gfx.SpriteOpt{})
	if !g.Opened {
		gfx.DrawSprite(dst, gfx.ImgCage, x, y+8, gfx.SpriteOpt{Alpha: 0.92})
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
