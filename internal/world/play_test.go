package world

import (
	"math"
	"testing"

	"superdario/internal/core"
)

// botInput is a crude autopilot: hold right, jump on a rhythm. It is not meant
// to finish a level, only to prove the level is traversable enough that a
// player who holds right makes real progress without falling through geometry.
func botInput(tick int) Input {
	phase := tick % 42
	return Input{
		Right:    true,
		Run:      tick%240 > 60,
		Jump:     phase == 0,
		JumpDown: phase < 14,
	}
}

// The player must stay inside the level and remain visible to the camera while
// alive. A regression here means the camera or the collision has broken.
func TestBotStaysOnScreen(t *testing.T) {
	for w := range Levels {
		for l := range Levels[w] {
			def := Levels[w][l]
			wd, err := NewWorld(def, 99, 0, 0, core.PowerScaled)
			if err != nil {
				t.Fatalf("%d-%d: %v", w+1, l+1, err)
			}
			startX := wd.Player.X
			maxX := startX

			for i := 0; i < 3600; i++ { // one minute of play
				wd.Update(botInput(i))
				p := wd.Player
				if p.X > maxX {
					maxX = p.X
				}
				if wd.Status != StatusPlaying {
					// death or clear: respawn and keep going so one bad jump
					// does not end the check early
					if wd.Status == StatusDead || wd.Status == StatusTimeUp {
						wd.Respawn()
					}
					if wd.Status == StatusCleared {
						break
					}
					continue
				}
				if p.Dead {
					continue
				}
				// horizontal containment
				if p.X < -TS || p.Right() > wd.Level.PixelWidth()+TS {
					t.Fatalf("%d-%d tick %d: player left the level (x=%.1f, width=%.1f)",
						w+1, l+1, i, p.X, wd.Level.PixelWidth())
				}
				// Visibility: while inside the level bounds the player must be
				// on camera. Falling down a pit legitimately leaves the view
				// for the few frames before the death plane kills him, so
				// that case is excluded.
				if p.Bottom() < wd.Level.PixelHeight() {
					cx, cy := wd.Camera()
					sx := p.CX() - cx
					sy := p.CY() - cy
					if sx < -TS || sx > core.ScreenW+TS || sy < -TS || sy > core.ScreenH+TS {
						t.Fatalf("%d-%d tick %d: player is off camera at screen (%.1f, %.1f)",
							w+1, l+1, i, sx, sy)
					}
				}
			}

			progress := maxX - startX
			if progress < 200 {
				t.Errorf("%d-%d (%s): bot only advanced %.0fpx in a minute; the stage may be blocked near the start",
					w+1, l+1, def.Name, progress)
			}
		}
	}
}

// Holding right must not let Dario tunnel through solid tiles.
func TestPlayerNeverInsideSolidTile(t *testing.T) {
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
				p := wd.Player
				if p.Dead || p.Transf > 0 {
					continue
				}
				// sample the middle of the hitbox; corners legitimately touch
				tx := ifloor(p.CX() / TS)
				ty := ifloor(p.CY() / TS)
				if wd.Level.At(tx, ty).Solid() {
					t.Fatalf("%d-%d tick %d: player centre is inside a solid tile at (%d,%d)",
						w+1, l+1, i, tx, ty)
				}
			}
		}
	}
}

// A stage with a boss must actually spawn one when the player reaches the gate.
func TestBossSpawnsAtGate(t *testing.T) {
	for w := range Levels {
		for l := range Levels[w] {
			def := Levels[w][l]
			if def.Boss < 0 {
				continue
			}
			wd, err := NewWorld(def, 99, 0, 0, core.PowerAlign)
			if err != nil {
				t.Fatal(err)
			}
			// teleport to the gate rather than relying on the bot's driving
			wd.Player.X = wd.Level.BossGateX + 4
			wd.Player.Y = float64((wd.Level.H-5)*TS) - wd.Player.H
			wd.Update(Input{})
			if wd.Boss == nil {
				t.Fatalf("%d-%d: crossing the boss gate did not spawn a boss", w+1, l+1)
			}
			b := wd.Boss
			if b.MaxHP <= 0 {
				t.Errorf("%d-%d: boss has no health", w+1, l+1)
			}
			// the boss must sit inside its arena and on the floor
			if b.X < b.arenaL-1 || b.Right() > b.arenaR+1 {
				t.Errorf("%d-%d: boss starts outside its arena (x=%.1f, arena %.1f..%.1f)",
					w+1, l+1, b.X, b.arenaL, b.arenaR)
			}
		}
	}
}

