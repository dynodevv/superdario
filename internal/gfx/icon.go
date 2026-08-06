package gfx

import (
	"image"
	"image/color"
	"math"
)

// AppIcon draws the application icon at any size. It returns a plain
// *image.RGBA and touches no GPU state, so the build-time icon generator can
// call it without a graphics context, and the game can hand the same images to
// ebiten.SetWindowIcon at runtime.
//
// The mark is Dario's silhouette — hair and glasses — over the coral sunburst.
// The hair-and-glasses shape is what still reads at 16 pixels.
func AppIcon(size int) *image.RGBA {
	if size < 8 {
		size = 8
	}
	c := NewCanvas(size, size)
	s := float64(size)
	u := s / 64 // one "design unit": the art is authored against a 64px grid

	// A coral field with a pale halo and a dark head on top. Three big blocks
	// of contrast is what survives being scaled down to 16 pixels; fine
	// detail just turns to mud.
	bgTop := color.RGBA{0xe8, 0x88, 0x63, 0xff}
	bgBot := color.RGBA{0xa8, 0x4c, 0x30, 0xff}
	radius := s * 0.22

	// --- rounded background ---------------------------------------------
	for y := 0; y < size; y++ {
		f := float64(y) / s
		row := Mix(bgTop, bgBot, f*f)
		for x := 0; x < size; x++ {
			if insideRounded(float64(x)+0.5, float64(y)+0.5, s, radius) {
				c.Px(x, y, row)
			}
		}
	}

	cx, cy := s/2, s/2

	// --- sunburst, as texture rather than structure ----------------------
	rays := 12
	for i := 0; i < rays; i++ {
		a := float64(i)/float64(rays)*2*math.Pi + 0.26
		for r := s * 0.16; r < s*0.52; r += 0.3 {
			w := math.Max(1, s*0.055*(1-(r-s*0.16)/(s*0.36)))
			x := cx + r*math.Cos(a)
			y := cy + r*math.Sin(a)
			col := Mix(bgTop, ColCoralDark, 0.45)
			for dy := -w / 2; dy <= w/2; dy++ {
				for dx := -w / 2; dx <= w/2; dx++ {
					px, py := int(x+dx), int(y+dy)
					if insideRounded(float64(px)+0.5, float64(py)+0.5, s, radius) {
						c.Px(px, py, col)
					}
				}
			}
		}
	}

	// --- pale halo the head sits against ----------------------------------
	fillDisc(c, cx, cy-0.5*u, s*0.315, Mix(ColCream, ColCoral, 0.10))
	fillDisc(c, cx, cy-0.5*u, s*0.295, ColCream)

	// --- Dario ------------------------------------------------------------
	hair := Pal['b']
	hairLo := Pal['a']
	skin := Pal['d']
	frame := color.RGBA{0x14, 0x11, 0x1a, 0xff}
	lens := color.RGBA{0xdf, 0xef, 0xf7, 0xff}

	headR := s * 0.215
	hy := cy - 1.6*u

	// hair: a wide mass with three bumps on top, so the silhouette reads curly
	fillDisc(c, cx, hy-0.5*u, headR*1.16, hairLo)
	fillDisc(c, cx, hy-1.2*u, headR*1.10, hair)
	for _, o := range []float64{-1, 0, 1} {
		fillDisc(c, cx+o*headR*0.62, hy-headR*0.76, headR*0.42, hair)
	}

	// face, sitting inside the hair
	fillDisc(c, cx, hy+2.0*u, headR*0.80, skin)
	// jaw
	fillRectC(c, cx-headR*0.62, hy+2.0*u, headR*1.24, headR*0.72, skin)

	// glasses: one bar, two lenses. This is the read at small sizes.
	gw := headR * 1.42
	gh := math.Max(1, headR*0.40)
	gy := hy + 1.1*u
	fillRectC(c, cx-gw/2, gy-gh/2, gw, gh, frame)
	lw := gw*0.42 - math.Max(1, u*1.2)
	lh := gh - math.Max(1, u*1.6)
	if lw > 0 && lh > 0 {
		fillRectC(c, cx-gw/2+math.Max(1, u*1.2)/1, gy-lh/2, lw, lh, lens)
		fillRectC(c, cx+gw/2-lw-math.Max(1, u*1.2)/1, gy-lh/2, lw, lh, lens)
	}

	// shoulders, so it does not look like a floating head
	fillRectC(c, cx-headR*1.15, hy+headR*1.02, headR*2.30, headR*0.85, Pal['h'])
	fillRectC(c, cx-headR*0.34, hy+headR*1.02, headR*0.68, headR*0.85, Pal['i'])

	// --- rim light --------------------------------------------------------
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if !insideRounded(float64(x)+0.5, float64(y)+0.5, s, radius) {
				continue
			}
			// 1px inner highlight along the top edge
			if !insideRounded(float64(x)+0.5, float64(y)+0.5-math.Max(1, u), s, radius) {
				p := c.At2(x, y)
				c.Px(x, y, Mix(p, ColCream, 0.28))
			}
		}
	}
	return c.RGBA
}

// insideRounded reports whether a point is inside a rounded square of side s.
func insideRounded(x, y, s, r float64) bool {
	if x < 0 || y < 0 || x > s || y > s {
		return false
	}
	// distance to the rounded-rect edge
	dx := math.Max(math.Max(r-x, x-(s-r)), 0)
	dy := math.Max(math.Max(r-y, y-(s-r)), 0)
	return dx*dx+dy*dy <= r*r
}

func fillDisc(c *Canvas, cx, cy, r float64, col color.RGBA) {
	if r <= 0 {
		return
	}
	y0, y1 := int(cy-r-1), int(cy+r+1)
	for y := y0; y <= y1; y++ {
		dy := float64(y) + 0.5 - cy
		if math.Abs(dy) > r {
			continue
		}
		half := math.Sqrt(r*r - dy*dy)
		for x := int(cx - half); x <= int(cx+half); x++ {
			c.Px(x, y, col)
		}
	}
}

func fillRectC(c *Canvas, x, y, w, h float64, col color.RGBA) {
	for j := int(math.Floor(y)); j < int(math.Ceil(y+h)); j++ {
		for i := int(math.Floor(x)); i < int(math.Ceil(x+w)); i++ {
			c.Px(i, j, col)
		}
	}
}

// IconSizes are the sizes shipped in the .ico and .icns bundles.
var IconSizes = []int{16, 24, 32, 48, 64, 128, 256, 512}

// AppIcons renders the set Ebitengine wants for the window icon.
func AppIcons() []image.Image {
	out := make([]image.Image, 0, 4)
	for _, s := range []int{16, 32, 48, 64, 128} {
		out = append(out, AppIcon(s))
	}
	return out
}
