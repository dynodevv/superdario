package gfx

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// A drawn Xbox-style controller for the Controls screen. Buttons are addressed
// by name so the game layer does not need to know the geometry.

// PadPart identifies one element of the diagram.
type PadPart int

const (
	PadA PadPart = iota
	PadB
	PadX
	PadY
	PadLB
	PadRB
	PadLT
	PadRT
	PadView
	PadMenu
	PadGuide
	PadDUp
	PadDDown
	PadDLeft
	PadDRight
	PadLStick
	PadRStick
	PadPartCount
)

// padGeom is the position and size of every part, in diagram-local pixels.
// The diagram is 200x120.
const (
	PadDiagramW = 200
	PadDiagramH = 120
)

type padRect struct {
	X, Y, W, H float64
	Round      bool
	Label      string
}

// Standard Xbox arrangement: left stick top-left, d-pad bottom-left, face
// buttons top-right, right stick bottom-right.
var padGeom = [PadPartCount]padRect{
	PadLStick: {51, 35, 26, 26, true, "LS"},
	PadRStick: {115, 77, 26, 26, true, "RS"},

	PadDUp:    {64, 79, 12, 11, false, ""},
	PadDDown:  {64, 101, 12, 11, false, ""},
	PadDLeft:  {53, 90, 11, 11, false, ""},
	PadDRight: {76, 90, 11, 11, false, ""},

	PadY: {140, 30, 14, 14, true, "Y"},
	PadX: {126, 44, 14, 14, true, "X"},
	PadB: {154, 44, 14, 14, true, "B"},
	PadA: {140, 58, 14, 14, true, "A"},

	PadLB: {34, 16, 32, 9, false, "LB"},
	PadRB: {136, 16, 32, 9, false, "RB"},
	PadLT: {40, 4, 24, 9, false, "LT"},
	PadRT: {138, 4, 24, 9, false, "RT"},

	PadView:  {80, 47, 13, 9, false, ""},
	PadMenu:  {109, 47, 13, 9, false, ""},
	PadGuide: {94, 28, 13, 13, true, ""},
}

// PadPartRect exposes a part's rectangle so the caller can point at it.
func PadPartRect(p PadPart) (x, y, w, h float64) {
	r := padGeom[p]
	return r.X, r.Y, r.W, r.H
}

// PadDiagramOpts controls how the controller is drawn.
type PadDiagramOpts struct {
	// Highlight is drawn in the accent colour (the action being rebound).
	Highlight PadPart
	HasHL     bool
	// Pressed reports live button state so the diagram reacts to the pad.
	Pressed func(PadPart) bool
	// Sticks are the live analog values, -1..1.
	LX, LY, RX, RY float64
	// Connected dims the whole diagram when no pad is plugged in.
	Connected bool
	Tick      int
}

