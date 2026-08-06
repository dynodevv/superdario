package gfx

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

const TS = 16 // tile size

// Theme carries every colour a world needs. Tiles are generated from it at
// startup, so a whole new world look is one struct literal.
type Theme struct {
	Name string

	SurfFace, SurfLight, SurfDark color.RGBA // the cap on top of ground
	SoilFace, SoilLight, SoilDark color.RGBA // ground body
	BrickFace, BrickMortar        color.RGBA
	BlockFace, BlockEdge          color.RGBA
	PipeFace, PipeLight, PipeDark color.RGBA
	Accent                        color.RGBA // question blocks, trim
	SkyTop, SkyBot                color.RGBA
	HazeCol                       color.RGBA
	Underground                   bool
	Liquid                        LiquidKind
}

type LiquidKind int

const (
	LiquidNone LiquidKind = iota
	LiquidWater
	LiquidLava
	LiquidVoid // "the null gradient" in the fortress
)

// Tileset is the finished GPU-side art for one theme.
type Tileset struct {
	Theme      Theme
	Surf       [3]*ebiten.Image // left cap, middle cap, right cap
	Soil       *ebiten.Image
	SoilEdgeL  *ebiten.Image
	SoilEdgeR  *ebiten.Image
	Brick      *ebiten.Image
	BrickSolid *ebiten.Image
	Block      *ebiten.Image
	Question   [4]*ebiten.Image // 3 idle frames + spent
	Pipe       [6]*ebiten.Image // TL TR BL BR, and horizontal mouth top/bottom
	SpikeFloor *ebiten.Image    // stands on the ground, points up
	SpikeCeil  *ebiten.Image    // hangs from the ceiling, points down
	Liquid     [3]*ebiten.Image
	LiquidTop  [3]*ebiten.Image
	Platform   *ebiten.Image
	Girder     *ebiten.Image // full-height structural tile (solid)
	GirderBeam *ebiten.Image // thin rail, for moving platforms
	Belt       [4]*ebiten.Image
	Deco       [4]*ebiten.Image
	Ladder     *ebiten.Image
	Flag       *ebiten.Image
	Cloud      *ebiten.Image
}

