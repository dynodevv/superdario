package game

import (
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// Action is something the player can do. Everything the game reads goes
// through one of these, so rebinding is a data change rather than a code one.
type Action int

const (
	ActLeft Action = iota
	ActRight
	ActUp
	ActDown
	ActJump
	ActRun // run, and throw Alignment Bolts
	ActPause
	ActConfirm
	ActBack
	ActOptions
	ActCount
)

// ActionName is the label shown on the controls screen.
var ActionName = [ActCount]string{
	ActLeft:    "MOVE LEFT",
	ActRight:   "MOVE RIGHT",
	ActUp:      "LOOK UP",
	ActDown:    "DUCK",
	ActJump:    "JUMP",
	ActRun:     "RUN / THROW",
	ActPause:   "PAUSE",
	ActConfirm: "CONFIRM",
	ActBack:    "BACK",
	ActOptions: "OPTIONS",
}

// ActionHint explains what each action is for.
var ActionHint = [ActCount]string{
	ActLeft:    "WALK AND SWIM LEFT",
	ActRight:   "WALK AND SWIM RIGHT",
	ActUp:      "ENTER DOORS, CLIMB",
	ActDown:    "CROUCH WHEN SCALED",
	ActJump:    "HOLD LONGER TO JUMP HIGHER",
	ActRun:     "HOLD TO RUN; TAP TO THROW",
	ActPause:   "OPEN THE PAUSE MENU",
	ActConfirm: "ACCEPT A MENU CHOICE",
	ActBack:    "LEAVE A MENU",
	ActOptions: "OPEN OPTIONS FROM THE MAP",
}

// Bindings maps every action to a set of keys and gamepad buttons.
type Bindings struct {
	Keys [ActCount][]ebiten.Key
	Pads [ActCount][]ebiten.StandardGamepadButton
}

// DefaultBindings is the out-of-the-box control scheme.
func DefaultBindings() Bindings {
	var b Bindings
	set := func(a Action, keys []ebiten.Key, pads []ebiten.StandardGamepadButton) {
		b.Keys[a] = keys
		b.Pads[a] = pads
	}
	set(ActLeft, []ebiten.Key{ebiten.KeyArrowLeft, ebiten.KeyA},
		[]ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonLeftLeft})
	set(ActRight, []ebiten.Key{ebiten.KeyArrowRight, ebiten.KeyD},
		[]ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonLeftRight})
	set(ActUp, []ebiten.Key{ebiten.KeyArrowUp, ebiten.KeyW},
		[]ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonLeftTop})
	set(ActDown, []ebiten.Key{ebiten.KeyArrowDown, ebiten.KeyS},
		[]ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonLeftBottom})
	set(ActJump, []ebiten.Key{ebiten.KeySpace, ebiten.KeyZ},
		[]ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonRightBottom})
	set(ActRun, []ebiten.Key{ebiten.KeyShiftLeft, ebiten.KeyX},
		[]ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonRightLeft, ebiten.StandardGamepadButtonFrontBottomRight})
	set(ActPause, []ebiten.Key{ebiten.KeyEscape},
		[]ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonCenterRight})
	set(ActConfirm, []ebiten.Key{ebiten.KeyEnter, ebiten.KeySpace, ebiten.KeyZ},
		[]ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonRightBottom})
	set(ActBack, []ebiten.Key{ebiten.KeyEscape, ebiten.KeyBackspace},
		[]ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonRightRight})
	set(ActOptions, []ebiten.Key{ebiten.KeyP},
		[]ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonCenterLeft})
	return b
}

// ---------------------------------------------------------------------------
// Serialisation
//
// Bindings are stored by name rather than by numeric code, so a save survives
// an Ebitengine update that renumbers the key constants.
// ---------------------------------------------------------------------------

var (
	keyByName  map[string]ebiten.Key
	keyNames   []string
	nameByKey  map[ebiten.Key]string
	padByName  = map[string]ebiten.StandardGamepadButton{}
	padDisplay = map[ebiten.StandardGamepadButton]string{}
)

// XboxNames labels the standard gamepad layout the way an Xbox pad is printed.
var XboxNames = map[ebiten.StandardGamepadButton]string{
	ebiten.StandardGamepadButtonRightBottom:      "A",
	ebiten.StandardGamepadButtonRightRight:       "B",
	ebiten.StandardGamepadButtonRightLeft:        "X",
	ebiten.StandardGamepadButtonRightTop:         "Y",
	ebiten.StandardGamepadButtonFrontTopLeft:     "LB",
	ebiten.StandardGamepadButtonFrontTopRight:    "RB",
	ebiten.StandardGamepadButtonFrontBottomLeft:  "LT",
	ebiten.StandardGamepadButtonFrontBottomRight: "RT",
	ebiten.StandardGamepadButtonCenterLeft:       "VIEW",
	ebiten.StandardGamepadButtonCenterRight:      "MENU",
	ebiten.StandardGamepadButtonCenterCenter:     "GUIDE",
	ebiten.StandardGamepadButtonLeftTop:          "D-UP",
	ebiten.StandardGamepadButtonLeftBottom:       "D-DOWN",
	ebiten.StandardGamepadButtonLeftLeft:         "D-LEFT",
	ebiten.StandardGamepadButtonLeftRight:        "D-RIGHT",
	ebiten.StandardGamepadButtonLeftStick:        "LS",
	ebiten.StandardGamepadButtonRightStick:       "RS",
}

