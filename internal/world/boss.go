package world

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"superdario/internal/gfx"
)

// Every boss shares this state machine; the per-kind behaviour lives in the
// switch inside think(). All of them have a vulnerable window that can be
// stomped, so a player who arrives without the Constitution Scroll can still
// win — bolts just make it faster.

type bossState uint8

const (
	bsIntro bossState = iota
	bsIdle
	bsAct     // performing the kind's signature move
	bsExposed // vulnerable: stompable, sitting still
	bsHurt
	bsDying
	bsDone
)

type Boss struct {
	Body
	Kind     int
	HP       int
	MaxHP    int
	Phase    int
	state    bossState
	timer    int
	animT    float64
	hurtT    int
	deadT    int
	home     [2]float64
	arenaL   float64
	arenaR   float64
	floorY   float64
	Defeated bool
	facing2  float64
	charge   float64
	Started  bool
}

func newBoss(kind int, x, y float64) *Boss {
	dims := gfx.BossDims[kind]
	b := &Boss{Kind: kind, state: bsIntro, timer: 90}
	b.W = float64(dims[0]) - 8
	b.H = float64(dims[1]) - 6
	b.X = x - b.W/2
	b.Y = y - b.H
	b.home = [2]float64{b.X, b.Y}
	b.Facing = -1
	switch kind {
	case gfx.BossScrapemaster:
		b.HP = 3
	case gfx.BossBaron:
		b.HP = 4
	case gfx.BossLeviathan:
		b.HP = 4
	case gfx.BossLunar:
		b.HP = 5
	default:
		b.HP = 6
	}
	b.MaxHP = b.HP
	return b
}

// SetArena confines the boss to the room it fights in.
func (b *Boss) SetArena(left, right, floor float64) {
	b.arenaL, b.arenaR, b.floorY = left, right, floor
	b.Y = floor - b.H
	b.home[1] = b.Y
}

func (b *Boss) Vulnerable() bool {
	return b.state == bsExposed || b.state == bsIdle || b.state == bsAct
}

// Stompable is true only in the exposed window, so the fight has rhythm.
func (b *Boss) Stompable() bool { return b.state == bsExposed }

func (b *Boss) Update(w *World) {
	b.animT += 0.06
	if b.hurtT > 0 {
		b.hurtT--
	}

	switch b.state {
	case bsIntro:
		b.timer--
		if b.timer <= 0 {
			b.state = bsIdle
			b.timer = 60
			b.Started = true
		}
		return
	case bsDying:
		b.deadT++
		b.VY += 0.18
		b.Y += b.VY
		b.X += math.Sin(float64(b.deadT)*0.3) * 1.6
		if b.deadT%9 == 0 {
			w.Particles.Burst(b.CX()+(w.rng.Float64()-0.5)*b.W, b.CY()+(w.rng.Float64()-0.5)*b.H, 12, gfx.ColGold, 2.4)
			w.sfxBossHit()
		}
		if b.deadT > 110 {
			b.state = bsDone
			b.Defeated = true
			w.onBossDefeated()
		}
		return
	case bsDone:
		return
	}

	b.think(w)
}

