package game

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"superdario/internal/core"
	"superdario/internal/gfx"
	"superdario/internal/save"
	"superdario/internal/sfx"
	"superdario/internal/world"
)

// MapScene is the level select: five world islands, three stages each.
type MapScene struct {
	t        int
	selW     int
	selL     int
	camX     float64
	markerX  float64
	markerY  float64
	firstRun bool
}

const (
	islandX0 = 56.0
	islandDX = 92.0
	nodeY0   = 104.0
	nodeDY   = 34.0
)

func nodePos(w, l int) (float64, float64) {
	return islandX0 + float64(w)*islandDX, nodeY0 + float64(l)*nodeDY
}

func (s *MapScene) Enter(g *Game) {
	sfx.PlayMusic(sfx.MusMap)
	s.selW = g.Run.World
	s.selL = g.Run.Level
	x, y := nodePos(s.selW, s.selL)
	s.markerX, s.markerY = x, y
	g.PersistRun()
}

func (s *MapScene) Update(g *Game) error {
	s.t++
	s.camX += 0.15
	gfx.Backgrounds[s.selW].Update()

	slot := g.Slot()

	move := func(dw, dl int) {
		nw := s.selW + dw
		nl := s.selL + dl
		if nl < 0 {
			nl = 0
		}
		if nl >= save.NumLevels {
			nl = save.NumLevels - 1
		}
		if nw < 0 || nw >= save.NumWorlds {
			return
		}
		s.selW, s.selL = nw, nl
		sfx.Play(sfx.SndMenuMove)
	}

	if g.Pad.MenuLeft() {
		move(-1, 0)
	}
	if g.Pad.MenuRight() {
		move(1, 0)
	}
	if g.Pad.MenuUp() {
		move(0, -1)
	}
	if g.Pad.MenuDown() {
		move(0, 1)
	}

	// glide the marker to the selection
	tx, ty := nodePos(s.selW, s.selL)
	s.markerX += (tx - s.markerX) * 0.22
	s.markerY += (ty - s.markerY) * 0.22

	if g.Pad.ConfirmPressed() {
		if slot.IsUnlocked(s.selW, s.selL) {
			sfx.Play(sfx.SndMenuPick)
			g.Run.World, g.Run.Level = s.selW, s.selL
			g.Go(&PlayScene{})
		} else {
			sfx.Play(sfx.SndBump)
		}
	}
	// Esc is bound to both "back" and "pause"; on the map it means back, so
	// return rather than letting the options check below also fire.
	if g.Pad.BackPressed() {
		sfx.Play(sfx.SndMenuMove)
		g.PersistRun()
		g.Go(&TitleScene{})
		return nil
	}
	if g.Pad.OptionsPressed() {
		sfx.Play(sfx.SndMenuPick)
		g.Go(&OptionsScene{back: func() Scene { return &MapScene{} }})
		return nil
	}
	return nil
}

