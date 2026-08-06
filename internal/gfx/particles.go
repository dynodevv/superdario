package gfx

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// A small pooled particle system. Everything from landing dust to boss
// explosions runs through it.

type PKind uint8

const (
	PSquare PKind = iota
	PSpark        // leaves a short trail
	PDebris       // spins
	PRing         // expanding hollow ring
	PText         // floating score popup
)

type Particle struct {
	X, Y, VX, VY   float64
	Life, MaxLife  float64
	Size           float64
	Grav           float64
	Col            color.RGBA
	Kind           PKind
	Rot, Spin      float64
	Text           string
	Fade           bool
	Shrink         bool
	active         bool
	prevX, prevY   float64
	scaleWithWorld bool
}

type System struct {
	pool []Particle
	rng  *rand.Rand
}

func NewSystem(capacity int) *System {
	return &System{pool: make([]Particle, capacity), rng: rand.New(rand.NewSource(0x5EED))}
}

func (s *System) Clear() {
	for i := range s.pool {
		s.pool[i].active = false
	}
}

func (s *System) alloc() *Particle {
	for i := range s.pool {
		if !s.pool[i].active {
			p := &s.pool[i]
			*p = Particle{active: true}
			return p
		}
	}
	// pool exhausted: recycle the oldest slot
	p := &s.pool[s.rng.Intn(len(s.pool))]
	*p = Particle{active: true}
	return p
}

func (s *System) Add(p Particle) {
	q := s.alloc()
	p.active = true
	if p.MaxLife == 0 {
		p.MaxLife = p.Life
	}
	if p.Size == 0 {
		p.Size = 2
	}
	p.prevX, p.prevY = p.X, p.Y
	*q = p
}

func (s *System) Update() {
	for i := range s.pool {
		p := &s.pool[i]
		if !p.active {
			continue
		}
		p.prevX, p.prevY = p.X, p.Y
		p.VY += p.Grav
		p.X += p.VX
		p.Y += p.VY
		p.Rot += p.Spin
		p.Life--
		if p.Life <= 0 {
			p.active = false
		}
	}
}

func (s *System) Draw(dst *ebiten.Image, camX, camY float64) {
	for i := range s.pool {
		p := &s.pool[i]
		if !p.active {
			continue
		}
		f := p.Life / p.MaxLife
		col := p.Col
		if p.Fade {
			col = Alpha(col, float64(col.A)/255*math.Min(1, f*1.6))
		}
		sz := p.Size
		if p.Shrink {
			sz = math.Max(1, p.Size*f)
		}
		x, y := p.X-camX, p.Y-camY
		switch p.Kind {
		case PSpark:
			LineF(dst, p.prevX-camX, p.prevY-camY, x, y, col)
			FillRectF(dst, x, y, sz, sz, col)
		case PRing:
			r := p.Size * (1 - f) * 3
			steps := 16
			for k := 0; k < steps; k++ {
				a := float64(k) / float64(steps) * 6.283
				FillRectF(dst, x+r*math.Cos(a), y+r*math.Sin(a), 2, 2, col)
			}
		case PDebris:
			// crude rotation: swap the axis every quarter turn
			q := int(math.Abs(p.Rot/1.5707)) % 2
			if q == 0 {
				FillRectF(dst, x, y, sz, sz*0.6, col)
			} else {
				FillRectF(dst, x, y, sz*0.6, sz, col)
			}
		case PText:
			TextShadow(dst, p.Text, int(x), int(y), col, Alpha(ColInkDeep, float64(col.A)/255*0.8), 1)
		default:
			FillRectF(dst, x, y, sz, sz, col)
		}
	}
}

// --- convenience emitters ---------------------------------------------------

func (s *System) Dust(x, y float64, dir float64) {
	for i := 0; i < 5; i++ {
		s.Add(Particle{
			X: x, Y: y,
			VX:   dir*(0.2+s.rng.Float64()*0.6) + (s.rng.Float64()-0.5)*0.4,
			VY:   -s.rng.Float64() * 0.7,
			Grav: 0.03,
			Life: 14 + s.rng.Float64()*10,
			Size: 1 + s.rng.Float64()*2,
			Col:  Alpha(ColCream, 0.75),
			Fade: true, Shrink: true,
		})
	}
}

func (s *System) Burst(x, y float64, n int, col color.RGBA, speed float64) {
	for i := 0; i < n; i++ {
		a := s.rng.Float64() * 6.283
		v := speed * (0.4 + s.rng.Float64()*0.9)
		s.Add(Particle{
			X: x, Y: y,
			VX: math.Cos(a) * v, VY: math.Sin(a)*v - 0.4,
			Grav: 0.10,
			Life: 20 + s.rng.Float64()*18,
			Size: 1 + s.rng.Float64()*2.4,
			Col:  col, Kind: PSpark,
			Fade: true, Shrink: true,
		})
	}
}

func (s *System) Debris(x, y float64, col color.RGBA) {
	for i := 0; i < 8; i++ {
		s.Add(Particle{
			X: x, Y: y,
			VX:   (s.rng.Float64() - 0.5) * 3.2,
			VY:   -1.4 - s.rng.Float64()*2.4,
			Grav: 0.22,
			Life: 40 + s.rng.Float64()*20,
			Size: 2 + s.rng.Float64()*2,
			Col:  col, Kind: PDebris,
			Spin: (s.rng.Float64() - 0.5) * 0.6,
		})
	}
}

func (s *System) Sparkle(x, y float64, col color.RGBA) {
	s.Add(Particle{
		X: x + (s.rng.Float64()-0.5)*8, Y: y + (s.rng.Float64()-0.5)*8,
		VX: (s.rng.Float64() - 0.5) * 0.4, VY: -0.3 - s.rng.Float64()*0.5,
		Life: 18 + s.rng.Float64()*12,
		Size: 1 + s.rng.Float64()*1.6,
		Col:  col, Fade: true, Shrink: true,
	})
}

func (s *System) Ring(x, y float64, col color.RGBA, size float64) {
	s.Add(Particle{X: x, Y: y, Life: 22, Size: size, Col: col, Kind: PRing, Fade: true})
}

func (s *System) Popup(x, y float64, text string, col color.RGBA) {
	s.Add(Particle{
		X: x - float64(TextWidth(text, 1))/2, Y: y,
		VY: -0.55, Life: 52, Col: col, Kind: PText, Text: text, Fade: true,
	})
}

func (s *System) Splash(x, y float64, col color.RGBA) {
	for i := 0; i < 10; i++ {
		s.Add(Particle{
			X: x, Y: y,
			VX:   (s.rng.Float64() - 0.5) * 2.4,
			VY:   -1.0 - s.rng.Float64()*1.8,
			Grav: 0.16,
			Life: 22 + s.rng.Float64()*14,
			Size: 1 + s.rng.Float64()*2,
			Col:  col, Fade: true, Shrink: true,
		})
	}
}