// Every boss must be beatable by stomping alone, so a player who arrives
// without the Constitution Scroll is never stuck.
func TestEveryBossBecomesStompable(t *testing.T) {
	for w := range Levels {
		for l := range Levels[w] {
			def := Levels[w][l]
			if def.Boss < 0 {
				continue
			}
			wd, err := NewWorld(def, 99, 0, 0, core.PowerSmall)
			if err != nil {
				t.Fatal(err)
			}
			wd.Player.X = wd.Level.BossGateX + 4
			wd.Player.Y = float64((wd.Level.H-5)*TS) - wd.Player.H
			wd.Update(Input{})
			if wd.Boss == nil {
				t.Fatalf("%d-%d: no boss spawned", w+1, l+1)
			}
			exposed := false
			for i := 0; i < 1800 && !exposed; i++ {
				wd.Boss.Update(wd)
				if wd.Boss.Stompable() {
					exposed = true
				}
			}
			if !exposed {
				t.Errorf("%d-%d (%s): boss never enters a stompable state within 30s",
					w+1, l+1, def.Name)
			}
		}
	}
}

// Damaging a boss enough must defeat it and open the exit.
func TestBossCanBeDefeated(t *testing.T) {
	def := Levels[0][2]
	wd, err := NewWorld(def, 99, 0, 0, core.PowerAlign)
	if err != nil {
		t.Fatal(err)
	}
	wd.Player.X = wd.Level.BossGateX + 4
	wd.Player.Y = float64((wd.Level.H-5)*TS) - wd.Player.H
	wd.Update(Input{})
	b := wd.Boss
	if b == nil {
		t.Fatal("no boss")
	}
	b.state = bsExposed
	for i := 0; i < b.MaxHP; i++ {
		b.hurtT = 0
		if !b.Damage(wd, 1) {
			t.Fatalf("hit %d did not land", i)
		}
	}
	for i := 0; i < 300 && !b.Defeated; i++ {
		b.Update(wd)
	}
	if !b.Defeated {
		t.Fatal("boss never finished dying")
	}
	if wd.Goal == nil {
		t.Error("defeating the boss did not open an exit")
	}
}

// Tokens must convert to lives at 100 and never go negative.
func TestTokenToLife(t *testing.T) {
	wd, err := NewWorld(Levels[0][0], 3, 0, 99, core.PowerSmall)
	if err != nil {
		t.Fatal(err)
	}
	wd.collectToken()
	if wd.Tokens != 0 || wd.Lives != 4 {
		t.Errorf("100 tokens should give a life: tokens=%d lives=%d", wd.Tokens, wd.Lives)
	}
}

// Taking damage steps down one power tier at a time, and only kills at Small.
func TestDamageStepsDownOneTier(t *testing.T) {
	wd, err := NewWorld(Levels[0][0], 3, 0, 0, core.PowerAlign)
	if err != nil {
		t.Fatal(err)
	}
	p := wd.Player
	if !p.Hurt(wd) || p.Power != core.PowerScaled {
		t.Fatalf("Aligned should drop to Scaled, got %v", p.Power)
	}
	p.Invuln, p.Transf = 0, 0
	if !p.Hurt(wd) || p.Power != core.PowerSmall {
		t.Fatalf("Scaled should drop to Small, got %v", p.Power)
	}
	p.Invuln, p.Transf = 0, 0
	if !p.Hurt(wd) || !p.Dead {
		t.Fatal("Small should die")
	}
}

// Invulnerability must actually prevent damage.
func TestInvulnerabilityBlocksDamage(t *testing.T) {
	wd, _ := NewWorld(Levels[0][0], 3, 0, 0, core.PowerScaled)
	p := wd.Player
	p.Invuln = 60
	if p.Hurt(wd) {
		t.Error("damage landed while invulnerable")
	}
	p.Invuln = 0
	p.Halo = 60
	if p.Hurt(wd) {
		t.Error("damage landed during Safety Halo")
	}
}

// The clock must run out and kill the player.
func TestTimerExpires(t *testing.T) {
	wd, _ := NewWorld(Levels[0][0], 3, 0, 0, core.PowerSmall)
	wd.TimeLeft = 2
	wd.Update(Input{})
	wd.Update(Input{})
	if wd.Status != StatusTimeUp {
		t.Errorf("status = %v, want StatusTimeUp", wd.Status)
	}
	if !wd.Player.Dead {
		t.Error("running out of time should kill the player")
	}
}

// Respawning must place the player somewhere solid and reset the clock.
func TestRespawnIsSafe(t *testing.T) {
	wd, _ := NewWorld(Levels[0][0], 3, 0, 0, core.PowerScaled)
	wd.Player.Kill(wd)
	for i := 0; i < 200 && wd.Status != StatusDead; i++ {
		wd.Update(Input{})
	}
	wd.Respawn()
	if wd.Status != StatusPlaying {
		t.Errorf("status after respawn = %v", wd.Status)
	}
	if wd.Player.Dead {
		t.Error("player still dead after respawn")
	}
	// let gravity settle, then check there is ground below
	for i := 0; i < 120; i++ {
		wd.Update(Input{})
	}
	if wd.Player.Y > wd.Level.PixelHeight() {
		t.Error("player fell out of the world after respawning")
	}
	if math.Abs(wd.Player.X-wd.Level.StartX) > wd.Level.PixelWidth() {
		t.Error("respawn position is nonsense")
	}
}
