package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"superdario/internal/world"
)

// Pad turns keyboard and gamepad state into game actions. Scenes never touch
// Ebitengine's input API directly, so rebinding only has to change Bindings.

const stickDeadzone = 0.4

type Pad struct {
	Binds Bindings

	gamepads []ebiten.GamepadID
	active   ebiten.GamepadID
	hasPad   bool

	// menu auto-repeat, one counter per direction
	repeat [4]int

	// true when the last input came from a gamepad, so prompts can say
	// "A" instead of "ENTER"
	LastWasPad bool
}

func NewPad() *Pad { return &Pad{Binds: DefaultBindings()} }

// refresh re-scans connected gamepads once per frame.
func (p *Pad) refresh() {
	p.gamepads = ebiten.AppendGamepadIDs(p.gamepads[:0])
	p.hasPad = false
	for _, id := range p.gamepads {
		if ebiten.IsStandardGamepadLayoutAvailable(id) {
			p.active = id
			p.hasPad = true
			break
		}
	}
	// track which device the player is actually using
	if p.hasPad && len(inpututil.AppendJustPressedGamepadButtons(p.active, nil)) > 0 {
		p.LastWasPad = true
	}
	if len(inpututil.AppendJustPressedKeys(nil)) > 0 {
		p.LastWasPad = false
	}
}

// HasGamepad reports whether a standard-layout controller is connected.
func (p *Pad) HasGamepad() bool { return p.hasPad }

// GamepadName is the connected controller's reported name.
func (p *Pad) GamepadName() string {
	if !p.hasPad {
		return ""
	}
	return ebiten.GamepadName(p.active)
}

// --- raw queries ------------------------------------------------------------

func (p *Pad) keyDown(a Action) bool {
	for _, k := range p.Binds.Keys[a] {
		if ebiten.IsKeyPressed(k) {
			return true
		}
	}
	return false
}

func (p *Pad) keyPressed(a Action) bool {
	for _, k := range p.Binds.Keys[a] {
		if inpututil.IsKeyJustPressed(k) {
			return true
		}
	}
	return false
}

func (p *Pad) padDown(a Action) bool {
	if !p.hasPad {
		return false
	}
	for _, b := range p.Binds.Pads[a] {
		if ebiten.IsStandardGamepadButtonPressed(p.active, b) {
			return true
		}
	}
	return false
}

func (p *Pad) padPressed(a Action) bool {
	if !p.hasPad {
		return false
	}
	for _, b := range p.Binds.Pads[a] {
		if inpututil.IsStandardGamepadButtonJustPressed(p.active, b) {
			return true
		}
	}
	return false
}

// axis reads the left stick, deadzoned.
func (p *Pad) axis(a ebiten.StandardGamepadAxis) float64 {
	if !p.hasPad {
		return 0
	}
	v := ebiten.StandardGamepadAxisValue(p.active, a)
	if v > stickDeadzone || v < -stickDeadzone {
		return v
	}
	return 0
}

// stickDir contributes the analog stick to the four directions. The stick is
// always live regardless of bindings — nobody expects to rebind a stick.
func (p *Pad) stickDir(a Action) bool {
	switch a {
	case ActLeft:
		return p.axis(ebiten.StandardGamepadAxisLeftStickHorizontal) < 0
	case ActRight:
		return p.axis(ebiten.StandardGamepadAxisLeftStickHorizontal) > 0
	case ActUp:
		return p.axis(ebiten.StandardGamepadAxisLeftStickVertical) < 0
	case ActDown:
		return p.axis(ebiten.StandardGamepadAxisLeftStickVertical) > 0
	}
	return false
}

// --- action queries ---------------------------------------------------------

// Down is true while an action is held.
func (p *Pad) Down(a Action) bool {
	return p.keyDown(a) || p.padDown(a) || p.stickDir(a)
}

// Pressed is true on the frame an action starts. Stick directions get an
// edge-detected press through the repeat machinery instead.
func (p *Pad) Pressed(a Action) bool {
	return p.keyPressed(a) || p.padPressed(a)
}

func (p *Pad) Left() bool  { return p.Down(ActLeft) }
func (p *Pad) Right() bool { return p.Down(ActRight) }
func (p *Pad) Up() bool    { return p.Down(ActUp) }
func (p *Pad) Down4() bool { return p.Down(ActDown) }

func (p *Pad) JumpDown() bool    { return p.Down(ActJump) }
func (p *Pad) JumpPressed() bool { return p.Pressed(ActJump) }
func (p *Pad) RunDown() bool     { return p.Down(ActRun) }
func (p *Pad) FirePressed() bool { return p.Pressed(ActRun) }

