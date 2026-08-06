package gfx

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

var white1 *ebiten.Image

func initWhite() {
	white1 = ebiten.NewImage(1, 1)
	white1.Fill(color.White)
}

// FillRectF fills an axis-aligned rectangle in float coordinates.
func FillRectF(dst *ebiten.Image, x, y, w, h float64, col color.RGBA) {
	if w <= 0 || h <= 0 || col.A == 0 {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w, h)
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(col)
	dst.DrawImage(white1, op)
}

// StrokeRectF draws a 1px rectangle outline.
func StrokeRectF(dst *ebiten.Image, x, y, w, h float64, col color.RGBA) {
	FillRectF(dst, x, y, w, 1, col)
	FillRectF(dst, x, y+h-1, w, 1, col)
	FillRectF(dst, x, y, 1, h, col)
	FillRectF(dst, x+w-1, y, 1, h, col)
}

// LineF draws a 1px line between two points.
func LineF(dst *ebiten.Image, x0, y0, x1, y1 float64, col color.RGBA) {
	dx, dy := x1-x0, y1-y0
	l := math.Hypot(dx, dy)
	if l < 0.001 {
		FillRectF(dst, x0, y0, 1, 1, col)
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(l, 1)
	op.GeoM.Rotate(math.Atan2(dy, dx))
	op.GeoM.Translate(x0, y0)
	op.ColorScale.ScaleWithColor(col)
	dst.DrawImage(white1, op)
}

// Panel draws a bordered UI panel with an optional shadow.
func Panel(dst *ebiten.Image, x, y, w, h float64, fill, border color.RGBA) {
	FillRectF(dst, x+2, y+3, w, h, Alpha(ColInkDeep, 0.45))
	FillRectF(dst, x, y, w, h, fill)
	StrokeRectF(dst, x, y, w, h, border)
	StrokeRectF(dst, x+1, y+1, w-2, h-2, Alpha(border, 0.35))
}

// Sprite draws img at (x,y), optionally mirrored, with a colour multiply.
type SpriteOpt struct {
	FlipX   bool
	FlipY   bool
	Alpha   float64
	Tint    color.RGBA // zero value = no tint
	ScaleX  float64
	ScaleY  float64
	Rot     float64
	OriginX float64 // rotation origin in image space
	OriginY float64
}

// DrawSprite is the workhorse used by every entity.
func DrawSprite(dst, img *ebiten.Image, x, y float64, o SpriteOpt) {
	if img == nil {
		return
	}
	w := float64(img.Bounds().Dx())
	h := float64(img.Bounds().Dy())
	sx, sy := o.ScaleX, o.ScaleY
	if sx == 0 {
		sx = 1
	}
	if sy == 0 {
		sy = 1
	}
	if o.FlipX {
		sx = -sx
	}
	if o.FlipY {
		sy = -sy
	}
	op := &ebiten.DrawImageOptions{}
	if o.Rot != 0 {
		op.GeoM.Translate(-o.OriginX, -o.OriginY)
		op.GeoM.Scale(sx, sy)
		op.GeoM.Rotate(o.Rot)
		op.GeoM.Translate(o.OriginX, o.OriginY)
	} else {
		op.GeoM.Scale(sx, sy)
	}
	ox, oy := 0.0, 0.0
	if sx < 0 {
		ox = w
	}
	if sy < 0 {
		oy = h
	}
	op.GeoM.Translate(math.Floor(x+ox), math.Floor(y+oy))
	if o.Tint.A > 0 {
		op.ColorScale.ScaleWithColor(o.Tint)
	}
	if o.Alpha > 0 && o.Alpha < 1 {
		op.ColorScale.ScaleAlpha(float32(o.Alpha))
	}
	dst.DrawImage(img, op)
}

// FlashTint returns a white-hot tint used when something takes damage.
func FlashTint(t float64) color.RGBA {
	v := clampU8(255 * (0.6 + 0.4*math.Sin(t*0.8)))
	return color.RGBA{255, v, v, 255}
}

// VGradient paints a vertical gradient band (menus, HUD backing).
func VGradient(dst *ebiten.Image, x, y, w, h float64, top, bot color.RGBA) {
	steps := int(h)
	if steps < 1 {
		return
	}
	for i := 0; i < steps; i++ {
		f := float64(i) / float64(steps)
		FillRectF(dst, x, y+float64(i), w, 1, Mix(top, bot, f))
	}
}
