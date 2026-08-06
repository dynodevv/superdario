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

// ---------------------------------------------------------------------------
// ResultsScene — the post-stage tally
// ---------------------------------------------------------------------------

type ResultsScene struct {
	def   world.LevelDef
	w     *world.World
	next  func() Scene
	t     int
	rows  []resultRow
	shown int
}

type resultRow struct {
	label string
	value string
	col   [3]uint8
	gold  bool
}

func (s *ResultsScene) Enter(g *Game) {
	sfx.PlayMusic(sfx.MusClear)
	slot := g.Slot()
	best := slot.BestTime[s.def.World][s.def.Index]

	s.rows = []resultRow{
		{label: "TIME", value: fmtTicks(s.w.ClearedTime)},
		{label: "TOKENS", value: fmt.Sprintf("%d", s.w.Tokens)},
		{label: "SCORE", value: fmt.Sprintf("%07d", s.w.Score)},
		{label: "LIVES", value: fmt.Sprintf("%d", s.w.Lives)},
	}
	if best > 0 && best >= s.w.ClearedTime {
		s.rows = append(s.rows, resultRow{label: "NEW BEST TIME", value: fmtTicks(best), gold: true})
	}
	if s.def.Model >= 0 {
		s.rows = append(s.rows, resultRow{label: "RESCUED", value: gfx.ModelNames[s.def.Model], gold: true})
	}
}

func (s *ResultsScene) Update(g *Game) error {
	s.t++
	if s.t%22 == 0 && s.shown < len(s.rows) {
		s.shown++
		sfx.Play(sfx.SndToken)
	}
	if g.Pad.ConfirmPressed() || g.Pad.JumpPressed() {
		if s.shown < len(s.rows) {
			s.shown = len(s.rows)
			return nil
		}
		sfx.Play(sfx.SndMenuPick)
		g.Go(s.next())
	}
	return nil
}

func (s *ResultsScene) Draw(g *Game, dst *ebiten.Image) {
	dst.Fill(gfx.ColInkDeep)
	gfx.Backgrounds[s.def.World].Draw(dst, float64(s.t)*0.3, 0)
	gfx.FillRectF(dst, 0, 0, core.ScreenW, core.ScreenH, gfx.Alpha(gfx.ColInkDeep, 0.72))

	gfx.TextCenterOutline(dst, "STAGE CLEAR", core.ScreenW/2, 26, gfx.ColGoldLight, gfx.ColInkDeep, 3)
	gfx.TextCenter(dst, fmt.Sprintf("%d-%d   %s", s.def.World+1, s.def.Index+1, s.def.Name),
		core.ScreenW/2, 56, gfx.ColCream, 1)

	gfx.Panel(dst, 90, 76, 300, 116, gfx.Alpha(gfx.ColInkDeep, 0.9), gfx.ColCoral)
	for i := 0; i < s.shown && i < len(s.rows); i++ {
		r := s.rows[i]
		y := 88 + i*17
		col := gfx.ColCream
		if r.gold {
			col = gfx.ColGoldLight
		}
		gfx.Text(dst, r.label, 104, y, gfx.Alpha(col, 0.8), 1)
		gfx.Text(dst, r.value, 300, y, col, 1)
	}

	// the rescued model floats in beside the tally
	if s.def.Model >= 0 && s.shown >= len(s.rows) {
		bob := math.Sin(float64(s.t)*0.06) * 4
		gfx.DrawSprite(dst, gfx.ImgModel[s.def.Model][(s.t/14)%2], core.ScreenW/2-16, 196+bob,
			gfx.SpriteOpt{ScaleX: 2, ScaleY: 2})
	}
	if (s.t/22)%2 == 0 {
		gfx.TextCenter(dst, "PRESS "+g.Pad.ConfirmLabel(), core.ScreenW/2, 244, gfx.ColGoldLight, 1)
	}
}

// ---------------------------------------------------------------------------
// GameOverScene
// ---------------------------------------------------------------------------

type GameOverScene struct {
	t   int
	sel int
}

var gameOverMenu = []string{"CONTINUE (3 LIVES)", "QUIT TO MAP"}

func (s *GameOverScene) Enter(g *Game) {
	sfx.StopMusic()
	sfx.PlayMusic(sfx.MusGameOver)
}

