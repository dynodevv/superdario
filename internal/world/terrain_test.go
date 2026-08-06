package world

import (
	"testing"

	"superdario/internal/core"
)

// An underwater stage must be water everywhere below its surface line.
//
// Water levels are assembled partly from shared chunks — "checkpoint" is drawn
// as open air — and dropping one into an ocean used to leave a full-height dry
// shaft the player could stand around in, unable to swim.
func TestUnderwaterLevelsHaveNoAirPockets(t *testing.T) {
	for w := range Levels {
		for l := range Levels[w] {
			def := Levels[w][l]
			if !def.Underwater {
				continue
			}
			lv, err := BuildLevel(def)
			if err != nil {
				t.Fatal(err)
			}
			// locate the surface
			surface := -1
			for ty := 0; ty < lv.H && surface < 0; ty++ {
				for tx := 0; tx < lv.W; tx++ {
					if lv.RawAt(tx, ty).Liquid() {
						surface = ty
						break
					}
				}
			}
			if surface < 0 {
				t.Fatalf("%d-%d (%s) is marked underwater but has no water", w+1, l+1, def.Name)
			}
			holes := 0
			var firstX, firstY int
			for ty := surface; ty < lv.H; ty++ {
				for tx := 0; tx < lv.W; tx++ {
					if lv.RawAt(tx, ty) == TEmpty {
						if holes == 0 {
							firstX, firstY = tx, ty
						}
						holes++
					}
				}
			}
			if holes > 0 {
				t.Errorf("%d-%d (%s): %d air pockets below the surface, first at tile (%d,%d)",
					w+1, l+1, def.Name, holes, firstX, firstY)
			}
		}
	}
}

// Placing an entity must not punch a hole in the surrounding terrain. A spawn
// character leaves its cell empty, which is correct in air and wrong in water.
func TestSpawnsDoNotPunchHolesInWater(t *testing.T) {
	for w := range Levels {
		for l := range Levels[w] {
			def := Levels[w][l]
			lv, err := BuildLevel(def)
			if err != nil {
				t.Fatal(err)
			}
			for ty := 1; ty < lv.H-1; ty++ {
				for tx := 1; tx < lv.W-1; tx++ {
					if lv.RawAt(tx, ty) != TEmpty {
						continue
					}
					// an empty cell with water on both sides is a pocket
					if lv.RawAt(tx-1, ty).Liquid() && lv.RawAt(tx+1, ty).Liquid() {
						t.Errorf("%d-%d (%s): air pocket at tile (%d,%d) with water either side",
							w+1, l+1, def.Name, tx, ty)
						return
					}
				}
			}
		}
	}
}

// A token tucked under a platform, with a pit beneath it, cannot be had
// without jumping into the pit — the platform caps the jump arc from the side
// and there is no floor to launch from underneath.
//
// Coins strung over a gap with open sky above them are fine and deliberate;
// you collect those on the way across. It is specifically the roofed-over
// ones that are a trap.
func TestNoTokensTrappedUnderPlatforms(t *testing.T) {
	reach := maxJumpTiles() + 1
	for w := range Levels {
		for l := range Levels[w] {
			def := Levels[w][l]
			wd, err := NewWorld(def, 3, 0, 0, core.PowerSmall)
			if err != nil {
				t.Fatal(err)
			}
			lv := wd.Level
			for _, it := range wd.Items {
				if it.Kind != ItToken {
					continue
				}
				tx := ifloor(it.CX() / TS)
				ty := ifloor(it.CY() / TS)
				if def.Underwater || lv.At(tx, ty).Liquid() {
					continue // swimmable from any direction
				}

				roofed := false
				for d := 1; d <= 2 && !roofed; d++ {
					if a := lv.At(tx, ty-d); a.OneWay() || a.Solid() {
						roofed = true
					}
				}
				if !roofed {
					continue
				}
				footing := false
				for d := 1; d <= reach && !footing; d++ {
					if b := lv.At(tx, ty+d); b.Solid() || b.OneWay() {
						footing = true
					}
				}
				if !footing {
					t.Errorf("%d-%d (%s): token at tile (%d,%d) sits under a platform with a pit below; "+
						"it cannot be collected without falling in",
						w+1, l+1, def.Name, tx, ty)
				}
			}
		}
	}
}

// Swimming has to actually climb. Holding the swim button from the seabed must
// carry Dario a useful distance upward — a stroke that lifts less than a tile
// makes the whole world feel broken.
func TestSwimmingClimbs(t *testing.T) {
	for w := range Levels {
		for l := range Levels[w] {
			def := Levels[w][l]
			if !def.Underwater {
				continue
			}
			wd, err := NewWorld(def, 3, 0, 0, core.PowerSmall)
			if err != nil {
				t.Fatal(err)
			}
			p := wd.Player
			// settle onto the bottom first
			for i := 0; i < 120; i++ {
				wd.Update(Input{})
			}
			startY := p.Y

			// two seconds of holding the swim button
			for i := 0; i < 120; i++ {
				wd.Update(Input{Jump: i == 0, JumpDown: true})
			}
			rise := startY - p.Y
			if rise < 4*TS {
				t.Errorf("%d-%d (%s): two seconds of swimming only climbed %.0fpx (%.1f tiles); want at least 4 tiles",
					w+1, l+1, def.Name, rise, rise/TS)
			}
		}
	}
}

// A single stroke should be worth making on its own.
func TestOneSwimStrokeClearsATile(t *testing.T) {
	wd, err := NewWorld(Levels[2][0], 3, 0, 0, core.PowerSmall)
	if err != nil {
		t.Fatal(err)
	}
	p := wd.Player
	for i := 0; i < 120; i++ {
		wd.Update(Input{})
	}
	startY := p.Y
	best := startY
	// one tap, then let it play out
	for i := 0; i < 40; i++ {
		wd.Update(Input{Jump: i == 0})
		if p.Y < best {
			best = p.Y
		}
	}
	if rise := startY - best; rise < TS {
		t.Errorf("a single stroke lifted %.0fpx, less than one %dpx tile", rise, TS)
	}
}
