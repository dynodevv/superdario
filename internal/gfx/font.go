package gfx

import (
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// A hand-drawn 5x7 bitmap font. Uppercase only, which is both authentic to the
// era and half the art. Lowercase input is folded to uppercase at draw time.

const (
	GlyphW  = 5
	GlyphH  = 7
	Advance = 6 // glyph width + 1px letter spacing
	LineH   = 9
)

var glyphArt = map[rune][]string{
	' ':  {".....", ".....", ".....", ".....", ".....", ".....", "....."},
	'A':  {".###.", "#...#", "#...#", "#####", "#...#", "#...#", "#...#"},
	'B':  {"####.", "#...#", "#...#", "####.", "#...#", "#...#", "####."},
	'C':  {".###.", "#...#", "#....", "#....", "#....", "#...#", ".###."},
	'D':  {"####.", "#...#", "#...#", "#...#", "#...#", "#...#", "####."},
	'E':  {"#####", "#....", "#....", "####.", "#....", "#....", "#####"},
	'F':  {"#####", "#....", "#....", "####.", "#....", "#....", "#...."},
	'G':  {".###.", "#...#", "#....", "#.###", "#...#", "#...#", ".###."},
	'H':  {"#...#", "#...#", "#...#", "#####", "#...#", "#...#", "#...#"},
	'I':  {"#####", "..#..", "..#..", "..#..", "..#..", "..#..", "#####"},
	'J':  {"..###", "...#.", "...#.", "...#.", "...#.", "#..#.", ".##.."},
	'K':  {"#...#", "#..#.", "#.#..", "##...", "#.#..", "#..#.", "#...#"},
	'L':  {"#....", "#....", "#....", "#....", "#....", "#....", "#####"},
	'M':  {"#...#", "##.##", "#.#.#", "#.#.#", "#...#", "#...#", "#...#"},
	'N':  {"#...#", "##..#", "#.#.#", "#.#.#", "#..##", "#...#", "#...#"},
	'O':  {".###.", "#...#", "#...#", "#...#", "#...#", "#...#", ".###."},
	'P':  {"####.", "#...#", "#...#", "####.", "#....", "#....", "#...."},
	'Q':  {".###.", "#...#", "#...#", "#...#", "#.#.#", "#..#.", ".##.#"},
	'R':  {"####.", "#...#", "#...#", "####.", "#.#..", "#..#.", "#...#"},
	'S':  {".####", "#....", "#....", ".###.", "....#", "....#", "####."},
	'T':  {"#####", "..#..", "..#..", "..#..", "..#..", "..#..", "..#.."},
	'U':  {"#...#", "#...#", "#...#", "#...#", "#...#", "#...#", ".###."},
	'V':  {"#...#", "#...#", "#...#", "#...#", "#...#", ".#.#.", "..#.."},
	'W':  {"#...#", "#...#", "#...#", "#.#.#", "#.#.#", "##.##", "#...#"},
	'X':  {"#...#", "#...#", ".#.#.", "..#..", ".#.#.", "#...#", "#...#"},
	'Y':  {"#...#", "#...#", ".#.#.", "..#..", "..#..", "..#..", "..#.."},
	'Z':  {"#####", "....#", "...#.", "..#..", ".#...", "#....", "#####"},
	'0':  {".###.", "#...#", "#..##", "#.#.#", "##..#", "#...#", ".###."},
	'1':  {"..#..", ".##..", "..#..", "..#..", "..#..", "..#..", ".###."},
	'2':  {".###.", "#...#", "....#", "...#.", "..#..", ".#...", "#####"},
	'3':  {"####.", "....#", "....#", ".###.", "....#", "....#", "####."},
	'4':  {"...#.", "..##.", ".#.#.", "#..#.", "#####", "...#.", "...#."},
	'5':  {"#####", "#....", "####.", "....#", "....#", "#...#", ".###."},
	'6':  {".###.", "#...#", "#....", "####.", "#...#", "#...#", ".###."},
	'7':  {"#####", "....#", "...#.", "..#..", ".#...", ".#...", ".#..."},
	'8':  {".###.", "#...#", "#...#", ".###.", "#...#", "#...#", ".###."},
	'9':  {".###.", "#...#", "#...#", ".####", "....#", "#...#", ".###."},
	'!':  {"..#..", "..#..", "..#..", "..#..", "..#..", ".....", "..#.."},
	'?':  {".###.", "#...#", "....#", "..##.", "..#..", ".....", "..#.."},
	'.':  {".....", ".....", ".....", ".....", ".....", ".....", "..#.."},
	',':  {".....", ".....", ".....", ".....", "..#..", "..#..", ".#..."},
	':':  {".....", "..#..", "..#..", ".....", "..#..", "..#..", "....."},
	';':  {".....", "..#..", "..#..", ".....", "..#..", "..#..", ".#..."},
	'-':  {".....", ".....", ".....", "#####", ".....", ".....", "....."},
	'+':  {".....", "..#..", "..#..", "#####", "..#..", "..#..", "....."},
	'=':  {".....", ".....", "#####", ".....", "#####", ".....", "....."},
	'*':  {".....", "#.#.#", ".###.", "#####", ".###.", "#.#.#", "....."},
	'/':  {"....#", "....#", "...#.", "..#..", ".#...", "#....", "#...."},
	'\\': {"#....", "#....", ".#...", "..#..", "...#.", "....#", "....#"},
	'\'': {"..#..", "..#..", ".....", ".....", ".....", ".....", "....."},
	'"':  {".#.#.", ".#.#.", ".....", ".....", ".....", ".....", "....."},
	'(':  {"...#.", "..#..", ".#...", ".#...", ".#...", "..#..", "...#."},
	')':  {".#...", "..#..", "...#.", "...#.", "...#.", "..#..", ".#..."},
	'[':  {"..###", "..#..", "..#..", "..#..", "..#..", "..#..", "..###"},
	']':  {"###..", "..#..", "..#..", "..#..", "..#..", "..#..", "###.."},
	'<':  {"...#.", "..#..", ".#...", "#....", ".#...", "..#..", "...#."},
	'>':  {".#...", "..#..", "...#.", "....#", "...#.", "..#..", ".#..."},
	'_':  {".....", ".....", ".....", ".....", ".....", ".....", "#####"},
	'@':  {".###.", "#...#", "#.###", "#.#.#", "#.###", "#....", ".###."},
	'#':  {".#.#.", ".#.#.", "#####", ".#.#.", "#####", ".#.#.", ".#.#."},
	'%':  {"##..#", "##.#.", "...#.", "..#..", ".#...", ".#.##", "#..##"},
	'&':  {".##..", "#..#.", "#.#..", ".#...", "#.#.#", "#..#.", ".##.#"},
	'$':  {"..#..", ".####", "#.#..", ".###.", "..#.#", "####.", "..#.."},
	'^':  {"..#..", ".#.#.", "#...#", ".....", ".....", ".....", "....."},
	'~':  {".....", ".....", ".##.#", "#..#.", ".....", ".....", "....."},
	'|':  {"..#..", "..#..", "..#..", "..#..", "..#..", "..#..", "..#.."},
}

var (
	fontAtlas *ebiten.Image
	glyphRect = map[rune]image.Rectangle{}
	fontOrder []rune
)

func buildFont() {
	fontOrder = fontOrder[:0]
	for r := range glyphArt {
		fontOrder = append(fontOrder, r)
	}
	// stable order so the atlas is deterministic
	for i := 1; i < len(fontOrder); i++ {
		for j := i; j > 0 && fontOrder[j] < fontOrder[j-1]; j-- {
			fontOrder[j], fontOrder[j-1] = fontOrder[j-1], fontOrder[j]
		}
	}
	c := NewCanvas(len(fontOrder)*Advance, GlyphH)
	white := color.RGBA{255, 255, 255, 255}
	for i, r := range fontOrder {
		art := glyphArt[r]
		ox := i * Advance
		for y, row := range art {
			for x := 0; x < len(row); x++ {
				if row[x] == '#' {
					c.Px(ox+x, y, white)
				}
			}
		}
		glyphRect[r] = image.Rect(ox, 0, ox+GlyphW, GlyphH)
	}
	fontAtlas = c.Img()
}

// TextWidth returns the pixel width of s at the given integer scale.
func TextWidth(s string, scale int) int {
	if s == "" {
		return 0
	}
	return (len(s)*Advance - 1) * scale
}

// Text draws s at (x, y) — y is the glyph top — tinted with col.
func Text(dst *ebiten.Image, s string, x, y int, col color.RGBA, scale int) {
	if scale < 1 {
		scale = 1
	}
	s = strings.ToUpper(s)
	fs := float32(col.A) / 255
	op := &ebiten.DrawImageOptions{}
	cx := float64(x)
	for _, r := range s {
		rect, ok := glyphRect[r]
		if !ok {
			rect = glyphRect['?']
		}
		if r != ' ' {
			op.GeoM.Reset()
			op.GeoM.Scale(float64(scale), float64(scale))
			op.GeoM.Translate(cx, float64(y))
			op.ColorScale.Reset()
			op.ColorScale.ScaleWithColor(color.RGBA{col.R, col.G, col.B, 255})
			op.ColorScale.ScaleAlpha(fs)
			dst.DrawImage(fontAtlas.SubImage(rect).(*ebiten.Image), op)
		}
		cx += float64(Advance * scale)
	}
}

// TextShadow draws text with a 1px (scaled) drop shadow for readability over
// busy backgrounds — used by the HUD and all menus.
func TextShadow(dst *ebiten.Image, s string, x, y int, col, shadow color.RGBA, scale int) {
	Text(dst, s, x+scale, y+scale, shadow, scale)
	Text(dst, s, x, y, col, scale)
}

// TextOutline draws text with a full 8-direction outline. Heavier but reads on
// anything; reserved for titles and cutscene captions.
func TextOutline(dst *ebiten.Image, s string, x, y int, col, outline color.RGBA, scale int) {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			Text(dst, s, x+dx*scale, y+dy*scale, outline, scale)
		}
	}
	Text(dst, s, x, y, col, scale)
}

// TextCenter centres s horizontally about cx.
func TextCenter(dst *ebiten.Image, s string, cx, y int, col color.RGBA, scale int) {
	Text(dst, s, cx-TextWidth(s, scale)/2, y, col, scale)
}

func TextCenterShadow(dst *ebiten.Image, s string, cx, y int, col, shadow color.RGBA, scale int) {
	TextShadow(dst, s, cx-TextWidth(s, scale)/2, y, col, shadow, scale)
}

func TextCenterOutline(dst *ebiten.Image, s string, cx, y int, col, outline color.RGBA, scale int) {
	TextOutline(dst, s, cx-TextWidth(s, scale)/2, y, col, outline, scale)
}

// WrapText breaks s into lines that fit within width pixels at the given scale.
func WrapText(s string, width, scale int) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	cur := words[0]
	for _, w := range words[1:] {
		if TextWidth(cur+" "+w, scale) <= width {
			cur += " " + w
		} else {
			lines = append(lines, cur)
			cur = w
		}
	}
	return append(lines, cur)
}
