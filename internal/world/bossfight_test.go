package world

import (
	"testing"

	"superdario/internal/core"
	"superdario/internal/gfx"
)

// startBossFight puts a player at the gate of a boss stage and runs the world
// until the boss has spawned and finished its intro.
func startBossFight(t *testing.T, world, level int, power core.PowerState) *World {
	t.Helper()
	def := Levels[world][level]
	if def.Boss < 0 {
		t.Fatalf("%d-%d is not a boss stage", world+1, level+1)
	}
	wd, err := NewWorld(def, 9, 0, 0, power)
	if err != nil {
		t.Fatal(err)
	}
	wd.Player.X = wd.Level.BossGateX + 4
	wd.Player.Y = float64((wd.Level.H-4)*TS) - wd.Player.H
	wd.Update(Input{})
	if wd.Boss == nil {
		t.Fatalf("%d-%d: no boss spawned at the gate", world+1, level+1)
	}
	for i := 0; i < 240 && !wd.Boss.Started; i++ {
		wd.Update(Input{})
	}
	return wd
}

// placeStomp drops the player onto the boss's head with downward velocity,
// exactly as a real landing would arrive.
func placeStomp(wd *World) {
	b := wd.Boss
	p := wd.Player
	p.X = b.CX() - p.W/2
	p.Y = b.Y - p.H + 2 // just breaking the surface of the boss's hitbox
	p.VY = 4
	p.OnGround = false
}

// moveClear puts the player back on the arena floor, well away from the boss —
// what a real player does after bouncing off it.
func moveClear(wd *World) {
	b := wd.Boss
	p := wd.Player
	if b.CX() > (b.arenaL+b.arenaR)/2 {
		p.X = b.arenaL + 8
	} else {
		p.X = b.arenaR - p.W - 8
	}
	p.Y = b.floorY - p.H
	p.VX, p.VY = 0, 0
	p.OnGround = true
}

// Landing a stomp must not cost the player health. This is the exact bug that
// made the World 1 boss feel impossible: damaging the boss knocked it out of
// its vulnerable state, and on the very next frame the still-overlapping
// player was treated as having walked into it.
func TestStompingABossDoesNotHurtThePlayer(t *testing.T) {
	for w := range Levels {
		for l := range Levels[w] {
			if Levels[w][l].Boss < 0 {
				continue
			}
			wd := startBossFight(t, w, l, core.PowerScaled)
			b := wd.Boss

			// wind forward to a vulnerable window
			found := false
			for i := 0; i < 1800 && !found; i++ {
				wd.Update(Input{})
				if b.Stompable() {
					found = true
				}
			}
			if !found {
				t.Fatalf("%d-%d: boss never became stompable", w+1, l+1)
			}

			hpBefore := b.HP
			powerBefore := wd.Player.Power
			placeStomp(wd)
			wd.Update(Input{})

			if b.HP != hpBefore-1 {
				t.Errorf("%d-%d: stomp did not damage the boss (hp %d -> %d)",
					w+1, l+1, hpBefore, b.HP)
			}
			// Ride out the bounce arc while still overlapping the boss. This
			// is the window the bug lived in: the boss leaves its vulnerable
			// state the instant it is hit, and the player has not moved yet.
			for i := 0; i < 45; i++ {
				wd.Update(Input{})
				if wd.Player.Dead {
					t.Fatalf("%d-%d: player died %d frames after a successful stomp", w+1, l+1, i)
				}
				if wd.Player.Power < powerBefore {
					t.Fatalf("%d-%d: player lost a power tier %d frames after a successful stomp",
						w+1, l+1, i)
				}
			}
		}
	}
}

