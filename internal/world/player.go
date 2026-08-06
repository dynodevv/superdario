package world

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"superdario/internal/core"
	"superdario/internal/gfx"
)

// Input is the abstract control state for one frame, filled in by the scene.
type Input struct {
	Left, Right, Up, Down bool
	Jump, JumpDown        bool // Jump = pressed this frame, JumpDown = held
	Run                   bool
	Fire                  bool // pressed this frame
}

// Player is Dario. Hitbox is narrower than the sprite so corners feel fair.
type Player struct {
	Body

	Power   core.PowerState
	Dead    bool
	DeadT   int
	Ducking bool
	Skid    bool

	Invuln int // post-damage blink
	Halo   int // Safety Halo invincibility

	coyote   int
	jumpBuf  int
	jumping  bool
	animT    float64
	swimT    int
	strokeCD int // ticks until the next stroke while the button is held
	boltCD   int
	Transf   int // transform freeze timer
	transfTo core.PowerState

	// pipe / goal cinematics
	Frozen  bool
	CheerT  int
	StartXY [2]float64
}

const (
	hbW      = 11.0
	hbSmallH = 14.0
	hbBigH   = 26.0
	hbDuckH  = 16.0
)

func NewPlayer(x, y float64, power core.PowerState) *Player {
	p := &Player{Power: power}
	p.W = hbW
	p.H = hbSmallH
	if power.Big() {
		p.H = hbBigH
	}
	// x,y is the tile corner the '@' sat in; stand the player on its floor
	p.X = x + (TS-hbW)/2
	p.Y = y + TS - p.H
	p.Facing = core.Right
	p.StartXY = [2]float64{p.X, p.Y}
	return p
}

func (p *Player) targetHeight() float64 {
	switch {
	case p.Ducking && p.Power.Big():
		return hbDuckH
	case p.Power.Big():
		return hbBigH
	default:
		return hbSmallH
	}
}

// setHeight resizes the hitbox keeping the feet planted.
func (p *Player) setHeight(h float64) {
	if p.H == h {
		return
	}
	p.Y += p.H - h
	p.H = h
}

