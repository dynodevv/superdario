// Command shots drives the game through its screens and writes PNG captures.
// Development tool: it is how the look of each scene gets reviewed without a
// human at the keyboard.
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"superdario/internal/core"
	"superdario/internal/game"
	"superdario/internal/sfx"
	"superdario/internal/world"
)

var outDir = "."

// A step runs for `frames` ticks, then the capture is written.
type step struct {
	name   string
	frames int
	setup  func(g *game.Game)
}

type shooter struct {
	g       *game.Game
	steps   []step
	idx     int
	frame   int
	started bool
}

// botInput walks right and jumps whenever forward motion stalls, which is
// enough to get through most of a stage for a screenshot.
func botInput(tick int) (world.Input, bool) {
	in := world.Input{Right: true, Run: tick%240 > 60}
	// jump on a rhythm so gaps and steps get cleared
	phase := tick % 42
	in.Jump = phase == 0
	in.JumpDown = phase < 14
	return in, true
}

func (s *shooter) Update() error {
	if s.idx >= len(s.steps) {
		return ebiten.Termination
	}
	st := s.steps[s.idx]
	if s.frame == 0 && st.setup != nil {
		st.setup(s.g)
	}
	s.frame++
	return s.g.Update()
}

func (s *shooter) Draw(screen *ebiten.Image) {
	if s.idx >= len(s.steps) {
		return
	}
	s.g.Draw(screen)
	st := s.steps[s.idx]
	if s.frame >= st.frames {
		save(st.name, screen)
		s.idx++
		s.frame = 0
	}
}

func (s *shooter) Layout(w, h int) (int, int) { return core.ScreenW, core.ScreenH }

func save(name string, img *ebiten.Image) {
	b := img.Bounds()
	buf := make([]byte, 4*b.Dx()*b.Dy())
	img.ReadPixels(buf)
	out := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	copy(out.Pix, buf)
	f, err := os.Create(filepath.Join(outDir, name+".png"))
	if err != nil {
		fmt.Println("create:", err)
		return
	}
	defer f.Close()
	if err := png.Encode(f, out); err != nil {
		fmt.Println("encode:", err)
		return
	}
	fmt.Println("wrote", name+".png")
}

func main() {
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}
	// This tool runs the engine uncapped and jumps between scenes, which would
	// otherwise fire thousands of overlapping sounds a second.
	sfx.Disable()

	g := game.New()

	gotoLevel := func(wld, lvl int, power core.PowerState) func(*game.Game) {
		return func(g *game.Game) {
			g.Run.World, g.Run.Level = wld, lvl
			g.Run.Power = power
			g.Run.Lives = 5
			g.Run.Score = 24500
			g.Run.Tokens = 37
			g.Autopilot = botInput
			g.GoNow(&game.PlayScene{})
		}
	}

	s := &shooter{g: g, steps: []step{
		{name: "01_title", frames: 90},
		{name: "01b_title_gamepad", frames: 40, setup: func(g *game.Game) {
			g.Pad.ForcePromptDevice(true)
			g.GoNow(&game.TitleScene{})
		}},
		{name: "02_files", frames: 60, setup: func(g *game.Game) { g.GoNow(&game.FileScene{}) }},
		{name: "03_story", frames: 300, setup: func(g *game.Game) {
			g.Run.SlotIdx = 0
			g.GoNow(game.NewStoryScene(game.Prologue, nil))
		}},
		// the longest caption in the game: seven lines
		{name: "03b_story_long", frames: 400, setup: func(g *game.Game) {
			g.GoNow(game.NewStoryScene(game.Interludes[3], nil))
		}},
		{name: "04_map", frames: 60, setup: func(g *game.Game) {
			sl := g.Slot()
			sl.Used = true
			sl.World, sl.Level = 3, 1
			sl.Cleared[0] = [3]bool{true, true, true}
			sl.Cleared[1] = [3]bool{true, true, true}
			sl.Cleared[2] = [3]bool{true, true, true}
			sl.Cleared[3] = [3]bool{true, false, false}
			sl.Rescued = [4]bool{true, true, true, false}
			sl.BestTime[0][0] = 4200
			g.Run.World, g.Run.Level = 3, 1
			g.Run.Rescued = sl.Rescued
			g.GoNow(&game.MapScene{})
		}},
		{name: "05_stage_card", frames: 60, setup: gotoLevel(0, 0, core.PowerScaled)},
		{name: "06_world1", frames: 210, setup: gotoLevel(0, 0, core.PowerScaled)},
		{name: "07_world1_far", frames: 900, setup: gotoLevel(0, 1, core.PowerAlign)},
		{name: "08_world2", frames: 560, setup: gotoLevel(1, 0, core.PowerScaled)},
		{name: "09_world3", frames: 560, setup: gotoLevel(2, 0, core.PowerScaled)},
		{name: "10_world4", frames: 560, setup: gotoLevel(3, 0, core.PowerAlign)},
		{name: "11_world5", frames: 560, setup: gotoLevel(4, 0, core.PowerAlign)},
		{name: "12_boss1", frames: 1500, setup: gotoLevel(0, 2, core.PowerAlign)},
		{name: "13_boss5", frames: 1500, setup: gotoLevel(4, 2, core.PowerAlign)},
		{name: "14_credits", frames: 200, setup: func(g *game.Game) {
			g.Autopilot = nil
			g.GoNow(&game.CreditsScene{})
		}},
		{name: "15_options", frames: 40, setup: func(g *game.Game) {
			g.GoNow(&game.OptionsScene{})
		}},
		{name: "16_controls_keyboard", frames: 40, setup: func(g *game.Game) {
			g.GoNow(&game.ControlsScene{})
		}},
		{name: "17_controls_gamepad", frames: 40, setup: func(g *game.Game) {
			cs := &game.ControlsScene{}
			cs.ShowGamepadTab()
			g.GoNow(cs)
		}},
	}}

	ebiten.SetWindowSize(core.ScreenW*2, core.ScreenH*2)
	ebiten.SetWindowTitle("shots")
	ebiten.SetTPS(ebiten.SyncWithFPS)
	ebiten.SetVsyncEnabled(false)
	if err := ebiten.RunGame(s); err != nil && err != ebiten.Termination {
		fmt.Println(err)
	}
}