// Every boss must be beatable with nothing but jumps — no Constitution Scroll,
// no Safety Halo — because a player can arrive at any boss having just been
// hit. This plays the whole fight: wait for the window, land a stomp, repeat.
func TestEveryBossIsBeatableByStompingAlone(t *testing.T) {
	for w := range Levels {
		for l := range Levels[w] {
			def := Levels[w][l]
			if def.Boss < 0 {
				continue
			}
			// PowerScaled, not Aligned: no bolts available at all.
			wd := startBossFight(t, w, l, core.PowerScaled)
			b := wd.Boss
			maxHP := b.MaxHP

			hits := 0
			cooldown := 0
			const limit = 60 * 60 * 4 // four simulated minutes
			for i := 0; i < limit && !b.Defeated; i++ {
				switch {
				case cooldown > 0:
					cooldown--
					if cooldown == 0 {
						// bounced off; run back to safe ground
						moveClear(wd)
					}
				case b.Stompable() && b.hurtT == 0:
					hp := b.HP
					placeStomp(wd)
					wd.Update(Input{})
					if b.HP < hp {
						hits++
						cooldown = 40 // the bounce arc
					}
					continue
				}
				wd.Update(Input{})
				if wd.Status == StatusDead || wd.Status == StatusTimeUp {
					t.Fatalf("%d-%d (%s): player died before the boss did (%d/%d hits landed)",
						w+1, l+1, def.Name, hits, maxHP)
				}
			}

			if !b.Defeated {
				t.Errorf("%d-%d (%s): boss survived four minutes of stomping (%d/%d hits landed, hp %d)",
					w+1, l+1, def.Name, hits, maxHP, b.HP)
				continue
			}
			if hits != maxHP {
				t.Errorf("%d-%d (%s): took %d stomps to fell a %d HP boss",
					w+1, l+1, def.Name, hits, maxHP)
			}
			if wd.Goal == nil {
				t.Errorf("%d-%d (%s): beating the boss did not open the exit", w+1, l+1, def.Name)
			}
			t.Logf("%d-%d %-20s beaten in %d stomps", w+1, l+1, def.Name, hits)
		}
	}
}

// A vulnerable window has to be long enough to actually react to and use. A
// jump takes roughly 40 ticks up and down, so anything under about a second
// is not a fair invitation.
func TestVulnerableWindowsAreUsable(t *testing.T) {
	const minTicks = 90 // 1.5 seconds
	for kind := 0; kind < gfx.BossCount; kind++ {
		if got := exposedTicks(kind); got < minTicks {
			t.Errorf("%s: vulnerable window is %d ticks, want at least %d",
				gfx.BossNames[kind], got, minTicks)
		}
	}
}

// The window must not get shorter as the boss weakens — that would make the
// last hit of every fight the hardest one.
func TestVulnerableWindowDoesNotShrinkWithDamage(t *testing.T) {
	wd := startBossFight(t, 0, 2, core.PowerScaled)
	b := wd.Boss

	windowAt := func() int {
		for i := 0; i < 1800; i++ {
			wd.Update(Input{})
			if b.Stompable() {
				return b.timer
			}
		}
		t.Fatal("boss never became stompable")
		return 0
	}

	full := windowAt()
	// hurt it and measure again
	b.hurtT = 0
	b.Damage(wd, 1)
	for i := 0; i < 60; i++ {
		wd.Update(Input{})
	}
	hurt := windowAt()

	if hurt < full {
		t.Errorf("vulnerable window shrank after damage: %d -> %d ticks", full, hurt)
	}
}

// Every boss stage must offer a way to re-arm, so losing the Constitution
// Scroll on the way in is a setback rather than a dead end.
func TestBossStagesLetYouReArm(t *testing.T) {
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
			// count power-up blocks after the last checkpoint
			var lastCP float64
			wd, _ := NewWorld(def, 3, 0, 0, core.PowerSmall)
			for _, c := range wd.Checkpoints {
				if c.X > lastCP {
					lastCP = c.X
				}
			}
			n := 0
			for ty := 0; ty < lv.H; ty++ {
				for tx := 0; tx < lv.W; tx++ {
					if float64(tx*TS) < lastCP {
						continue
					}
					if lv.RawAt(tx, ty) == TQPower {
						n++
					}
				}
			}
			if n == 0 {
				t.Errorf("%d-%d (%s): no power-up block after the last checkpoint; "+
					"a player who arrives hurt cannot re-arm", w+1, l+1, def.Name)
			}
		}
	}
}

// Respawning at a checkpoint must restore the power-up blocks, or repeated
// deaths at a boss leave the player permanently unable to re-arm.
func TestRespawnRefillsPowerUpBlocks(t *testing.T) {
	wd := startBossFight(t, 0, 2, core.PowerScaled)
	lv := wd.Level

	// spend every ? block in the stage
	spent := 0
	for ty := 0; ty < lv.H; ty++ {
		for tx := 0; tx < lv.W; tx++ {
			if lv.RawAt(tx, ty).Question() {
				lv.Set(tx, ty, TQUsed)
				spent++
			}
		}
	}
	if spent == 0 {
		t.Fatal("stage has no ? blocks to test with")
	}

	wd.Respawn()

	back := 0
	for ty := 0; ty < lv.H; ty++ {
		for tx := 0; tx < lv.W; tx++ {
			if lv.RawAt(tx, ty).Question() {
				back++
			}
		}
	}
	if back != spent {
		t.Errorf("respawn restored %d of %d ? blocks", back, spent)
	}
}