// Update runs one tick of Dario. Anything with world-level consequences
// (blocks, pickups, enemies) is handled by World; this is movement only.
func (p *Player) Update(w *World, in Input) {
	l := w.Level

	if p.Invuln > 0 {
		p.Invuln--
	}
	if p.Halo > 0 {
		p.Halo--
		if p.Halo == 0 {
			w.onHaloEnd()
		}
	}
	if p.boltCD > 0 {
		p.boltCD--
	}

	if p.Dead {
		p.updateDeath(w)
		return
	}
	if p.Transf > 0 {
		p.Transf--
		p.VX = 0
		return
	}
	if p.Frozen {
		// level-clear walk: keep drifting right until the cinematic ends
		p.VX = 1.1
		p.ApplyGravity(l.Def, false)
		p.MoveX(l, p.VX)
		p.MoveY(l, p.VY, true)
		p.animT += 0.25
		return
	}

	inWater := p.InLiquid(l)

	// --- horizontal -----------------------------------------------------
	accel := core.WalkAccel
	maxSpd := core.MaxWalk
	if in.Run && !inWater {
		accel = core.RunAccel
		maxSpd = core.MaxRun
	}
	if inWater {
		maxSpd = core.MaxWalk * 0.75
		accel = core.WalkAccel * 0.7
	}

	dir := 0
	if in.Left {
		dir = -1
	}
	if in.Right {
		dir = 1
	}
	// ducking on the ground locks movement (you can still slide out momentum)
	if p.Ducking && p.OnGround {
		dir = 0
	}

	p.Skid = false
	switch {
	case dir != 0:
		if p.VX != 0 && sign(p.VX) != float64(dir) && p.OnGround {
			// skidding: decelerate hard and show the turn-around pose
			p.VX += float64(dir) * core.SkidDecel
			p.Skid = true
		} else {
			p.VX += float64(dir) * accel
		}
		p.Facing = dir
	case p.OnGround:
		if math.Abs(p.VX) <= core.Friction {
			p.VX = 0
		} else {
			p.VX -= sign(p.VX) * core.Friction
		}
	default:
		if math.Abs(p.VX) <= core.AirDrag {
			p.VX = 0
		} else {
			p.VX -= sign(p.VX) * core.AirDrag
		}
	}
	if math.Abs(p.VX) > maxSpd {
		p.VX = sign(p.VX) * maxSpd
	}
	p.VX += p.OnBelt(l)

	// --- ducking --------------------------------------------------------
	wantDuck := in.Down && p.OnGround && p.Power.Big() && !inWater
	if wantDuck != p.Ducking {
		if !wantDuck && !p.canStand(l) {
			// something overhead: stay crouched
			wantDuck = true
		}
		p.Ducking = wantDuck
	}
	p.setHeight(p.targetHeight())

	// --- jumping --------------------------------------------------------
	if p.OnGround {
		p.coyote = core.CoyoteTicks
		p.jumping = false
	} else if p.coyote > 0 {
		p.coyote--
	}
	if in.Jump {
		p.jumpBuf = core.BufferTicks
	} else if p.jumpBuf > 0 {
		p.jumpBuf--
	}

	if inWater {
		p.swimT++
		if p.strokeCD > 0 {
			p.strokeCD--
		}
		// Tap to stroke, or hold to keep stroking. Requiring a mash to gain
		// height is miserable, especially on a d-pad.
		stroke := in.Jump || (in.JumpDown && p.strokeCD == 0)
		if stroke {
			p.VY = core.SwimStroke
			p.strokeCD = core.SwimHoldGap
			p.swimT = 0
			w.playSwim()
		}
		// cap the climb so chained strokes rise smoothly rather than rocketing
		if p.VY < core.SwimRise {
			p.VY = core.SwimRise
		}
		p.jumping = false
		p.coyote = 0
	} else if p.jumpBuf > 0 && p.coyote > 0 {
		v := core.JumpVel
		// a running jump goes noticeably higher
		v += core.JumpRunBonus * math.Min(1, math.Abs(p.VX)/core.MaxRun)
		p.VY = v
		p.jumpBuf = 0
		p.coyote = 0
		p.jumping = true
		p.OnGround = false
		w.playJump(p.Power.Big())
	}
	// releasing jump early cuts the arc
	if p.jumping && !in.JumpDown && p.VY < 0 {
		p.VY *= core.JumpCut
		p.jumping = false
	}

	// --- integrate ------------------------------------------------------
	p.ApplyGravity(l.Def, inWater)
	p.MoveX(l, p.VX)
	p.MoveY(l, p.VY, true)

	// --- fire -----------------------------------------------------------
	if in.Fire && p.Power == core.PowerAlign && p.boltCD == 0 && !p.Ducking {
		if w.spawnBolt(p) {
			p.boltCD = 16
		}
	}

	// --- animation ------------------------------------------------------
	if p.OnGround {
		p.animT += math.Abs(p.VX) * 0.09
	} else {
		p.animT += 0.06
	}
}

// canStand checks headroom before standing up from a duck.
func (p *Player) canStand(l *Level) bool {
	top := p.Bottom() - hbBigH
	tx0 := ifloor(p.X / TS)
	tx1 := ifloor((p.Right() - 0.001) / TS)
	ty0 := ifloor(top / TS)
	ty1 := ifloor((p.Bottom() - hbDuckH - 0.001) / TS)
	for ty := ty0; ty <= ty1; ty++ {
		for tx := tx0; tx <= tx1; tx++ {
			if l.At(tx, ty).Solid() {
				return false
			}
		}
	}
	return true
}

func (p *Player) updateDeath(w *World) {
	p.DeadT++
	if p.DeadT < 30 {
		return // brief pause before the fall
	}
	p.VY += 0.30
	if p.VY > 8 {
		p.VY = 8
	}
	p.Y += p.VY
}

// Kill starts the death animation.
func (p *Player) Kill(w *World) {
	if p.Dead {
		return
	}
	p.Dead = true
	p.DeadT = 0
	p.VY = -6.0
	p.VX = 0
	p.Ducking = false
	p.setHeight(hbSmallH)
	w.onPlayerDeath()
}

