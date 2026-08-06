// Package game hosts the Ebitengine game loop, the scene stack and all the
// screens that are not the platforming itself.
package game

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"superdario/internal/core"
	"superdario/internal/gfx"
	"superdario/internal/save"
	"superdario/internal/sfx"
	"superdario/internal/world"
)

// Scene is one screen of the game.
type Scene interface {
	Enter(g *Game)
	Update(g *Game) error
	Draw(g *Game, dst *ebiten.Image)
}

// RunState is the progress carried from level to level within one session.
type RunState struct {
	World, Level int
	Lives        int
	Score        int
	Tokens       int
	Power        core.PowerState
	Rescued      [4]bool
	SlotIdx      int
	sessionStart time.Time
}

// Game is the Ebitengine entry point and scene manager.
type Game struct {
	Pad  *Pad
	Save *save.File
	Run  RunState

	scene   Scene
	pending Scene

	fade     float64 // 1 = fully black
	fadeDir  int     // -1 fading in, +1 fading out
	tick     int
	inited   bool
	quitting bool

	// a one-line toast for save errors and the like
	toast  string
	toastT int

	// Autopilot, when set, supplies play input instead of the controller.
	// Used by the screenshot tool and available for an attract-mode demo.
	Autopilot func(tick int) (world.Input, bool)
}

func New() *Game {
	g := &Game{Pad: NewPad(), Save: save.Load()}
	g.Run.Lives = 3
	g.Run.sessionStart = time.Now()
	g.fade = 1
	g.fadeDir = -1
	return g
}

// Layout fixes the internal resolution; Ebitengine scales it to the window.
func (g *Game) Layout(outW, outH int) (int, int) { return core.ScreenW, core.ScreenH }

func (g *Game) init() {
	if g.inited {
		return
	}
	g.inited = true
	gfx.Init()
	sfx.Init()
	sfx.SetMusicVolume(g.Save.Settings.MusicVolume)
	sfx.SetSFXVolume(g.Save.Settings.SFXVolume)
	g.Pad.Binds = BindingsFromStrings(g.Save.Settings.KeyBinds, g.Save.Settings.PadBinds)
	if err := g.Save.Err(); err != nil {
		g.Toast(err.Error())
	}
	g.scene = &TitleScene{}
	g.scene.Enter(g)
}

func (g *Game) Update() error {
	g.init()
	g.tick++
	g.Pad.refresh()
	sfx.Update()

	if g.toastT > 0 {
		g.toastT--
	}

	// scene transition: fade out, swap, fade in
	if g.fadeDir > 0 {
		g.fade += 0.08
		if g.fade >= 1 {
			g.fade = 1
			if g.pending != nil {
				g.scene = g.pending
				g.pending = nil
				g.scene.Enter(g)
			}
			g.fadeDir = -1
		}
		return nil
	}
	if g.fadeDir < 0 {
		g.fade -= 0.08
		if g.fade <= 0 {
			g.fade = 0
			g.fadeDir = 0
		}
	}

	if g.quitting {
		return ebiten.Termination
	}
	if g.scene == nil {
		return nil
	}
	return g.scene.Update(g)
}

func (g *Game) Draw(dst *ebiten.Image) {
	if g.scene != nil {
		g.scene.Draw(g, dst)
	}
	if g.toastT > 0 {
		w := gfx.TextWidth(g.toast, 1) + 12
		gfx.Panel(dst, float64(core.ScreenW/2-w/2), 6, float64(w), 15, gfx.Alpha(gfx.ColInkDeep, 0.9), gfx.ColCoral)
		gfx.TextCenter(dst, g.toast, core.ScreenW/2, 11, gfx.ColCream, 1)
	}
	if g.Save.Settings.Scanlines {
		drawScanlines(dst)
	}
	if g.fade > 0 {
		gfx.FillRectF(dst, 0, 0, core.ScreenW, core.ScreenH, gfx.Alpha(gfx.ColInkDeep, g.fade))
	}
}

// Go queues a scene change behind a fade.
func (g *Game) Go(s Scene) {
	g.pending = s
	g.fadeDir = 1
}

// GoNow swaps scenes immediately (used within a screen, e.g. pause overlays).
func (g *Game) GoNow(s Scene) {
	g.scene = s
	g.scene.Enter(g)
}

func (g *Game) Quit() { g.quitting = true }

// Toast shows a brief message at the top of the screen.
func (g *Game) Toast(msg string) {
	if len(msg) > 62 {
		msg = msg[:62]
	}
	g.toast = msg
	g.toastT = 300
}

// Slot is the currently selected save slot.
func (g *Game) Slot() *save.Slot { return &g.Save.Slots[g.Run.SlotIdx] }

// PersistRun copies the live run state into the save slot and writes it out.
func (g *Game) PersistRun() {
	s := g.Slot()
	s.Used = true
	s.Played = time.Now()
	s.Lives = g.Run.Lives
	s.Score = g.Run.Score
	s.Tokens = g.Run.Tokens
	s.Power = int(g.Run.Power)
	s.Rescued = g.Run.Rescued
	s.Seconds += int(time.Since(g.Run.sessionStart).Seconds())
	g.Run.sessionStart = time.Now()
	if err := g.Save.Save(); err != nil {
		g.Toast("COULD NOT SAVE: " + err.Error())
	}
}

// LoadRunFromSlot seeds the live run from the chosen save slot.
func (g *Game) LoadRunFromSlot(idx int) {
	g.Run.SlotIdx = idx
	s := &g.Save.Slots[idx]
	g.Run.World = s.World
	g.Run.Level = s.Level
	g.Run.Lives = s.Lives
	if g.Run.Lives <= 0 {
		g.Run.Lives = 3
	}
	g.Run.Score = s.Score
	g.Run.Tokens = s.Tokens
	g.Run.Power = core.PowerState(s.Power)
	g.Run.Rescued = s.Rescued
	g.Run.sessionStart = time.Now()
}

// PersistBindings writes the current control scheme into the save file.
func (g *Game) PersistBindings() {
	k, p := g.Pad.Binds.ToStrings()
	g.Save.Settings.KeyBinds = k
	g.Save.Settings.PadBinds = p
	g.PersistSettingsOnly()
}

// savePathForDisplay returns a shortened save path for the options screen.
func savePathForDisplay() (string, error) {
	p, err := save.Path()
	if err != nil {
		return "", err
	}
	if len(p) > 58 {
		p = "..." + p[len(p)-55:]
	}
	return p, nil
}

// Scanlines give the whole game a subtle CRT feel. The overlay is built once
// at full screen size, so it costs a single draw call per frame.
var scanlineImg *ebiten.Image

func drawScanlines(dst *ebiten.Image) {
	if scanlineImg == nil {
		c := gfx.NewCanvas(core.ScreenW, core.ScreenH)
		for y := 1; y < core.ScreenH; y += 2 {
			c.HLine(0, y, core.ScreenW, gfx.Alpha(gfx.ColInkDeep, 0.15))
		}
		// a touch of vignette at the very top and bottom
		for y := 0; y < 4; y++ {
			a := 0.10 * float64(4-y) / 4
			c.HLine(0, y, core.ScreenW, gfx.Alpha(gfx.ColInkDeep, a))
			c.HLine(0, core.ScreenH-1-y, core.ScreenW, gfx.Alpha(gfx.ColInkDeep, a))
		}
		scanlineImg = c.Img()
	}
	dst.DrawImage(scanlineImg, nil)
}