func (s *MapScene) Draw(g *Game, dst *ebiten.Image) {
	gfx.Backgrounds[s.selW].Draw(dst, s.camX, 0)
	gfx.FillRectF(dst, 0, 0, core.ScreenW, core.ScreenH, gfx.Alpha(gfx.ColInkDeep, 0.55))

	slot := g.Slot()

	// header
	gfx.VGradient(dst, 0, 0, core.ScreenW, 28, gfx.Alpha(gfx.ColInkDeep, 0.9), gfx.Alpha(gfx.ColInkDeep, 0.4))
	gfx.TextShadow(dst, "SELECT A STAGE", 10, 8, gfx.ColCream, gfx.ColInkDeep, 2)

	// lives / score / rescued
	gfx.DrawSprite(dst, gfx.ImgHeart, 300, 6, gfx.SpriteOpt{})
	gfx.TextShadow(dst, fmt.Sprintf("x%d", g.Run.Lives), 313, 9, gfx.ColCream, gfx.ColInkDeep, 1)
	gfx.TextShadow(dst, fmt.Sprintf("SCORE %07d", g.Run.Score), 340, 9, gfx.ColGoldLight, gfx.ColInkDeep, 1)
	for m := 0; m < 4; m++ {
		x := 340.0 + float64(m)*18
		if g.Run.Rescued[m] {
			gfx.DrawSprite(dst, gfx.ImgModel[m][(s.t/16)%2], x, 18, gfx.SpriteOpt{})
		} else {
			gfx.DrawSprite(dst, gfx.ImgLock, x+3, 20, gfx.SpriteOpt{Alpha: 0.4})
		}
	}

	// --- paths between islands ------------------------------------------
	for w := 0; w < save.NumWorlds-1; w++ {
		x0, y0 := nodePos(w, save.NumLevels-1)
		x1, y1 := nodePos(w+1, 0)
		unlocked := slot.IsUnlocked(w+1, 0)
		col := gfx.Alpha(gfx.ColCream, 0.18)
		if unlocked {
			col = gfx.Alpha(gfx.ColGoldLight, 0.55)
		}
		steps := 14
		for i := 0; i <= steps; i++ {
			f := float64(i) / float64(steps)
			// arc downward then up to the next island
			x := x0 + (x1-x0)*f
			y := y0 + (y1-y0)*f + math.Sin(f*math.Pi)*22
			gfx.FillRectF(dst, x-1, y-1, 2, 2, col)
		}
	}
	// paths within an island
	for w := 0; w < save.NumWorlds; w++ {
		for l := 0; l < save.NumLevels-1; l++ {
			x, y := nodePos(w, l)
			col := gfx.Alpha(gfx.ColCream, 0.18)
			if slot.IsUnlocked(w, l+1) {
				col = gfx.Alpha(gfx.ColGoldLight, 0.55)
			}
			for i := 0; i < 8; i++ {
				gfx.FillRectF(dst, x-1, y+6+float64(i)*3.4, 2, 2, col)
			}
		}
	}

	// --- nodes -----------------------------------------------------------
	for w := 0; w < save.NumWorlds; w++ {
		theme := gfx.Themes[w]
		for l := 0; l < save.NumLevels; l++ {
			x, y := nodePos(w, l)
			def := world.LevelAt(w, l)
			unlocked := slot.IsUnlocked(w, l)
			cleared := slot.Cleared[w][l]

			r := 9.0
			fill := gfx.Alpha(gfx.ColInkDeep, 0.85)
			edge := gfx.Alpha(gfx.ColCream, 0.3)
			if unlocked {
				fill = gfx.Scale(theme.Accent, 0.55)
				edge = theme.Accent
			}
			if cleared {
				fill = gfx.Scale(theme.SurfFace, 0.9)
			}
			if def.Boss >= 0 {
				// boss nodes are square and red-rimmed
				gfx.FillRectF(dst, x-r, y-r, r*2, r*2, fill)
				gfx.StrokeRectF(dst, x-r, y-r, r*2, r*2, edge)
				if unlocked {
					gfx.StrokeRectF(dst, x-r-1, y-r-1, r*2+2, r*2+2, gfx.Alpha(gfx.ColRed, 0.8))
				}
			} else {
				gfx.FillRectF(dst, x-r, y-r+2, r*2, r*2-4, fill)
				gfx.FillRectF(dst, x-r+2, y-r, r*2-4, r*2, fill)
				gfx.StrokeRectF(dst, x-r, y-r+2, r*2, r*2-4, edge)
			}

			switch {
			case !unlocked:
				gfx.DrawSprite(dst, gfx.ImgLock, x-5, y-6, gfx.SpriteOpt{Alpha: 0.75})
			case cleared:
				gfx.DrawSprite(dst, gfx.ImgStarIcon, x-5, y-5, gfx.SpriteOpt{})
			default:
				gfx.TextCenter(dst, fmt.Sprintf("%d", l+1), int(x), int(y)-3, gfx.ColCream, 1)
			}
		}
		// world label
		name := fmt.Sprintf("W%d", w+1)
		gfx.TextCenterShadow(dst, name, int(islandX0+float64(w)*islandDX), 84, gfx.Alpha(gfx.ColCream, 0.8), gfx.ColInkDeep, 1)
	}

	// --- marker ----------------------------------------------------------
	bob := math.Sin(float64(s.t)*0.12) * 2
	gfx.DrawSprite(dst, gfx.Dario[0][gfx.FrStand], s.markerX-8, s.markerY-24+bob, gfx.SpriteOpt{})

	// --- info panel ------------------------------------------------------
	def := world.LevelAt(s.selW, s.selL)
	unlocked := slot.IsUnlocked(s.selW, s.selL)
	gfx.Panel(dst, 20, 210, core.ScreenW-40, 46, gfx.Alpha(gfx.ColInkDeep, 0.9), gfx.Themes[s.selW].Accent)

	title := fmt.Sprintf("%d-%d  %s", s.selW+1, s.selL+1, def.Name)
	if !unlocked {
		title = fmt.Sprintf("%d-%d  ???", s.selW+1, s.selL+1)
	}
	gfx.TextShadow(dst, title, 30, 216, gfx.ColGoldLight, gfx.ColInkDeep, 1)
	gfx.Text(dst, world.WorldNames[s.selW], 30, 228, gfx.Alpha(gfx.ColCream, 0.75), 1)
	gfx.Text(dst, world.WorldBlurbs[s.selW], 30, 240, gfx.Alpha(gfx.ColCream, 0.45), 1)

	if unlocked {
		if bt := slot.BestTime[s.selW][s.selL]; bt > 0 {
			gfx.Text(dst, fmt.Sprintf("BEST %s", fmtTicks(bt)), core.ScreenW-104, 216, gfx.ColCyan, 1)
		}
		if def.Boss >= 0 {
			gfx.Text(dst, "BOSS STAGE", core.ScreenW-104, 228, gfx.ColRed, 1)
		}
		if def.Model >= 0 && !g.Run.Rescued[def.Model] {
			gfx.Text(dst, gfx.ModelNames[def.Model]+" HELD HERE", core.ScreenW-136, 240, gfx.ModelColors[def.Model], 1)
		}
	} else {
		gfx.Text(dst, "LOCKED", core.ScreenW-104, 216, gfx.Alpha(gfx.ColRed, 0.9), 1)
	}

	gfx.TextCenter(dst, "ENTER PLAY   ESC TITLE   P OPTIONS", core.ScreenW/2, 262, gfx.Alpha(gfx.ColCream, 0.4), 1)
}

// fmtTicks renders a tick count as M:SS.
func fmtTicks(t int) string {
	sec := t / core.TPS
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}
