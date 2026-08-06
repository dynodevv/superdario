package gfx

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// The five members of the Distillation Syndicate. Each is drawn from
// primitives rather than ASCII art — they're large, and parametric shapes give
// them cleaner curves than a hand-typed grid would.

const (
	BossScrapemaster = iota // W1 — a scaled-up scraper drone
	BossBaron               // W2 — the Scraper Baron's drill rig
	BossLeviathan           // W3 — DeepSeek's mechanical whale
	BossLunar               // W4 — Moonshot's crescent lander
	BossCloudlord           // W5 — the Qwen cloud engine
	BossDistiller           // final — The Distiller itself
	BossCount
)

type BossArt struct {
	Frames []*ebiten.Image // idle animation
	W, H   int
}

// BossDims are the sprite sizes, kept as plain data so gameplay code can size
// a boss's hitbox without needing a graphics context.
var BossDims = [BossCount][2]int{
	BossScrapemaster: {44, 44},
	BossBaron:        {48, 46},
	BossLeviathan:    {64, 40},
	BossLunar:        {48, 48},
	BossCloudlord:    {56, 48},
	BossDistiller:    {60, 60},
}

// eye draws a glowing lens with a pupil that looks toward `look`.
func eye(c *Canvas, cx, cy, r int, iris, pupil color.RGBA, look float64) {
	c.FillCircle(cx, cy, r, color.RGBA{0xf4, 0xf2, 0xea, 0xff})
	c.Circle(cx, cy, r, Scale(iris, 0.5))
	px := cx + int(float64(r)*0.4*math.Cos(look))
	py := cy + int(float64(r)*0.4*math.Sin(look))
	c.FillCircle(px, py, maxi(1, r/2), iris)
	c.FillCircle(px, py, maxi(1, r/4), pupil)
	c.Px(cx-r/2, cy-r/2, color.RGBA{255, 255, 255, 255})
}

func maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// riveted panel used across all the machines
func panel(c *Canvas, x, y, w, h int, face color.RGBA) {
	c.Bevel(x, y, w, h, face, Scale(face, 1.3), Scale(face, 0.55))
	for i := x + 2; i < x+w-1; i += 6 {
		c.Px(i, y+2, Scale(face, 1.55))
		c.Px(i, y+h-3, Scale(face, 0.45))
	}
}

// BuildBoss renders one Syndicate boss.
func BuildBoss(kind int) *BossArt {
	rng := rand.New(rand.NewSource(int64(kind)*7919 + 13))
	switch kind {
	case BossScrapemaster:
		return buildScrapemaster(rng)
	case BossBaron:
		return buildBaron(rng)
	case BossLeviathan:
		return buildLeviathan(rng)
	case BossLunar:
		return buildLunar(rng)
	case BossCloudlord:
		return buildCloudlord(rng)
	default:
		return buildDistiller(rng)
	}
}

// --- W1: Scrapemaster Prime -------------------------------------------------
func buildScrapemaster(rng *rand.Rand) *BossArt {
	w, h := BossDims[BossScrapemaster][0], BossDims[BossScrapemaster][1]
	a := &BossArt{W: w, H: h}
	for f := 0; f < 2; f++ {
		c := NewCanvas(w, h)
		body := color.RGBA{0x77, 0x7d, 0x8a, 0xff}
		trim := ColCoral
		// treads
		c.FillRect(2, h-9, w-4, 8, color.RGBA{0x33, 0x36, 0x3d, 0xff})
		for x := 4 + f*3; x < w-4; x += 6 {
			c.FillRect(x, h-8, 3, 6, color.RGBA{0x55, 0x59, 0x63, 0xff})
		}
		// chassis
		panel(c, 4, 18, w-8, h-27, body)
		c.FillRect(9, 24, w-18, 8, Scale(body, 0.6))
		for i := 0; i < 5; i++ {
			col := trim
			if (i+f)%2 == 0 {
				col = Scale(trim, 1.4)
			}
			c.FillRect(11+i*5, 26, 3, 4, col)
		}
		// funnel head
		c.FillRect(12, 2, w-24, 4, color.RGBA{0xb6, 0xbc, 0xc8, 0xff})
		for y := 0; y < 8; y++ {
			wd := 20 - y*2
			c.HLine(w/2-wd/2, 6+y, wd, Mix(color.RGBA{0xd6, 0xdc, 0xe8, 0xff}, body, float64(y)/8))
		}
		panel(c, 8, 12, w-16, 10, Scale(body, 1.1))
		eye(c, w/2-7, 17, 4, ColCyan, ColInk, float64(f)*0.9-0.4)
		eye(c, w/2+7, 17, 4, ColCyan, ColInk, float64(f)*0.9-0.4)
		// arms
		armY := 22 + f*2
		panel(c, 0, armY, 6, 12, Scale(body, 0.85))
		panel(c, w-6, armY, 6, 12, Scale(body, 0.85))
		c.Speckle(0, 0, w, h, rng, 0.08, 0.25)
		c.Outline(ColInkDeep)
		a.Frames = append(a.Frames, c.Img())
	}
	return a
}

