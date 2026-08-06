package game

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"superdario/internal/core"
	"superdario/internal/gfx"
	"superdario/internal/sfx"
	"superdario/internal/world"
)

// PlayScene runs a level: the world, the HUD, the title card and the pause menu.
type PlayScene struct {
	w    *world.World
	def  world.LevelDef
	t    int
	card int // title-card countdown before play starts

	paused   bool
	pauseSel int

	clearT   int
	deadT    int
	buildErr error
}

var pauseMenu = []string{"RESUME", "RESTART STAGE", "OPTIONS", "QUIT TO MAP"}

func (s *PlayScene) Enter(g *Game) {
	// Returning from an overlay (Options, Controls) must not restart the
	// stage — Enter is called again on the *same* scene instance, so keep the
	// world that is already running.
	if s.w != nil {
		s.paused = false
		sfx.PlayMusic(sfx.MusicFor(s.def.World))
		return
	}
	s.def = world.LevelAt(g.Run.World, g.Run.Level)
	w, err := world.NewWorld(s.def, g.Run.Lives, g.Run.Score, g.Run.Tokens, g.Run.Power)
	if err != nil {
		s.buildErr = err
		g.Toast("LEVEL FAILED TO LOAD")
		return
	}
	s.w = w
	s.card = 150
	s.t = 0
	s.clearT = 0
	s.deadT = 0
	s.paused = false
	sfx.PlayMusic(sfx.MusicFor(s.def.World))
}

func (s *PlayScene) Update(g *Game) error {
	s.t++
	if s.buildErr != nil {
		if g.Pad.ConfirmPressed() || g.Pad.BackPressed() {
			g.Go(&MapScene{})
		}
		return nil
	}

	// --- pause -----------------------------------------------------------
	if s.paused {
		if g.Pad.MenuUp() {
			s.pauseSel = (s.pauseSel + len(pauseMenu) - 1) % len(pauseMenu)
			sfx.Play(sfx.SndMenuMove)
		}
		if g.Pad.MenuDown() {
			s.pauseSel = (s.pauseSel + 1) % len(pauseMenu)
			sfx.Play(sfx.SndMenuMove)
		}
		if g.Pad.PausePressed() {
			s.setPaused(false)
			return nil
		}
		if g.Pad.ConfirmPressed() {
			sfx.Play(sfx.SndMenuPick)
			switch s.pauseSel {
			case 0:
				s.setPaused(false)
			case 1:
				s.setPaused(false)
				g.Go(&PlayScene{}) // a fresh scene really does restart
			case 2:
				// Keep the music running so volume changes are audible, and
				// hand back *this* scene so the stage survives the detour.
				s.setPaused(false)
				g.Go(&OptionsScene{back: func() Scene { return s }})
			case 3:
				s.setPaused(false)
				s.syncRun(g)
				g.PersistRun()
				g.Go(&MapScene{})
			}
		}
		return nil
	}

	// --- title card ------------------------------------------------------
	if s.card > 0 {
		s.card--
		if g.Pad.ConfirmPressed() || g.Pad.JumpPressed() {
			s.card = 0
		}
		return nil
	}

	if g.Pad.PausePressed() {
		s.setPaused(true)
		return nil
	}

	// --- run the world ----------------------------------------------------
	in := g.Pad.GameInput()
	if g.Autopilot != nil {
		if auto, ok := g.Autopilot(s.t); ok {
			in = auto
		}
	}
	s.w.Update(in)

	switch s.w.Status {
	case world.StatusCleared:
		s.clearT++
		if s.clearT > 180 {
			s.onCleared(g)
		}
	case world.StatusDead, world.StatusTimeUp:
		s.deadT++
		if s.deadT > 40 {
			s.onDeath(g)
		}
	}
	return nil
}

func (s *PlayScene) setPaused(p bool) {
	s.paused = p
	if p {
		s.pauseSel = 0
		sfx.Play(sfx.SndPause)
		sfx.PauseMusic()
	} else {
		sfx.ResumeMusic()
	}
}

// syncRun copies live world state back into the session.
func (s *PlayScene) syncRun(g *Game) {
	if s.w == nil {
		return
	}
	g.Run.Lives = s.w.Lives
	g.Run.Score = s.w.Score
	g.Run.Tokens = s.w.Tokens
	g.Run.Power = s.w.Player.Power
}