// Hurt applies damage. Returns true if it actually landed.
func (p *Player) Hurt(w *World) bool {
	if p.Dead || p.Invuln > 0 || p.Halo > 0 || p.Transf > 0 {
		return false
	}
	if p.Power == core.PowerSmall {
		p.Kill(w)
		return true
	}
	p.setPower(p.Power - 1) // Aligned -> Scaled -> Small
	p.Invuln = 110
	p.Transf = 26
	w.onPlayerShrink()
	return true
}

// setPower changes tier and resizes cleanly.
func (p *Player) setPower(to core.PowerState) {
	if to < core.PowerSmall {
		to = core.PowerSmall
	}
	if to > core.PowerAlign {
		to = core.PowerAlign
	}
	wasBig := p.Power.Big()
	p.Power = to
	if p.Power.Big() != wasBig {
		p.Ducking = false
		p.setHeight(p.targetHeight())
	}
}

// Grow applies a power-up. Returns the score awarded.
func (p *Player) Grow(w *World, to core.PowerState) int {
	if to <= p.Power {
		// already at or above this tier: award points instead
		return 200
	}
	p.setPower(to)
	p.Transf = 30
	w.onPlayerGrow()
	return 1000
}

// GiveHalo starts invincibility.
func (p *Player) GiveHalo(ticks int) {
	p.Halo = ticks
}

// Bounce is the pop Dario gets from stomping something.
func (p *Player) Bounce(hold bool) {
	if hold {
		p.VY = core.StompHold
	} else {
		p.VY = core.StompBounce
	}
	p.jumping = true
	p.OnGround = false
}

// SpriteFrame picks the animation frame for the current state.
func (p *Player) SpriteFrame(l *Level) int {
	switch {
	case p.Dead:
		return gfx.FrDead
	case p.CheerT > 0:
		return gfx.FrCheer
	case p.InLiquid(l):
		if (p.swimT/8)%2 == 0 {
			return gfx.FrSwim1
		}
		return gfx.FrSwim2
	case p.Ducking && p.Power.Big():
		return gfx.FrDuck
	case !p.OnGround:
		return gfx.FrJump
	case p.Skid:
		return gfx.FrSkid
	case math.Abs(p.VX) < 0.12:
		return gfx.FrStand
	default:
		switch int(p.animT) % 4 {
		case 0:
			return gfx.FrWalk1
		case 1, 3:
			return gfx.FrWalk2
		default:
			return gfx.FrWalk3
		}
	}
}

// Draw paints Dario, honouring invincibility flashing and the transform pop.
func (p *Player) Draw(dst *ebiten.Image, l *Level, camX, camY float64, tick int) {
	// blink out on alternating frames while briefly invulnerable
	if p.Invuln > 0 && p.Transf == 0 && (tick/3)%2 == 0 {
		return
	}
	state := int(p.Power)
	// during a transform, flicker between the two sizes
	if p.Transf > 0 {
		if (p.Transf/4)%2 == 0 {
			if state > 0 {
				state = 0
			} else {
				state = 1
			}
		}
	}
	if state >= len(gfx.Dario) {
		state = len(gfx.Dario) - 1
	}
	frames := gfx.Dario[state]
	fr := p.SpriteFrame(l)
	if fr >= len(frames) {
		fr = gfx.FrStand
	}
	img := frames[fr]

	sw := float64(img.Bounds().Dx())
	sh := float64(img.Bounds().Dy())
	x := p.CX() - sw/2 - camX
	y := p.Bottom() - sh - camY

	op := gfx.SpriteOpt{FlipX: p.Facing == core.Left}
	if p.Halo > 0 {
		op.Tint = gfx.TintFor(tick)
		if p.Halo < 90 && (p.Halo/4)%2 == 0 {
			op.Tint = gfx.ColCream
		}
	}
	gfx.DrawSprite(dst, img, x, y, op)
}

func sign(v float64) float64 {
	if v > 0 {
		return 1
	}
	if v < 0 {
		return -1
	}
	return 0
}