// Themes for the five worlds.
var Themes = []Theme{
	{ // 1 — Anthropic Campus: warm cream stone, clipped lawns, blue sky
		Name:      "ANTHROPIC CAMPUS",
		SurfFace:  color.RGBA{0x4f, 0xb0, 0x6e, 0xff},
		SurfLight: color.RGBA{0x86, 0xdd, 0x95, 0xff},
		SurfDark:  color.RGBA{0x27, 0x77, 0x46, 0xff},
		SoilFace:  color.RGBA{0xb5, 0x84, 0x53, 0xff},
		SoilLight: color.RGBA{0xd7, 0xa9, 0x73, 0xff},
		SoilDark:  color.RGBA{0x74, 0x4e, 0x2c, 0xff},
		BrickFace: color.RGBA{0xc9, 0x6f, 0x4a, 0xff}, BrickMortar: color.RGBA{0x7b, 0x3a, 0x24, 0xff},
		BlockFace: color.RGBA{0xef, 0xea, 0xdc, 0xff}, BlockEdge: color.RGBA{0xa8, 0xa0, 0x8d, 0xff},
		PipeFace: color.RGBA{0x2f, 0xa0, 0x76, 0xff}, PipeLight: color.RGBA{0x6b, 0xd8, 0xa6, 0xff}, PipeDark: color.RGBA{0x14, 0x54, 0x3d, 0xff},
		Accent: color.RGBA{0xe0, 0xba, 0x45, 0xff},
		SkyTop: color.RGBA{0x54, 0x9d, 0xe0, 0xff}, SkyBot: color.RGBA{0xbf, 0xe4, 0xf2, 0xff},
		HazeCol: color.RGBA{0xff, 0xf4, 0xdd, 0xff},
		Liquid:  LiquidWater,
	},
	{ // 2 — The Data Mines: cold rock, glowing indices
		Name:      "THE DATA MINES",
		SurfFace:  color.RGBA{0x46, 0x5a, 0x74, 0xff},
		SurfLight: color.RGBA{0x74, 0x91, 0xb0, 0xff},
		SurfDark:  color.RGBA{0x22, 0x2e, 0x40, 0xff},
		SoilFace:  color.RGBA{0x33, 0x3d, 0x52, 0xff},
		SoilLight: color.RGBA{0x4f, 0x5d, 0x77, 0xff},
		SoilDark:  color.RGBA{0x18, 0x1e, 0x2c, 0xff},
		BrickFace: color.RGBA{0x53, 0x4a, 0x63, 0xff}, BrickMortar: color.RGBA{0x2a, 0x24, 0x33, 0xff},
		BlockFace: color.RGBA{0x69, 0x75, 0x8c, 0xff}, BlockEdge: color.RGBA{0x2c, 0x35, 0x47, 0xff},
		PipeFace: color.RGBA{0x2a, 0x8b, 0x99, 0xff}, PipeLight: color.RGBA{0x62, 0xd6, 0xe0, 0xff}, PipeDark: color.RGBA{0x11, 0x44, 0x4d, 0xff},
		Accent: color.RGBA{0x62, 0xd6, 0xe0, 0xff},
		SkyTop: color.RGBA{0x0b, 0x10, 0x1c, 0xff}, SkyBot: color.RGBA{0x1c, 0x2a, 0x44, 0xff},
		HazeCol:     color.RGBA{0x2a, 0x8b, 0x99, 0xff},
		Underground: true,
		Liquid:      LiquidWater,
	},
	{ // 3 — Whale Depths: DeepSeek's trench
		Name:      "WHALE DEPTHS",
		SurfFace:  color.RGBA{0x2b, 0x6f, 0x8c, 0xff},
		SurfLight: color.RGBA{0x54, 0xa8, 0xc4, 0xff},
		SurfDark:  color.RGBA{0x13, 0x3c, 0x52, 0xff},
		SoilFace:  color.RGBA{0x1c, 0x4a, 0x63, 0xff},
		SoilLight: color.RGBA{0x2f, 0x6c, 0x8a, 0xff},
		SoilDark:  color.RGBA{0x0c, 0x26, 0x38, 0xff},
		BrickFace: color.RGBA{0x25, 0x59, 0x77, 0xff}, BrickMortar: color.RGBA{0x0e, 0x2b, 0x3d, 0xff},
		BlockFace: color.RGBA{0x3d, 0x7f, 0x9c, 0xff}, BlockEdge: color.RGBA{0x14, 0x38, 0x4d, 0xff},
		PipeFace: color.RGBA{0x2c, 0x7d, 0x62, 0xff}, PipeLight: color.RGBA{0x63, 0xc9, 0xa4, 0xff}, PipeDark: color.RGBA{0x0f, 0x3a, 0x2c, 0xff},
		Accent: color.RGBA{0x7f, 0xe4, 0xf0, 0xff},
		SkyTop: color.RGBA{0x04, 0x14, 0x2c, 0xff}, SkyBot: color.RGBA{0x0d, 0x39, 0x60, 0xff},
		HazeCol: color.RGBA{0x1a, 0x74, 0xa8, 0xff},
		Liquid:  LiquidWater,
	},
	{ // 4 — Moonshot Orbital: low gravity, cold panels, starfield
		Name:      "MOONSHOT ORBITAL",
		SurfFace:  color.RGBA{0x8e, 0x8a, 0x9e, 0xff},
		SurfLight: color.RGBA{0xc8, 0xc4, 0xd8, 0xff},
		SurfDark:  color.RGBA{0x4c, 0x48, 0x5c, 0xff},
		SoilFace:  color.RGBA{0x4a, 0x46, 0x5a, 0xff},
		SoilLight: color.RGBA{0x6d, 0x68, 0x80, 0xff},
		SoilDark:  color.RGBA{0x25, 0x22, 0x30, 0xff},
		BrickFace: color.RGBA{0x59, 0x53, 0x7a, 0xff}, BrickMortar: color.RGBA{0x2b, 0x27, 0x3f, 0xff},
		BlockFace: color.RGBA{0xa9, 0xa4, 0xc2, 0xff}, BlockEdge: color.RGBA{0x4a, 0x45, 0x63, 0xff},
		PipeFace: color.RGBA{0x6b, 0x4f, 0xb5, 0xff}, PipeLight: color.RGBA{0xb0, 0x96, 0xf0, 0xff}, PipeDark: color.RGBA{0x30, 0x1e, 0x60, 0xff},
		Accent: color.RGBA{0xba, 0x96, 0xec, 0xff},
		SkyTop: color.RGBA{0x05, 0x03, 0x14, 0xff}, SkyBot: color.RGBA{0x1c, 0x12, 0x3c, 0xff},
		HazeCol: color.RGBA{0x6b, 0x4f, 0xb5, 0xff},
		Liquid:  LiquidVoid,
	},
	{ // 5 — Qwen Cloud Fortress: the final furnace
		Name:      "CLOUD FORTRESS",
		SurfFace:  color.RGBA{0x6a, 0x3b, 0x3b, 0xff},
		SurfLight: color.RGBA{0x9c, 0x5c, 0x55, 0xff},
		SurfDark:  color.RGBA{0x38, 0x1c, 0x1e, 0xff},
		SoilFace:  color.RGBA{0x45, 0x25, 0x28, 0xff},
		SoilLight: color.RGBA{0x66, 0x39, 0x3a, 0xff},
		SoilDark:  color.RGBA{0x22, 0x10, 0x13, 0xff},
		BrickFace: color.RGBA{0x5b, 0x33, 0x35, 0xff}, BrickMortar: color.RGBA{0x24, 0x11, 0x14, 0xff},
		BlockFace: color.RGBA{0x7d, 0x4b, 0x46, 0xff}, BlockEdge: color.RGBA{0x33, 0x18, 0x1a, 0xff},
		PipeFace: color.RGBA{0xa8, 0x4a, 0x22, 0xff}, PipeLight: color.RGBA{0xf0, 0x93, 0x4a, 0xff}, PipeDark: color.RGBA{0x50, 0x1c, 0x0c, 0xff},
		Accent: color.RGBA{0xff, 0x9d, 0x3d, 0xff},
		SkyTop: color.RGBA{0x14, 0x04, 0x08, 0xff}, SkyBot: color.RGBA{0x54, 0x12, 0x14, 0xff},
		HazeCol: color.RGBA{0xff, 0x6b, 0x30, 0xff},
		Liquid:  LiquidLava,
	},
}

