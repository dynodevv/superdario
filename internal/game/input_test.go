package game

import (
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// withPad fakes a connected controller. refresh() needs real hardware, so the
// device state is set directly; everything downstream of it is pure logic.
func withPad(connected, lastWasPad bool) *Pad {
	p := NewPad()
	p.hasPad = connected
	p.LastWasPad = lastWasPad
	return p
}

// Prompts must be written in the vocabulary of the device actually in use.
func TestPromptsFollowTheActiveDevice(t *testing.T) {
	kb := withPad(false, false)
	pad := withPad(true, true)

	for a := Action(0); a < ActCount; a++ {
		if got, want := kb.Label(a), kb.Binds.KeysLabel(a); got != want {
			t.Errorf("%s on keyboard: label %q, want %q", ActionName[a], got, want)
		}
		if got, want := pad.Label(a), pad.Binds.PadsLabel(a); got != want {
			t.Errorf("%s on gamepad: label %q, want %q", ActionName[a], got, want)
		}
	}

	// the specific case from the title screen
	if pad.Prompt(ActJump, "JUMP") != "A JUMP" {
		t.Errorf("gamepad jump prompt = %q, want %q", pad.Prompt(ActJump, "JUMP"), "A JUMP")
	}
	if !strings.Contains(kb.Prompt(ActJump, "JUMP"), "SPACE") {
		t.Errorf("keyboard jump prompt = %q, expected it to mention SPACE", kb.Prompt(ActJump, "JUMP"))
	}
	if kb.MoveLabel() == pad.MoveLabel() {
		t.Errorf("move label is %q on both devices", kb.MoveLabel())
	}
	if pad.MoveLabel() != "D-PAD / STICK" {
		t.Errorf("gamepad move label = %q", pad.MoveLabel())
	}
}

// A connected controller takes over the prompts without waiting for a press,
// and the keyboard takes them back the moment it is used.
func TestPromptDeviceHandover(t *testing.T) {
	p := withPad(true, true) // as refresh() sets it on first detection
	if !p.UsingPad() {
		t.Error("a freshly connected pad should own the prompts")
	}

	p.LastWasPad = false // player touched the keyboard
	if p.UsingPad() {
		t.Error("using the keyboard should hand the prompts back to it")
	}

	p.LastWasPad = true
	p.hasPad = false // controller unplugged
	if p.UsingPad() {
		t.Error("an unplugged pad must not keep the prompts")
	}
	if p.Label(ActConfirm) != p.Binds.KeysLabel(ActConfirm) {
		t.Error("prompts should be keyboard labels with no pad connected")
	}
}

// Rebinding an action has to change what the prompt advertises, on both
// devices — otherwise the hints lie after the player customises the controls.
func TestPromptsFollowRebinds(t *testing.T) {
	p := withPad(false, false)
	p.Binds.BindKey(ActJump, ebiten.KeyW)
	if got := p.Label(ActJump); got != "W" {
		t.Errorf("after rebinding jump to W, label = %q", got)
	}

	p.hasPad, p.LastWasPad = true, true
	p.Binds.BindPad(ActJump, ebiten.StandardGamepadButtonRightTop)
	if got := p.Label(ActJump); got != "Y" {
		t.Errorf("after rebinding jump to Y, label = %q", got)
	}
}

// Every action must render a real label on both devices; a prompt reading "-"
// or "?" means an action shipped without a usable default.
func TestEveryActionHasLabelsOnBothDevices(t *testing.T) {
	b := DefaultBindings()
	for a := Action(0); a < ActCount; a++ {
		k, g := b.KeysLabel(a), b.PadsLabel(a)
		if k == "-" || strings.Contains(k, "?") {
			t.Errorf("%s has no usable keyboard binding (%q)", ActionName[a], k)
		}
		if g == "-" || strings.Contains(g, "?") {
			t.Errorf("%s has no usable gamepad binding (%q)", ActionName[a], g)
		}
	}
}

// Bindings are stored by name so a save survives an engine update that
// renumbers the key constants; that round trip has to be lossless.
func TestBindingsRoundTripThroughSave(t *testing.T) {
	b := DefaultBindings()
	b.BindKey(ActJump, ebiten.KeyW)
	b.BindPad(ActRun, ebiten.StandardGamepadButtonFrontTopLeft)

	keys, pads := b.ToStrings()
	got := BindingsFromStrings(keys, pads)

	for a := Action(0); a < ActCount; a++ {
		if got.KeysLabel(a) != b.KeysLabel(a) {
			t.Errorf("%s keys: %q -> %q", ActionName[a], b.KeysLabel(a), got.KeysLabel(a))
		}
		if got.PadsLabel(a) != b.PadsLabel(a) {
			t.Errorf("%s pads: %q -> %q", ActionName[a], b.PadsLabel(a), got.PadsLabel(a))
		}
	}
}