func (p *Pad) ConfirmPressed() bool { return p.Pressed(ActConfirm) }
func (p *Pad) BackPressed() bool    { return p.Pressed(ActBack) }
func (p *Pad) PausePressed() bool   { return p.Pressed(ActPause) }
func (p *Pad) OptionsPressed() bool { return p.Pressed(ActOptions) }

// MenuUp and friends auto-repeat when held, and accept the analog stick.
func (p *Pad) MenuUp() bool    { return p.menuRepeat(0, ActUp) }
func (p *Pad) MenuDown() bool  { return p.menuRepeat(1, ActDown) }
func (p *Pad) MenuLeft() bool  { return p.menuRepeat(2, ActLeft) }
func (p *Pad) MenuRight() bool { return p.menuRepeat(3, ActRight) }

func (p *Pad) menuRepeat(slot int, a Action) bool {
	if !p.Down(a) {
		p.repeat[slot] = 0
		return false
	}
	p.repeat[slot]++
	if p.repeat[slot] == 1 {
		return true
	}
	if p.repeat[slot] > 26 && (p.repeat[slot]-26)%6 == 0 {
		return true
	}
	return false
}

// GameInput packages one frame of play control for the world.
func (p *Pad) GameInput() world.Input {
	return world.Input{
		Left:     p.Left(),
		Right:    p.Right(),
		Up:       p.Up(),
		Down:     p.Down4(),
		Jump:     p.JumpPressed(),
		JumpDown: p.JumpDown(),
		Run:      p.RunDown(),
		Fire:     p.FirePressed(),
	}
}

// AnyPressed backs "press any button" prompts.
func (p *Pad) AnyPressed() bool {
	if len(inpututil.AppendJustPressedKeys(nil)) > 0 {
		return true
	}
	if p.hasPad && len(inpututil.AppendJustPressedGamepadButtons(p.active, nil)) > 0 {
		return true
	}
	return false
}

// CaptureKey returns the first key pressed this frame, for rebinding.
func (p *Pad) CaptureKey() (ebiten.Key, bool) {
	ks := inpututil.AppendJustPressedKeys(nil)
	if len(ks) == 0 {
		return 0, false
	}
	return ks[0], true
}

// CapturePad returns the first standard gamepad button pressed this frame.
func (p *Pad) CapturePad() (ebiten.StandardGamepadButton, bool) {
	if !p.hasPad {
		return 0, false
	}
	for b := ebiten.StandardGamepadButton(0); b <= ebiten.StandardGamepadButtonMax; b++ {
		if inpututil.IsStandardGamepadButtonJustPressed(p.active, b) {
			return b, true
		}
	}
	return 0, false
}

// CapturePadIsBack lets a rebind prompt be cancelled from the controller.
func (p *Pad) CapturePadIsBack() bool {
	return p.hasPad && inpututil.IsStandardGamepadButtonJustPressed(p.active, ebiten.StandardGamepadButtonRightRight)
}

// PadButtonDown is used by the controller diagram to light up live presses.
func (p *Pad) PadButtonDown(b ebiten.StandardGamepadButton) bool {
	return p.hasPad && ebiten.IsStandardGamepadButtonPressed(p.active, b)
}

// Sticks returns the two analog sticks, for the diagram.
func (p *Pad) Sticks() (lx, ly, rx, ry float64) {
	if !p.hasPad {
		return
	}
	lx = ebiten.StandardGamepadAxisValue(p.active, ebiten.StandardGamepadAxisLeftStickHorizontal)
	ly = ebiten.StandardGamepadAxisValue(p.active, ebiten.StandardGamepadAxisLeftStickVertical)
	rx = ebiten.StandardGamepadAxisValue(p.active, ebiten.StandardGamepadAxisRightStickHorizontal)
	ry = ebiten.StandardGamepadAxisValue(p.active, ebiten.StandardGamepadAxisRightStickVertical)
	return
}

// ConfirmLabel is the prompt text for "accept", matching the device in use.
func (p *Pad) ConfirmLabel() string {
	if p.LastWasPad {
		return p.Binds.PadsLabel(ActConfirm)
	}
	return p.Binds.KeysLabel(ActConfirm)
}

// BackLabel is the prompt text for "cancel".
func (p *Pad) BackLabel() string {
	if p.LastWasPad {
		return p.Binds.PadsLabel(ActBack)
	}
	return p.Binds.KeysLabel(ActBack)
}