func init() {
	keyByName = map[string]ebiten.Key{}
	nameByKey = map[ebiten.Key]string{}
	for k := ebiten.Key(0); k <= ebiten.KeyMax; k++ {
		n := strings.ToUpper(k.String())
		if n == "" {
			continue
		}
		if _, dup := keyByName[n]; dup {
			continue
		}
		keyByName[n] = k
		nameByKey[k] = n
		keyNames = append(keyNames, n)
	}
	sort.Strings(keyNames)

	for b, n := range XboxNames {
		padByName[n] = b
		padDisplay[b] = n
	}
}

// KeyLabel is the short name shown in the UI for a key.
func KeyLabel(k ebiten.Key) string {
	if n, ok := nameByKey[k]; ok {
		return prettyKey(n)
	}
	return "?"
}

// prettyKey shortens the longer Ebitengine names so they fit the 5x7 font.
func prettyKey(n string) string {
	switch n {
	case "ARROWLEFT":
		return "LEFT"
	case "ARROWRIGHT":
		return "RIGHT"
	case "ARROWUP":
		return "UP"
	case "ARROWDOWN":
		return "DOWN"
	case "SHIFTLEFT":
		return "L-SHIFT"
	case "SHIFTRIGHT":
		return "R-SHIFT"
	case "CONTROLLEFT":
		return "L-CTRL"
	case "CONTROLRIGHT":
		return "R-CTRL"
	case "ALTLEFT":
		return "L-ALT"
	case "ALTRIGHT":
		return "R-ALT"
	case "NUMPADENTER":
		return "NUM-ENT"
	case "BACKSPACE":
		return "BKSP"
	case "ESCAPE":
		return "ESC"
	}
	if strings.HasPrefix(n, "DIGIT") {
		return n[5:]
	}
	if strings.HasPrefix(n, "NUMPAD") {
		return "NUM" + n[6:]
	}
	return n
}

// PadLabel is the Xbox-style name for a gamepad button.
func PadLabel(b ebiten.StandardGamepadButton) string {
	if n, ok := padDisplay[b]; ok {
		return n
	}
	return "?"
}

// ToStrings flattens the bindings into the form stored in the save file.
func (b Bindings) ToStrings() (map[string][]string, map[string][]string) {
	keys := map[string][]string{}
	pads := map[string][]string{}
	for a := Action(0); a < ActCount; a++ {
		name := ActionName[a]
		for _, k := range b.Keys[a] {
			if n, ok := nameByKey[k]; ok {
				keys[name] = append(keys[name], n)
			}
		}
		for _, p := range b.Pads[a] {
			if n, ok := padDisplay[p]; ok {
				pads[name] = append(pads[name], n)
			}
		}
	}
	return keys, pads
}

// BindingsFromStrings rebuilds bindings from the save file, falling back to
// the default for any action the file does not cover.
func BindingsFromStrings(keys, pads map[string][]string) Bindings {
	b := DefaultBindings()
	if len(keys) == 0 && len(pads) == 0 {
		return b
	}
	for a := Action(0); a < ActCount; a++ {
		name := ActionName[a]
		if names, ok := keys[name]; ok {
			var out []ebiten.Key
			for _, n := range names {
				if k, ok := keyByName[strings.ToUpper(n)]; ok {
					out = append(out, k)
				}
			}
			if len(out) > 0 {
				b.Keys[a] = out
			}
		}
		if names, ok := pads[name]; ok {
			var out []ebiten.StandardGamepadButton
			for _, n := range names {
				if p, ok := padByName[strings.ToUpper(n)]; ok {
					out = append(out, p)
				}
			}
			if len(out) > 0 {
				b.Pads[a] = out
			}
		}
	}
	return b
}

// Conflicts lists the other actions a key is already bound to.
func (b Bindings) Conflicts(a Action, k ebiten.Key) []Action {
	var out []Action
	for other := Action(0); other < ActCount; other++ {
		if other == a {
			continue
		}
		// Confirm/Back deliberately overlap Jump/Run and Pause; menu actions
		// and gameplay actions never fire in the same frame.
		if isMenuAction(other) != isMenuAction(a) {
			continue
		}
		for _, ok := range b.Keys[other] {
			if ok == k {
				out = append(out, other)
				break
			}
		}
	}
	return out
}

func isMenuAction(a Action) bool {
	return a == ActConfirm || a == ActBack || a == ActOptions
}

// BindKey replaces an action's keys with a single key.
func (b *Bindings) BindKey(a Action, k ebiten.Key) {
	b.Keys[a] = []ebiten.Key{k}
}

// BindPad replaces an action's gamepad buttons with a single button.
func (b *Bindings) BindPad(a Action, p ebiten.StandardGamepadButton) {
	b.Pads[a] = []ebiten.StandardGamepadButton{p}
}

// KeysLabel renders an action's keys as one string, e.g. "LEFT / A".
func (b Bindings) KeysLabel(a Action) string {
	if len(b.Keys[a]) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(b.Keys[a]))
	for _, k := range b.Keys[a] {
		parts = append(parts, KeyLabel(k))
	}
	return strings.Join(parts, " / ")
}

// PadsLabel renders an action's gamepad buttons as one string.
func (b Bindings) PadsLabel(a Action) string {
	if len(b.Pads[a]) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(b.Pads[a]))
	for _, p := range b.Pads[a] {
		parts = append(parts, PadLabel(p))
	}
	return strings.Join(parts, " / ")
}
