package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"superdario/internal/core"
	"superdario/internal/gfx"
	"superdario/internal/sfx"
)

// ControlsScene shows and edits the bindings. Two tabs: keyboard, and a drawn
// Xbox pad that lights up as you press buttons on a real controller.
type ControlsScene struct {
	t    int
	tab  int // 0 keyboard, 1 gamepad
	sel  int // selected action row
	back func() Scene

	capturing bool
	captureT  int // grace ticks so the key that opened capture is not eaten
	conflict  string
	conflictT int
	resetT    int // confirmation countdown for "reset to defaults"
}

// The order actions appear on the screen.
var controlRows = []Action{
	ActLeft, ActRight, ActDown, ActJump, ActRun, ActPause, ActConfirm, ActBack, ActOptions,
}

// padPartFor maps a gamepad button to the diagram part, for highlighting.
var padPartFor = map[ebiten.StandardGamepadButton]gfx.PadPart{
	ebiten.StandardGamepadButtonRightBottom:      gfx.PadA,
	ebiten.StandardGamepadButtonRightRight:       gfx.PadB,
	ebiten.StandardGamepadButtonRightLeft:        gfx.PadX,
	ebiten.StandardGamepadButtonRightTop:         gfx.PadY,
	ebiten.StandardGamepadButtonFrontTopLeft:     gfx.PadLB,
	ebiten.StandardGamepadButtonFrontTopRight:    gfx.PadRB,
	ebiten.StandardGamepadButtonFrontBottomLeft:  gfx.PadLT,
	ebiten.StandardGamepadButtonFrontBottomRight: gfx.PadRT,
	ebiten.StandardGamepadButtonCenterLeft:       gfx.PadView,
	ebiten.StandardGamepadButtonCenterRight:      gfx.PadMenu,
	ebiten.StandardGamepadButtonCenterCenter:     gfx.PadGuide,
	ebiten.StandardGamepadButtonLeftTop:          gfx.PadDUp,
	ebiten.StandardGamepadButtonLeftBottom:       gfx.PadDDown,
	ebiten.StandardGamepadButtonLeftLeft:         gfx.PadDLeft,
	ebiten.StandardGamepadButtonLeftRight:        gfx.PadDRight,
	ebiten.StandardGamepadButtonLeftStick:        gfx.PadLStick,
	ebiten.StandardGamepadButtonRightStick:       gfx.PadRStick,
}

func (s *ControlsScene) Enter(g *Game) {
	if s.back == nil {
		s.back = func() Scene { return &TitleScene{} }
	}
}

// ShowGamepadTab opens the scene on the controller tab.
func (s *ControlsScene) ShowGamepadTab() { s.tab = 1 }

func (s *ControlsScene) Update(g *Game) error {
	s.t++
	if s.conflictT > 0 {
		s.conflictT--
	}
	if s.resetT > 0 {
		s.resetT--
	}

	// --- rebinding capture ------------------------------------------------
	if s.capturing {
		if s.captureT > 0 {
			s.captureT--
			return nil
		}
		a := controlRows[s.sel]
		if s.tab == 0 {
			if k, ok := g.Pad.CaptureKey(); ok {
				if k == ebiten.KeyEscape {
					s.capturing = false
					sfx.Play(sfx.SndMenuMove)
					return nil
				}
				if c := g.Pad.Binds.Conflicts(a, k); len(c) > 0 {
					s.conflict = KeyLabel(k) + " IS ALREADY " + ActionName[c[0]]
					s.conflictT = 150
					sfx.Play(sfx.SndBump)
					return nil
				}
				g.Pad.Binds.BindKey(a, k)
				g.PersistBindings()
				s.capturing = false
				sfx.Play(sfx.SndMenuPick)
			}
			// let the player back out with the pad too
			if g.Pad.CapturePadIsBack() {
				s.capturing = false
			}
		} else {
			if b, ok := g.Pad.CapturePad(); ok {
				g.Pad.Binds.BindPad(a, b)
				g.PersistBindings()
				s.capturing = false
				sfx.Play(sfx.SndMenuPick)
				return nil
			}
			if k, ok := g.Pad.CaptureKey(); ok && k == ebiten.KeyEscape {
				s.capturing = false
				sfx.Play(sfx.SndMenuMove)
			}
		}
		return nil
	}

	// --- navigation --------------------------------------------------------
	n := len(controlRows) + 2 // + "reset" + "back"
	if g.Pad.MenuUp() {
		s.sel = (s.sel + n - 1) % n
		sfx.Play(sfx.SndMenuMove)
	}
	if g.Pad.MenuDown() {
		s.sel = (s.sel + 1) % n
		sfx.Play(sfx.SndMenuMove)
	}
	if g.Pad.MenuLeft() || g.Pad.MenuRight() {
		s.tab = 1 - s.tab
		sfx.Play(sfx.SndMenuMove)
	}

	if g.Pad.ConfirmPressed() {
		switch {
		case s.sel < len(controlRows):
			if s.tab == 1 && !g.Pad.HasGamepad() {
				s.conflict = "NO CONTROLLER CONNECTED"
				s.conflictT = 150
				sfx.Play(sfx.SndBump)
				break
			}
			s.capturing = true
			s.captureT = 2
			sfx.Play(sfx.SndMenuPick)
		case s.sel == len(controlRows): // reset
			if s.resetT > 0 {
				g.Pad.Binds = DefaultBindings()
				g.PersistBindings()
				s.resetT = 0
				s.conflict = "CONTROLS RESET TO DEFAULTS"
				s.conflictT = 150
				sfx.Play(sfx.SndUnlock)
			} else {
				s.resetT = 150
				sfx.Play(sfx.SndMenuPick)
			}
		default: // back
			sfx.Play(sfx.SndMenuPick)
			g.PersistBindings()
			g.Go(s.back())
		}
		return nil
	}

	if g.Pad.BackPressed() {
		g.PersistBindings()
		g.Go(s.back())
	}
	return nil
}

