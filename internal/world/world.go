package world

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"superdario/internal/core"
	"superdario/internal/gfx"
	"superdario/internal/sfx"
)

// Status is what the owning scene polls to know when the stage is over.
type Status int

const (
	StatusPlaying Status = iota
	StatusDying          // death animation running
	StatusDead           // animation finished; scene should respawn or game-over
	StatusCleared        // stage complete
	StatusTimeUp
)

// World is one playthrough of one level.
type World struct {
	Level  *Level
	Player *Player

	Enemies     []*Enemy
	Items       []*Item
	Bolts       []*Bolt
	Shots       []*Shot
	Platforms   []*Platform
	FireBars    []*FireBar
	Springs     []*Spring
	Checkpoints []*Checkpoint
	Goal        *Goal
	Boss        *Boss

	Particles *gfx.System

	camX, camY float64
	shakeT     int
	shakeMag   float64
	tick       int
	rng        *rand.Rand

	Score    int
	Tokens   int
	Lives    int
	TimeLeft int // ticks

	Status Status
	stateT int

	// checkpoint
	cpX, cpY float64
	hasCP    bool

	// boss fight
	bossLocked bool
	bossLeft   float64
	bossRight  float64
	BossIntroT int

	// counters for multi-coin bricks
	brickCoins map[int]int

	// flags the scene reads
	ClearedTime int
	Muted       bool
	haloWasOn   bool
}

// NewWorld builds a playable world from a level definition.
func NewWorld(def LevelDef, lives, score, tokens int, power core.PowerState) (*World, error) {
	l, err := BuildLevel(def)
	if err != nil {
		return nil, err
	}
	w := &World{
		Level:      l,
		Particles:  gfx.NewSystem(420),
		rng:        rand.New(rand.NewSource(int64(def.World)*97 + int64(def.Index)*13 + 5)),
		Lives:      lives,
		Score:      score,
		Tokens:     tokens,
		TimeLeft:   def.Time * core.TPS,
		brickCoins: map[int]int{},
	}
	l.AttachArt()
	w.Player = NewPlayer(l.StartX, l.StartY, power)
	w.cpX, w.cpY = w.Player.X, w.Player.Y

	for _, s := range l.Spawns {
		w.spawn(s)
	}
	if l.HasGoal {
		w.Goal = newGoal(l.GoalX, l.GoalY, def.Model)
	}
	w.snapCamera()
	return w, nil
}

