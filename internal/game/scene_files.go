package game

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"superdario/internal/core"
	"superdario/internal/gfx"
	"superdario/internal/save"
	"superdario/internal/sfx"
)

// FileScene is the save-slot picker. Three slots, plus an erase confirmation.
type FileScene struct {
	t       int
	sel     int // 0..2 slots, 3 = back
	erasing bool
	camX    float64
}

func (s *FileScene) Enter(g *Game) { sfx.PlayMusic(sfx.MusTitle) }

func (s *FileScene) Update(g *Game) error {
	s.t++
	s.camX += 0.2
	gfx.Backgrounds[0].Update()

	if s.erasing {
		if g.Pad.ConfirmPressed() {
			g.Save.Slots[s.sel] = save.Slot{}
			g.PersistSettingsOnly()
			s.erasing = false
			sfx.Play(sfx.SndMenuPick)
		}
		if g.Pad.BackPressed() {
			s.erasing = false
			sfx.Play(sfx.SndMenuMove)
		}
		return nil
	}

	if g.Pad.MenuUp() {
		s.sel = (s.sel + 3) % 4
		sfx.Play(sfx.SndMenuMove)
	}
	if g.Pad.MenuDown() {
		s.sel = (s.sel + 1) % 4
		sfx.Play(sfx.SndMenuMove)
	}
	// erase with the run button on an occupied slot
	if s.sel < 3 && g.Pad.FirePressed() && g.Save.Slots[s.sel].Used {
		s.erasing = true
		sfx.Play(sfx.SndMenuPick)
		return nil
	}
	if g.Pad.BackPressed() {
		sfx.Play(sfx.SndMenuMove)
		g.Go(&TitleScene{})
		return nil
	}
	if g.Pad.ConfirmPressed() {
		sfx.Play(sfx.SndMenuPick)
		if s.sel == 3 {
			g.Go(&TitleScene{})
			return nil
		}
		slot := &g.Save.Slots[s.sel]
		fresh := !slot.Used
		if fresh {
			*slot = save.NewSlot()
		}
		g.LoadRunFromSlot(s.sel)
		g.PersistRun()
		if fresh {
			g.Go(&StoryScene{cut: Prologue, next: func() Scene { return &MapScene{} }})
		} else {
			g.Go(&MapScene{})
		}
	}
	return nil
}

func (s *FileScene) Draw(g *Game, dst *ebiten.Image) {
	gfx.Backgrounds[0].Draw(dst, s.camX, 0)
	gfx.FillRectF(dst, 0, 0, core.ScreenW, core.ScreenH, gfx.Alpha(gfx.ColInkDeep, 0.62))

	gfx.TextCenterOutline(dst, "SELECT A FILE", core.ScreenW/2, 14, gfx.ColCream, gfx.ColInkDeep, 2)

	for i := 0; i < 3; i++ {
		slot := g.Save.Slots[i]
		y := 44.0 + float64(i)*58
		focused := s.sel == i

		border := gfx.Alpha(gfx.ColCream, 0.35)
		fill := gfx.Alpha(gfx.ColInkDeep, 0.78)
		if focused {
			border = gfx.ColCoral
			fill = gfx.Alpha(gfx.ColSlate, 0.85)
		}
		gfx.Panel(dst, 36, y, core.ScreenW-72, 50, fill, border)

		label := fmt.Sprintf("FILE %d", i+1)
		gfx.TextShadow(dst, label, 48, int(y)+7, gfx.ColGoldLight, gfx.ColInkDeep, 1)

		if !slot.Used {
			gfx.Text(dst, "- NEW GAME -", 48, int(y)+24, gfx.Alpha(gfx.ColCream, 0.7), 1)
		} else {
			// world / level and completion
			gfx.Text(dst, fmt.Sprintf("WORLD %d-%d", slot.World+1, slot.Level+1), 110, int(y)+7, gfx.ColCream, 1)
			gfx.Text(dst, fmt.Sprintf("SCORE %d", slot.Score), 200, int(y)+7, gfx.ColCream, 1)
			gfx.Text(dst, "TIME "+slot.PlayTime(), 320, int(y)+7, gfx.ColCream, 1)

			// rescued models
			gfx.Text(dst, "RESCUED", 48, int(y)+24, gfx.Alpha(gfx.ColCream, 0.6), 1)
			for m := 0; m < 4; m++ {
				mx := 110 + float64(m)*22
				if slot.Rescued[m] {
					gfx.DrawSprite(dst, gfx.ImgModel[m][(s.t/16)%2], mx, y+18, gfx.SpriteOpt{})
				} else {
					gfx.DrawSprite(dst, gfx.ImgLock, mx+3, y+20, gfx.SpriteOpt{Alpha: 0.5})
				}
			}

			// completion bar
			pct := slot.Completion()
			gfx.FillRectF(dst, 210, y+26, 120, 6, gfx.Alpha(gfx.ColInkDeep, 0.8))
			gfx.FillRectF(dst, 210, y+26, 120*pct, 6, gfx.ColCoral)
			gfx.StrokeRectF(dst, 210, y+26, 120, 6, gfx.Alpha(gfx.ColCream, 0.5))
			gfx.Text(dst, fmt.Sprintf("%d%%", int(pct*100+0.5)), 338, int(y)+25, gfx.ColCream, 1)

			if slot.Finished {
				gfx.DrawSprite(dst, gfx.ImgStarIcon, core.ScreenW-64, y+20, gfx.SpriteOpt{})
			}
			if focused {
				gfx.Text(dst, "X: ERASE", core.ScreenW-108, int(y)+38, gfx.Alpha(gfx.ColRed, 0.9), 1)
			}
		}
		if focused && (s.t/16)%2 == 0 {
			gfx.Text(dst, ">", 26, int(y)+20, gfx.ColCoral, 2)
		}
	}

	backCol := gfx.Alpha(gfx.ColCream, 0.7)
	if s.sel == 3 {
		backCol = gfx.ColGoldLight
		if (s.t/16)%2 == 0 {
			gfx.Text(dst, ">", core.ScreenW/2-40, 228, gfx.ColCoral, 2)
		}
	}
	gfx.TextCenterShadow(dst, "BACK", core.ScreenW/2, 228, backCol, gfx.ColInkDeep, 2)
	gfx.TextCenter(dst, "ENTER SELECT     ESC BACK", core.ScreenW/2, 256, gfx.Alpha(gfx.ColCream, 0.45), 1)

	if s.erasing {
		gfx.FillRectF(dst, 0, 0, core.ScreenW, core.ScreenH, gfx.Alpha(gfx.ColInkDeep, 0.8))
		gfx.Panel(dst, 90, 100, 300, 70, gfx.ColSlate, gfx.ColRed)
		gfx.TextCenter(dst, fmt.Sprintf("ERASE FILE %d?", s.sel+1), core.ScreenW/2, 114, gfx.ColCream, 2)
		gfx.TextCenter(dst, "THIS CANNOT BE UNDONE.", core.ScreenW/2, 136, gfx.Alpha(gfx.ColCream, 0.75), 1)
		gfx.TextCenter(dst, "ENTER: ERASE      ESC: KEEP", core.ScreenW/2, 152, gfx.ColGoldLight, 1)
	}
}

// PersistSettingsOnly writes the file without touching the live run state,
// used when erasing a slot or changing options.
func (g *Game) PersistSettingsOnly() {
	if err := g.Save.Save(); err != nil {
		g.Toast("COULD NOT SAVE: " + err.Error())
	}
}