func (s *ControlsScene) Draw(g *Game, dst *ebiten.Image) {
	dst.Fill(gfx.ColInkDeep)
	gfx.Backgrounds[3].Draw(dst, float64(s.t)*0.15, 0)
	gfx.FillRectF(dst, 0, 0, core.ScreenW, core.ScreenH, gfx.Alpha(gfx.ColInkDeep, 0.80))

	gfx.TextCenterOutline(dst, "CONTROLS", core.ScreenW/2, 8, gfx.ColCream, gfx.ColInkDeep, 2)

	// --- tabs --------------------------------------------------------------
	tabs := []string{"KEYBOARD", "GAMEPAD"}
	for i, name := range tabs {
		w := 108.0
		x := float64(core.ScreenW)/2 - w - 4 + float64(i)*(w+8)
		fill := gfx.Alpha(gfx.ColInkDeep, 0.8)
		col := gfx.Alpha(gfx.ColCream, 0.55)
		if i == s.tab {
			fill = gfx.Alpha(gfx.ColSlate, 0.95)
			col = gfx.ColGoldLight
		}
		gfx.FillRectF(dst, x, 26, w, 16, fill)
		gfx.StrokeRectF(dst, x, 26, w, 16, gfx.Alpha(gfx.ColCoral, 0.9))
		if i == s.tab {
			gfx.FillRectF(dst, x, 40, w, 2, gfx.ColCoral)
		}
		label := name
		if i == 1 && g.Pad.HasGamepad() {
			label = name + " *"
		}
		gfx.TextCenter(dst, label, int(x+w/2), 31, col, 1)
	}

	if s.tab == 0 {
		s.drawKeyboardTab(g, dst)
	} else {
		s.drawGamepadTab(g, dst)
	}

	// --- footer -------------------------------------------------------------
	if s.conflictT > 0 {
		w := float64(gfx.TextWidth(s.conflict, 1) + 14)
		gfx.Panel(dst, float64(core.ScreenW)/2-w/2, 236, w, 14, gfx.ColInkDeep, gfx.ColRed)
		gfx.TextCenter(dst, s.conflict, core.ScreenW/2, 240, gfx.ColCream, 1)
	}
	hint := "UP/DOWN SELECT   LEFT/RIGHT TAB   ENTER REBIND   ESC BACK"
	if s.capturing {
		hint = "PRESS A NEW INPUT...   ESC CANCELS"
	}
	gfx.TextCenter(dst, hint, core.ScreenW/2, 258, gfx.Alpha(gfx.ColCream, 0.45), 1)
}

func (s *ControlsScene) drawKeyboardTab(g *Game, dst *ebiten.Image) {
	gfx.Panel(dst, 26, 50, core.ScreenW-52, 180, gfx.Alpha(gfx.ColInkDeep, 0.92), gfx.ColCoral)

	for i, a := range controlRows {
		y := 56 + i*15
		col := gfx.Alpha(gfx.ColCream, 0.8)
		if i == s.sel {
			gfx.FillRectF(dst, 30, float64(y)-3, core.ScreenW-60, 13, gfx.Alpha(gfx.ColCoral, 0.22))
			col = gfx.ColGoldLight
			if (s.t/16)%2 == 0 {
				gfx.Text(dst, ">", 34, y, gfx.ColCoral, 1)
			}
		}
		gfx.Text(dst, ActionName[a], 46, y, col, 1)
		gfx.Text(dst, ActionHint[a], 150, y, gfx.Alpha(gfx.ColCream, 0.32), 1)

		val := g.Pad.Binds.KeysLabel(a)
		if s.capturing && i == s.sel {
			val = "PRESS A KEY"
			if (s.t/8)%2 == 0 {
				val = "PRESS A KEY_"
			}
			gfx.Text(dst, val, core.ScreenW-150, y, gfx.ColCoral, 1)
		} else {
			gfx.Text(dst, val, core.ScreenW-150, y, col, 1)
		}
	}
	s.drawTail(g, dst, 56+len(controlRows)*15+6)
}