func (w *World) spawn(s Spawn) {
	switch s.Kind {
	case SpawnToken:
		w.Items = append(w.Items, newItem(ItToken, s.X, s.Y))
	case SpawnCompute:
		w.Items = append(w.Items, newItem(ItCompute, s.X, s.Y))
	case SpawnDistill:
		w.Enemies = append(w.Enemies, newEnemy(EnDistill, s.X, s.Y))
	case SpawnCrawler:
		w.Enemies = append(w.Enemies, newEnemy(EnCrawler, s.X, s.Y))
	case SpawnSpike:
		w.Enemies = append(w.Enemies, newEnemy(EnSpikeBot, s.X, s.Y))
	case SpawnFlyer:
		w.Enemies = append(w.Enemies, newEnemy(EnFlyer, s.X, s.Y))
	case SpawnLeech:
		w.Enemies = append(w.Enemies, newEnemy(EnLeech, s.X, s.Y))
	case SpawnPlant:
		w.Enemies = append(w.Enemies, newEnemy(EnPlant, s.X, s.Y))
	case SpawnSpring:
		w.Springs = append(w.Springs, newSpring(s.X, s.Y))
	case SpawnPlatH:
		w.Platforms = append(w.Platforms, newPlatform(PlatH, s.X, s.Y))
	case SpawnPlatV:
		w.Platforms = append(w.Platforms, newPlatform(PlatV, s.X, s.Y))
	case SpawnFalling:
		w.Platforms = append(w.Platforms, newPlatform(PlatFall, s.X, s.Y))
	case SpawnFireBar:
		w.FireBars = append(w.FireBars, newFireBar(s.X, s.Y))
	case SpawnCheckpoint:
		w.Checkpoints = append(w.Checkpoints, newCheckpoint(s.X, s.Y))
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func (w *World) Update(in Input) {
	w.tick++
	w.Level.Update()
	if w.Level.BG != nil { // nil when running headless (tests)
		w.Level.BG.Update()
	}
	w.Particles.Update()
	if w.shakeT > 0 {
		w.shakeT--
	}

	switch w.Status {
	case StatusDying:
		w.Player.Update(w, Input{})
		w.stateT++
		if w.stateT > 150 || w.Player.Y > w.Level.PixelHeight()+120 {
			w.Status = StatusDead
		}
		return
	case StatusDead, StatusCleared, StatusTimeUp:
		// the clear cinematic keeps running so Dario walks off screen
		if w.Status == StatusCleared {
			w.stateT++
			w.Player.Update(w, Input{})
			w.updateCamera()
		}
		return
	}

	// --- clock ----------------------------------------------------------
	if w.TimeLeft > 0 {
		w.TimeLeft--
		if w.TimeLeft == 100*core.TPS {
			// hurry-up warning
			sfx.Play(sfx.SndCheckpoint)
		}
		if w.TimeLeft == 0 {
			w.Player.Kill(w)
			w.Status = StatusTimeUp
			return
		}
	}

	// --- player ---------------------------------------------------------
	w.Player.Update(w, in)
	w.handleBump()
	w.ridePlatforms()

	// --- actors ---------------------------------------------------------
	for _, e := range w.Enemies {
		e.Update(w)
	}
	for _, it := range w.Items {
		it.Update(w)
	}
	for _, b := range w.Bolts {
		b.Update(w)
	}
	for _, s := range w.Shots {
		s.Update(w)
	}
	for _, p := range w.Platforms {
		p.Update(w)
	}
	for _, f := range w.FireBars {
		f.Update()
	}
	for _, s := range w.Springs {
		s.Update()
	}
	if w.Boss != nil {
		w.Boss.Update(w)
	}
	w.maybeStartBoss()

	// --- interactions ---------------------------------------------------
	w.collideItems()
	w.collideEnemies()
	w.collideBolts()
	w.collideShots()
	w.collideHazards()
	w.collideSprings()
	w.collideCheckpoints()
	w.collideGoal()

	w.reap()
	w.updateCamera()

	// fell out of the world
	if w.Player.Y > w.Level.PixelHeight()+40 && !w.Player.Dead {
		w.Player.Kill(w)
	}
}

// handleBump resolves Dario headbutting a block.
func (w *World) handleBump() {
	p := w.Player
	if !p.DidBump {
		return
	}
	tx, ty := p.BumpTX, p.BumpTY
	l := w.Level
	t := l.RawAt(tx, ty)
	idx := ty*l.W + tx
	above := float64(ty * TS)

	switch {
	case t.Question():
		l.Bump(tx, ty)
		l.Set(tx, ty, TQUsed)
		switch t {
		case TQCoin:
			w.popToken(tx, ty)
		case TQ1Up:
			it := newItem(ItSnapshot, float64(tx*TS), above)
			it.spawnFromBlock()
			w.Items = append(w.Items, it)
			sfx.Play(sfx.SndPowerUp)
		case TQHalo:
			it := newItem(ItHalo, float64(tx*TS), above)
			it.spawnFromBlock()
			w.Items = append(w.Items, it)
			sfx.Play(sfx.SndPowerUp)
		default: // TQPower
			kind := ItCompute
			if p.Power.Big() {
				kind = ItScroll
			}
			it := newItem(kind, float64(tx*TS), above)
			it.spawnFromBlock()
			w.Items = append(w.Items, it)
			sfx.Play(sfx.SndPowerUp)
		}
		w.bumpEnemiesOn(tx, ty)

	case t == THidden:
		l.Reveal(tx, ty)
		l.Set(tx, ty, TBlock)
		l.Bump(tx, ty)
		w.popToken(tx, ty)

	case t == TBrickCoin:
		l.Bump(tx, ty)
		w.brickCoins[idx]++
		w.popToken(tx, ty)
		if w.brickCoins[idx] >= 6 {
			l.Set(tx, ty, TQUsed)
		}
		w.bumpEnemiesOn(tx, ty)

	case t == TBrick:
		if p.Power.Big() {
			l.Set(tx, ty, TEmpty)
			w.addScore(50, float64(tx*TS)+8, above)
			sfx.Play(sfx.SndBreak)
			col := w.brickColor()
			w.Particles.Debris(float64(tx*TS)+8, above+8, col)
			w.Particles.Debris(float64(tx*TS)+8, above+8, gfx.Scale(col, 0.7))
		} else {
			l.Bump(tx, ty)
			sfx.Play(sfx.SndBump)
		}
		w.bumpEnemiesOn(tx, ty)

	default:
		sfx.Play(sfx.SndBump)
	}
}

// bumpEnemiesOn flips any enemy standing on the block that was just hit.
func (w *World) bumpEnemiesOn(tx, ty int) {
	top := float64(ty * TS)
	for _, e := range w.Enemies {
		if !e.Alive || e.flipping {
			continue
		}
		if math.Abs(e.Bottom()-top) < 4 && e.Right() > float64(tx*TS) && e.X < float64((tx+1)*TS) {
			e.KnockOut(w, sign(e.CX()-float64(tx*TS+8)))
			w.addScore(100, e.CX(), e.Y)
		}
	}
}

func (w *World) popToken(tx, ty int) {
	w.collectToken()
	x := float64(tx*TS) + 8
	y := float64(ty*TS) - 4
	w.Particles.Popup(x, y-6, "200", gfx.ColGoldLight)
	for i := 0; i < 6; i++ {
		w.Particles.Sparkle(x, y, gfx.ColGold)
	}
	w.Score += 200
}

// ridePlatforms lets Dario stand on and be carried by moving platforms.
func (w *World) ridePlatforms() {
	p := w.Player
	for _, pl := range w.Platforms {
		if !pl.Alive {
			continue
		}
		// only land from above, and only while falling
		if p.VY >= 0 &&
			p.Right() > pl.X+1 && p.X < pl.Right()-1 &&
			p.Bottom() >= pl.Y && p.Bottom() <= pl.Y+pl.H+math.Max(4, p.VY+2) {
			p.Y = pl.Y - p.H
			p.VY = 0
			p.OnGround = true
			p.X += pl.DX
			p.Y += pl.DY
			pl.Trigger()
			pl.Carrying = true
		} else {
			pl.Carrying = false
		}
	}
}

func (w *World) maybeStartBoss() {
	l := w.Level
	if w.Boss != nil || !l.HasBossGate || l.Def.Boss < 0 {
		return
	}
	if w.Player.CX() < l.BossGateX {
		return
	}
	// the arena is the last chunk; lock the camera to it
	arenaW := float64(28 * TS)
	left := l.PixelWidth() - arenaW
	w.bossLeft = left + 2*TS
	w.bossRight = l.PixelWidth() - 2*TS
	w.bossLocked = true

	floor := float64((l.H - 3) * TS)
	b := newBoss(l.Def.Boss, left+arenaW*0.72, floor)
	b.SetArena(l, w.bossLeft, w.bossRight, floor)
	w.Boss = b
	w.BossIntroT = 120
	sfx.PlayMusic(sfx.MusBoss)
}

// ---------------------------------------------------------------------------
// Collisions
// ---------------------------------------------------------------------------

func (w *World) collideItems() {
	p := w.Player
	if p.Dead {
		return
	}
	for _, it := range w.Items {
		if !it.Alive || it.emerge > 0 {
			continue
		}
		if !p.Overlaps(&it.Body) {
			continue
		}
		it.Alive = false
		switch it.Kind {
		case ItToken:
			w.collectToken()
			w.addScore(200, it.CX(), it.Y)
			for i := 0; i < 5; i++ {
				w.Particles.Sparkle(it.CX(), it.CY(), gfx.ColGoldLight)
			}
		case ItCompute:
			w.addScore(p.Grow(w, core.PowerScaled), it.CX(), it.Y)
		case ItScroll:
			w.addScore(p.Grow(w, core.PowerAlign), it.CX(), it.Y)
		case ItSnapshot:
			w.Lives++
			sfx.Play(sfx.SndOneUp)
			w.Particles.Popup(it.CX(), it.Y-8, "1UP", gfx.ColGreen)
		case ItHalo:
			p.GiveHalo(11 * core.TPS)
			w.haloWasOn = true
			sfx.PushMusic(sfx.MusHalo)
			w.Particles.Ring(it.CX(), it.CY(), gfx.ColGoldLight, 8)
			w.addScore(1000, it.CX(), it.Y)
		}
	}
}

func (w *World) collideEnemies() {
	p := w.Player
	if p.Dead || p.Transf > 0 {
		return
	}
	for _, e := range w.Enemies {
		if !e.Alive || !e.active || e.squashT > 0 || e.flipping {
			continue
		}
		if !p.Overlaps(&e.Body) {
			continue
		}

		// Safety Halo flattens everything it touches
		if p.Halo > 0 {
			e.KnockOut(w, sign(e.CX()-p.CX()))
			w.addScore(200, e.CX(), e.Y)
			continue
		}

		// a stationary shell gets kicked rather than hurting Dario
		if e.Kind == EnShell && !e.kicked {
			falling := p.VY > 0 && p.Bottom()-p.VY <= e.Y+8
			if falling {
				e.Stomp(w)
				p.Bounce(false)
			} else {
				e.kick(w, sign(p.CX()-e.CX()))
			}
			continue
		}

		falling := p.VY > 0 && (p.Bottom()-p.VY) <= e.Y+7
		if falling && e.Stompable() {
			if e.Stomp(w) {
				p.Bounce(p.jumping)
				sfx.Play(sfx.SndStomp)
			}
			continue
		}
		p.Hurt(w)
	}

	// boss body contact
	if w.Boss != nil && w.Boss.state != bsDying && w.Boss.state != bsDone && w.Boss.state != bsIntro {
		b := w.Boss
		if p.Overlaps(&b.Body) {
			switch {
			case b.hurtT > 0:
				// The boss is reeling. It cannot hurt you while it flashes —
				// without this, landing a stomp knocks it out of its
				// vulnerable state and the very next frame, with the player
				// still overlapping it, counts as a body hit. Every
				// successful stomp used to cost a life.

			case p.Halo > 0:
				b.Damage(w, 1)

			case b.Stompable() && p.VY > 0 && p.Bottom()-p.VY <= b.Y+12:
				if b.Damage(w, 1) {
					p.Bounce(p.jumping)
					// A full second of grace: the bounce arc alone is about
					// 40 ticks, so this leaves room to land and run clear
					// before the boss is dangerous to touch again.
					if p.Invuln < 60 {
						p.Invuln = 60
					}
					sfx.Play(sfx.SndStomp)
				}

			default:
				p.Hurt(w)
			}
		}
	}
}

func (w *World) collideBolts() {
	for _, bo := range w.Bolts {
		if !bo.Alive {
			continue
		}
		for _, e := range w.Enemies {
			if !e.Alive || !e.active || e.flipping || e.squashT > 0 {
				continue
			}
			if bo.Overlaps(&e.Body) {
				if e.Kind == EnSpikeBot {
					// armoured: the bolt just pops
					bo.explode(w)
					break
				}
				e.KnockOut(w, sign(bo.VX))
				w.addScore(200, e.CX(), e.Y)
				bo.explode(w)
				break
			}
		}
		if bo.Alive && w.Boss != nil && w.Boss.Vulnerable() {
			if bo.Overlaps(&w.Boss.Body) {
				w.Boss.Damage(w, 1)
				bo.explode(w)
			}
		}
	}
}

func (w *World) collideShots() {
	p := w.Player
	if p.Dead {
		return
	}
	for _, s := range w.Shots {
		if !s.Dangerous() {
			continue
		}
		if p.Overlaps(&s.Body) {
			if p.Halo > 0 {
				s.Alive = false
				continue
			}
			if p.Hurt(w) {
				s.Alive = false
			}
		}
	}
}

func (w *World) collideHazards() {
	p := w.Player
	if p.Dead || p.Halo > 0 {
		return
	}
	l := w.Level
	if p.TouchingHazard(l) {
		p.Hurt(w)
		return
	}
	// lava is instant death regardless of power state
	if l.Def.LavaKills && p.InLiquid(l) {
		p.Kill(w)
		return
	}
	for _, f := range w.FireBars {
		if f.Hits(&p.Body) {
			p.Hurt(w)
			return
		}
	}
}

func (w *World) collideSprings() {
	p := w.Player
	for _, s := range w.Springs {
		if p.VY >= 0 && p.Overlaps(&s.Body) && p.Bottom() <= s.Y+s.H*0.7+p.VY {
			p.Y = s.Y - p.H
			p.VY = -9.4
			p.jumping = true
			p.OnGround = false
			s.compress = 12
			sfx.Play(sfx.SndSpring)
		}
	}
}

func (w *World) collideCheckpoints() {
	p := w.Player
	for _, c := range w.Checkpoints {
		if c.Active || !p.Overlaps(&c.Body) {
			continue
		}
		c.Active = true
		w.cpX, w.cpY = p.X, p.Y
		w.hasCP = true
		sfx.Play(sfx.SndCheckpoint)
		w.Particles.Popup(c.CX(), c.Y-6, "CHECKPOINT", gfx.ColCream)
		for i := 0; i < 14; i++ {
			w.Particles.Sparkle(c.CX(), c.Y+float64(i*3), gfx.ColCoral)
		}
	}
}

func (w *World) collideGoal() {
	if w.Goal == nil || w.Status != StatusPlaying {
		return
	}
	p := w.Player
	// a boss level is only clear once the boss is down
	if w.Level.Def.Boss >= 0 && (w.Boss == nil || !w.Boss.Defeated) {
		return
	}
	if !p.Overlaps(&w.Goal.Body) {
		return
	}
	w.clearLevel()
}

func (w *World) clearLevel() {
	w.Status = StatusCleared
	w.stateT = 0
	w.ClearedTime = w.Level.Def.Time*core.TPS - w.TimeLeft
	if w.Goal != nil {
		w.Goal.Opened = true
		for i := 0; i < 30; i++ {
			w.Particles.Sparkle(w.Goal.CX(), w.Goal.CY(), gfx.ColGoldLight)
		}
		w.Particles.Ring(w.Goal.CX(), w.Goal.CY(), gfx.ColCream, 12)
	}
	w.Player.CheerT = 240
	w.Player.Frozen = false
	w.Player.VX = 0
	sfx.StopMusic()
	sfx.Play(sfx.SndRescue)
	sfx.PlayMusic(sfx.MusClear)
	// remaining time converts to score
	w.Score += (w.TimeLeft / core.TPS) * 10
}

// ---------------------------------------------------------------------------
// Helpers used by the actors
// ---------------------------------------------------------------------------

// brickColor is the debris colour for a broken brick. It falls back to a
// neutral tone when running without a graphics context (tests).
func (w *World) brickColor() color.RGBA {
	if w.Level.Theme != nil {
		return w.Level.Theme.Theme.BrickFace
	}
	return gfx.ColCoralDark
}

func (w *World) collectToken() {
	w.Tokens++
	sfx.Play(sfx.SndToken)
	if w.Tokens >= 100 {
		w.Tokens -= 100
		w.Lives++
		sfx.Play(sfx.SndOneUp)
	}
}

func (w *World) addScore(n int, x, y float64) {
	w.Score += n
	if n >= 100 {
		w.Particles.Popup(x, y-6, itoa(n), gfx.ColCream)
	}
}

func (w *World) spawnBolt(p *Player) bool {
	live := 0
	for _, b := range w.Bolts {
		if b.Alive {
			live++
		}
	}
	if live >= 2 {
		return false
	}
	x := p.CX() + float64(p.Facing)*6
	w.Bolts = append(w.Bolts, newBolt(x-4, p.Y+p.H*0.35, p.Facing))
	sfx.Play(sfx.SndShoot)
	return true
}

func (w *World) shake(mag float64) {
	w.shakeT = 14
	w.shakeMag = mag
}

func (w *World) playJump(big bool) {
	if big {
		sfx.Play(sfx.SndJumpBig)
	} else {
		sfx.Play(sfx.SndJump)
	}
	w.Particles.Dust(w.Player.CX(), w.Player.Bottom(), 0)
}

func (w *World) playSwim()   { sfx.Play(sfx.SndSwim) }
func (w *World) sfxBump()    { sfx.Play(sfx.SndBump) }
func (w *World) sfxKick()    { sfx.Play(sfx.SndKick) }
func (w *World) sfxShoot()   { sfx.Play(sfx.SndShoot) }
func (w *World) sfxBossHit() { sfx.Play(sfx.SndBossHit) }

func (w *World) onPlayerGrow()   { /* sound already played by the pickup */ }
func (w *World) onPlayerShrink() { sfx.Play(sfx.SndPowerDown) }

func (w *World) onPlayerDeath() {
	w.Status = StatusDying
	w.stateT = 0
	sfx.StopMusic()
	sfx.Play(sfx.SndDie)
}

func (w *World) onHaloEnd() {
	if w.haloWasOn {
		w.haloWasOn = false
		sfx.PopMusic()
	}
}

func (w *World) onBossDying() {
	// Everything the boss threw fizzles out with it. Landing the final blow
	// and then being killed by a shockwave already in flight is a rotten way
	// to lose a fight you have won.
	for _, s := range w.Shots {
		s.Alive = false
		w.Particles.Burst(s.CX(), s.CY(), 5, gfx.ColCream, 1.2)
	}
	w.Shots = nil
	sfx.Play(sfx.SndBossDie)
	sfx.StopMusic()
}

func (w *World) onBossPhase(b *Boss) {
	sfx.Play(sfx.SndBossHit)
	w.Particles.Ring(b.CX(), b.CY(), gfx.ColCoral, 14)
}

func (w *World) onBossDefeated() {
	w.Score += 5000
	w.bossLocked = false
	sfx.PlayMusic(sfx.MusClear)
	if w.Goal == nil && w.Level.HasGoal {
		w.Goal = newGoal(w.Level.GoalX, w.Level.GoalY, w.Level.Def.Model)
	}
	// the boss room's exit opens: place the rescue pod where the boss stood
	if w.Goal == nil {
		w.Goal = newGoal(w.Boss.CX()-TS/2, w.Boss.floorY-2*TS, w.Level.Def.Model)
	}
}

// Respawn puts Dario back at the last checkpoint.
func (w *World) Respawn() {
	def := w.Level.Def
	power := core.PowerSmall
	p := NewPlayer(0, 0, power)
	p.X, p.Y = w.cpX, w.cpY
	if !w.hasCP {
		p.X, p.Y = w.Level.StartX+(TS-hbW)/2, w.Level.StartY+TS-p.H
	}
	p.Invuln = 90
	w.Player = p
	w.TimeLeft = def.Time * core.TPS
	w.Status = StatusPlaying
	w.stateT = 0
	w.Shots = nil
	w.Bolts = nil
	w.Boss = nil
	w.bossLocked = false
	// Dario respawns Small, so give him back the power-up blocks he already
	// opened — otherwise a death at the boss leaves him with no way to re-arm.
	w.Level.RefillQuestionBlocks()
	w.brickCoins = map[int]int{}
	w.resetPlatforms()
	w.resetEnemies()
	w.snapCamera()
	sfx.PlayMusic(sfx.MusicFor(def.World))
}

// resetPlatforms rebuilds every platform from the level's spawn list.
//
// A falling platform drops away once stepped on and is reaped when it leaves
// the screen. Without this, dying after triggering them left the gap they
// spanned with nothing to cross it — and castle_lavarun is three of them over
// a lava pit, so the stage became unwinnable until you quit out to the map.
// Moving platforms are rebuilt too, which puts their phase back to the start
// so a retry plays out the same way as the first attempt.
func (w *World) resetPlatforms() {
	w.Platforms = w.Platforms[:0]
	for _, s := range w.Level.Spawns {
		switch s.Kind {
		case SpawnPlatH, SpawnPlatV, SpawnFalling:
			w.spawn(s)
		}
	}
	for _, s := range w.Springs {
		s.compress = 0
	}
}

// resetEnemies rebuilds every enemy from the level's spawn list.
//
// Stomped and flipped enemies are reaped out of w.Enemies, so simply waking the
// survivors back up left a retry with only whatever Dario had not yet killed —
// walk back through a cleared stretch and it was empty. Rebuilding also drops
// shells and the crawlers that climb out of them, and puts the survivors back
// at their spawn points facing their original way, so a retry plays out the
// same as the first attempt.
func (w *World) resetEnemies() {
	w.Enemies = w.Enemies[:0]
	for _, s := range w.Level.Spawns {
		switch s.Kind {
		case SpawnDistill, SpawnCrawler, SpawnSpike, SpawnFlyer, SpawnLeech, SpawnPlant:
			w.spawn(s)
		}
	}
}

// reap drops dead entities so the slices do not grow without bound.
func (w *World) reap() {
	w.Enemies = compact(w.Enemies, func(e *Enemy) bool { return e.Alive })
	w.Items = compact(w.Items, func(i *Item) bool { return i.Alive })
	w.Bolts = compact(w.Bolts, func(b *Bolt) bool { return b.Alive })
	w.Shots = compact(w.Shots, func(s *Shot) bool { return s.Alive })
	w.Platforms = compact(w.Platforms, func(p *Platform) bool { return p.Alive })
}

func compact[T any](xs []T, keep func(T) bool) []T {
	out := xs[:0]
	for _, x := range xs {
		if keep(x) {
			out = append(out, x)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Camera
// ---------------------------------------------------------------------------

func (w *World) updateCamera() {
	p := w.Player
	targetX := p.CX() - core.ScreenW*0.42
	if p.Facing < 0 {
		targetX = p.CX() - core.ScreenW*0.58
	}
	targetY := p.CY() - core.ScreenH*0.55

	if w.bossLocked {
		targetX = w.Level.PixelWidth() - float64(28*TS)
		targetY = w.Level.PixelHeight() - core.ScreenH
	}

	w.camX += (targetX - w.camX) * 0.11
	w.camY += (targetY - w.camY) * 0.09
	w.clampCamera()
}

func (w *World) snapCamera() {
	p := w.Player
	w.camX = p.CX() - core.ScreenW*0.42
	w.camY = p.CY() - core.ScreenH*0.55
	w.clampCamera()
}

func (w *World) clampCamera() {
	maxX := w.Level.PixelWidth() - core.ScreenW
	maxY := w.Level.PixelHeight() - core.ScreenH
	if maxX < 0 {
		maxX = 0
	}
	if maxY < 0 {
		maxY = 0
	}
	w.camX = clampF(w.camX, 0, maxX)
	w.camY = clampF(w.camY, 0, maxY)
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Camera exposes the current view origin (used by the HUD for effects).
func (w *World) Camera() (float64, float64) { return w.camX, w.camY }

// Tick is the world's frame counter.
func (w *World) Tick() int { return w.tick }

// ---------------------------------------------------------------------------
// Draw
// ---------------------------------------------------------------------------

func (w *World) Draw(dst *ebiten.Image) {
	camX, camY := w.camX, w.camY
	if w.shakeT > 0 {
		f := float64(w.shakeT) / 14
		camX += (w.rng.Float64()*2 - 1) * w.shakeMag * f
		camY += (w.rng.Float64()*2 - 1) * w.shakeMag * f
	}
	camX = math.Floor(camX)
	camY = math.Floor(camY)

	w.Level.BG.Draw(dst, camX, camY)

	// data-pipe plants live behind the tiles so the pipe hides their base
	for _, e := range w.Enemies {
		if e.Kind == EnPlant {
			e.Draw(dst, camX, camY, w.tick)
		}
	}

	w.Level.Draw(dst, camX, camY, w.tick)

	for _, p := range w.Platforms {
		p.Draw(dst, w.Level.Theme, camX, camY)
	}
	for _, s := range w.Springs {
		s.Draw(dst, w.Level.Theme, camX, camY)
	}
	for _, c := range w.Checkpoints {
		c.Draw(dst, camX, camY, w.tick)
	}
	if w.Goal != nil {
		w.Goal.Draw(dst, camX, camY, w.tick)
	}
	for _, it := range w.Items {
		it.Draw(dst, camX, camY)
	}
	for _, e := range w.Enemies {
		if e.Kind != EnPlant {
			e.Draw(dst, camX, camY, w.tick)
		}
	}
	if w.Boss != nil {
		w.Boss.Draw(dst, camX, camY, w.tick)
	}
	for _, b := range w.Bolts {
		b.Draw(dst, camX, camY)
	}
	for _, s := range w.Shots {
		s.Draw(dst, camX, camY, w.tick)
	}
	for _, f := range w.FireBars {
		f.Draw(dst, camX, camY, w.tick)
	}

	w.Player.Draw(dst, w.Level, camX, camY, w.tick)
	w.Particles.Draw(dst, camX, camY)

	// underwater tint
	if w.Level.Def.Underwater {
		gfx.FillRectF(dst, 0, 0, core.ScreenW, core.ScreenH, gfx.Alpha(gfx.ColSlate, 0.18))
	}
}

// itoa avoids pulling strconv into hot paths for tiny numbers.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
