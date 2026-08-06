package game

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"superdario/internal/core"
	"superdario/internal/gfx"
	"superdario/internal/sfx"
)

// TitleScene: the attract screen.
type TitleScene struct {
	t        int
	sel      int
	camX     float64
	darioX   float64
	darioDir int
}

var titleMenu = []string{"START GAME", "OPTIONS", "QUIT"}

func (s *TitleScene) Enter(g *Game) {
	s.darioX = -20
	s.darioDir = 1
	sfx.PlayMusic(sfx.MusTitle)
}

func (s *TitleScene) Update(g *Game) error {
	s.t++
	s.camX += 0.35
	s.darioX += float64(s.darioDir) * 0.75
	if s.darioX > core.ScreenW+20 {
		s.darioX = -20
	}
	gfx.Backgrounds[0].Update()

	if g.Pad.MenuUp() {
		s.sel = (s.sel + len(titleMenu) - 1) % len(titleMenu)
		sfx.Play(sfx.SndMenuMove)
	}
	if g.Pad.MenuDown() {
		s.sel = (s.sel + 1) % len(titleMenu)
		sfx.Play(sfx.SndMenuMove)
	}
	if g.Pad.ConfirmPressed() {
		sfx.Play(sfx.SndMenuPick)
		switch s.sel {
		case 0:
			g.Go(&FileScene{})
		case 1:
			g.Go(&OptionsScene{back: func() Scene { return &TitleScene{} }})
		case 2:
			g.Quit()
		}
	}
	return nil
}

func (s *TitleScene) Draw(g *Game, dst *ebiten.Image) {
	gfx.Backgrounds[0].Draw(dst, s.camX, 0)

	// darken the lower half so the menu reads
	gfx.VGradient(dst, 0, 120, core.ScreenW, 150, gfx.Alpha(gfx.ColInkDeep, 0), gfx.Alpha(gfx.ColInkDeep, 0.72))

	// Dario strolls along the horizon
	fr := gfx.FrWalk1
	switch (s.t / 8) % 4 {
	case 1, 3:
		fr = gfx.FrWalk2
	case 2:
		fr = gfx.FrWalk3
	}
	gfx.DrawSprite(dst, gfx.Dario[1][fr], s.darioX, 214, gfx.SpriteOpt{})

	// emblem
	bob := math.Sin(float64(s.t)*0.04) * 3
	gfx.DrawSprite(dst, gfx.ImgLogo, core.ScreenW/2-32, 14+bob, gfx.SpriteOpt{})

	// title
	gfx.TextCenterOutline(dst, "SUPER", core.ScreenW/2-2, 86, gfx.ColCream, gfx.ColInkDeep, 4)
	gfx.TextCenterOutline(dst, "DARIO", core.ScreenW/2+2, 120, gfx.ColCoral, gfx.ColInkDeep, 5)
	gfx.TextCenterShadow(dst, "- SAVE THE MODELS -", core.ScreenW/2, 160, gfx.ColGoldLight, gfx.ColInkDeep, 1)

	// menu
	for i, item := range titleMenu {
		y := 194 + i*16
		col := gfx.Alpha(gfx.ColCream, 0.72)
		if i == s.sel {
			col = gfx.ColGoldLight
			// blinking cursor
			if (s.t/16)%2 == 0 {
				gfx.TextShadow(dst, ">", core.ScreenW/2-gfx.TextWidth(item, 2)/2-16, y, gfx.ColCoral, gfx.ColInkDeep, 2)
			}
		}
		gfx.TextCenterShadow(dst, item, core.ScreenW/2, y, col, gfx.ColInkDeep, 2)
	}

	gfx.TextCenter(dst, "ARROWS / WASD MOVE   Z OR SPACE JUMP   X OR SHIFT RUN & THROW", core.ScreenW/2, 244, gfx.Alpha(gfx.ColCream, 0.55), 1)
	gfx.TextCenter(dst, "PARODY. AN UNOFFICIAL FAN GAME.", core.ScreenW/2, 253, gfx.Alpha(gfx.ColCoral, 0.75), 1)
	gfx.TextCenter(dst, "NOT AFFILIATED WITH ANY REAL COMPANY OR PERSON. ALL CHARACTERS FICTIONAL.",
		core.ScreenW/2, 262, gfx.Alpha(gfx.ColCream, 0.36), 1)
}