func (s *ControlsScene) drawGamepadTab(g *Game, dst *ebiten.Image) {
	gfx.Panel(dst, 12, 50, core.ScreenW-24, 180, gfx.Alpha(gfx.ColInkDeep, 0.92), gfx.ColCoral)

	// --- the diagram --------------------------------------------------------
	hl := gfx.PadA
	hasHL := false
	if s.sel < len(controlRows) {
		a := controlRows[s.sel]
		if len(g.Pad.Binds.Pads[a]) > 0 {
			if part, ok := padPartFor[g.Pad.Binds.Pads[a][0]]; ok {
				hl, hasHL = part, true
			}
		}
	}
	lx, ly, rx, ry := g.Pad.Sticks()
	gfx.DrawGamepad(dst, 252, 74, gfx.PadDiagramOpts{
		Highlight: hl, HasHL: hasHL,
		Pressed: func(p gfx.PadPart) bool {
			for b, part := range padPartFor {
				if part == p {
					return g.Pad.PadButtonDown(b)
				}
			}
			return false
		},
		LX: lx, LY: ly, RX: rx, RY: ry,
		Connected: g.Pad.HasGamepad(),
		Tick:      s.t,
	})

	status := "NO CONTROLLER CONNECTED"
	statusCol := gfx.Alpha(gfx.ColCream, 0.45)
	if g.Pad.HasGamepad() {
		status = g.Pad.GamepadName()
		if status == "" {
			status = "CONTROLLER CONNECTED"
		}
		if len(status) > 26 {
			status = status[:26]
		}
		statusCol = gfx.ColGreen
	}
	gfx.TextCenter(dst, status, 352, 200, statusCol, 1)
	gfx.TextCenter(dst, "PRESS BUTTONS TO TEST", 352, 212, gfx.Alpha(gfx.ColCream, 0.3), 1)

	// --- the binding list ---------------------------------------------------
	for i, a := range controlRows {
		y := 56 + i*15
		col := gfx.Alpha(gfx.ColCream, 0.8)
		if i == s.sel {
			gfx.FillRectF(dst, 16, float64(y)-3, 224, 13, gfx.Alpha(gfx.ColCoral, 0.22))
			col = gfx.ColGoldLight
			if (s.t/16)%2 == 0 {
				gfx.Text(dst, ">", 20, y, gfx.ColCoral, 1)
			}
		}
		gfx.Text(dst, ActionName[a], 32, y, col, 1)

		val := g.Pad.Binds.PadsLabel(a)
		if s.capturing && i == s.sel {
			val = "PRESS A BUTTON"
			if (s.t/8)%2 == 0 {
				val = "PRESS A BUTTON_"
			}
			gfx.Text(dst, val, 138, y, gfx.ColCoral, 1)
		} else {
			gfx.Text(dst, val, 138, y, col, 1)
		}
	}
	// the analog stick is not rebindable, so say so
	gfx.Text(dst, "LEFT STICK ALWAYS MOVES", 32, 56+len(controlRows)*15+2, gfx.Alpha(gfx.ColCream, 0.3), 1)
	s.drawTail(g, dst, 56+len(controlRows)*15+14)
}

// drawTail renders the shared "reset" and "back" rows.
func (s *ControlsScene) drawTail(g *Game, dst *ebiten.Image, y int) {
	x := 46
	if s.tab == 1 {
		x = 32
	}

	resetIdx := len(controlRows)
	label := "RESET TO DEFAULTS"
	col := gfx.Alpha(gfx.ColCream, 0.7)
	if s.resetT > 0 {
		label = "PRESS AGAIN TO CONFIRM"
		col = gfx.ColRed
	}
	if s.sel == resetIdx {
		if s.resetT == 0 {
			col = gfx.ColGoldLight
		}
		if (s.t/16)%2 == 0 {
			gfx.Text(dst, ">", x-12, y, gfx.ColCoral, 1)
		}
	}
	gfx.Text(dst, label, x, y, col, 1)

	backCol := gfx.Alpha(gfx.ColCream, 0.7)
	if s.sel == resetIdx+1 {
		backCol = gfx.ColGoldLight
		if (s.t/16)%2 == 0 {
			gfx.Text(dst, ">", x-12, y+13, gfx.ColCoral, 1)
		}
	}
	gfx.Text(dst, "BACK", x, y+13, backCol, 1)
}
