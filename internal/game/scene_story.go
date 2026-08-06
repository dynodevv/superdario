package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"superdario/internal/core"
	"superdario/internal/gfx"
	"superdario/internal/sfx"
)

// StoryScene plays a cutscene: an illustration, a typed-out caption, and a
// prompt to continue.
type StoryScene struct {
	cut   Cutscene
	next  func() Scene
	panel int
	chars int // characters revealed so far, for the typewriter
	t     int
	done  bool
}

// NewStoryScene builds a cutscene that hands off to next when it finishes.
func NewStoryScene(c Cutscene, next func() Scene) *StoryScene {
	if next == nil {
		next = func() Scene { return &MapScene{} }
	}
	return &StoryScene{cut: c, next: next}
}

func (s *StoryScene) Enter(g *Game) {
	sfx.PlayMusic(sfx.MusStory)
	s.panel = 0
	s.chars = 0
}

// totalChars is how many characters the current panel has to reveal.
func (s *StoryScene) totalChars() int {
	if s.panel >= len(s.cut.Panels) {
		return 0
	}
	n := 0
	for _, l := range s.cut.Panels[s.panel].Lines {
		n += len(l)
	}
	return n
}

func (s *StoryScene) Update(g *Game) error {
	s.t++
	if len(s.cut.Panels) == 0 {
		g.Go(s.next())
		return nil
	}

	total := s.totalChars()
	if s.chars < total {
		s.chars += 2
		if s.chars > total {
			s.chars = total
		}
		if s.t%3 == 0 {
			sfx.Play(sfx.SndBlip)
		}
	}

	advance := g.Pad.ConfirmPressed() || g.Pad.JumpPressed()
	skip := g.Pad.BackPressed()

	if skip {
		g.Go(s.next())
		return nil
	}
	if advance {
		if s.chars < total {
			s.chars = total // first press finishes the typewriter
			return nil
		}
		s.panel++
		s.chars = 0
		sfx.Play(sfx.SndMenuPick)
		if s.panel >= len(s.cut.Panels) {
			g.Go(s.next())
		}
	}
	return nil
}

func (s *StoryScene) Draw(g *Game, dst *ebiten.Image) {
	dst.Fill(color.RGBA{0x0a, 0x08, 0x10, 0xff})
	if s.panel >= len(s.cut.Panels) {
		return
	}
	p := s.cut.Panels[s.panel]

	// --- illustration ---------------------------------------------------
	drawPanelArt(dst, p.Art, s.cut.Model, s.t)

	// --- caption box ----------------------------------------------------
	// Sized to its content and anchored above the prompt row. A fixed height
	// clipped the longest panels, whose last line ran into the prompt.
	const (
		lineH     = 11
		promptY   = 259
		boxTopMin = 88.0
	)
	boxH := float64(13 + len(p.Lines)*lineH)
	boxY := float64(promptY-5) - boxH
	if boxY < boxTopMin {
		boxY = boxTopMin
		boxH = float64(promptY-5) - boxTopMin
	}
	gfx.Panel(dst, 24, boxY, core.ScreenW-48, boxH, gfx.Alpha(gfx.ColInkDeep, 0.92), gfx.ColCoral)

	if name := speakerName(p.Speak); name != "" {
		nameCol := gfx.ColGoldLight
		if p.Speak >= 0 && p.Speak < 4 {
			nameCol = gfx.ModelColors[p.Speak]
		} else if p.Speak == 5 {
			nameCol = gfx.ColRed
		}
		w := gfx.TextWidth(name, 1) + 10
		gfx.Panel(dst, 30, boxY-9, float64(w), 12, gfx.ColInkDeep, nameCol)
		gfx.Text(dst, name, 35, int(boxY)-6, nameCol, 1)
	}

	// typewriter reveal, spending the character budget line by line
	left := s.chars
	for i, line := range p.Lines {
		if left <= 0 {
			break
		}
		shown := line
		if left < len(line) {
			shown = line[:left]
		}
		left -= len(line)
		gfx.Text(dst, shown, 34, int(boxY)+7+i*lineH, gfx.ColCream, 1)
	}

	// --- prompt ---------------------------------------------------------
	if s.chars >= s.totalChars() && (s.t/20)%2 == 0 {
		gfx.TextShadow(dst, ">", core.ScreenW-46, int(boxY+boxH)-15, gfx.ColGoldLight, gfx.ColInkDeep, 2)
	}
	gfx.Text(dst, g.Pad.Hints(
		g.Pad.Prompt(ActConfirm, "NEXT"),
		g.Pad.Prompt(ActBack, "SKIP"),
	), 30, promptY, gfx.Alpha(gfx.ColCream, 0.40), 1)
}

