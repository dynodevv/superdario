package world

import (
	"math"
	"testing"

	"superdario/internal/core"
)

// maxJumpTiles is how high a running jump actually clears.
//
//	v  = JumpVel + JumpRunBonus  (the best case, at full run speed)
//	h  = v^2 / (2*g)
//
// Landing *on* a ledge exactly h tall is frame-perfect, so anything above
// maxJumpTiles is treated as a wall a player cannot be expected to clear.
func maxJumpTiles() int {
	v := math.Abs(core.JumpVel + core.JumpRunBonus)
	h := v * v / (2 * core.Gravity)
	return int(h / TS) // floor: the marginal tile does not count
}

// surfaceRow returns, for each column, the row of the walkable surface — the
// top of the solid stack that reaches the floor. Columns that are open all the
// way down (pits) report -1.
func surfaceRows(l *Level) []int {
	rows := make([]int, l.W)
	for x := 0; x < l.W; x++ {
		rows[x] = -1
		// walk up from the bottom while tiles are solid
		y := l.H - 1
		for y >= 0 && !l.At(x, y).Solid() {
			y--
		}
		if y < 0 {
			continue // nothing solid in this column at all
		}
		// climb to the top of that contiguous stack
		for y > 0 && l.At(x, y-1).Solid() {
			y--
		}
		rows[x] = y
	}
	return rows
}

// hasHelpAbove reports whether there is anything standable in the columns
// around a step-up — a platform, a floating block, a spring — that would let a
// player gain height without a single jump.
func hasHelpAbove(l *Level, x, fromRow, toRow int) bool {
	for dx := -4; dx <= 4; dx++ {
		cx := x + dx
		if cx < 0 || cx >= l.W {
			continue
		}
		for y := toRow; y < fromRow; y++ {
			t := l.At(cx, y)
			if t.Solid() || t.OneWay() || t.Climbable() {
				return true
			}
		}
	}
	return false
}

// No stage may present a step up taller than a running jump with nothing to
// help. This is the check that would have caught the five-tile wall in 1-1.
func TestNoUnjumpableWalls(t *testing.T) {
	limit := maxJumpTiles()
	if limit < 3 {
		t.Fatalf("jump height calculation looks wrong: %d tiles", limit)
	}
	t.Logf("a running jump clears %d tiles", limit)

	for w := range Levels {
		for l := range Levels[w] {
			def := Levels[w][l]
			lv, err := BuildLevel(def)
			if err != nil {
				t.Fatalf("%d-%d: %v", w+1, l+1, err)
			}
			// Spring chunks legitimately launch the player higher; and the
			// player is only ever asked to travel left-to-right.
			rows := surfaceRows(lv)
			for x := 0; x+1 < lv.W; x++ {
				from, to := rows[x], rows[x+1]
				if from < 0 || to < 0 {
					continue // a pit: falling is a different problem
				}
				rise := from - to // positive means the ground steps up
				if rise <= limit {
					continue
				}
				if hasHelpAbove(lv, x+1, from, to) {
					continue
				}
				t.Errorf("%d-%d (%s): column %d steps up %d tiles (row %d -> %d); a running jump only clears %d",
					w+1, l+1, def.Name, x+1, rise, from, to, limit)
			}
		}
	}
}

// The boss arena must be enterable on foot from the approach chunk.
func TestBossArenaIsEnterable(t *testing.T) {
	limit := maxJumpTiles()
	for w := range Levels {
		for l := range Levels[w] {
			def := Levels[w][l]
			if def.Boss < 0 {
				continue
			}
			lv, err := BuildLevel(def)
			if err != nil {
				t.Fatal(err)
			}
			// Walk a body right from the gate and check it actually gets into
			// the arena rather than jamming against a wall.
			wd, err := NewWorld(def, 9, 0, 0, core.PowerScaled)
			if err != nil {
				t.Fatal(err)
			}
			place := func() {
				wd.Player.X = lv.BossGateX - 32
				wd.Player.Y = float64((lv.H-4)*TS) - wd.Player.H
				wd.Player.Invuln = 600 // the boss is not what is under test here
			}
			place()
			startX := wd.Player.X
			maxX := startX
			for i := 0; i < 900; i++ {
				wd.Update(botInput(i))
				if wd.Player.X > maxX {
					maxX = wd.Player.X
				}
				if wd.Status != StatusPlaying {
					wd.Respawn()
					place()
				}
				wd.Player.Invuln = 600
			}
			// the arena occupies the last 28 tiles; require getting well inside
			arenaStart := lv.PixelWidth() - float64(28*TS)
			if maxX < arenaStart+float64(6*TS) {
				t.Errorf("%d-%d (%s): walking right from the gate never got inside the arena (x=%.0f -> %.0f, arena starts at %.0f)",
					w+1, l+1, def.Name, startX, maxX, arenaStart)
			}
			_ = limit
		}
	}
}

// Springs must have headroom: a launch into a ceiling is a wasted mechanic.
func TestSpringsHaveHeadroom(t *testing.T) {
	for w := range Levels {
		for l := range Levels[w] {
			def := Levels[w][l]
			lv, err := BuildLevel(def)
			if err != nil {
				t.Fatal(err)
			}
			wd, err := NewWorld(def, 9, 0, 0, core.PowerSmall)
			if err != nil {
				t.Fatal(err)
			}
			for _, s := range wd.Springs {
				tx := ifloor(s.CX() / TS)
				ty := ifloor(s.Y / TS)
				clear := 0
				for y := ty - 1; y >= 0 && !lv.At(tx, y).Solid(); y-- {
					clear++
				}
				if clear < 4 {
					t.Errorf("%d-%d: spring at tile (%d,%d) has only %d tiles of headroom",
						w+1, l+1, tx, ty, clear)
				}
			}
		}
	}
}
