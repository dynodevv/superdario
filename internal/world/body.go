package world

import (
	"math"

	"superdario/internal/core"
)

// Body is an axis-aligned box with velocity. X,Y is the top-left of the
// *hitbox*, which is usually narrower than the sprite.
type Body struct {
	X, Y   float64
	VX, VY float64
	W, H   float64

	OnGround bool
	HitCeil  bool
	HitLeft  bool
	HitRight bool
	Facing   int

	// set by the last MoveY when the ceiling tile could be bumped
	BumpTX, BumpTY int
	DidBump        bool
}

func ifloor(v float64) int { return int(math.Floor(v)) }

func (b *Body) Right() float64  { return b.X + b.W }
func (b *Body) Bottom() float64 { return b.Y + b.H }
func (b *Body) CX() float64     { return b.X + b.W/2 }
func (b *Body) CY() float64     { return b.Y + b.H/2 }

// Overlaps reports AABB intersection with another body.
func (b *Body) Overlaps(o *Body) bool {
	return b.X < o.Right() && b.Right() > o.X && b.Y < o.Bottom() && b.Bottom() > o.Y
}

// OverlapsRect tests against a loose rectangle.
func (b *Body) OverlapsRect(x, y, w, h float64) bool {
	return b.X < x+w && b.Right() > x && b.Y < y+h && b.Bottom() > y
}

// MoveX advances horizontally and resolves against solid tiles.
func (b *Body) MoveX(l *Level, dx float64) {
	b.HitLeft, b.HitRight = false, false
	if dx == 0 {
		return
	}
	b.X += dx

	ty0 := ifloor(b.Y / TS)
	ty1 := ifloor((b.Bottom() - 0.001) / TS)

	if dx > 0 {
		tx := ifloor((b.Right() - 0.001) / TS)
		for ty := ty0; ty <= ty1; ty++ {
			if l.SolidAt(tx, ty) {
				b.X = float64(tx*TS) - b.W
				b.VX = 0
				b.HitRight = true
				return
			}
		}
	} else {
		tx := ifloor(b.X / TS)
		for ty := ty0; ty <= ty1; ty++ {
			if l.SolidAt(tx, ty) {
				b.X = float64((tx + 1) * TS)
				b.VX = 0
				b.HitLeft = true
				return
			}
		}
	}
}

// MoveY advances vertically. One-way platforms only stop a downward move that
// began above the platform's surface.
func (b *Body) MoveY(l *Level, dy float64, useOneWay bool) {
	b.HitCeil = false
	b.DidBump = false
	wasBottom := b.Bottom()
	b.OnGround = false
	if dy == 0 {
		// still need to know whether we are resting on something
		b.checkGround(l, useOneWay)
		return
	}
	b.Y += dy

	tx0 := ifloor(b.X / TS)
	tx1 := ifloor((b.Right() - 0.001) / TS)

	if dy > 0 { // falling
		ty := ifloor((b.Bottom() - 0.001) / TS)
		for tx := tx0; tx <= tx1; tx++ {
			t := l.At(tx, ty)
			hit := t.Solid()
			if !hit && useOneWay && t.OneWay() {
				top := float64(ty * TS)
				// only land if we were fully above the platform last frame
				if wasBottom <= top+1 {
					hit = true
				}
			}
			if hit {
				b.Y = float64(ty*TS) - b.H
				b.VY = 0
				b.OnGround = true
				return
			}
		}
	} else { // rising
		ty := ifloor(b.Y / TS)
		for tx := tx0; tx <= tx1; tx++ {
			if l.At(tx, ty).Solid() {
				b.Y = float64((ty + 1) * TS)
				b.VY = 0
				b.HitCeil = true
				b.BumpTX, b.BumpTY = tx, ty
				b.DidBump = true
				return
			}
		}
	}
	b.checkGround(l, useOneWay)
}

// checkGround sets OnGround when resting exactly on a surface.
func (b *Body) checkGround(l *Level, useOneWay bool) {
	if b.VY < 0 {
		return
	}
	ty := ifloor((b.Bottom() + 0.5) / TS)
	tx0 := ifloor(b.X / TS)
	tx1 := ifloor((b.Right() - 0.001) / TS)
	for tx := tx0; tx <= tx1; tx++ {
		t := l.At(tx, ty)
		if t.Solid() {
			b.OnGround = true
			return
		}
		if useOneWay && t.OneWay() && b.Bottom() <= float64(ty*TS)+1 {
			b.OnGround = true
			return
		}
	}
}

// ApplyGravity integrates vertical velocity for the level's physics rules.
func (b *Body) ApplyGravity(def LevelDef, inWater bool) {
	g := core.Gravity
	maxFall := core.MaxFall
	switch {
	case inWater:
		g, maxFall = core.WaterGravity, core.WaterMaxFall
	case def.LowGrav:
		g = core.GravityLow
	}
	b.VY += g
	if b.VY > maxFall {
		b.VY = maxFall
	}
}

// InLiquid reports whether the body's centre is submerged.
func (b *Body) InLiquid(l *Level) bool {
	return l.TileAtPx(b.CX(), b.CY()).Liquid()
}

// FeetInLiquid is used for splash effects at the surface.
func (b *Body) FeetInLiquid(l *Level) bool {
	return l.TileAtPx(b.CX(), b.Bottom()-2).Liquid()
}

// TouchingHazard reports contact with spikes.
func (b *Body) TouchingHazard(l *Level) bool {
	tx0 := ifloor((b.X + 2) / TS)
	tx1 := ifloor((b.Right() - 2.001) / TS)
	ty0 := ifloor((b.Y + 2) / TS)
	ty1 := ifloor((b.Bottom() - 2.001) / TS)
	for ty := ty0; ty <= ty1; ty++ {
		for tx := tx0; tx <= tx1; tx++ {
			if l.At(tx, ty).Hazard() {
				return true
			}
		}
	}
	return false
}

// OnBelt reports the conveyor direction under the body, or 0.
func (b *Body) OnBelt(l *Level) float64 {
	if !b.OnGround {
		return 0
	}
	ty := ifloor((b.Bottom() + 1) / TS)
	tx0 := ifloor(b.X / TS)
	tx1 := ifloor((b.Right() - 0.001) / TS)
	for tx := tx0; tx <= tx1; tx++ {
		if l.At(tx, ty) == TBelt {
			return 0.55
		}
	}
	return 0
}

// WouldFall is used by cautious enemies to avoid walking off ledges.
func (b *Body) WouldFall(l *Level, dir int) bool {
	probeX := b.X - 2
	if dir > 0 {
		probeX = b.Right() + 2
	}
	tx := ifloor(probeX / TS)
	ty := ifloor((b.Bottom() + 2) / TS)
	t := l.At(tx, ty)
	return !t.Solid() && !t.OneWay()
}