// BuildTileset renders every tile for a theme.
func BuildTileset(t Theme, seed int64) *Tileset {
	rng := rand.New(rand.NewSource(seed))
	ts := &Tileset{Theme: t}

	// --- ground surface caps (left / middle / right) --------------------
	for i := 0; i < 3; i++ {
		c := NewCanvas(TS, TS)
		c.FillRect(0, 0, TS, TS, t.SoilFace)
		// soil texture
		c.Speckle(0, 0, TS, TS, rng, 0.16, 0.5)
		for k := 0; k < 5; k++ {
			x, y := rng.Intn(TS-3), 5+rng.Intn(TS-7)
			c.FillRect(x, y, 2+rng.Intn(2), 2, t.SoilDark)
		}
		// cap
		capH := 5
		c.FillRect(0, 0, TS, capH, t.SurfFace)
		c.HLine(0, 0, TS, t.SurfLight)
		c.HLine(0, capH-1, TS, t.SurfDark)
		// tufts poking down into the soil so the seam isn't a straight line
		for x := 0; x < TS; x += 2 {
			if rng.Intn(2) == 0 {
				c.Px(x, capH, t.SurfFace)
			}
		}
		if !t.Underground {
			for x := 1; x < TS; x += 4 {
				c.Px(x, 1, t.SurfLight)
				c.Px(x+1, 2, t.SurfLight)
			}
		}
		if i == 0 { // left edge highlight
			c.VLine(0, 0, TS, t.SurfLight)
			c.VLine(0, capH, TS-capH, Scale(t.SoilLight, 1.0))
		}
		if i == 2 { // right edge shadow
			c.VLine(TS-1, 0, TS, t.SurfDark)
			c.VLine(TS-1, capH, TS-capH, t.SoilDark)
		}
		ts.Surf[i] = c.Img()
	}

	// --- soil body ------------------------------------------------------
	mkSoil := func(edgeL, edgeR bool) *ebiten.Image {
		c := NewCanvas(TS, TS)
		c.FillRect(0, 0, TS, TS, t.SoilFace)
		c.Speckle(0, 0, TS, TS, rng, 0.18, 0.55)
		for k := 0; k < 6; k++ {
			x, y := rng.Intn(TS-3), rng.Intn(TS-3)
			c.FillRect(x, y, 2, 2, t.SoilDark)
			c.Px(x, y, t.SoilLight)
		}
		if edgeL {
			c.VLine(0, 0, TS, t.SoilLight)
		}
		if edgeR {
			c.VLine(TS-1, 0, TS, t.SoilDark)
		}
		return c.Img()
	}
	ts.Soil = mkSoil(false, false)
	ts.SoilEdgeL = mkSoil(true, false)
	ts.SoilEdgeR = mkSoil(false, true)

	// --- brick ----------------------------------------------------------
	{
		c := NewCanvas(TS, TS)
		c.FillRect(0, 0, TS, TS, t.BrickMortar)
		drawBrick := func(x, y, w, h int) {
			c.Bevel(x, y, w, h, t.BrickFace, Scale(t.BrickFace, 1.28), Scale(t.BrickFace, 0.7))
		}
		drawBrick(0, 0, 7, 7)
		drawBrick(8, 0, 8, 7)
		drawBrick(0, 8, 11, 7)
		drawBrick(12, 8, 4, 7)
		c.Speckle(0, 0, TS, TS, rng, 0.09, 0.35)
		ts.Brick = c.Img()
	}
	{ // solid variant used for indestructible walls
		c := NewCanvas(TS, TS)
		c.Bevel(0, 0, TS, TS, t.BlockFace, Scale(t.BlockFace, 1.25), Scale(t.BlockFace, 0.62))
		c.Rect(2, 2, TS-4, TS-4, t.BlockEdge)
		for _, p := range [][2]int{{3, 3}, {TS - 4, 3}, {3, TS - 4}, {TS - 4, TS - 4}} {
			c.Px(p[0], p[1], Scale(t.BlockFace, 1.4))
		}
		c.Speckle(0, 0, TS, TS, rng, 0.08, 0.3)
		ts.BrickSolid = c.Img()
		ts.Block = ts.BrickSolid
	}

	// --- question block -------------------------------------------------
	for f := 0; f < 4; f++ {
		c := NewCanvas(TS, TS)
		base := t.Accent
		if f == 3 {
			base = Scale(t.BlockFace, 0.75) // spent
		} else {
			// gentle brightness pulse across the three idle frames
			base = Scale(base, 1.0+0.11*math.Sin(float64(f)*2.094))
		}
		c.Bevel(0, 0, TS, TS, base, Scale(base, 1.35), Scale(base, 0.55))
		c.Rect(1, 1, TS-2, TS-2, Scale(base, 0.72))
		for _, p := range [][2]int{{2, 2}, {TS - 3, 2}, {2, TS - 3}, {TS - 3, TS - 3}} {
			c.Px(p[0], p[1], Scale(base, 1.5))
		}
		if f != 3 {
			mark := color.RGBA{0x3a, 0x24, 0x10, 0xff}
			glow := Scale(base, 1.6)
			for y, row := range glyphArt['?'] {
				for x := 0; x < len(row); x++ {
					if row[x] == '#' {
						c.Px(5+x, 4+y+1, glow)
						c.Px(5+x, 4+y, mark)
					}
				}
			}
		} else {
			c.FillRect(4, 6, 8, 4, Scale(base, 0.85))
		}
		ts.Question[f] = c.Img()
	}

	// --- pipes ----------------------------------------------------------
	{
		mk := func(kind int) *ebiten.Image {
			c := NewCanvas(TS, TS)
			// kind: 0 TL, 1 TR, 2 BL, 3 BR
			top := kind < 2
			left := kind%2 == 0
			x0, w := 0, TS
			if top {
				// lip overhangs by 2px on the outer side
				c.FillRect(x0, 0, w, 6, t.PipeFace)
				c.HLine(x0, 0, w, t.PipeLight)
				c.HLine(x0, 5, w, t.PipeDark)
				if left {
					c.VLine(0, 0, 6, t.PipeLight)
					c.FillRect(2, 1, 2, 4, t.PipeLight)
				} else {
					c.VLine(TS-1, 0, 6, t.PipeDark)
					c.FillRect(TS-5, 1, 1, 4, Scale(t.PipeFace, 0.8))
				}
			}
			y0 := 0
			if top {
				y0 = 6
			}
			bw, bx := TS-2, 0
			if left {
				bx = 2
			}
			c.FillRect(bx, y0, bw, TS-y0, t.PipeFace)
			if left {
				c.VLine(2, y0, TS-y0, t.PipeDark)
				c.FillRect(4, y0, 2, TS-y0, t.PipeLight)
			} else {
				c.VLine(TS-1, y0, TS-y0, t.PipeDark)
				c.FillRect(TS-4, y0, 1, TS-y0, Scale(t.PipeFace, 0.78))
			}
			c.Speckle(0, 0, TS, TS, rng, 0.07, 0.25)
			return c.Img()
		}
		for i := 0; i < 4; i++ {
			ts.Pipe[i] = mk(i)
		}
		// two spare slots reserved for horizontal pipe mouths
		ts.Pipe[4] = ts.Pipe[0]
		ts.Pipe[5] = ts.Pipe[2]
	}

	// --- spikes ---------------------------------------------------------
	{
		// Three spikes on a mounting plate. The old version computed its
		// width as (11-h)/2 while stepping y downward, which drew each spike
		// widest at the *top* — floor spikes rendered as hanging wedges, 11px
		// across, overlapping each other into a row of blunt teeth.
		mkSpikes := func(ceiling bool) *ebiten.Image {
			c := NewCanvas(TS, TS)
			metal := color.RGBA{0xc4, 0xc9, 0xd2, 0xff}
			hi := color.RGBA{0xee, 0xf2, 0xf7, 0xff}
			edge := color.RGBA{0x6e, 0x76, 0x86, 0xff}
			base := color.RGBA{0x4a, 0x50, 0x5c, 0xff}
			if t.Underground || t.Liquid == LiquidLava {
				metal = color.RGBA{0xdc, 0xcd, 0xb2, 0xff}
				hi = color.RGBA{0xf6, 0xec, 0xd8, 0xff}
				edge = color.RGBA{0x82, 0x70, 0x56, 0xff}
				base = color.RGBA{0x5c, 0x4d, 0x3a, 0xff}
			}

			const plate = 3      // mounting plate
			const h = TS - plate // spike height
			// Two broad spikes per tile rather than three narrow ones: at this
			// size three tapered to near-needles and read as a picket fence.
			for _, cx := range [2]int{4, 11} {
				for i := 0; i < h; i++ {
					// i counts from the base to the tip: 7px wide down to 1px
					half := 3 * (h - 1 - i) / (h - 1)
					y := TS - 1 - plate - i // standing on the floor
					if ceiling {
						y = plate + i // hanging from the roof
					}
					c.HLine(cx-half, y, half*2+1, metal)
					if half > 0 {
						c.Px(cx-half, y, hi)   // lit edge
						c.Px(cx+half, y, edge) // shaded edge
					}
				}
			}
			if ceiling {
				c.FillRect(0, 0, TS, plate, base)
				c.HLine(0, plate-1, TS, Scale(base, 1.35))
			} else {
				c.FillRect(0, TS-plate, TS, plate, base)
				c.HLine(0, TS-plate, TS, Scale(base, 1.35))
			}
			return c.Img()
		}
		ts.SpikeFloor = mkSpikes(false)
		ts.SpikeCeil = mkSpikes(true)
	}

	// --- liquid ---------------------------------------------------------
	{
		var base, hi color.RGBA
		switch t.Liquid {
		case LiquidLava:
			base, hi = color.RGBA{0xe0, 0x42, 0x10, 0xff}, color.RGBA{0xff, 0xd0, 0x60, 0xff}
		case LiquidVoid:
			base, hi = color.RGBA{0x2a, 0x12, 0x50, 0xff}, color.RGBA{0xba, 0x96, 0xec, 0xff}
		default:
			base, hi = color.RGBA{0x1e, 0x64, 0xb4, 0xff}, color.RGBA{0x86, 0xd4, 0xf0, 0xff}
		}
		for f := 0; f < 3; f++ {
			c := NewCanvas(TS, TS)
			c.FillRect(0, 0, TS, TS, base)
			for y := 0; y < TS; y++ {
				w := math.Sin(float64(y)*0.6 + float64(f)*2.1)
				c.HLine(0, y, TS, Mix(base, hi, 0.10+0.10*w))
			}
			ts.Liquid[f] = c.Img()

			s := NewCanvas(TS, TS)
			s.FillRect(0, 0, TS, TS, base)
			for y := 0; y < TS; y++ {
				w := math.Sin(float64(y)*0.6 + float64(f)*2.1)
				s.HLine(0, y, TS, Mix(base, hi, 0.10+0.10*w))
			}
			// wavy top surface
			for x := 0; x < TS; x++ {
				h := int(2 + 1.8*math.Sin(float64(x)*0.55+float64(f)*2.1))
				for y := 0; y < h; y++ {
					s.Px(x, y, color.RGBA{})
				}
				s.Px(x, h, hi)
				s.Px(x, h+1, Mix(hi, base, 0.5))
			}
			ts.LiquidTop[f] = s.Img()
		}
	}

	// --- one-way platform ----------------------------------------------
	{
		c := NewCanvas(TS, TS)
		face := t.BlockFace
		c.FillRect(0, 0, TS, 6, face)
		c.HLine(0, 0, TS, Scale(face, 1.35))
		c.HLine(0, 5, TS, Scale(face, 0.5))
		for x := 1; x < TS; x += 5 {
			c.VLine(x, 1, 4, Scale(face, 0.8))
		}
		c.FillRect(3, 6, 2, 3, Scale(face, 0.6))
		c.FillRect(TS-5, 6, 2, 3, Scale(face, 0.6))
		ts.Platform = c.Img()
	}

	// --- girder ----------------------------------------------------------
	// A full-height structural tile. Girders are solid and world 4's floors
	// are built from them, so drawing only the top 8px left a transparent
	// slot under the player's feet — the ground looked like it was floating.
	{
		c := NewCanvas(TS, TS)
		face := Mix(t.BlockFace, t.Accent, 0.25)
		web := Scale(face, 0.42)
		c.FillRect(0, 0, TS, TS, web)
		// lattice between the flanges
		c.Line(0, TS-5, TS/2, 5, Scale(face, 0.85))
		c.Line(TS/2, 5, TS-1, TS-5, Scale(face, 0.85))
		c.VLine(TS/2, 5, TS-9, Scale(face, 0.7))
		// top and bottom flanges
		c.Bevel(0, 0, TS, 5, face, Scale(face, 1.35), Scale(face, 0.55))
		c.Bevel(0, TS-4, TS, 4, face, Scale(face, 1.2), Scale(face, 0.5))
		for x := 2; x < TS-2; x += 5 {
			c.Px(x, 2, Scale(face, 1.55))
		}
		ts.Girder = c.Img()
	}
	// A thin beam for moving platforms, which should read as a floating rail
	// rather than a slab.
	{
		c := NewCanvas(TS, TS)
		face := Mix(t.BlockFace, t.Accent, 0.25)
		c.Bevel(0, 0, TS, 8, face, Scale(face, 1.3), Scale(face, 0.55))
		for x := 2; x < TS-2; x += 4 {
			c.Px(x, 2, Scale(face, 1.5))
			c.Px(x, 5, Scale(face, 0.5))
		}
		ts.GirderBeam = c.Img()
	}

	// --- conveyor belt (4 scroll frames) --------------------------------
	// Also full height: a belt is a solid tile, and a half-height sprite left
	// a see-through gap beneath it.
	for f := 0; f < 4; f++ {
		c := NewCanvas(TS, TS)
		face := Scale(t.BlockFace, 0.8)
		housing := Scale(t.BlockFace, 0.40)
		c.FillRect(0, 0, TS, TS, housing)
		// rollers turning inside the housing
		for x := 3; x < TS; x += 7 {
			c.FillCircle(x, 11, 2, Scale(t.BlockFace, 0.62))
			c.Px(x-1+f%2, 10, Scale(t.BlockFace, 0.95))
		}
		c.HLine(0, TS-1, TS, Scale(t.BlockFace, 0.28))
		// the moving surface
		c.FillRect(0, 0, TS, 8, face)
		c.HLine(0, 0, TS, Scale(face, 1.4))
		c.HLine(0, 7, TS, Scale(face, 0.45))
		for x := -8; x < TS+8; x += 8 {
			c.Line(x+f*2, 6, x+f*2+4, 1, t.Accent)
		}
		ts.Belt[f] = c.Img()
	}

	// --- decorations ----------------------------------------------------
	for i := 0; i < 4; i++ {
		c := NewCanvas(TS, TS)
		switch t.Name {
		case "ANTHROPIC CAMPUS":
			// tufts of grass and small shrubs
			g := t.SurfFace
			for k := 0; k < 4+i; k++ {
				x := 2 + rng.Intn(12)
				h := 3 + rng.Intn(5)
				c.VLine(x, TS-h, h, Mix(g, t.SurfLight, rng.Float64()))
			}
		case "THE DATA MINES":
			// glowing index crystals
			for k := 0; k < 2+i; k++ {
				x, y := 3+rng.Intn(10), 6+rng.Intn(8)
				c.Line(x, TS-1, x, y, t.Accent)
				c.Px(x, y, color.RGBA{0xff, 0xff, 0xff, 0xff})
			}
		case "WHALE DEPTHS":
			// kelp
			for k := 0; k < 2+i; k++ {
				x := 3 + rng.Intn(10)
				for y := TS - 1; y > 2; y-- {
					c.Px(x+int(1.4*math.Sin(float64(y)*0.5)), y, t.PipeFace)
				}
			}
		case "MOONSHOT ORBITAL":
			// antenna arrays
			x := 4 + i*2
			c.VLine(x, 4, TS-4, t.SurfLight)
			c.HLine(x-3, 5, 7, t.Accent)
			c.HLine(x-2, 8, 5, t.Accent)
		default:
			// fortress: chains and embers
			x := 4 + i*3
			for y := 0; y < TS; y += 3 {
				c.FillRect(x, y, 2, 2, t.SurfLight)
			}
		}
		ts.Deco[i] = c.Img()
	}

	// --- ladder / vine ---------------------------------------------------
	{
		c := NewCanvas(TS, TS)
		col := Mix(t.Accent, ColCream, 0.3)
		c.VLine(4, 0, TS, col)
		c.VLine(11, 0, TS, col)
		for y := 2; y < TS; y += 5 {
			c.HLine(4, y, 8, col)
		}
		ts.Ladder = c.Img()
	}

	// --- rescue beacon post ---------------------------------------------
	{
		c := NewCanvas(TS, TS)
		c.FillRect(6, 0, 4, TS, Scale(t.BlockFace, 0.9))
		c.VLine(6, 0, TS, Scale(t.BlockFace, 1.3))
		c.VLine(9, 0, TS, Scale(t.BlockFace, 0.6))
		ts.Flag = c.Img()
	}

	// --- cloud / floating platform block ---------------------------------
	{
		c := NewCanvas(TS, TS)
		white := ColCream
		if t.Underground {
			white = Mix(ColCream, t.Accent, 0.4)
		}
		c.FillEllipse(5, 9, 5, 4, white)
		c.FillEllipse(11, 9, 5, 4, white)
		c.FillEllipse(8, 6, 5, 4, white)
		c.FillRect(0, 9, TS, 5, white)
		c.Shade(0, 11, TS, 4, 0.85)
		ts.Cloud = c.Img()
	}

	return ts
}