// --- W2: The Scraper Baron --------------------------------------------------
func buildBaron(rng *rand.Rand) *BossArt {
	w, h := BossDims[BossBaron][0], BossDims[BossBaron][1]
	a := &BossArt{W: w, H: h}
	for f := 0; f < 2; f++ {
		c := NewCanvas(w, h)
		body := color.RGBA{0x5a, 0x45, 0x7e, 0xff}
		metal := color.RGBA{0x8c, 0x93, 0xa6, 0xff}
		// drill (rotating stripes)
		for y := 0; y < 14; y++ {
			wd := 14 - y
			x0 := 2
			c.HLine(x0, 16+y/2, wd, metal)
			c.HLine(x0, 30-y/2, wd, metal)
		}
		for k := 0; k < 6; k++ {
			y := 17 + ((k*3 + f*2) % 12)
			c.HLine(2, y, 14-abs(y-23), Scale(metal, 0.6))
		}
		// hull
		panel(c, 14, 14, w-16, 22, body)
		c.FillRect(18, 20, w-24, 6, Scale(body, 0.55))
		for i := 0; i < 4; i++ {
			c.FillRect(20+i*6, 21, 4, 4, Scale(ColCyan, 0.6+0.5*float64((i+f)%2)))
		}
		// top hat + monocle: the Baron himself, riding up top
		c.FillRect(26, 2, 12, 8, ColInk)
		c.FillRect(22, 9, 20, 2, ColInk)
		c.FillRect(26, 6, 12, 2, ColCoral)
		c.FillEllipse(32, 13, 7, 4, color.RGBA{0xf2, 0xc6, 0xa0, 0xff})
		eye(c, 30, 13, 2, ColInk, ColInk, 0)
		c.Circle(35, 13, 3, ColGold)
		c.Px(35, 13, ColCyan)
		// treads
		c.FillRect(14, h-9, w-16, 8, color.RGBA{0x2a, 0x28, 0x36, 0xff})
		for x := 16 + f*3; x < w-4; x += 6 {
			c.FillRect(x, h-8, 3, 6, color.RGBA{0x4b, 0x47, 0x5e, 0xff})
		}
		c.Speckle(14, 14, w-16, 22, rng, 0.08, 0.25)
		c.Outline(ColInkDeep)
		a.Frames = append(a.Frames, c.Img())
	}
	return a
}

// --- W3: The DeepSeek Leviathan ---------------------------------------------
func buildLeviathan(rng *rand.Rand) *BossArt {
	w, h := BossDims[BossLeviathan][0], BossDims[BossLeviathan][1]
	a := &BossArt{W: w, H: h}
	for f := 0; f < 2; f++ {
		c := NewCanvas(w, h)
		skin := color.RGBA{0x2a, 0x4e, 0x7a, 0xff}
		belly := color.RGBA{0x8d, 0xb8, 0xd8, 0xff}
		// tail flukes
		c.FillEllipse(58, 12+f, 7, 5, skin)
		c.FillEllipse(58, 26-f, 7, 5, skin)
		// body
		c.FillEllipse(28, 20, 26, 14, skin)
		c.FillEllipse(26, 26, 20, 7, belly)
		for i := 0; i < 7; i++ {
			c.VLine(16+i*3, 24, 8, Scale(belly, 0.8))
		}
		// head plating
		panel(c, 2, 12, 18, 16, Scale(skin, 1.25))
		// jaw
		jaw := 4 + f*3
		c.FillRect(2, 26+jaw/2, 20, 4, Scale(skin, 0.7))
		for i := 0; i < 6; i++ {
			c.FillRect(4+i*3, 24+jaw/2, 2, 3, ColCream)
		}
		eye(c, 12, 18, 4, color.RGBA{0x3d, 0xd6, 0xe8, 0xff}, ColInk, 3.14)
		// blowhole jet
		if f == 1 {
			for k := 0; k < 5; k++ {
				c.FillCircle(24+rng.Intn(4), 4+k, 2-k/3, Alpha(ColCream, 0.8-float64(k)*0.15))
			}
		}
		// fin
		c.FillEllipse(30, 30, 8, 4, Scale(skin, 0.85))
		c.Speckle(4, 8, 56, 28, rng, 0.07, 0.2)
		c.Outline(ColInkDeep)
		a.Frames = append(a.Frames, c.Img())
	}
	return a
}