func (b *Boss) think(w *World) {
	p := w.Player
	b.timer--

	// The fight speeds up as the boss loses health.
	rage := 1.0 + float64(b.MaxHP-b.HP)*0.28

	switch b.Kind {

	// ---- W1: Scrapemaster Prime — paces, then leaps and slams ----------
	case gfx.BossScrapemaster:
		switch b.state {
		case bsIdle:
			b.VX = float64(b.Facing) * 0.9 * rage
			b.MoveX(w.Level, b.VX)
			b.clampArena()
			if b.HitLeft || b.HitRight || b.X <= b.arenaL || b.Right() >= b.arenaR {
				b.Facing = -b.Facing
			}
			if b.timer <= 0 {
				b.state = bsAct
				b.timer = 0
				b.VY = -6.5
			}
		case bsAct:
			b.VY += 0.34
			b.Y += b.VY
			b.X += sign(p.CX()-b.CX()) * 1.4
			b.clampArena()
			if b.Y >= b.floorY-b.H {
				b.Y = b.floorY - b.H
				b.VY = 0
				w.shake(9)
				w.sfxBossHit()
				// slam sends shockwaves both ways along the floor
				for _, d := range []float64{-1, 1} {
					s := newShot(b.CX(), b.floorY-8, d*2.4, 0, shotWave)
					w.Shots = append(w.Shots, s)
				}
				b.state = bsExposed
				b.timer = int(90 / rage)
			}
		case bsExposed:
			b.VX = 0
			if b.timer <= 0 {
				b.state = bsIdle
				b.timer = int(140 / rage)
			}
		}

	// ---- W2: The Scraper Baron — charges across the room ---------------
	case gfx.BossBaron:
		switch b.state {
		case bsIdle:
			b.VX = 0
			if b.timer <= 0 {
				b.Facing = int(sign(p.CX() - b.CX()))
				if b.Facing == 0 {
					b.Facing = -1
				}
				b.state = bsAct
				b.timer = 150
				b.charge = 0
			}
		case bsAct:
			b.charge = math.Min(1, b.charge+0.03)
			b.VX = float64(b.Facing) * 3.6 * b.charge * rage
			b.MoveX(w.Level, b.VX)
			b.clampArena()
			if b.X <= b.arenaL+1 || b.Right() >= b.arenaR-1 || b.timer <= 0 {
				w.shake(7)
				w.sfxBump()
				b.state = bsExposed
				b.timer = int(110 / rage)
				// the impact shakes mines loose from the ceiling
				for i := 0; i < 3; i++ {
					x := b.arenaL + (b.arenaR-b.arenaL)*w.rng.Float64()
					w.Shots = append(w.Shots, newShot(x, b.floorY-190, 0, 1.0, shotMeteor))
				}
			}
		case bsExposed:
			b.VX = 0
			if b.timer <= 0 {
				b.state = bsIdle
				b.timer = int(70 / rage)
			}
		}

	// ---- W3: The Leviathan — dives and surfaces, spitting arcs ---------
	case gfx.BossLeviathan:
		switch b.state {
		case bsIdle:
			// hovers, tracking Dario
			b.Y = b.home[1] - 30 + math.Sin(b.animT*1.4)*10
			b.X += sign(p.CX()-b.CX()) * 0.55 * rage
			b.clampArena()
			if b.timer <= 0 {
				b.state = bsAct
				b.timer = 70
			}
		case bsAct:
			b.Y = b.home[1] - 30 + math.Sin(b.animT*1.4)*10
			if b.timer%18 == 0 {
				dir := sign(p.CX() - b.CX())
				if dir == 0 {
					dir = -1
				}
				w.Shots = append(w.Shots, newShot(b.CX()+dir*20, b.CY(), dir*2.6, -2.2, shotOrb))
				w.sfxShoot()
			}
			if b.timer <= 0 {
				// drops to the floor, exposed
				b.state = bsExposed
				b.timer = int(100 / rage)
			}
		case bsExposed:
			if b.Y < b.floorY-b.H {
				b.Y += 3
			} else {
				b.Y = b.floorY - b.H
			}
			if b.timer <= 0 {
				b.state = bsIdle
				b.timer = int(120 / rage)
			}
		}

	// ---- W4: The Lunar Lord — blinks about and rains meteors ----------
	case gfx.BossLunar:
		switch b.state {
		case bsIdle:
			b.Y = b.home[1] - 60 + math.Sin(b.animT*1.1)*16
			b.X += math.Cos(b.animT*0.8) * 1.2
			b.clampArena()
			if b.timer <= 0 {
				b.state = bsAct
				b.timer = 90
			}
		case bsAct:
			b.Y = b.home[1] - 60 + math.Sin(b.animT*1.1)*16
			if b.timer%22 == 0 {
				x := b.arenaL + 20 + (b.arenaR-b.arenaL-40)*w.rng.Float64()
				w.Shots = append(w.Shots, newShot(x, b.floorY-200, 0, 0.8, shotMeteor))
			}
			if b.timer <= 0 {
				// teleports down to gloat — the only chance to hit it
				b.X = p.CX() - b.W/2 + (w.rng.Float64()-0.5)*40
				b.clampArena()
				b.state = bsExposed
				b.timer = int(90 / rage)
				w.Particles.Ring(b.CX(), b.CY(), gfx.ColPurple, 8)
			}
		case bsExposed:
			if b.Y < b.floorY-b.H {
				b.Y += 2.6
			} else {
				b.Y = b.floorY - b.H
			}
			if b.timer <= 0 {
				b.state = bsIdle
				b.timer = int(110 / rage)
				w.Particles.Ring(b.CX(), b.CY(), gfx.ColPurple, 8)
			}
		}

	// ---- W5: The Cloudlord / The Distiller — three phases -------------
	default:
		// swaps its art to The Distiller for the final third
		if b.HP <= b.MaxHP/3 && b.Phase < 2 {
			b.Phase = 2
			w.onBossPhase(b)
			w.shake(16)
		} else if b.HP <= b.MaxHP*2/3 && b.Phase < 1 {
			b.Phase = 1
			w.onBossPhase(b)
		}
		switch b.state {
		case bsIdle:
			b.Y = b.home[1] - 48 + math.Sin(b.animT*1.2)*12
			b.X += sign(p.CX()-b.CX()) * (0.5 + 0.3*float64(b.Phase))
			b.clampArena()
			if b.timer <= 0 {
				b.state = bsAct
				b.timer = 100
			}
		case bsAct:
			b.Y = b.home[1] - 48 + math.Sin(b.animT*1.2)*12
			interval := 26 - b.Phase*6
			if b.timer%interval == 0 {
				// lightning falls where Dario is standing
				w.Shots = append(w.Shots, newShot(p.CX(), b.floorY-200, 0, 1.4, shotBolt))
				w.sfxShoot()
			}
			if b.Phase >= 1 && b.timer%40 == 0 {
				dir := sign(p.CX() - b.CX())
				if dir == 0 {
					dir = -1
				}
				w.Shots = append(w.Shots, newShot(b.CX(), b.CY(), dir*2.2, -1.0, shotOrb))
			}
			if b.timer <= 0 {
				b.state = bsExposed
				b.timer = int(96 / rage)
			}
		case bsExposed:
			if b.Y < b.floorY-b.H {
				b.Y += 2.4
			} else {
				b.Y = b.floorY - b.H
			}
			if b.timer <= 0 {
				b.state = bsIdle
				b.timer = int(100 / rage)
			}
		}
	}

	if b.timer < 0 {
		b.timer = 0
	}
}

