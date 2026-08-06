package gfx

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"superdario/internal/core"
)

// Background is a scrolling parallax stack plus an animated ambient layer
// (clouds, bubbles, embers, stars) generated per world.
type Background struct {
	Sky      *ebiten.Image
	Layers   []*ebiten.Image
	Factors  []float64
	YOffs    []int
	Ambient  ambientKind
	Motes    []mote
	Theme    Theme
	glowTint color.RGBA
}

type ambientKind int

const (
	ambNone ambientKind = iota
	ambClouds
	ambDust
	ambBubbles
	ambStars
	ambEmbers
)

type mote struct {
	x, y, vx, vy, size, phase float64
	col                       color.RGBA
}

const bgW = core.ScreenW * 2 // parallax layers tile at twice screen width

// BuildBackground generates the full backdrop for a world.
func BuildBackground(t Theme, worldIdx int) *Background {
	rng := rand.New(rand.NewSource(int64(worldIdx)*104729 + 7))
	bg := &Background{Theme: t, glowTint: t.HazeCol}

	// --- sky gradient ---------------------------------------------------
	sky := NewCanvas(1, core.ScreenH)
	for y := 0; y < core.ScreenH; y++ {
		f := float64(y) / float64(core.ScreenH-1)
		sky.HLine(0, y, 1, Mix(t.SkyTop, t.SkyBot, math.Pow(f, 0.85)))
	}
	bg.Sky = sky.Img()

	switch worldIdx {
	case 0:
		bg.Ambient = ambClouds
		bg.Layers = append(bg.Layers, campusFar(t, rng), campusMid(t, rng), campusNear(t, rng))
		bg.Factors = append(bg.Factors, 0.10, 0.28, 0.55)
		bg.YOffs = append(bg.YOffs, 0, 0, 0)
	case 1:
		bg.Ambient = ambDust
		bg.Layers = append(bg.Layers, caveFar(t, rng), caveMid(t, rng))
		bg.Factors = append(bg.Factors, 0.14, 0.36)
		bg.YOffs = append(bg.YOffs, 0, 0)
	case 2:
		bg.Ambient = ambBubbles
		bg.Layers = append(bg.Layers, depthsFar(t, rng), depthsMid(t, rng))
		bg.Factors = append(bg.Factors, 0.12, 0.34)
		bg.YOffs = append(bg.YOffs, 0, 0)
	case 3:
		bg.Ambient = ambStars
		bg.Layers = append(bg.Layers, orbitFar(t, rng), orbitMid(t, rng))
		bg.Factors = append(bg.Factors, 0.06, 0.30)
		bg.YOffs = append(bg.YOffs, 0, 0)
	default:
		bg.Ambient = ambEmbers
		bg.Layers = append(bg.Layers, fortressFar(t, rng), fortressMid(t, rng))
		bg.Factors = append(bg.Factors, 0.12, 0.33)
		bg.YOffs = append(bg.YOffs, 0, 0)
	}

	// --- ambient motes ---------------------------------------------------
	n := 0
	switch bg.Ambient {
	case ambClouds:
		n = 7
	case ambDust:
		n = 46
	case ambBubbles:
		n = 40
	case ambStars:
		n = 110
	case ambEmbers:
		n = 60
	}
	for i := 0; i < n; i++ {
		m := mote{
			x:     rng.Float64() * core.ScreenW,
			y:     rng.Float64() * core.ScreenH,
			phase: rng.Float64() * 6.283,
			size:  1 + rng.Float64()*2,
		}
		switch bg.Ambient {
		case ambClouds:
			m.vx = -0.10 - rng.Float64()*0.14
			m.y = rng.Float64() * 110
			m.size = 10 + rng.Float64()*16
			m.col = Alpha(ColCream, 0.55+rng.Float64()*0.35)
		case ambDust:
			m.vx = -0.16 - rng.Float64()*0.2
			m.vy = -0.05 - rng.Float64()*0.08
			m.col = Alpha(t.Accent, 0.20+rng.Float64()*0.4)
		case ambBubbles:
			m.vy = -0.24 - rng.Float64()*0.4
			m.col = Alpha(color.RGBA{0xcf, 0xf2, 0xff, 0xff}, 0.30+rng.Float64()*0.4)
		case ambStars:
			m.col = Alpha(ColCream, 0.35+rng.Float64()*0.65)
			m.size = 1 + rng.Float64()*1.4
		case ambEmbers:
			m.vy = -0.3 - rng.Float64()*0.7
			m.vx = -0.08 + rng.Float64()*0.16
			m.col = Alpha(Mix(color.RGBA{0xff, 0xd0, 0x60, 0xff}, ColCoral, rng.Float64()), 0.55+rng.Float64()*0.45)
		}
		bg.Motes = append(bg.Motes, m)
	}
	return bg
}