// drawPanelArt renders the simple illustrations used by the cutscenes.
func drawPanelArt(dst *ebiten.Image, art PanelArt, model int, t int) {
	cx := float64(core.ScreenW / 2)
	switch art {

	case ArtHQ:
		// night sky, then a dark office block with one lit window
		gfx.VGradient(dst, 0, 0, core.ScreenW, 170, color.RGBA{0x06, 0x08, 0x18, 0xff}, color.RGBA{0x18, 0x1e, 0x38, 0xff})
		for i := 0; i < 60; i++ {
			x := float64((i*137)%core.ScreenW) + 0.5
			y := float64((i * 71) % 120)
			a := 0.3 + 0.5*math.Abs(math.Sin(float64(t)*0.02+float64(i)))
			gfx.FillRectF(dst, x, y, 1, 1, gfx.Alpha(gfx.ColCream, a))
		}
		gfx.FillRectF(dst, cx-70, 46, 140, 124, color.RGBA{0x14, 0x14, 0x1e, 0xff})
		gfx.StrokeRectF(dst, cx-70, 46, 140, 124, color.RGBA{0x2a, 0x2a, 0x3a, 0xff})
		for r := 0; r < 7; r++ {
			for c := 0; c < 7; c++ {
				lit := (r*7+c)%11 == 3
				col := color.RGBA{0x20, 0x20, 0x2c, 0xff}
				if lit {
					col = gfx.Scale(gfx.ColGoldLight, 0.9)
				}
				gfx.FillRectF(dst, cx-58+float64(c)*17, 58+float64(r)*16, 10, 10, col)
			}
		}
		gfx.DrawSprite(dst, gfx.ImgLogo, cx-16, 12, gfx.SpriteOpt{ScaleX: 0.5, ScaleY: 0.5, Alpha: 0.85})

	case ArtTheft:
		gfx.VGradient(dst, 0, 0, core.ScreenW, 170, color.RGBA{0x10, 0x0a, 0x14, 0xff}, color.RGBA{0x30, 0x18, 0x20, 0xff})
		// crates on a conveyor, each holding a model
		for i := 0; i < 4; i++ {
			x := cx - 150 + float64(i)*80 + math.Mod(float64(t)*0.4, 80)
			if x > core.ScreenW+40 {
				x -= 400
			}
			gfx.FillRectF(dst, x-20, 92, 40, 40, color.RGBA{0x4a, 0x35, 0x28, 0xff})
			gfx.StrokeRectF(dst, x-20, 92, 40, 40, color.RGBA{0x27, 0x1a, 0x12, 0xff})
			gfx.DrawSprite(dst, gfx.ImgModel[i%4][(t/16)%2], x-8, 104, gfx.SpriteOpt{Alpha: 0.85})
			gfx.DrawSprite(dst, gfx.ImgCage, x-8, 100, gfx.SpriteOpt{Alpha: 0.7})
		}
		gfx.FillRectF(dst, 0, 132, core.ScreenW, 10, color.RGBA{0x28, 0x28, 0x30, 0xff})
		for x := math.Mod(float64(t)*1.5, 12) - 12; x < core.ScreenW; x += 12 {
			gfx.FillRectF(dst, x, 134, 6, 6, color.RGBA{0x40, 0x40, 0x4c, 0xff})
		}

	case ArtSyndicate:
		gfx.VGradient(dst, 0, 0, core.ScreenW, 170, color.RGBA{0x0c, 0x06, 0x0a, 0xff}, color.RGBA{0x28, 0x10, 0x14, 0xff})
		// five silhouettes at a long table
		for i := 0; i < 5; i++ {
			x := cx - 160 + float64(i)*80
			h := 54.0 + float64((i*13)%12)
			sway := math.Sin(float64(t)*0.02+float64(i)) * 1.5
			gfx.FillRectF(dst, x-16+sway, 130-h, 32, h, color.RGBA{0x0a, 0x06, 0x08, 0xff})
			gfx.FillRectF(dst, x-11+sway, 122-h, 22, 18, color.RGBA{0x0a, 0x06, 0x08, 0xff})
			// glowing eyes
			blink := (t/40+i*7)%50 > 3
			if blink {
				gfx.FillRectF(dst, x-6+sway, 116-h, 4, 2, gfx.ColRed)
				gfx.FillRectF(dst, x+3+sway, 116-h, 4, 2, gfx.ColRed)
			}
		}
		gfx.FillRectF(dst, 0, 130, core.ScreenW, 8, color.RGBA{0x30, 0x18, 0x16, 0xff})
		gfx.FillRectF(dst, 0, 130, core.ScreenW, 1, color.RGBA{0x60, 0x30, 0x28, 0xff})

	case ArtDario:
		gfx.VGradient(dst, 0, 0, core.ScreenW, 170, color.RGBA{0x1a, 0x12, 0x18, 0xff}, color.RGBA{0x3a, 0x22, 0x1c, 0xff})
		// a spotlight and Dario, large
		for r := 90.0; r > 0; r -= 6 {
			gfx.FillRectF(dst, cx-r, 20, r*2, 150, gfx.Alpha(gfx.ColGoldLight, 0.012))
		}
		gfx.DrawSprite(dst, gfx.Dario[1][gfx.FrStand], cx-32, 50, gfx.SpriteOpt{ScaleX: 4, ScaleY: 4})

	case ArtModel:
		idx := 0
		if model >= 0 {
			idx = model
		}
		col := gfx.ModelColors[idx%4]
		gfx.VGradient(dst, 0, 0, core.ScreenW, 170, gfx.Scale(col, 0.10), gfx.Scale(col, 0.30))
		pulse := 0.5 + 0.5*math.Sin(float64(t)*0.05)
		for r := 80.0; r > 10; r -= 8 {
			gfx.FillRectF(dst, cx-r, 90-r*0.8, r*2, r*1.6, gfx.Alpha(col, 0.02+0.02*pulse))
		}
		bob := math.Sin(float64(t)*0.04) * 4
		gfx.DrawSprite(dst, gfx.ImgModel[idx%4][(t/14)%2], cx-32, 56+bob, gfx.SpriteOpt{ScaleX: 4, ScaleY: 4})
		gfx.TextCenterOutline(dst, gfx.ModelNames[idx%4], int(cx), 140, col, gfx.ColInkDeep, 2)

	case ArtTower:
		gfx.VGradient(dst, 0, 0, core.ScreenW, 170, color.RGBA{0x14, 0x04, 0x08, 0xff}, color.RGBA{0x5a, 0x14, 0x10, 0xff})
		// a monolithic tower with a furnace mouth
		gfx.FillRectF(dst, cx-42, 20, 84, 150, color.RGBA{0x1c, 0x10, 0x14, 0xff})
		gfx.StrokeRectF(dst, cx-42, 20, 84, 150, color.RGBA{0x40, 0x22, 0x22, 0xff})
		for r := 0; r < 8; r++ {
			for c := 0; c < 4; c++ {
				flick := 0.6 + 0.4*math.Sin(float64(t)*0.1+float64(r*4+c))
				gfx.FillRectF(dst, cx-32+float64(c)*19, 32+float64(r)*17, 10, 9, gfx.Scale(gfx.ColGold, flick))
			}
		}
		glow := 0.6 + 0.4*math.Sin(float64(t)*0.06)
		gfx.FillRectF(dst, cx-30, 150, 60, 20, gfx.Scale(color.RGBA{0xff, 0x6b, 0x30, 0xff}, glow))
		for i := 0; i < 20; i++ {
			x := cx - 30 + float64((i*37)%60)
			y := 150 - math.Mod(float64(t)*0.8+float64(i*13), 120)
			gfx.FillRectF(dst, x, y, 2, 2, gfx.Alpha(gfx.ColGoldLight, 0.7))
		}

	case ArtFreed:
		gfx.VGradient(dst, 0, 0, core.ScreenW, 170, color.RGBA{0x1c, 0x30, 0x50, 0xff}, color.RGBA{0x8a, 0x9e, 0xb4, 0xff})
		for i := 0; i < 4; i++ {
			x := cx - 132 + float64(i)*88
			bob := math.Sin(float64(t)*0.05+float64(i)*1.2) * 5
			col := gfx.ModelColors[i]
			for r := 30.0; r > 6; r -= 5 {
				gfx.FillRectF(dst, x-r/2, 84+bob-r/2, r, r, gfx.Alpha(col, 0.05))
			}
			gfx.DrawSprite(dst, gfx.ImgModel[i][(t/14+i)%2], x-16, 70+bob, gfx.SpriteOpt{ScaleX: 2, ScaleY: 2})
			gfx.TextCenter(dst, gfx.ModelNames[i], int(x), 118, col, 1)
		}
		gfx.DrawSprite(dst, gfx.Dario[1][gfx.FrCheer], cx-16, 118, gfx.SpriteOpt{ScaleX: 2, ScaleY: 2})

	default:
		gfx.VGradient(dst, 0, 0, core.ScreenW, 170, color.RGBA{0x0a, 0x08, 0x10, 0xff}, color.RGBA{0x18, 0x14, 0x22, 0xff})
	}
}