func (s *PlayScene) onCleared(g *Game) {
	s.syncRun(g)
	slot := g.Slot()
	slot.Unlock(s.def.World, s.def.Index)
	slot.RecordTime(s.def.World, s.def.Index, s.w.ClearedTime)
	slot.Deaths += 0
	if s.def.Model >= 0 {
		g.Run.Rescued[s.def.Model] = true
	}
	g.PersistRun()

	next := func() Scene { return &MapScene{} }

	// after a boss, play that world's interlude (or the ending)
	if s.def.Boss >= 0 {
		if s.def.World == len(world.Levels)-1 {
			g.Go(&ResultsScene{def: s.def, w: s.w, next: func() Scene {
				return &StoryScene{cut: Ending, next: func() Scene { return &CreditsScene{} }}
			}})
			return
		}
		cut := Interludes[s.def.World]
		if len(cut.Panels) > 0 {
			next = func() Scene {
				return &StoryScene{cut: cut, next: func() Scene { return &MapScene{} }}
			}
		}
	}
	// advance the cursor to the next stage
	nw, nl := s.def.World, s.def.Index+1
	if nl >= len(world.Levels[nw]) {
		nw, nl = nw+1, 0
	}
	if nw < len(world.Levels) {
		g.Run.World, g.Run.Level = nw, nl
	}
	g.Go(&ResultsScene{def: s.def, w: s.w, next: next})
}

func (s *PlayScene) onDeath(g *Game) {
	g.Slot().Deaths++
	s.w.Lives--
	g.Run.Lives = s.w.Lives
	g.Run.Score = s.w.Score
	g.Run.Tokens = s.w.Tokens
	g.Run.Power = core.PowerSmall
	g.PersistRun()

	if s.w.Lives <= 0 {
		g.Go(&GameOverScene{})
		return
	}
	// respawn in place at the last checkpoint
	s.w.Respawn()
	s.deadT = 0
	s.card = 90
}

func (s *PlayScene) Draw(g *Game, dst *ebiten.Image) {
	if s.buildErr != nil {
		dst.Fill(gfx.ColInkDeep)
		gfx.TextCenter(dst, "THIS STAGE FAILED TO LOAD", core.ScreenW/2, 110, gfx.ColRed, 2)
		for i, line := range gfx.WrapText(s.buildErr.Error(), core.ScreenW-60, 1) {
			gfx.TextCenter(dst, line, core.ScreenW/2, 140+i*10, gfx.Alpha(gfx.ColCream, 0.8), 1)
		}
		gfx.TextCenter(dst, "PRESS ENTER TO RETURN TO THE MAP", core.ScreenW/2, 200, gfx.ColGoldLight, 1)
		return
	}

	s.w.Draw(dst)
	s.drawHUD(g, dst)

	if s.w.BossIntroT > 0 && s.w.Boss != nil {
		s.w.BossIntroT--
		s.drawBossIntro(dst)
	}
	if s.w.Boss != nil && s.w.Boss.Started && !s.w.Boss.Defeated {
		s.drawBossBar(dst)
	}
	if s.card > 0 {
		s.drawTitleCard(dst)
	}
	if s.paused {
		s.drawPause(dst)
	}
	if s.w.Status == world.StatusCleared {
		s.drawClearBanner(dst)
	}
	if s.w.Status == world.StatusTimeUp {
		gfx.FillRectF(dst, 0, 0, core.ScreenW, core.ScreenH, gfx.Alpha(gfx.ColInkDeep, 0.5))
		gfx.TextCenterOutline(dst, "TIME UP", core.ScreenW/2, 110, gfx.ColRed, gfx.ColInkDeep, 4)
	}
}