// Update advances the ambient layer.
func (b *Background) Update() {
	for i := range b.Motes {
		m := &b.Motes[i]
		m.x += m.vx
		m.y += m.vy
		m.phase += 0.03
		if b.Ambient == ambEmbers || b.Ambient == ambDust {
			m.x += math.Sin(m.phase) * 0.16
		}
		if m.x < -40 {
			m.x = core.ScreenW + 20
		}
		if m.x > core.ScreenW+40 {
			m.x = -20
		}
		if m.y < -20 {
			m.y = core.ScreenH + 10
			m.x = rand.Float64() * core.ScreenW
		}
		if m.y > core.ScreenH+20 {
			m.y = -10
		}
	}
}

// Draw paints sky, ambient motes and every parallax layer.
func (b *Background) Draw(dst *ebiten.Image, camX, camY float64) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(core.ScreenW, 1)
	dst.DrawImage(b.Sky, op)

	// stars sit behind the parallax silhouettes; everything else in front
	if b.Ambient == ambStars {
		b.drawMotes(dst)
	}

	for i, layer := range b.Layers {
		f := b.Factors[i]
		off := math.Mod(camX*f, bgW)
		if off < 0 {
			off += bgW
		}
		lw := float64(layer.Bounds().Dx())
		for x := -off; x < core.ScreenW; x += lw {
			o := &ebiten.DrawImageOptions{}
			o.GeoM.Translate(math.Floor(x), float64(b.YOffs[i])-camY*f*0.35)
			dst.DrawImage(layer, o)
		}
	}

	if b.Ambient != ambStars {
		b.drawMotes(dst)
	}
}

func (b *Background) drawMotes(dst *ebiten.Image) {
	for _, m := range b.Motes {
		switch b.Ambient {
		case ambClouds:
			drawSoftCloud(dst, m.x, m.y, m.size, m.col)
		case ambStars:
			tw := 0.55 + 0.45*math.Sin(m.phase*2)
			FillRectF(dst, m.x, m.y, m.size, m.size, Alpha(m.col, float64(m.col.A)/255*tw))
		case ambBubbles:
			r := m.size
			FillRectF(dst, m.x, m.y, r, r, m.col)
			FillRectF(dst, m.x, m.y-r, r, 1, Alpha(m.col, 0.4))
		default:
			FillRectF(dst, m.x, m.y, m.size, m.size, m.col)
		}
	}
}

func drawSoftCloud(dst *ebiten.Image, x, y, s float64, col color.RGBA) {
	FillRectF(dst, x, y, s*2.2, s*0.55, col)
	FillRectF(dst, x+s*0.35, y-s*0.36, s*1.3, s*0.6, col)
	FillRectF(dst, x+s*1.1, y-s*0.18, s*0.8, s*0.5, Alpha(col, float64(col.A)/255*0.85))
}

// ---------------------------------------------------------------------------
// Layer painters
// ---------------------------------------------------------------------------

func newLayer() *Canvas { return NewCanvas(bgW, core.ScreenH) }

func campusFar(t Theme, rng *rand.Rand) *ebiten.Image {
	c := newLayer()
	hill := Mix(t.SurfDark, t.SkyBot, 0.55)
	base := core.ScreenH - 46
	for x := 0; x < bgW; x++ {
		hgt := 26 + 16*math.Sin(float64(x)*0.008) + 9*math.Sin(float64(x)*0.021+1.4)
		for y := base - int(hgt); y < core.ScreenH; y++ {
			c.Px(x, y, hill)
		}
		c.Px(x, base-int(hgt), Mix(hill, ColCream, 0.25))
	}
	return c.Img()
}

func campusMid(t Theme, rng *rand.Rand) *ebiten.Image {
	c := newLayer()
	wall := Mix(t.BlockFace, t.SkyBot, 0.3)
	roof := Mix(ColCoral, t.SkyBot, 0.25)
	ground := core.ScreenH - 40
	x := 6
	for x < bgW-40 {
		w := 26 + rng.Intn(30)
		h := 34 + rng.Intn(52)
		top := ground - h
		c.FillRect(x, top, w, h, wall)
		c.HLine(x, top, w, Scale(wall, 1.2))
		// pitched roof
		for i := 0; i < 6; i++ {
			c.HLine(x-3+i, top-6+i, w+6-i*2, roof)
		}
		// lit windows
		for wy := top + 8; wy < ground-6; wy += 10 {
			for wx := x + 4; wx < x+w-5; wx += 8 {
				col := Scale(wall, 0.55)
				if rng.Intn(3) == 0 {
					col = Alpha(ColGoldLight, 0.8)
				}
				c.FillRect(wx, wy, 4, 5, col)
			}
		}
		x += w + 8 + rng.Intn(24)
	}
	c.FillRect(0, ground, bgW, core.ScreenH-ground, Mix(t.SurfDark, t.SkyBot, 0.3))
	return c.Img()
}