// --- W4: The Lunar Lord -----------------------------------------------------
func buildLunar(rng *rand.Rand) *BossArt {
	w, h := BossDims[BossLunar][0], BossDims[BossLunar][1]
	a := &BossArt{W: w, H: h}
	for f := 0; f < 2; f++ {
		c := NewCanvas(w, h)
		moon := color.RGBA{0xd9, 0xd4, 0xe6, 0xff}
		shade := color.RGBA{0x8a, 0x84, 0xa2, 0xff}
		// Crescent: a disc minus an offset disc. Done in one pass so the bite
		// has a clean edge instead of being painted over.
		const mx, my, mr = 22, 24, 19 // main disc
		const bx, by, br = 36, 21, 17 // the bite
		for j := -mr; j <= mr; j++ {
			for i := -mr; i <= mr; i++ {
				if i*i+j*j > mr*mr {
					continue
				}
				px, py := mx+i, my+j
				dx, dy := px-bx, py-by
				if dx*dx+dy*dy < br*br {
					continue
				}
				// shade toward the inner edge of the crescent
				d := math.Sqrt(float64(dx*dx+dy*dy)) - br
				c.Px(px, py, Mix(shade, moon, math.Min(1, d/9)))
			}
		}
		// craters
		for k := 0; k < 9; k++ {
			cx, cy := 8+rng.Intn(14), 10+rng.Intn(28)
			dx, dy := cx-bx, cy-by
			if dx*dx+dy*dy < br*br {
				continue
			}
			r := 1 + rng.Intn(3)
			c.FillCircle(cx, cy, r, shade)
			c.FillCircle(cx, cy-1, r, Scale(shade, 1.25))
		}
		// face, tucked into the thick part of the crescent
		eye(c, 13, 17, 4, ColPurple, ColInk, 0.4+float64(f)*0.5)
		eye(c, 13, 31, 4, ColPurple, ColInk, -0.4-float64(f)*0.5)
		c.FillRect(7, 23, 6, 3, Scale(shade, 0.65))
		c.HLine(7, 23, 6, Scale(shade, 0.4))
		// thruster ring around the outside
		for k := 0; k < 10; k++ {
			ang := float64(k)/10*6.283 + float64(f)*0.3
			cx := mx + int(21*math.Cos(ang))
			cy := my + int(21*math.Sin(ang))
			c.FillCircle(cx, cy, 2, Alpha(ColPurple, 0.85))
			c.Px(cx, cy, ColCream)
		}
		c.Outline(ColInkDeep)
		a.Frames = append(a.Frames, c.Img())
	}
	return a
}

// --- W5: The Qwen Cloudlord -------------------------------------------------
func buildCloudlord(rng *rand.Rand) *BossArt {
	w, h := BossDims[BossCloudlord][0], BossDims[BossCloudlord][1]
	a := &BossArt{W: w, H: h}
	for f := 0; f < 2; f++ {
		c := NewCanvas(w, h)
		cloud := color.RGBA{0x4a, 0x3c, 0x4e, 0xff}
		hot := color.RGBA{0xff, 0x9d, 0x3d, 0xff}
		// thunderhead body
		c.FillEllipse(16, 22, 14, 12, cloud)
		c.FillEllipse(30, 18, 16, 13, cloud)
		c.FillEllipse(43, 24, 12, 10, cloud)
		c.FillRect(6, 22, 44, 14, cloud)
		c.FillEllipse(28, 34, 24, 8, Scale(cloud, 0.75))
		// server rack core glowing inside
		panel(c, 20, 16, 18, 20, color.RGBA{0x2c, 0x22, 0x30, 0xff})
		for r := 0; r < 5; r++ {
			for cc := 0; cc < 4; cc++ {
				on := (r*4+cc+f*3)%5 != 0
				col := Scale(hot, 0.35)
				if on {
					col = hot
				}
				c.FillRect(22+cc*4, 18+r*4, 3, 2, col)
			}
		}
		// three eyes in the cloud
		eye(c, 12, 20, 3, hot, ColInk, float64(f))
		eye(c, 44, 22, 3, hot, ColInk, float64(f)+1)
		eye(c, 28, 10, 4, ColGoldLight, ColInk, -1.57)
		// lightning
		if f == 1 {
			c.Line(20, 38, 16, 44, ColGoldLight)
			c.Line(16, 44, 20, 46, ColGoldLight)
			c.Line(38, 38, 42, 45, ColGoldLight)
		}
		c.Speckle(4, 8, 48, 32, rng, 0.1, 0.25)
		c.Outline(ColInkDeep)
		a.Frames = append(a.Frames, c.Img())
	}
	return a
}