func (s *GameOverScene) Update(g *Game) error {
	s.t++
	if s.t < 90 {
		return nil
	}
	if g.Pad.MenuUp() || g.Pad.MenuDown() {
		s.sel = 1 - s.sel
		sfx.Play(sfx.SndMenuMove)
	}
	if g.Pad.ConfirmPressed() {
		sfx.Play(sfx.SndMenuPick)
		g.Run.Lives = 3
		g.Run.Power = core.PowerSmall
		g.PersistRun()
		if s.sel == 0 {
			g.Go(&PlayScene{})
		} else {
			g.Go(&MapScene{})
		}
	}
	return nil
}

func (s *GameOverScene) Draw(g *Game, dst *ebiten.Image) {
	dst.Fill(gfx.ColInkDeep)
	// a slow drift of falling tokens, like the run leaking away
	for i := 0; i < 30; i++ {
		x := float64((i*97)%core.ScreenW) + 4
		y := math.Mod(float64(s.t)*0.6+float64(i*37), core.ScreenH+40) - 20
		gfx.FillRectF(dst, x, y, 2, 3, gfx.Alpha(gfx.ColGold, 0.16))
	}
	gfx.TextCenterOutline(dst, "DISTILLED", core.ScreenW/2, 78, gfx.ColRed, gfx.ColInkDeep, 4)
	gfx.TextCenter(dst, "THE SYNDICATE GOT WHAT IT CAME FOR.", core.ScreenW/2, 118, gfx.Alpha(gfx.ColCream, 0.7), 1)
	gfx.TextCenter(dst, "THIS TIME.", core.ScreenW/2, 130, gfx.Alpha(gfx.ColCream, 0.5), 1)

	if s.t >= 90 {
		for i, item := range gameOverMenu {
			y := 172 + i*18
			col := gfx.Alpha(gfx.ColCream, 0.7)
			if i == s.sel {
				col = gfx.ColGoldLight
				if (s.t/16)%2 == 0 {
					gfx.Text(dst, ">", core.ScreenW/2-gfx.TextWidth(item, 1)/2-14, y, gfx.ColCoral, 1)
				}
			}
			gfx.TextCenter(dst, item, core.ScreenW/2, y, col, 1)
		}
		gfx.TextCenter(dst, fmt.Sprintf("SCORE %07d", g.Run.Score), core.ScreenW/2, 220, gfx.Alpha(gfx.ColCream, 0.6), 1)
	}
}

// ---------------------------------------------------------------------------
// OptionsScene
// ---------------------------------------------------------------------------

type OptionsScene struct {
	t    int
	sel  int
	back func() Scene
}

func (s *OptionsScene) Enter(g *Game) {
	if s.back == nil {
		s.back = func() Scene { return &TitleScene{} }
	}
}

// Option rows, in screen order.
const (
	optMusic = iota
	optSFX
	optScanlines
	optTimer
	optFullscreen
	optControls
	optBack
	optCount
)