func campusNear(t Theme, rng *rand.Rand) *ebiten.Image {
	c := newLayer()
	ground := core.ScreenH - 26
	trunk := Mix(t.SoilDark, t.SkyBot, 0.15)
	leaf := Mix(t.SurfFace, t.SkyBot, 0.15)
	for x := 4; x < bgW-10; x += 22 + rng.Intn(30) {
		h := 20 + rng.Intn(16)
		c.FillRect(x, ground-h, 4, h+8, trunk)
		c.FillEllipse(x+2, ground-h-4, 9+rng.Intn(4), 8+rng.Intn(3), leaf)
		c.FillEllipse(x-3, ground-h+1, 6, 5, Scale(leaf, 0.9))
		c.FillEllipse(x+8, ground-h+1, 6, 5, Scale(leaf, 1.08))
	}
	c.FillRect(0, ground+4, bgW, core.ScreenH-ground-4, Mix(t.SurfDark, t.SkyBot, 0.12))
	return c.Img()
}

func caveFar(t Theme, rng *rand.Rand) *ebiten.Image {
	c := newLayer()
	rock := Mix(t.SoilDark, t.SkyBot, 0.45)
	for x := 0; x < bgW; x++ {
		top := 40 + 22*math.Sin(float64(x)*0.012) + 10*math.Sin(float64(x)*0.037)
		bot := core.ScreenH - 30 - 20*math.Sin(float64(x)*0.009+2.0)
		for y := 0; y < int(top); y++ {
			c.Px(x, y, rock)
		}
		for y := int(bot); y < core.ScreenH; y++ {
			c.Px(x, y, rock)
		}
	}
	// stalactites
	for k := 0; k < 40; k++ {
		x := rng.Intn(bgW)
		h := 8 + rng.Intn(26)
		for i := 0; i < h; i++ {
			w := (h - i) / 5
			c.HLine(x-w, 40+i, w*2+1, rock)
		}
	}
	return c.Img()
}

func caveMid(t Theme, rng *rand.Rand) *ebiten.Image {
	c := newLayer()
	rock := Mix(t.SoilFace, t.SkyTop, 0.35)
	for x := 0; x < bgW; x++ {
		top := 70 + 26*math.Sin(float64(x)*0.017+1.1)
		for y := 0; y < int(top); y++ {
			c.Px(x, y, rock)
		}
	}
	// glowing index crystals embedded in the ceiling
	for k := 0; k < 70; k++ {
		x := rng.Intn(bgW)
		y := 20 + rng.Intn(70)
		h := 4 + rng.Intn(9)
		col := Alpha(t.Accent, 0.5+rng.Float64()*0.5)
		for i := 0; i < h; i++ {
			w := (h - i) / 4
			c.HLine(x-w, y+i, w*2+1, col)
		}
		c.Px(x, y+h/2, ColCream)
	}
	return c.Img()
}

func depthsFar(t Theme, rng *rand.Rand) *ebiten.Image {
	c := newLayer()
	// god rays from the distant surface
	for k := 0; k < 14; k++ {
		x := rng.Intn(bgW)
		w := 8 + rng.Intn(18)
		for y := 0; y < core.ScreenH; y++ {
			a := 0.16 * (1 - float64(y)/float64(core.ScreenH))
			for i := 0; i < w; i++ {
				c.Blend(x+i+y/4, y, color.RGBA{0xbf, 0xef, 0xff, 0xff}, a)
			}
		}
	}
	ridge := Mix(t.SoilDark, t.SkyBot, 0.4)
	base := core.ScreenH - 34
	for x := 0; x < bgW; x++ {
		h := 22 + 18*math.Sin(float64(x)*0.010+0.6) + 8*math.Sin(float64(x)*0.031)
		for y := base - int(h); y < core.ScreenH; y++ {
			c.Px(x, y, ridge)
		}
	}
	return c.Img()
}

func depthsMid(t Theme, rng *rand.Rand) *ebiten.Image {
	c := newLayer()
	ground := core.ScreenH - 22
	// kelp forest
	for x := 0; x < bgW; x += 5 + rng.Intn(9) {
		h := 30 + rng.Intn(80)
		col := Mix(t.PipeDark, t.SkyBot, 0.25)
		for y := 0; y < h; y++ {
			xx := x + int(4*math.Sin(float64(y)*0.06+float64(x)))
			c.Px(xx, ground-y, col)
			c.Px(xx+1, ground-y, Scale(col, 0.85))
		}
	}
	// coral clumps
	for k := 0; k < 26; k++ {
		x := rng.Intn(bgW)
		col := Mix(t.Accent, t.SkyBot, 0.45)
		c.FillEllipse(x, ground-4, 5+rng.Intn(6), 4+rng.Intn(4), col)
	}
	c.FillRect(0, ground, bgW, core.ScreenH-ground, Mix(t.SoilFace, t.SkyBot, 0.35))
	return c.Img()
}

