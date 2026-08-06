package world

import (
	"testing"

	"superdario/internal/core"
)

// Every chunk must be rectangular, fit inside the level height, and use only
// characters the builder understands. A typo here would otherwise show up as
// an invisible hole in a level.
func TestChunksWellFormed(t *testing.T) {
	for name, c := range chunks {
		if err := validateChunk(name, c); err != nil {
			t.Errorf("%v", err)
			continue
		}
		for ry, row := range c.Rows {
			for rx := 0; rx < len(row); rx++ {
				ch := row[rx]
				if isSpawn(ch) {
					continue
				}
				if _, ok := runeToTile[ch]; !ok {
					t.Errorf("chunk %q row %d col %d: unknown character %q", name, ry, rx, string(ch))
				}
			}
		}
	}
}

// Pipes are drawn from four quadrant tiles; a mismatched pair looks broken.
func TestPipesArePaired(t *testing.T) {
	for name, c := range chunks {
		for ry, row := range c.Rows {
			for rx := 0; rx < len(row); rx++ {
				switch row[rx] {
				case '[':
					if rx+1 >= len(row) || row[rx+1] != ']' {
						t.Errorf("chunk %q row %d: '[' at %d has no matching ']'", name, ry, rx)
					}
				case '{':
					if rx+1 >= len(row) || row[rx+1] != '}' {
						t.Errorf("chunk %q row %d: '{' at %d has no matching '}'", name, ry, rx)
					}
				}
			}
		}
	}
}

// Data-pipe plants only make sense sitting on a pipe mouth.
func TestPlantsSitOnPipes(t *testing.T) {
	for name, c := range chunks {
		for ry, row := range c.Rows {
			for rx := 0; rx < len(row); rx++ {
				if row[rx] != SpawnPlant {
					continue
				}
				if ry+1 >= len(c.Rows) {
					t.Errorf("chunk %q: plant at row %d has nothing below it", name, ry)
					continue
				}
				below := c.Rows[ry+1]
				if rx >= len(below) || below[rx] != '[' {
					t.Errorf("chunk %q: plant at row %d col %d is not above a pipe mouth", name, ry, rx)
				}
			}
		}
	}
}

// Every level must build, be wide enough to play, and have a reachable exit.
func TestAllLevelsBuild(t *testing.T) {
	for w := range Levels {
		for l := range Levels[w] {
			def := Levels[w][l]
			lv, err := BuildLevel(def)
			if err != nil {
				t.Errorf("level %d-%d (%s): %v", w+1, l+1, def.Name, err)
				continue
			}
			if lv.W < 60 {
				t.Errorf("level %d-%d is only %d tiles wide; that is too short", w+1, l+1, lv.W)
			}
			if lv.H != core.LevelH {
				t.Errorf("level %d-%d height = %d, want %d", w+1, l+1, lv.H, core.LevelH)
			}
			if lv.StartX == 0 && lv.StartY == 0 {
				t.Errorf("level %d-%d has no player spawn", w+1, l+1)
			}
			hasExit := lv.HasGoal || def.Boss >= 0
			if !hasExit {
				t.Errorf("level %d-%d has neither a goal nor a boss", w+1, l+1)
			}
			if def.Boss >= 0 && !lv.HasBossGate {
				t.Errorf("level %d-%d has a boss but no boss gate marker", w+1, l+1)
			}
			if def.Time < 200 {
				t.Errorf("level %d-%d has only %ds on the clock", w+1, l+1, def.Time)
			}
		}
	}
}

// Dario must actually be able to stand where he spawns.
func TestSpawnHasGroundBeneath(t *testing.T) {
	for w := range Levels {
		for l := range Levels[w] {
			def := Levels[w][l]
			lv, err := BuildLevel(def)
			if err != nil {
				continue
			}
			tx := int(lv.StartX) / TS
			found := false
			for ty := int(lv.StartY) / TS; ty < lv.H; ty++ {
				if lv.At(tx, ty).Solid() {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("level %d-%d (%s): nothing solid below the spawn point", w+1, l+1, def.Name)
			}
		}
	}
}

// Boss levels put the arena last so the camera lock lines up with it.
func TestBossArenaIsLastChunk(t *testing.T) {
	for w := range Levels {
		for l := range Levels[w] {
			def := Levels[w][l]
			if def.Boss < 0 {
				continue
			}
			last := def.Chunks[len(def.Chunks)-1]
			if last != "boss_arena" && last != "boss_arena_lava" {
				t.Errorf("level %d-%d ends with %q, expected a boss arena", w+1, l+1, last)
			}
			if chunks[last].Width() != 28 {
				t.Errorf("boss arena %q is %d tiles wide, the camera lock expects 28", last, chunks[last].Width())
			}
		}
	}
}

// The four Claude models must each be rescued exactly once across the game.
func TestEachModelRescuedOnce(t *testing.T) {
	seen := map[int]int{}
	for w := range Levels {
		for l := range Levels[w] {
			if m := Levels[w][l].Model; m >= 0 {
				seen[m]++
			}
		}
	}
	for m := 0; m < 4; m++ {
		if seen[m] != 1 {
			t.Errorf("model %d is rescued %d times, want exactly 1", m, seen[m])
		}
	}
}

// A body resting on a solid tile should report OnGround and not sink into it.
func TestBodyLandsOnSolidGround(t *testing.T) {
	lv, err := BuildLevel(Levels[0][0])
	if err != nil {
		t.Fatal(err)
	}
	var b Body
	b.W, b.H = 11, 14
	b.X = lv.StartX
	b.Y = 0
	for i := 0; i < 200; i++ {
		b.ApplyGravity(lv.Def, false)
		b.MoveY(lv, b.VY, true)
		if b.OnGround {
			break
		}
	}
	if !b.OnGround {
		t.Fatal("body never landed")
	}
	ty := ifloor((b.Bottom() + 0.5) / TS)
	if !lv.At(ifloor(b.CX()/TS), ty).Solid() {
		t.Errorf("landed at y=%.1f but there is no solid tile beneath", b.Bottom())
	}
	if b.VY != 0 {
		t.Errorf("resting body still has vertical velocity %.2f", b.VY)
	}
}

// Walls must stop horizontal movement rather than letting a body tunnel.
func TestBodyStopsAtWall(t *testing.T) {
	lv, err := BuildLevel(Levels[0][0])
	if err != nil {
		t.Fatal(err)
	}
	// Build a body just left of the level's right edge and push it outward.
	var b Body
	b.W, b.H = 11, 14
	b.X = lv.PixelWidth() - 40
	b.Y = float64((lv.H - 4) * TS)
	for i := 0; i < 300; i++ {
		b.MoveX(lv, 4)
	}
	if b.Right() > lv.PixelWidth() {
		t.Errorf("body escaped the right edge: right=%.1f, level width=%.1f", b.Right(), lv.PixelWidth())
	}
}
