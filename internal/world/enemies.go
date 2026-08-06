package world

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"superdario/internal/core"
	"superdario/internal/gfx"
)

type EnemyKind uint8

const (
	EnDistill  EnemyKind = iota // walker, stompable
	EnCrawler                   // walker that retracts into a kickable shell
	EnShell                     // a retracted crawler
	EnSpikeBot                  // walker, cannot be stomped
	EnFlyer                     // hops/flies along a path
	EnLeech                     // drifts in a sine wave, ignores terrain
	EnPlant                     // lunges out of a pipe
)

// Enemy covers every non-boss hostile. One struct with a kind switch beats a
// pile of near-identical types at this scale.
type Enemy struct {
	Body
	Kind  EnemyKind
	Alive bool

	squashT  int  // >0 while playing the stomped animation
	flipping bool // knocked over by a bolt/shell — tumbles off screen
	shellT   int  // ticks a shell has been still (before waking up)
	kicked   bool // shell is sliding
	animT    float64
	homeY    float64 // for flyers/leeches/plants
	phase    float64
	active   bool // woken up once on screen
	stunned  int
}

func newEnemy(kind EnemyKind, x, y float64) *Enemy {
	e := &Enemy{Kind: kind, Alive: true, Body: Body{Facing: core.Left}}
	switch kind {
	case EnDistill:
		e.W, e.H = 13, 13
	case EnCrawler:
		e.W, e.H = 12, 15
	case EnShell:
		e.W, e.H = 13, 10
	case EnSpikeBot:
		e.W, e.H = 13, 13
	case EnFlyer:
		e.W, e.H = 13, 13
	case EnLeech:
		e.W, e.H = 12, 9
	case EnPlant:
		e.W, e.H = 12, 20
	}
	e.X = x + (TS-e.W)/2
	e.Y = y + TS - e.H
	if kind == EnPlant {
		// plants sit inside the pipe below and rise out of it
		e.Y = y + TS
	}
	e.homeY = e.Y
	return e
}

// walkSpeed per kind.
func (e *Enemy) walkSpeed() float64 {
	switch e.Kind {
	case EnDistill:
		return 0.42
	case EnCrawler:
		return 0.36
	case EnSpikeBot:
		return 0.30
	case EnFlyer:
		return 0.55
	default:
		return 0.4
	}
}

// Update advances one enemy.
func (e *Enemy) Update(w *World) {
	if !e.Alive {
		return
	}
	l := w.Level

	// Enemies stay dormant until they come near the camera, so a level does
	// not have everything marching about from tick zero.
	if !e.active {
		if e.X < w.camX+core.ScreenW+64 && e.X > w.camX-96 {
			e.active = true
		} else {
			return
		}
	}

	if e.squashT > 0 {
		e.squashT--
		if e.squashT == 0 {
			e.Alive = false
		}
		return
	}
	if e.flipping {
		e.VY += 0.32
		e.X += e.VX
		e.Y += e.VY
		e.animT += 0.4
		if e.Y > l.PixelHeight()+64 {
			e.Alive = false
		}
		return
	}
	if e.stunned > 0 {
		e.stunned--
	}

	switch e.Kind {
	case EnDistill, EnSpikeBot:
		e.walkerUpdate(w, false)
	case EnCrawler:
		e.walkerUpdate(w, true)
	case EnShell:
		e.shellUpdate(w)
	case EnFlyer:
		e.flyerUpdate(w)
	case EnLeech:
		e.leechUpdate(w)
	case EnPlant:
		e.plantUpdate(w)
	}
	e.animT += 0.12
}

func (e *Enemy) walkerUpdate(w *World, cautious bool) {
	l := w.Level
	e.VX = float64(e.Facing) * e.walkSpeed()
	e.ApplyGravity(l.Def, e.InLiquid(l))
	e.MoveX(l, e.VX)
	if e.HitLeft || e.HitRight {
		e.Facing = -e.Facing
	} else if cautious && e.OnGround && e.WouldFall(l, e.Facing) {
		// crawlers are careful; distill bots happily march off cliffs
		e.Facing = -e.Facing
	}
	e.MoveY(l, e.VY, false)
	if e.Y > l.PixelHeight()+64 {
		e.Alive = false
	}
}

func (e *Enemy) shellUpdate(w *World) {
	l := w.Level
	if e.kicked {
		e.VX = sign(e.VX) * 3.4
	} else {
		e.VX = 0
		e.shellT++
		if e.shellT > 420 {
			// the crawler climbs back out
			ne := newEnemy(EnCrawler, e.X-(12-e.W)/2, e.Bottom()-TS)
			ne.active = true
			ne.Y = e.Bottom() - ne.H
			ne.X = e.CX() - ne.W/2
			w.Enemies = append(w.Enemies, ne)
			e.Alive = false
			return
		}
	}
	e.ApplyGravity(l.Def, e.InLiquid(l))
	e.MoveX(l, e.VX)
	if e.HitLeft || e.HitRight {
		e.VX = -e.VX
		w.sfxBump()
	}
	e.MoveY(l, e.VY, false)

	if e.kicked {
		// a moving shell mows down anything it touches
		for _, o := range w.Enemies {
			if o == e || !o.Alive || o.squashT > 0 || o.flipping {
				continue
			}
			if e.Overlaps(&o.Body) {
				o.KnockOut(w, sign(e.VX))
				w.addScore(200, o.CX(), o.Y)
			}
		}
	}
	if e.Y > l.PixelHeight()+64 {
		e.Alive = false
	}
}