// DrawGamepad paints the controller at (ox, oy).
func DrawGamepad(dst *ebiten.Image, ox, oy float64, o PadDiagramOpts) {
	// Dim only slightly when nothing is plugged in: the diagram still has to
	// be readable, it just should not look live.
	dim := 1.0
	if !o.Connected {
		dim = 0.72
	}

	shell := Scale(color.RGBA{0x4a, 0x50, 0x5e, 0xff}, dim)
	shellHi := Scale(color.RGBA{0x6c, 0x74, 0x86, 0xff}, dim)
	shellLo := Scale(color.RGBA{0x1e, 0x21, 0x29, 0xff}, dim)

	// --- body -----------------------------------------------------------
	// grips, angled outward
	fillRound(dst, ox+10, oy+40, 54, 72, 18, shell)
	fillRound(dst, ox+136, oy+40, 54, 72, 18, shell)
	// centre slab
	fillRound(dst, ox+36, oy+26, 128, 62, 20, shell)
	// top highlight
	fillRound(dst, ox+40, oy+28, 120, 16, 8, shellHi)
	// underside shadow
	fillRound(dst, ox+44, oy+82, 112, 6, 3, shellLo)

	// --- shoulder buttons ------------------------------------------------
	for _, p := range []PadPart{PadLT, PadRT, PadLB, PadRB} {
		r := padGeom[p]
		col := Scale(color.RGBA{0x4d, 0x53, 0x62, 0xff}, dim)
		if o.Pressed != nil && o.Pressed(p) {
			col = ColGoldLight
		}
		if o.HasHL && o.Highlight == p {
			col = pulseCol(ColCoral, o.Tick)
		}
		fillRound(dst, ox+r.X, oy+r.Y, r.W, r.H, 4, col)
		strokeRound(dst, ox+r.X, oy+r.Y, r.W, r.H, 4, shellLo)
		TextCenter(dst, r.Label, int(ox+r.X+r.W/2), int(oy+r.Y+r.H/2)-3, Scale(ColCream, dim), 1)
	}

	// --- face buttons ----------------------------------------------------
	faceCols := map[PadPart]color.RGBA{
		PadA: {0x6a, 0xc4, 0x6a, 0xff},
		PadB: {0xd0, 0x54, 0x4c, 0xff},
		PadX: {0x53, 0x87, 0xd8, 0xff},
		PadY: {0xe0, 0xba, 0x45, 0xff},
	}
	for _, p := range []PadPart{PadA, PadB, PadX, PadY} {
		r := padGeom[p]
		col := Scale(faceCols[p], dim)
		if o.Pressed != nil && o.Pressed(p) {
			col = Mix(col, ColCream, 0.6)
		}
		if o.HasHL && o.Highlight == p {
			drawRing(dst, ox+r.X+r.W/2, oy+r.Y+r.H/2, r.W/2+4, pulseCol(ColCoral, o.Tick))
		}
		fillCircleF(dst, ox+r.X+r.W/2, oy+r.Y+r.H/2, r.W/2, Scale(col, 0.55))
		fillCircleF(dst, ox+r.X+r.W/2, oy+r.Y+r.H/2-1, r.W/2-1, col)
		TextCenter(dst, r.Label, int(ox+r.X+r.W/2), int(oy+r.Y+r.H/2)-3, ColInkDeep, 1)
	}

	// --- d-pad ------------------------------------------------------------
	dpadCol := Scale(color.RGBA{0x9a, 0xa2, 0xb2, 0xff}, dim)
	for _, p := range []PadPart{PadDUp, PadDDown, PadDLeft, PadDRight} {
		r := padGeom[p]
		col := dpadCol
		if o.Pressed != nil && o.Pressed(p) {
			col = ColGoldLight
		}
		if o.HasHL && o.Highlight == p {
			col = pulseCol(ColCoral, o.Tick)
		}
		FillRectF(dst, ox+r.X, oy+r.Y+1, r.W, r.H, Scale(col, 0.45)) // drop shadow
		FillRectF(dst, ox+r.X, oy+r.Y, r.W, r.H, col)
		StrokeRectF(dst, ox+r.X, oy+r.Y, r.W, r.H, shellLo)
	}
	// d-pad hub
	FillRectF(dst, ox+64, oy+90, 12, 11, Scale(dpadCol, 0.8))
	StrokeRectF(dst, ox+64, oy+90, 12, 11, shellLo)

	// --- centre buttons ---------------------------------------------------
	// No text labels here: at this size they collide with each other and with
	// the guide button. The binding list beside the diagram already names them.
	for _, p := range []PadPart{PadView, PadMenu} {
		r := padGeom[p]
		col := Scale(color.RGBA{0x8e, 0x96, 0xa6, 0xff}, dim)
		if o.Pressed != nil && o.Pressed(p) {
			col = ColGoldLight
		}
		if o.HasHL && o.Highlight == p {
			col = pulseCol(ColCoral, o.Tick)
		}
		fillRound(dst, ox+r.X, oy+r.Y, r.W, r.H, 4, col)
		strokeRound(dst, ox+r.X, oy+r.Y, r.W, r.H, 4, shellLo)
	}
	// guide
	{
		r := padGeom[PadGuide]
		glow := 0.5 + 0.5*math.Sin(float64(o.Tick)*0.05)
		if !o.Connected {
			glow = 0.15
		}
		fillCircleF(dst, ox+r.X+r.W/2, oy+r.Y+r.H/2, r.W/2, Scale(shellLo, 1.2))
		fillCircleF(dst, ox+r.X+r.W/2, oy+r.Y+r.H/2, r.W/2-3, Alpha(ColGreen, 0.35+0.55*glow))
	}

	// --- sticks ------------------------------------------------------------
	drawStick := func(p PadPart, dx, dy float64) {
		r := padGeom[p]
		cx := ox + r.X + r.W/2
		cy := oy + r.Y + r.H/2
		if o.HasHL && o.Highlight == p {
			drawRing(dst, cx, cy, r.W/2+3, pulseCol(ColCoral, o.Tick))
		}
		// recessed well
		fillCircleF(dst, cx, cy, r.W/2, shellLo)
		fillCircleF(dst, cx, cy, r.W/2-1, Scale(color.RGBA{0x33, 0x38, 0x44, 0xff}, dim))
		// cap, offset by the live axis values
		capX := cx + dx*4
		capY := cy + dy*4
		fillCircleF(dst, capX, capY+1, r.W/2-3, shellLo)
		fillCircleF(dst, capX, capY, r.W/2-3, Scale(color.RGBA{0x8e, 0x96, 0xa6, 0xff}, dim))
		fillCircleF(dst, capX, capY-1, r.W/2-5, Scale(color.RGBA{0xb4, 0xbc, 0xcb, 0xff}, dim))
		if o.Pressed != nil && o.Pressed(p) {
			fillCircleF(dst, capX, capY, r.W/2-5, ColGoldLight)
		}
	}
	drawStick(PadLStick, o.LX, o.LY)
	drawStick(PadRStick, o.RX, o.RY)
}