// --- Final: The Distiller ---------------------------------------------------
func buildDistiller(rng *rand.Rand) *BossArt {
	w, h := BossDims[BossDistiller][0], BossDims[BossDistiller][1]
	a := &BossArt{W: w, H: h}
	for f := 0; f < 2; f++ {
		c := NewCanvas(w, h)
		iron := color.RGBA{0x3a, 0x36, 0x44, 0xff}
		glass := color.RGBA{0x7f, 0xe4, 0xf0, 0xff}
		// furnace base
		panel(c, 6, 40, w-12, 18, iron)
		mouth := 6 + f*4
		c.FillRect(16, 46, w-32, mouth, color.RGBA{0xff, 0x6b, 0x30, 0xff})
		for k := 0; k < 8; k++ {
			c.FillCircle(18+rng.Intn(w-36), 46+rng.Intn(mouth), 1, ColGoldLight)
		}
		// alembic bulb full of stolen weights
		c.FillEllipse(30, 26, 20, 18, Alpha(glass, 0.55))
		c.Circle(30, 26, 18, Scale(glass, 0.8))
		for k := 0; k < 14; k++ {
			ang := float64(k)*0.9 + float64(f)*0.4
			r := 4.0 + float64(k%5)*2.4
			cx := 30 + int(r*math.Cos(ang))
			cy := 26 + int(r*math.Sin(ang)*0.85)
			c.FillCircle(cx, cy, 2, ModelColors[k%4])
		}
		// condenser neck + face plate
		panel(c, 24, 2, 12, 10, iron)
		c.FillRect(20, 10, 20, 4, Scale(iron, 1.2))
		eye(c, 22, 24, 5, ColCoral, ColInk, 0.5*float64(f))
		eye(c, 38, 24, 5, ColCoral, ColInk, 3.14-0.5*float64(f))
		// pipes running down each side
		for _, x := range []int{2, w - 6} {
			c.FillRect(x, 14, 4, 30, Scale(iron, 1.15))
			for y := 16; y < 42; y += 6 {
				c.FillRect(x-1, y, 6, 2, Scale(iron, 0.7))
			}
		}
		c.Outline(ColInkDeep)
		a.Frames = append(a.Frames, c.Img())
	}
	return a
}

// ---------------------------------------------------------------------------
// The captured Claude models
// ---------------------------------------------------------------------------

// ModelSprite draws a rescued Claude as a small glowing entity: a rounded
// bubble body with a calm face and a drifting halo of tokens.
func ModelSprite(idx int, frame int) *ebiten.Image {
	col := ModelColors[idx%4]
	c := NewCanvas(16, 16)
	bob := 0
	if frame%2 == 1 {
		bob = 1
	}
	// glow
	for r := 8; r >= 6; r-- {
		c.FillCircle(8, 8+bob, r, Alpha(col, 0.18))
	}
	// body
	c.FillEllipse(8, 8+bob, 6, 6, col)
	c.FillEllipse(8, 6+bob, 5, 4, Scale(col, 1.28))
	// speech-bubble tail — these are language models, after all
	c.FillRect(4, 13+bob, 3, 2, col)
	c.Px(3, 14+bob, col)
	// face
	c.FillRect(6, 7+bob, 1, 2, ColInkDeep)
	c.FillRect(10, 7+bob, 1, 2, ColInkDeep)
	c.HLine(7, 10+bob, 3, ColInkDeep)
	c.Px(6, 9+bob, ColInkDeep)
	c.Px(10, 9+bob, ColInkDeep)
	c.Outline(Alpha(ColInkDeep, 0.85))
	return c.Img()
}

// ModelNames as they appear in the story, in rescue order.
var ModelNames = [4]string{"HAIKU", "SONNET", "OPUS", "FABLE"}

// BossNames and the labs behind them.
var BossNames = [BossCount]string{
	"SCRAPEMASTER PRIME",
	"THE SCRAPER BARON",
	"THE LEVIATHAN",
	"THE LUNAR LORD",
	"THE CLOUDLORD",
	"THE DISTILLER",
}