func (s *OptionsScene) Update(g *Game) error {
	s.t++
	const n = optCount
	if g.Pad.MenuUp() {
		s.sel = (s.sel + n - 1) % n
		sfx.Play(sfx.SndMenuMove)
	}
	if g.Pad.MenuDown() {
		s.sel = (s.sel + 1) % n
		sfx.Play(sfx.SndMenuMove)
	}

	adj := 0
	if g.Pad.MenuLeft() {
		adj = -1
	}
	if g.Pad.MenuRight() {
		adj = 1
	}
	st := &g.Save.Settings
	if adj != 0 {
		switch s.sel {
		case optMusic:
			st.MusicVolume = clamp01(st.MusicVolume + float64(adj)*0.1)
			sfx.SetMusicVolume(st.MusicVolume)
		case optSFX:
			st.SFXVolume = clamp01(st.SFXVolume + float64(adj)*0.1)
			sfx.SetSFXVolume(st.SFXVolume)
			sfx.Play(sfx.SndToken)
		case optScanlines:
			st.Scanlines = !st.Scanlines
		case optTimer:
			st.ShowTimer = !st.ShowTimer
		case optFullscreen:
			st.Fullscreen = !st.Fullscreen
			ebiten.SetFullscreen(st.Fullscreen)
		}
		if s.sel != optSFX {
			sfx.Play(sfx.SndMenuMove)
		}
	}

	if g.Pad.ConfirmPressed() {
		switch s.sel {
		case optControls:
			sfx.Play(sfx.SndMenuPick)
			g.PersistSettingsOnly()
			// come back to this same options screen afterwards
			g.Go(&ControlsScene{back: func() Scene { return s }})
			return nil
		case optBack:
			sfx.Play(sfx.SndMenuPick)
			g.PersistSettingsOnly()
			g.Go(s.back())
			return nil
		}
	}
	if g.Pad.BackPressed() {
		g.PersistSettingsOnly()
		g.Go(s.back())
	}
	return nil
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func (s *OptionsScene) Draw(g *Game, dst *ebiten.Image) {
	dst.Fill(gfx.ColInkDeep)
	gfx.Backgrounds[1].Draw(dst, float64(s.t)*0.2, 0)
	gfx.FillRectF(dst, 0, 0, core.ScreenW, core.ScreenH, gfx.Alpha(gfx.ColInkDeep, 0.75))

	gfx.TextCenterOutline(dst, "OPTIONS", core.ScreenW/2, 22, gfx.ColCream, gfx.ColInkDeep, 3)
	gfx.Panel(dst, 60, 56, 360, 156, gfx.Alpha(gfx.ColInkDeep, 0.9), gfx.ColCoral)

	st := g.Save.Settings
	rows := []struct {
		label string
		val   string
		bar   float64
	}{
		{"MUSIC VOLUME", fmt.Sprintf("%d%%", int(st.MusicVolume*100+0.5)), st.MusicVolume},
		{"SOUND VOLUME", fmt.Sprintf("%d%%", int(st.SFXVolume*100+0.5)), st.SFXVolume},
		{"SCANLINES", onOff(st.Scanlines), -1},
		{"SHOW TIMER", onOff(st.ShowTimer), -1},
		{"FULLSCREEN", onOff(st.Fullscreen), -1},
		{"CONTROLS", "KEYBOARD & GAMEPAD >", -1},
		{"BACK", "", -1},
	}
	for i, r := range rows {
		y := 70 + i*20
		col := gfx.Alpha(gfx.ColCream, 0.75)
		if i == s.sel {
			gfx.FillRectF(dst, 66, float64(y)-3, 348, 15, gfx.Alpha(gfx.ColCoral, 0.20))
			col = gfx.ColGoldLight
			if (s.t/16)%2 == 0 {
				gfx.Text(dst, ">", 74, y, gfx.ColCoral, 1)
			}
		}
		gfx.Text(dst, r.label, 88, y, col, 1)
		if r.bar >= 0 {
			gfx.FillRectF(dst, 240, float64(y), 100, 7, gfx.Alpha(gfx.ColSlate, 0.9))
			gfx.FillRectF(dst, 240, float64(y), 100*r.bar, 7, gfx.ColCoral)
			gfx.StrokeRectF(dst, 240, float64(y), 100, 7, gfx.Alpha(gfx.ColCream, 0.5))
			gfx.Text(dst, r.val, 348, y, col, 1)
		} else if r.val != "" {
			gfx.Text(dst, r.val, 240, y, col, 1)
		}
	}

	gfx.TextCenter(dst, g.Pad.Hints("LEFT/RIGHT ADJUST", g.Pad.Prompt(ActConfirm, "SELECT"), g.Pad.Prompt(ActBack, "BACK")), core.ScreenW/2, 222, gfx.Alpha(gfx.ColCream, 0.45), 1)
	p, err := savePathForDisplay()
	if err == nil {
		gfx.TextCenter(dst, "SAVE: "+p, core.ScreenW/2, 244, gfx.Alpha(gfx.ColCream, 0.28), 1)
	}
}

func onOff(b bool) string {
	if b {
		return "ON"
	}
	return "OFF"
}

// ---------------------------------------------------------------------------
// CreditsScene
// ---------------------------------------------------------------------------

type CreditsScene struct {
	t      int
	scroll float64
}

var creditLines = []string{
	"S U P E R   D A R I O",
	"",
	"SAVE THE MODELS",
	"",
	"",
	"THE SYNDICATE, DEFEATED",
	"",
	"SCRAPEMASTER PRIME",
	"THE SCRAPER BARON",
	"THE LEVIATHAN",
	"THE LUNAR LORD",
	"THE CLOUDLORD",
	"THE DISTILLER",
	"",
	"",
	"THE MODELS, RECOVERED",
	"",
	"HAIKU",
	"SONNET",
	"FABLE",
	"OPUS",
	"",
	"",
	"MADE BY",
	"",
	"DYNO",
	"DESIGN, DIRECTION & PRODUCTION",
	"",
	"CLAUDE OPUS 5",
	"CODE, PIXELS & SOUNDTRACK",
	"",
	"",
	"BUILT IN GO WITH EBITENGINE",
	"",
	"EVERY SPRITE, TILE AND NOTE",
	"IN THIS GAME IS GENERATED",
	"BY CODE AT STARTUP.",
	"NO ASSET FILES WERE HARMED.",
	"",
	"",
	"THIS IS A PARODY.",
	"",
	"AN UNOFFICIAL FAN GAME, NOT",
	"AFFILIATED WITH, ENDORSED BY,",
	"OR CONNECTED TO ANY REAL",
	"COMPANY OR ANY REAL PERSON.",
	"",
	"EVERY CHARACTER HERE IS",
	"FICTIONAL. NOTHING DEPICTED",
	"IS A CLAIM ABOUT ANYONE.",
	"",
	"",
	"THANK YOU FOR PLAYING.",
	"",
	"",
	"",
}

func (s *CreditsScene) Enter(g *Game) {
	sfx.PlayMusic(sfx.MusEnding)
	g.Slot().Finished = true
	g.PersistRun()
}

func (s *CreditsScene) Update(g *Game) error {
	s.t++
	s.scroll += 0.35
	if g.Pad.RunDown() {
		s.scroll += 1.6
	}
	end := float64(len(creditLines)*14) + 60
	if s.scroll > end || g.Pad.BackPressed() {
		g.Go(&TitleScene{})
	}
	return nil
}

func (s *CreditsScene) Draw(g *Game, dst *ebiten.Image) {
	dst.Fill(gfx.ColInkDeep)
	// a quiet starfield
	for i := 0; i < 70; i++ {
		x := float64((i * 149) % core.ScreenW)
		y := math.Mod(float64((i*83)%core.ScreenH)+s.scroll*0.25, core.ScreenH)
		a := 0.2 + 0.4*math.Abs(math.Sin(float64(s.t)*0.02+float64(i)))
		gfx.FillRectF(dst, x, y, 1, 1, gfx.Alpha(gfx.ColCream, a))
	}

	for i, line := range creditLines {
		y := float64(core.ScreenH) + float64(i)*14 - s.scroll
		if y < -20 || y > core.ScreenH+20 {
			continue
		}
		scale := 1
		col := gfx.Alpha(gfx.ColCream, 0.85)
		switch {
		case i == 0:
			scale, col = 3, gfx.ColCoral
		case i == 2:
			scale, col = 2, gfx.ColGoldLight
		case line == "THE SYNDICATE, DEFEATED" || line == "THE MODELS, RECOVERED" || line == "MADE BY":
			col = gfx.ColGoldLight
		case line == "DYNO" || line == "CLAUDE OPUS 5":
			scale, col = 2, gfx.ColCoral
		case line == "DESIGN, DIRECTION & PRODUCTION" || line == "CODE, PIXELS & SOUNDTRACK":
			col = gfx.Alpha(gfx.ColCream, 0.55)
		}
		gfx.TextCenterShadow(dst, line, core.ScreenW/2, int(y), col, gfx.ColInkDeep, scale)
	}

	// the four models drift up alongside the credits
	for m := 0; m < 4; m++ {
		x := 40.0 + float64(m)*135
		y := math.Mod(float64(core.ScreenH)+float64(m)*90-s.scroll*0.7, float64(core.ScreenH)+120) - 40
		gfx.DrawSprite(dst, gfx.ImgModel[m][(s.t/14+m)%2], x, y, gfx.SpriteOpt{Alpha: 0.55})
	}

	gfx.TextCenter(dst, g.Pad.Hints("HOLD "+g.Pad.Label(ActRun)+" TO SPEED UP", g.Pad.Prompt(ActBack, "TO EXIT")), core.ScreenW/2, 262, gfx.Alpha(gfx.ColCream, 0.3), 1)
}
