package gfx

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// ---------------------------------------------------------------------------
// Canvas: a tiny software drawing surface. All art is generated at startup on
// these, then uploaded once to GPU textures. No asset files ship with the game.
// ---------------------------------------------------------------------------

type Canvas struct {
	*image.RGBA
	W, H int
}

func NewCanvas(w, h int) *Canvas {
	return &Canvas{RGBA: image.NewRGBA(image.Rect(0, 0, w, h)), W: w, H: h}
}

func (c *Canvas) Px(x, y int, col color.RGBA) {
	if x < 0 || y < 0 || x >= c.W || y >= c.H || col.A == 0 {
		return
	}
	i := c.PixOffset(x, y)
	c.Pix[i] = col.R
	c.Pix[i+1] = col.G
	c.Pix[i+2] = col.B
	c.Pix[i+3] = col.A
}

// Blend draws col over the existing pixel with alpha a (0..1).
func (c *Canvas) Blend(x, y int, col color.RGBA, a float64) {
	if x < 0 || y < 0 || x >= c.W || y >= c.H || a <= 0 {
		return
	}
	if a > 1 {
		a = 1
	}
	i := c.PixOffset(x, y)
	dr, dg, db, da := float64(c.Pix[i]), float64(c.Pix[i+1]), float64(c.Pix[i+2]), float64(c.Pix[i+3])
	c.Pix[i] = uint8(float64(col.R)*a + dr*(1-a))
	c.Pix[i+1] = uint8(float64(col.G)*a + dg*(1-a))
	c.Pix[i+2] = uint8(float64(col.B)*a + db*(1-a))
	c.Pix[i+3] = uint8(math.Min(255, float64(col.A)*a+da))
}

func (c *Canvas) At2(x, y int) color.RGBA {
	if x < 0 || y < 0 || x >= c.W || y >= c.H {
		return color.RGBA{}
	}
	i := c.PixOffset(x, y)
	return color.RGBA{c.Pix[i], c.Pix[i+1], c.Pix[i+2], c.Pix[i+3]}
}

func (c *Canvas) FillRect(x, y, w, h int, col color.RGBA) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			c.Px(i, j, col)
		}
	}
}

func (c *Canvas) Rect(x, y, w, h int, col color.RGBA) {
	for i := x; i < x+w; i++ {
		c.Px(i, y, col)
		c.Px(i, y+h-1, col)
	}
	for j := y; j < y+h; j++ {
		c.Px(x, j, col)
		c.Px(x+w-1, j, col)
	}
}

func (c *Canvas) HLine(x, y, w int, col color.RGBA) {
	for i := x; i < x+w; i++ {
		c.Px(i, y, col)
	}
}

func (c *Canvas) VLine(x, y, h int, col color.RGBA) {
	for j := y; j < y+h; j++ {
		c.Px(x, j, col)
	}
}

func (c *Canvas) Line(x0, y0, x1, y1 int, col color.RGBA) {
	dx, dy := abs(x1-x0), -abs(y1-y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		c.Px(x0, y0, col)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func (c *Canvas) FillCircle(cx, cy, r int, col color.RGBA) {
	for j := -r; j <= r; j++ {
		for i := -r; i <= r; i++ {
			if i*i+j*j <= r*r {
				c.Px(cx+i, cy+j, col)
			}
		}
	}
}

func (c *Canvas) Circle(cx, cy, r int, col color.RGBA) {
	for j := -r; j <= r; j++ {
		for i := -r; i <= r; i++ {
			d := i*i + j*j
			if d <= r*r && d > (r-1)*(r-1) {
				c.Px(cx+i, cy+j, col)
			}
		}
	}
}

// FillEllipse is used a lot for round robot bodies and boss blobs.
func (c *Canvas) FillEllipse(cx, cy, rx, ry int, col color.RGBA) {
	for j := -ry; j <= ry; j++ {
		for i := -rx; i <= rx; i++ {
			fx := float64(i) / float64(rx)
			fy := float64(j) / float64(ry)
			if fx*fx+fy*fy <= 1.0 {
				c.Px(cx+i, cy+j, col)
			}
		}
	}
}

// Bevel draws a raised block: light top/left edge, dark bottom/right edge.
func (c *Canvas) Bevel(x, y, w, h int, face, light, dark color.RGBA) {
	c.FillRect(x, y, w, h, face)
	c.HLine(x, y, w, light)
	c.VLine(x, y, h, light)
	c.HLine(x, y+h-1, w, dark)
	c.VLine(x+w-1, y, h, dark)
}

// Outline traces a 1px border around every opaque cluster, giving sprites the
// crisp readable silhouette that classic platformers rely on.
func (c *Canvas) Outline(col color.RGBA) {
	src := make([]uint8, len(c.Pix))
	copy(src, c.Pix)
	opaque := func(x, y int) bool {
		if x < 0 || y < 0 || x >= c.W || y >= c.H {
			return false
		}
		return src[c.PixOffset(x, y)+3] > 0
	}
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			if opaque(x, y) {
				continue
			}
			if opaque(x-1, y) || opaque(x+1, y) || opaque(x, y-1) || opaque(x, y+1) {
				c.Px(x, y, col)
			}
		}
	}
}