func (e *Enemy) flyerUpdate(w *World) {
	l := w.Level
	e.phase += 0.045
	e.VX = float64(e.Facing) * e.walkSpeed()
	e.MoveX(l, e.VX)
	if e.HitLeft || e.HitRight {
		e.Facing = -e.Facing
	}
	e.Y = e.homeY + math.Sin(e.phase)*26
}

func (e *Enemy) leechUpdate(w *World) {
	l := w.Level
	e.phase += 0.06
	// drifts toward Dario, slowly, bobbing
	dx := w.Player.CX() - e.CX()
	e.VX += sign(dx) * 0.008
	if math.Abs(e.VX) > 0.55 {
		e.VX = sign(e.VX) * 0.55
	}
	e.X += e.VX
	e.Y = e.homeY + math.Sin(e.phase)*18
	e.Facing = int(sign(e.VX))
	if e.Facing == 0 {
		e.Facing = core.Right
	}
	if e.X < 0 || e.X > l.PixelWidth() {
		e.Alive = false
	}
}

func (e *Enemy) plantUpdate(w *World) {
	e.phase += 0.02
	// stay down while Dario is standing right on top of the pipe
	near := math.Abs(w.Player.CX()-e.CX()) < 26 && w.Player.Bottom() < e.homeY+4
	cycle := math.Sin(e.phase)
	extend := 0.0
	if !near || cycle < 0 {
		extend = math.Max(0, cycle) * 22
	}
	e.Y = e.homeY - extend
	e.H = 20
}

// Stompable reports whether landing on this enemy kills it.
func (e *Enemy) Stompable() bool {
	switch e.Kind {
	case EnSpikeBot, EnPlant:
		return false
	}
	return true
}

// Stomp handles Dario landing on the enemy. Returns true if Dario should bounce.
func (e *Enemy) Stomp(w *World) bool {
	switch e.Kind {
	case EnCrawler:
		// collapses into a shell rather than dying
		sh := newEnemy(EnShell, 0, 0)
		sh.X = e.CX() - sh.W/2
		sh.Y = e.Bottom() - sh.H
		sh.active = true
		w.Enemies = append(w.Enemies, sh)
		e.Alive = false
		w.addScore(100, e.CX(), e.Y)
		return true
	case EnShell:
		if e.kicked {
			e.kicked = false
			e.VX = 0
			e.shellT = 0
		} else {
			e.kick(w, sign(w.Player.CX()-e.CX()))
		}
		w.addScore(100, e.CX(), e.Y)
		return true
	case EnLeech, EnFlyer, EnDistill:
		e.squashT = 22
		e.VX = 0
		w.addScore(100, e.CX(), e.Y)
		w.Particles.Burst(e.CX(), e.CY(), 8, gfx.ColCyan, 1.4)
		return true
	}
	return false
}

func (e *Enemy) kick(w *World, dir float64) {
	if dir == 0 {
		dir = 1
	}
	e.kicked = true
	e.VX = dir * 3.4
	// nudge clear of Dario so he does not immediately re-collide
	e.X += dir * 3
	w.sfxKick()
}

// KnockOut flips an enemy off the screen (bolts, shells, Halo contact).
func (e *Enemy) KnockOut(w *World, dir float64) {
	if !e.Alive || e.flipping {
		return
	}
	e.flipping = true
	e.VX = dir * 1.2
	e.VY = -3.4
	w.Particles.Burst(e.CX(), e.CY(), 10, gfx.ColGoldLight, 1.8)
}

// Sprite returns the image and its draw offset for the current frame.
func (e *Enemy) Sprite(tick int) *ebiten.Image {
	f := int(e.animT) % 2
	switch e.Kind {
	case EnDistill:
		if e.squashT > 0 {
			return gfx.ImgDistill[2]
		}
		return gfx.ImgDistill[f]
	case EnCrawler:
		return gfx.ImgCrawler[f]
	case EnShell:
		if e.kicked {
			return gfx.ImgShell[(tick/3)%2]
		}
		return gfx.ImgShell[0]
	case EnSpikeBot:
		return gfx.ImgSpikeBot[f]
	case EnFlyer:
		return gfx.ImgFlyer[f]
	case EnLeech:
		return gfx.ImgLeech[f]
	case EnPlant:
		return gfx.ImgDataPipe[f]
	}
	return nil
}

func (e *Enemy) Draw(dst *ebiten.Image, camX, camY float64, tick int) {
	if !e.Alive || !e.active {
		return
	}
	img := e.Sprite(tick)
	if img == nil {
		return
	}
	sw := float64(img.Bounds().Dx())
	sh := float64(img.Bounds().Dy())
	x := e.CX() - sw/2 - camX
	y := e.Bottom() - sh - camY

	op := gfx.SpriteOpt{FlipX: e.Facing == core.Right}
	if e.Kind == EnLeech || e.Kind == EnPlant {
		op.FlipX = false
	}
	if e.flipping {
		op.FlipY = true
	}
	if e.squashT > 0 {
		op.Alpha = math.Min(1, float64(e.squashT)/8)
	}
	gfx.DrawSprite(dst, img, x, y, op)
}