func orbitFar(t Theme, rng *rand.Rand) *ebiten.Image {
	c := newLayer()
	// a gas giant hanging in the void
	cx, cy, r := bgW/3, 90, 58
	for j := -r; j <= r; j++ {
		for i := -r; i <= r; i++ {
			if i*i+j*j > r*r {
				continue
			}
			band := 0.5 + 0.5*math.Sin(float64(j)*0.22)
			col := Mix(color.RGBA{0x4a, 0x2e, 0x6e, 0xff}, color.RGBA{0x8a, 0x5a, 0xc4, 0xff}, band)
			// terminator shading
			lit := 1.0 - 0.55*float64(i+r)/float64(2*r)
			c.Px(cx+i, cy+j, Scale(col, 0.55+lit*0.8))
		}
	}
	// ring
	for a := 0.0; a < 6.283; a += 0.004 {
		for rr := r + 10; rr < r+20; rr++ {
			x := cx + int(float64(rr)*math.Cos(a))
			y := cy + int(float64(rr)*math.Sin(a)*0.22)
			c.Blend(x, y, Alpha(ColPurple, 0.5), 0.35)
		}
	}
	return c.Img()
}

func orbitMid(t Theme, rng *rand.Rand) *ebiten.Image {
	c := newLayer()
	// distant station trusses along the bottom
	metal := Mix(t.SurfDark, t.SkyBot, 0.35)
	y0 := core.ScreenH - 60
	c.FillRect(0, y0+30, bgW, 30, metal)
	for x := 0; x < bgW; x += 16 {
		c.Line(x, y0+30, x+16, y0+10, Scale(metal, 1.3))
		c.Line(x, y0+10, x+16, y0+30, Scale(metal, 1.3))
		c.VLine(x, y0+10, 20, Scale(metal, 1.1))
	}
	c.FillRect(0, y0+6, bgW, 5, Scale(metal, 1.45))
	// solar arrays
	for k := 0; k < 8; k++ {
		x := rng.Intn(bgW - 40)
		c.FillRect(x, y0-26, 34, 18, Mix(color.RGBA{0x24, 0x2a, 0x50, 0xff}, t.Accent, 0.25))
		for i := 0; i < 34; i += 4 {
			c.VLine(x+i, y0-26, 18, Scale(t.Accent, 0.5))
		}
		c.FillRect(x+15, y0-8, 4, 14, metal)
	}
	return c.Img()
}

func fortressFar(t Theme, rng *rand.Rand) *ebiten.Image {
	c := newLayer()
	glow := color.RGBA{0xff, 0x6b, 0x30, 0xff}
	// lava horizon glow
	for y := core.ScreenH - 70; y < core.ScreenH; y++ {
		f := float64(y-(core.ScreenH-70)) / 70
		c.HLine(0, y, bgW, Mix(t.SkyBot, glow, f*f*0.7))
	}
	// jagged mountains
	dark := color.RGBA{0x1a, 0x08, 0x0c, 0xff}
	base := core.ScreenH - 52
	for x := 0; x < bgW; x++ {
		h := 34 + 28*math.Sin(float64(x)*0.011) + 16*math.Sin(float64(x)*0.043+2.1)
		for y := base - int(h); y < core.ScreenH; y++ {
			c.Px(x, y, dark)
		}
	}
	return c.Img()
}

func fortressMid(t Theme, rng *rand.Rand) *ebiten.Image {
	c := newLayer()
	stone := Mix(t.SoilFace, t.SkyBot, 0.25)
	ground := core.ScreenH - 30
	x := 0
	for x < bgW {
		w := 22 + rng.Intn(20)
		h := 50 + rng.Intn(80)
		c.FillRect(x, ground-h, w, h, stone)
		c.HLine(x, ground-h, w, Scale(stone, 1.3))
		// crenellations
		for i := 0; i < w; i += 8 {
			c.FillRect(x+i, ground-h-6, 5, 6, stone)
		}
		// windows lit by furnace light
		for wy := ground - h + 14; wy < ground-10; wy += 16 {
			for wx := x + 5; wx < x+w-6; wx += 11 {
				c.FillRect(wx, wy, 4, 7, Alpha(color.RGBA{0xff, 0x9d, 0x3d, 0xff}, 0.9))
			}
		}
		x += w + 4 + rng.Intn(18)
	}
	c.FillRect(0, ground, bgW, core.ScreenH-ground, Scale(stone, 0.55))
	return c.Img()
}
