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

// Enemies must not end up buried in the terrain. Flying enemies used to set
// their hover height by assigning Y outright, bypassing collision entirely, so
// one spawned near the floor drove its whole sine amplitude into the seabed
// and sat half-submerged in the rock.
func TestEnemiesStayOutOfTerrain(t *testing.T) {
	for w := range Levels {
		for l := range Levels[w] {
			def := Levels[w][l]
			wd, err := NewWorld(def, 99, 0, 0, core.PowerScaled)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 1800; i++ {
				wd.Update(botInput(i))
				if wd.Status == StatusDead || wd.Status == StatusTimeUp {
					wd.Respawn()
					continue
				}
				for _, e := range wd.Enemies {
					// Plants live inside their pipe by design; anything
					// dying or knocked off the screen is on its way out.
					if !e.Alive || !e.active || e.Kind == EnPlant ||
						e.flipping || e.squashT > 0 {
						continue
					}
					tx := ifloor(e.CX() / TS)
					ty := ifloor(e.CY() / TS)
					if wd.Level.At(tx, ty).Solid() {
						t.Fatalf("%d-%d (%s) tick %d: %s is inside a solid tile at (%d,%d)",
							w+1, l+1, def.Name, i, enemyName(e.Kind), tx, ty)
					}
				}
			}
		}
	}
}

func enemyName(k EnemyKind) string {
	switch k {
	case EnDistill:
		return "DistillBot"
	case EnCrawler:
		return "Crawler"
	case EnShell:
		return "Shell"
	case EnSpikeBot:
		return "SpikeBot"
	case EnFlyer:
		return "Flyer"
	case EnLeech:
		return "Leech"
	case EnPlant:
		return "DataPipe"
	}
	return "enemy"
}

// Falling platforms must come back when Dario does. They drop away once
// stepped on and are reaped when they leave the screen, so without a reset a
// death after triggering them leaves the gap they spanned uncrossable —
// castle_lavarun is three of them over lava, which made the stage unwinnable
// until you quit out to the map.
func TestRespawnRestoresPlatforms(t *testing.T) {
	for w := range Levels {
		for l := range Levels[w] {
			def := Levels[w][l]
			wd, err := NewWorld(def, 9, 0, 0, core.PowerScaled)
			if err != nil {
				t.Fatal(err)
			}
			want := len(wd.Platforms)
			if want == 0 {
				continue
			}

			// trigger every falling platform and let them all drop away
			for _, p := range wd.Platforms {
				p.Trigger()
			}
			for i := 0; i < 900; i++ {
				wd.Update(Input{})
				if wd.Status == StatusDead || wd.Status == StatusTimeUp {
					break
				}
			}

			wd.Respawn()
			if got := len(wd.Platforms); got != want {
				t.Errorf("%d-%d (%s): %d platforms after respawn, want %d",
					w+1, l+1, def.Name, got, want)
				continue
			}
			for _, p := range wd.Platforms {
				if !p.Alive {
					t.Errorf("%d-%d (%s): platform restored dead", w+1, l+1, def.Name)
					break
				}
				if p.Kind == PlatFall && (p.fallT != 0 || p.VY != 0) {
					t.Errorf("%d-%d (%s): falling platform restored mid-drop (fallT=%d vy=%.2f)",
						w+1, l+1, def.Name, p.fallT, p.VY)
					break
				}
			}
		}
	}
}

// A stage that depends on platforms to cross a hazard must still have them
// after a death, or the retry is unwinnable.
func TestNoStageBecomesUnwinnableAfterDeath(t *testing.T) {
	for w := range Levels {
		for l := range Levels[w] {
			def := Levels[w][l]
			wd, err := NewWorld(def, 9, 0, 0, core.PowerScaled)
			if err != nil {
				t.Fatal(err)
			}
			before := len(wd.Platforms)

			// play for a while, dying and respawning as it goes
			for i := 0; i < 3000; i++ {
				wd.Update(botInput(i))
				if wd.Status == StatusDead || wd.Status == StatusTimeUp {
					wd.Respawn()
				}
				if wd.Status == StatusCleared {
					break
				}
			}
			wd.Respawn()
			if got := len(wd.Platforms); got != before {
				t.Errorf("%d-%d (%s): started with %d platforms, has %d after a run with deaths",
					w+1, l+1, def.Name, before, got)
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