func (b *Boss) clampArena() {
	if b.arenaR <= b.arenaL {
		return
	}
	if b.X < b.arenaL {
		b.X = b.arenaL
	}
	if b.Right() > b.arenaR {
		b.X = b.arenaR - b.W
	}
}

// Damage applies a hit. Returns true if it landed.
func (b *Boss) Damage(w *World, n int) bool {
	if b.state == bsDying || b.state == bsDone || b.hurtT > 0 {
		return false
	}
	b.HP -= n
	b.hurtT = 30
	w.sfxBossHit()
	w.Particles.Burst(b.CX(), b.CY(), 14, gfx.ColCoral, 2.2)
	w.shake(6)
	if b.HP <= 0 {
		b.HP = 0
		b.state = bsDying
		b.deadT = 0
		b.VY = -3
		w.onBossDying()
		return true
	}
	// knock it back into its pacing state
	if b.state == bsExposed {
		b.state = bsIdle
		b.timer = 50
	}
	return true
}

// artKind picks which sprite set to draw — the fortress boss transforms.
func (b *Boss) artKind() int {
	if b.Kind == gfx.BossCloudlord && b.Phase >= 2 {
		return gfx.BossDistiller
	}
	return b.Kind
}

func (b *Boss) Name() string {
	return gfx.BossNames[b.artKind()]
}

func (b *Boss) Draw(dst *ebiten.Image, camX, camY float64, tick int) {
	if b.state == bsDone {
		return
	}
	art := gfx.Bosses[b.artKind()]
	img := art.Frames[(tick/16)%len(art.Frames)]
	sw := float64(img.Bounds().Dx())
	sh := float64(img.Bounds().Dy())
	x := b.CX() - sw/2 - camX
	y := b.Bottom() - sh - camY

	op := gfx.SpriteOpt{FlipX: b.Facing > 0}
	if b.hurtT > 0 && (b.hurtT/3)%2 == 0 {
		op.Tint = color.RGBA{255, 180, 180, 255}
	}
	if b.state == bsExposed {
		// a subtle bob so the vulnerable window is legible
		y += math.Sin(b.animT*6) * 1.5
	}
	if b.state == bsDying {
		op.Alpha = math.Max(0.2, 1-float64(b.deadT)/110)
	}
	gfx.DrawSprite(dst, img, x, y, op)

	// a marker over the head during the vulnerable window
	if b.state == bsExposed && (tick/8)%2 == 0 {
		gfx.FillRectF(dst, b.CX()-camX-3, y-9, 6, 3, gfx.ColGoldLight)
		gfx.FillRectF(dst, b.CX()-camX-1, y-6, 2, 3, gfx.ColGoldLight)
	}
}