func (s *PlayScene) drawHUD(g *Game, dst *ebiten.Image) {
	w := s.w
	gfx.VGradient(dst, 0, 0, core.ScreenW, 24, gfx.Alpha(gfx.ColInkDeep, 0.78), gfx.Alpha(gfx.ColInkDeep, 0))

	// score
	gfx.TextShadow(dst, "DARIO", 8, 3, gfx.ColCoral, gfx.ColInkDeep, 1)
	gfx.TextShadow(dst, fmt.Sprintf("%07d", w.Score), 8, 13, gfx.ColCream, gfx.ColInkDeep, 1)

	// tokens
	gfx.DrawSprite(dst, gfx.ImgTokenHUD, 86, 11, gfx.SpriteOpt{})
	gfx.TextShadow(dst, fmt.Sprintf("x%02d", w.Tokens), 96, 13, gfx.ColGoldLight, gfx.ColInkDeep, 1)

	// lives
	gfx.DrawSprite(dst, gfx.ImgHeart, 138, 9, gfx.SpriteOpt{})
	gfx.TextShadow(dst, fmt.Sprintf("x%d", w.Lives), 150, 13, gfx.ColCream, gfx.ColInkDeep, 1)

	// stage
	gfx.TextShadow(dst, "STAGE", 196, 3, gfx.Alpha(gfx.ColCream, 0.7), gfx.ColInkDeep, 1)
	gfx.TextShadow(dst, fmt.Sprintf("%d-%d", s.def.World+1, s.def.Index+1), 196, 13, gfx.ColCream, gfx.ColInkDeep, 1)

	// power state
	label := "SMALL"
	col := gfx.Alpha(gfx.ColCream, 0.8)
	switch w.Player.Power {
	case core.PowerScaled:
		label, col = "SCALED", gfx.ColCyan
	case core.PowerAlign:
		label, col = "ALIGNED", gfx.ColGoldLight
	}
	gfx.TextShadow(dst, label, 244, 13, col, gfx.ColInkDeep, 1)

	// timer
	if g.Save.Settings.ShowTimer {
		secs := w.TimeLeft / core.TPS
		tcol := gfx.ColCream
		if secs <= 30 {
			tcol = gfx.ColRed
			if (s.t/10)%2 == 0 {
				tcol = gfx.ColGoldLight
			}
		}
		gfx.TextShadow(dst, "TIME", core.ScreenW-60, 3, gfx.Alpha(gfx.ColCream, 0.7), gfx.ColInkDeep, 1)
		gfx.TextShadow(dst, fmt.Sprintf("%03d", secs), core.ScreenW-60, 13, tcol, gfx.ColInkDeep, 1)
	}

	// Safety Halo countdown
	if w.Player.Halo > 0 {
		frac := float64(w.Player.Halo) / float64(11*core.TPS)
		gfx.FillRectF(dst, 300, 8, 60, 5, gfx.Alpha(gfx.ColInkDeep, 0.7))
		gfx.FillRectF(dst, 300, 8, 60*frac, 5, gfx.TintFor(s.t))
		gfx.StrokeRectF(dst, 300, 8, 60, 5, gfx.Alpha(gfx.ColCream, 0.6))
		gfx.Text(dst, "HALO", 300, 0, gfx.ColGoldLight, 1)
	}
}

func (s *PlayScene) drawTitleCard(dst *ebiten.Image) {
	a := 1.0
	if s.card < 30 {
		a = float64(s.card) / 30
	}
	gfx.FillRectF(dst, 0, 0, core.ScreenW, core.ScreenH, gfx.Alpha(gfx.ColInkDeep, 0.86*a))
	gfx.TextCenterOutline(dst, fmt.Sprintf("STAGE %d-%d", s.def.World+1, s.def.Index+1),
		core.ScreenW/2, 84, gfx.Alpha(gfx.ColGoldLight, a), gfx.Alpha(gfx.ColInkDeep, a), 3)
	gfx.TextCenterOutline(dst, s.def.Name, core.ScreenW/2, 118, gfx.Alpha(gfx.ColCream, a), gfx.Alpha(gfx.ColInkDeep, a), 2)

	if s.def.Intro != "" {
		for i, line := range gfx.WrapText(s.def.Intro, core.ScreenW-80, 1) {
			gfx.TextCenter(dst, line, core.ScreenW/2, 144+i*11, gfx.Alpha(gfx.ColCream, 0.75*a), 1)
		}
	}

	// lives readout, the classic way
	gfx.DrawSprite(dst, gfx.Dario[int(s.w.Player.Power)][gfx.FrStand], core.ScreenW/2-30, 178, gfx.SpriteOpt{Alpha: a})
	gfx.TextShadow(dst, fmt.Sprintf("x  %d", s.w.Lives), core.ScreenW/2-4, 190,
		gfx.Alpha(gfx.ColCream, a), gfx.Alpha(gfx.ColInkDeep, a), 2)
}

