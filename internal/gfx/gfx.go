package gfx

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// Every texture in the game is generated here at startup. Nothing is loaded
// from disk, which keeps the build a single self-contained binary on every
// platform.

var (
	// Dario: [power state][frame]
	Dario [3][]*ebiten.Image

	ImgDistill  [3]*ebiten.Image
	ImgCrawler  [2]*ebiten.Image
	ImgShell    [2]*ebiten.Image
	ImgSpikeBot [2]*ebiten.Image
	ImgFlyer    [2]*ebiten.Image
	ImgLeech    [2]*ebiten.Image
	ImgDataPipe [2]*ebiten.Image

	ImgCompute  *ebiten.Image
	ImgScroll   *ebiten.Image
	ImgSnapshot *ebiten.Image
	ImgHalo     [2]*ebiten.Image
	ImgBolt     [2]*ebiten.Image
	ImgCage     *ebiten.Image
	ImgToken    [6]*ebiten.Image
	ImgTokenHUD *ebiten.Image
	ImgModel    [4][2]*ebiten.Image
	ImgLogo     *ebiten.Image
	ImgHeart    *ebiten.Image
	ImgLock     *ebiten.Image
	ImgStarIcon *ebiten.Image

	Tilesets    []*Tileset
	Backgrounds []*Background
	Bosses      [BossCount]*BossArt
)

var inited bool

// Init builds the whole texture set. Safe to call once, from the game loop's
// first frame (Ebitengine needs a graphics context before images can exist).
func Init() {
	if inited {
		return
	}
	inited = true

	initWhite()
	buildFont()

	// --- Dario ----------------------------------------------------------
	for _, rows := range darioSmall {
		Dario[0] = append(Dario[0], Art(rows, nil))
	}
	for _, rows := range darioBig {
		Dario[1] = append(Dario[1], Art(rows, nil))
	}
	Dario[1] = append(Dario[1], Art(darioDuck, nil))
	for _, rows := range darioBig {
		Dario[2] = append(Dario[2], Art(rows, alignedSwap))
	}
	Dario[2] = append(Dario[2], Art(darioDuck, alignedSwap))
	// small Dario cannot duck; point the slot at his stand frame so callers
	// never have to special-case the index
	Dario[0] = append(Dario[0], Dario[0][FrStand])

	// --- enemies ---------------------------------------------------------
	for i, rows := range artDistillBot {
		ImgDistill[i] = Art(rows, nil)
	}
	for i, rows := range artCrawler {
		ImgCrawler[i] = Art(rows, nil)
	}
	for i, rows := range artShell {
		ImgShell[i] = Art(rows, nil)
	}
	for i, rows := range artSpikeBot {
		ImgSpikeBot[i] = Art(rows, nil)
	}
	for i, rows := range artFlyer {
		ImgFlyer[i] = Art(rows, nil)
	}
	for i, rows := range artLeech {
		ImgLeech[i] = Art(rows, nil)
	}
	for i, rows := range artDataPipe {
		ImgDataPipe[i] = Art(rows, nil)
	}

	// --- items -----------------------------------------------------------
	ImgCompute = Art(artCompute, nil)
	ImgScroll = Art(artScroll, nil)
	ImgSnapshot = Art(artSnapshot, nil)
	for i, rows := range artHalo {
		ImgHalo[i] = Art(rows, nil)
	}
	for i, rows := range artBolt {
		ImgBolt[i] = ArtRaw(rows, nil)
	}
	ImgCage = ArtRaw(artCage, nil)
	buildTokens()
	buildIcons()

	for i := 0; i < 4; i++ {
		for f := 0; f < 2; f++ {
			ImgModel[i][f] = ModelSprite(i, f)
		}
	}

	// --- worlds ----------------------------------------------------------
	for i, t := range Themes {
		Tilesets = append(Tilesets, BuildTileset(t, int64(i)*1013+17))
		Backgrounds = append(Backgrounds, BuildBackground(t, i))
	}
	for i := 0; i < BossCount; i++ {
		Bosses[i] = BuildBoss(i)
	}

	buildLogo()
}