// Shade multiplies a region's brightness, used for tile depth and dark caves.
func (c *Canvas) Shade(x, y, w, h int, f float64) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			p := c.At2(i, j)
			if p.A == 0 {
				continue
			}
			c.Px(i, j, color.RGBA{clampU8(float64(p.R) * f), clampU8(float64(p.G) * f), clampU8(float64(p.B) * f), p.A})
		}
	}
}

// Speckle adds deterministic grain so large flat tiles do not look plasticky.
func (c *Canvas) Speckle(x, y, w, h int, rng *rand.Rand, amount float64, density float64) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			if rng.Float64() > density {
				continue
			}
			p := c.At2(i, j)
			if p.A == 0 {
				continue
			}
			f := 1 + (rng.Float64()*2-1)*amount
			c.Px(i, j, color.RGBA{clampU8(float64(p.R) * f), clampU8(float64(p.G) * f), clampU8(float64(p.B) * f), p.A})
		}
	}
}

func (c *Canvas) Img() *ebiten.Image {
	img := ebiten.NewImage(c.W, c.H)
	img.WritePixels(c.Pix)
	return img
}

// Sub returns a copy of a region as its own image.
func (c *Canvas) Sub(x, y, w, h int) *Canvas {
	out := NewCanvas(w, h)
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			out.Px(i, j, c.At2(x+i, y+j))
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// ASCII sprite decoding
// ---------------------------------------------------------------------------

// ParseArt turns ASCII art rows into a canvas using the master palette.
// swap optionally remaps palette characters (used for power-state recolours).
func ParseArt(rows []string, swap map[byte]byte) *Canvas {
	h := len(rows)
	w := 0
	for _, r := range rows {
		if len(r) > w {
			w = len(r)
		}
	}
	c := NewCanvas(w, h)
	for y, row := range rows {
		for x := 0; x < len(row); x++ {
			ch := row[x]
			if swap != nil {
				if s, ok := swap[ch]; ok {
					ch = s
				}
			}
			if col, ok := Pal[ch]; ok {
				c.Px(x, y, col)
			}
		}
	}
	return c
}

// Art builds a finished, outlined sprite from ASCII art.
func Art(rows []string, swap map[byte]byte) *ebiten.Image {
	c := ParseArt(rows, swap)
	c.Outline(Pal['0'])
	return c.Img()
}

// ArtRaw builds a sprite without the automatic outline.
func ArtRaw(rows []string, swap map[byte]byte) *ebiten.Image {
	return ParseArt(rows, swap).Img()
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func clampU8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// Mix linearly interpolates two colours.
func Mix(a, b color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return color.RGBA{
		clampU8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		clampU8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		clampU8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		clampU8(float64(a.A) + (float64(b.A)-float64(a.A))*t),
	}
}

// Scale brightens/darkens a colour keeping alpha.
func Scale(c color.RGBA, f float64) color.RGBA {
	return color.RGBA{clampU8(float64(c.R) * f), clampU8(float64(c.G) * f), clampU8(float64(c.B) * f), c.A}
}

// Alpha returns c with the given alpha (0..1).
func Alpha(c color.RGBA, a float64) color.RGBA {
	return color.RGBA{c.R, c.G, c.B, clampU8(255 * a)}
}