func (s *PlayScene) drawBossIntro(dst *ebiten.Image) {
	b := s.w.Boss
	f := float64(s.w.BossIntroT) / 120
	a := math.Min(1, f*2.5)
	gfx.FillRectF(dst, 0, 96, core.ScreenW, 40, gfx.Alpha(gfx.ColInkDeep, 0.85*a))
	gfx.FillRectF(dst, 0, 96, core.ScreenW, 1, gfx.Alpha(gfx.ColRed, a))
	gfx.FillRectF(dst, 0, 135, core.ScreenW, 1, gfx.Alpha(gfx.ColRed, a))
	gfx.TextCenterOutline(dst, b.Name(), core.ScreenW/2, 104, gfx.Alpha(gfx.ColRed, a), gfx.Alpha(gfx.ColInkDeep, a), 3)
	gfx.TextCenter(dst, "OF THE DISTILLATION SYNDICATE", core.ScreenW/2, 126, gfx.Alpha(gfx.ColCream, 0.8*a), 1)
}

func (s *PlayScene) drawBossBar(dst *ebiten.Image) {
	b := s.w.Boss
	bw := 200.0
	x := float64(core.ScreenW)/2 - bw/2
	gfx.FillRectF(dst, x-2, 28, bw+4, 12, gfx.Alpha(gfx.ColInkDeep, 0.8))
	gfx.FillRectF(dst, x, 30, bw, 8, gfx.Alpha(gfx.ColSlate, 0.9))
	frac := float64(b.HP) / float64(b.MaxHP)
	col := gfx.ColRed
	if frac > 0.6 {
		col = gfx.ColCoral
	} else if frac < 0.34 {
		col = gfx.ColGold
	}
	gfx.FillRectF(dst, x, 30, bw*frac, 8, col)
	gfx.StrokeRectF(dst, x, 30, bw, 8, gfx.Alpha(gfx.ColCream, 0.7))
	gfx.TextCenter(dst, b.Name(), core.ScreenW/2, 42, gfx.Alpha(gfx.ColCream, 0.85), 1)
}

func (s *PlayScene) drawClearBanner(dst *ebiten.Image) {
	a := math.Min(1, float64(s.clearT)/30)
	gfx.FillRectF(dst, 0, 84, core.ScreenW, 56, gfx.Alpha(gfx.ColInkDeep, 0.7*a))
	gfx.TextCenterOutline(dst, "RESCUED!", core.ScreenW/2, 92, gfx.Alpha(gfx.ColGoldLight, a), gfx.Alpha(gfx.ColInkDeep, a), 3)
	if s.def.Model >= 0 {
		gfx.TextCenter(dst, gfx.ModelNames[s.def.Model]+" IS SAFE", core.ScreenW/2, 122,
			gfx.Alpha(gfx.ModelColors[s.def.Model], a), 1)
	} else {
		gfx.TextCenter(dst, "STAGE CLEAR", core.ScreenW/2, 122, gfx.Alpha(gfx.ColCream, a), 1)
	}
}

func (s *PlayScene) drawPause(dst *ebiten.Image) {
	gfx.FillRectF(dst, 0, 0, core.ScreenW, core.ScreenH, gfx.Alpha(gfx.ColInkDeep, 0.75))
	gfx.Panel(dst, 140, 76, 200, 118, gfx.ColSlate, gfx.ColCoral)
	gfx.TextCenterOutline(dst, "PAUSED", core.ScreenW/2, 86, gfx.ColCream, gfx.ColInkDeep, 2)
	for i, item := range pauseMenu {
		y := 112 + i*18
		col := gfx.Alpha(gfx.ColCream, 0.7)
		if i == s.pauseSel {
			col = gfx.ColGoldLight
			if (s.t/16)%2 == 0 {
				gfx.Text(dst, ">", 156, y, gfx.ColCoral, 1)
			}
		}
		gfx.Text(dst, item, 170, y, col, 1)
	}
	gfx.TextCenter(dst, "ESC RESUME", core.ScreenW/2, 200, gfx.Alpha(gfx.ColCream, 0.4), 1)
}