func pulseCol(c color.RGBA, tick int) color.RGBA {
	f := 0.6 + 0.4*math.Sin(float64(tick)*0.16)
	return Mix(c, ColCream, f*0.5)
}

// --- small vector helpers ---------------------------------------------------

func fillRound(dst *ebiten.Image, x, y, w, h, r float64, col color.RGBA) {
	if r*2 > h {
		r = h / 2
	}
	if r*2 > w {
		r = w / 2
	}
	FillRectF(dst, x+r, y, w-2*r, h, col)
	FillRectF(dst, x, y+r, w, h-2*r, col)
	fillCircleF(dst, x+r, y+r, r, col)
	fillCircleF(dst, x+w-r, y+r, r, col)
	fillCircleF(dst, x+r, y+h-r, r, col)
	fillCircleF(dst, x+w-r, y+h-r, r, col)
}

func strokeRound(dst *ebiten.Image, x, y, w, h, r float64, col color.RGBA) {
	FillRectF(dst, x+r, y, w-2*r, 1, col)
	FillRectF(dst, x+r, y+h-1, w-2*r, 1, col)
	FillRectF(dst, x, y+r, 1, h-2*r, col)
	FillRectF(dst, x+w-1, y+r, 1, h-2*r, col)
}

func fillCircleF(dst *ebiten.Image, cx, cy, r float64, col color.RGBA) {
	if r <= 0 {
		return
	}
	ri := int(math.Ceil(r))
	for j := -ri; j <= ri; j++ {
		dy := float64(j)
		if math.Abs(dy) > r {
			continue
		}
		half := math.Sqrt(r*r - dy*dy)
		FillRectF(dst, cx-half, cy+dy, half*2, 1, col)
	}
}

func drawRing(dst *ebiten.Image, cx, cy, r float64, col color.RGBA) {
	steps := 22
	for i := 0; i < steps; i++ {
		a := float64(i) / float64(steps) * 2 * math.Pi
		FillRectF(dst, cx+r*math.Cos(a)-1, cy+r*math.Sin(a)-1, 2, 2, col)
	}
}