// ---------------------------------------------------------------------------
// Boss projectiles
// ---------------------------------------------------------------------------

type shotKind uint8

const (
	shotOrb shotKind = iota
	shotMeteor
	shotWave
	shotBolt
)

type Shot struct {
	Body
	Kind  shotKind
	Alive bool
	life  int
	animT float64
	warn  int // telegraph before a falling attack becomes lethal
}

func newShot(x, y, vx, vy float64, k shotKind) *Shot {
	s := &Shot{Kind: k, Alive: true, life: 400}
	switch k {
	case shotOrb:
		s.W, s.H = 9, 9
	case shotMeteor:
		s.W, s.H = 12, 12
		s.warn = 34
	case shotWave:
		s.W, s.H = 12, 14
		s.life = 200
	case shotBolt:
		s.W, s.H = 8, 20
		s.warn = 40
	}
	s.X, s.Y = x-s.W/2, y-s.H/2
	s.VX, s.VY = vx, vy
	return s
}

func (s *Shot) Update(w *World) {
	if !s.Alive {
		return
	}
	s.animT += 0.3
	s.life--
	if s.life <= 0 {
		s.Alive = false
		return
	}
	if s.warn > 0 {
		s.warn--
		return // still telegraphing; harmless and stationary
	}
	l := w.Level
	switch s.Kind {
	case shotOrb:
		s.VY += 0.10
		s.X += s.VX
		s.Y += s.VY
	case shotMeteor, shotBolt:
		s.VY += 0.22
		if s.VY > 6 {
			s.VY = 6
		}
		s.Y += s.VY
		s.X += s.VX
		if s.Y > s.floorHit(l) {
			s.Alive = false
			w.Particles.Burst(s.CX(), s.Bottom(), 10, gfx.ColGold, 2.0)
		}
	case shotWave:
		s.X += s.VX
		// hug the floor
		s.MoveY(l, 2, false)
	}
	if s.X < -40 || s.X > l.PixelWidth()+40 || s.Y > l.PixelHeight()+40 {
		s.Alive = false
	}
}

func (s *Shot) floorHit(l *Level) float64 {
	tx := ifloor(s.CX() / TS)
	for ty := ifloor(s.Bottom() / TS); ty < l.H; ty++ {
		if l.At(tx, ty).Solid() {
			return float64(ty*TS) - s.H
		}
	}
	return l.PixelHeight() + 100
}

// Dangerous reports whether the shot can currently hurt Dario.
func (s *Shot) Dangerous() bool { return s.Alive && s.warn == 0 }

func (s *Shot) Draw(dst *ebiten.Image, camX, camY float64, tick int) {
	if !s.Alive {
		return
	}
	x, y := s.X-camX, s.Y-camY
	if s.warn > 0 {
		// telegraph: a blinking marker where it is about to land
		if (s.warn/4)%2 == 0 {
			gfx.StrokeRectF(dst, x-3, y-3, s.W+6, s.H+6, gfx.ColGoldLight)
			gfx.FillRectF(dst, x, y, s.W, s.H, gfx.Alpha(gfx.ColRed, 0.35))
		}
		return
	}
	switch s.Kind {
	case shotOrb:
		gfx.FillRectF(dst, x-1, y-1, s.W+2, s.H+2, gfx.Alpha(gfx.ColCyan, 0.45))
		gfx.FillRectF(dst, x, y, s.W, s.H, gfx.ColCyan)
		gfx.FillRectF(dst, x+2, y+2, 3, 3, gfx.ColCream)
	case shotMeteor:
		gfx.FillRectF(dst, x-2, y-6, s.W+4, 6, gfx.Alpha(gfx.ColCoral, 0.4))
		gfx.FillRectF(dst, x, y, s.W, s.H, gfx.ColCoral)
		gfx.FillRectF(dst, x+2, y+2, 4, 4, gfx.ColGoldLight)
	case shotWave:
		h := s.H * (0.7 + 0.3*math.Sin(s.animT*2))
		gfx.FillRectF(dst, x, y+s.H-h, s.W, h, gfx.Alpha(gfx.ColGoldLight, 0.85))
		gfx.FillRectF(dst, x+2, y+s.H-h+2, s.W-4, h-4, gfx.ColGold)
	case shotBolt:
		gfx.FillRectF(dst, x+1, y, s.W-2, s.H, gfx.ColGoldLight)
		gfx.FillRectF(dst, x+3, y, 2, s.H, gfx.ColCream)
	}
}