// Tokens are the game's coins: hexagonal "context tokens" that spin.
func buildTokens() {
	for f := 0; f < 6; f++ {
		c := NewCanvas(16, 16)
		// six frames of a spin: full -> edge -> full (mirrored)
		phase := float64(f) / 6 * math.Pi
		// Never go fully edge-on: a 1px-wide coin reads as a stick rather than
		// a spinning disc.
		rx := int(math.Round(2 + math.Abs(math.Cos(phase))*3))
		gold := ColGold
		lightSide := ColGoldLight
		if math.Cos(phase) < 0 {
			lightSide = Scale(ColGold, 0.75)
		}
		c.FillEllipse(8, 8, rx, 7, Scale(gold, 0.6))
		if rx > 1 {
			c.FillEllipse(8, 8, rx-1, 6, gold)
		}
		c.FillEllipse(8-rx/3, 7, maxi(1, rx/2), 4, lightSide)
		if rx >= 4 {
			// the token glyph: a small coral burst
			for k := 0; k < 6; k++ {
				a := float64(k)/6*6.283 + 0.3
				c.Px(8+int(2.6*math.Cos(a)), 8+int(3.0*math.Sin(a)), ColCoralDark)
			}
			c.Px(8, 8, ColCoralDark)
		}
		c.Outline(Alpha(ColInkDeep, 0.9))
		ImgToken[f] = c.Img()
	}
	// flattened HUD version
	h := NewCanvas(8, 8)
	h.FillEllipse(4, 4, 3, 3, ColGold)
	h.FillEllipse(3, 3, 1, 1, ColGoldLight)
	h.Circle(4, 4, 3, ColCoralDark)
	ImgTokenHUD = h.Img()
}

func buildIcons() {
	// life icon: a tiny Dario head
	c := NewCanvas(10, 10)
	c.FillEllipse(5, 5, 4, 4, Pal['b'])
	c.FillEllipse(5, 6, 3, 3, Pal['d'])
	c.Px(4, 5, ColInkDeep)
	c.Px(7, 5, ColInkDeep)
	c.HLine(4, 8, 4, Pal['b'])
	c.Outline(ColInkDeep)
	ImgHeart = c.Img()

	// padlock for locked world-map nodes
	l := NewCanvas(10, 12)
	l.Rect(3, 2, 4, 5, Pal['u'])
	l.Rect(2, 1, 6, 6, Pal['u'])
	l.FillRect(1, 5, 8, 6, Pal['s'])
	l.HLine(1, 5, 8, Pal['v'])
	l.Px(5, 7, ColInkDeep)
	l.Px(5, 8, ColInkDeep)
	l.Outline(ColInkDeep)
	ImgLock = l.Img()

	// star used for level-complete rank
	s := NewCanvas(11, 11)
	for k := 0; k < 5; k++ {
		a := -1.5708 + float64(k)*1.2566
		x1 := 5 + int(4.6*math.Cos(a))
		y1 := 5 + int(4.6*math.Sin(a))
		s.Line(5, 5, x1, y1, ColGold)
		b := a + 0.6283
		x2 := 5 + int(2.2*math.Cos(b))
		y2 := 5 + int(2.2*math.Sin(b))
		s.Line(5, 5, x2, y2, ColGold)
		s.Line(x1, y1, x2, y2, ColGold)
	}
	s.FillCircle(5, 5, 2, ColGoldLight)
	s.Outline(ColInkDeep)
	ImgStarIcon = s.Img()
}

// The title emblem: a coral sunburst, echoing Anthropic's mark.
func buildLogo() {
	const S = 64
	c := NewCanvas(S, S)
	rng := rand.New(rand.NewSource(99))
	cx, cy := S/2, S/2
	for k := 0; k < 12; k++ {
		a := float64(k) / 12 * 6.283
		for r := 6.0; r < 30; r += 0.4 {
			w := 1 + int((30-r)/10)
			x := cx + int(r*math.Cos(a))
			y := cy + int(r*math.Sin(a))
			col := Mix(ColCoral, ColCoralDark, r/30)
			c.FillRect(x-w/2, y-w/2, w, w, col)
		}
	}
	c.FillCircle(cx, cy, 9, ColCoral)
	c.FillCircle(cx, cy, 7, ColCoralDark)
	c.FillCircle(cx-2, cy-2, 3, Mix(ColCoral, ColCream, 0.5))
	c.Speckle(0, 0, S, S, rng, 0.08, 0.3)
	c.Outline(ColInkDeep)
	ImgLogo = c.Img()
}

// TintFor returns the flashing rainbow tint used while the Safety Halo is up.
func TintFor(tick int) color.RGBA {
	h := float64(tick%30) / 30
	switch {
	case h < 0.33:
		return Mix(ColGoldLight, ColCoral, h*3)
	case h < 0.66:
		return Mix(ColCoral, ColCyan, (h-0.33)*3)
	default:
		return Mix(ColCyan, ColGoldLight, (h-0.66)*3)
	}
}
