// Command artdump renders the generated sprite atlases to PNG files so the art
// can be reviewed outside the game. Development tool; not part of the game.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"superdario/internal/gfx"
)

var outDir = "."

type dumper struct{ done bool }

func (d *dumper) Update() error {
	if d.done {
		return ebiten.Termination
	}
	return nil
}

func (d *dumper) Layout(w, h int) (int, int) { return 320, 240 }

func save(name string, img *ebiten.Image) {
	b := img.Bounds()
	buf := make([]byte, 4*b.Dx()*b.Dy())
	img.ReadPixels(buf)
	out := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	copy(out.Pix, buf)
	f, err := os.Create(filepath.Join(outDir, name))
	if err != nil {
		fmt.Println("create:", err)
		return
	}
	defer f.Close()
	png.Encode(f, out)
	fmt.Println("wrote", name, b.Dx(), "x", b.Dy())
}

// sheet lays out a set of images on a checkerboard with labels.
func sheet(title string, cols int, cellW, cellH, scale int, items []labeled) *ebiten.Image {
	rows := (len(items) + cols - 1) / cols
	w := cols * (cellW*scale + 8)
	h := rows*(cellH*scale+18) + 20
	img := ebiten.NewImage(w, h)
	img.Fill(color.RGBA{0x18, 0x18, 0x20, 0xff})
	for y := 0; y < h; y += 8 {
		for x := 0; x < w; x += 8 {
			if (x/8+y/8)%2 == 0 {
				gfx.FillRectF(img, float64(x), float64(y), 8, 8, color.RGBA{0x22, 0x22, 0x2c, 0xff})
			}
		}
	}
	gfx.Text(img, title, 4, 4, gfx.ColGoldLight, 1)
	for i, it := range items {
		cx := (i % cols) * (cellW*scale + 8)
		cy := (i/cols)*(cellH*scale+18) + 20
		gfx.StrokeRectF(img, float64(cx), float64(cy), float64(cellW*scale+2), float64(cellH*scale+2), color.RGBA{0x44, 0x44, 0x55, 0xff})
		if it.img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(float64(scale), float64(scale))
			op.GeoM.Translate(float64(cx+1), float64(cy+1))
			img.DrawImage(it.img, op)
		}
		gfx.Text(img, it.name, cx+1, cy+cellH*scale+4, gfx.ColCream, 1)
	}
	return img
}

type labeled struct {
	name string
	img  *ebiten.Image
}

func (d *dumper) Draw(screen *ebiten.Image) {
	if d.done {
		return
	}
	d.done = true
	gfx.Init()

	names := []string{"STAND", "WALK1", "WALK2", "WALK3", "JUMP", "SKID", "DEAD", "SWIM1", "SWIM2", "CHEER", "DUCK"}

	var small, big, aligned []labeled
	for i, im := range gfx.Dario[0] {
		small = append(small, labeled{names[i], im})
	}
	for i, im := range gfx.Dario[1] {
		big = append(big, labeled{names[i], im})
	}
	for i, im := range gfx.Dario[2] {
		aligned = append(aligned, labeled{names[i], im})
	}
	save("art_dario_small.png", sheet("DARIO - SMALL", 10, 18, 18, 5, small))
	save("art_dario_big.png", sheet("DARIO - COMPUTE CLUSTER", 10, 18, 30, 4, big))
	save("art_dario_aligned.png", sheet("DARIO - CONSTITUTION", 10, 18, 30, 4, aligned))

	var en []labeled
	add := func(n string, im *ebiten.Image) { en = append(en, labeled{n, im}) }
	add("DBOT0", gfx.ImgDistill[0])
	add("DBOT1", gfx.ImgDistill[1])
	add("DBOTSQ", gfx.ImgDistill[2])
	add("CRAWL0", gfx.ImgCrawler[0])
	add("CRAWL1", gfx.ImgCrawler[1])
	add("SHELL0", gfx.ImgShell[0])
	add("SHELL1", gfx.ImgShell[1])
	add("SPIKE0", gfx.ImgSpikeBot[0])
	add("SPIKE1", gfx.ImgSpikeBot[1])
	add("FLY0", gfx.ImgFlyer[0])
	add("FLY1", gfx.ImgFlyer[1])
	add("LEECH0", gfx.ImgLeech[0])
	add("LEECH1", gfx.ImgLeech[1])
	add("COMPUTE", gfx.ImgCompute)
	add("SCROLL", gfx.ImgScroll)
	add("SNAPSHOT", gfx.ImgSnapshot)
	add("HALO0", gfx.ImgHalo[0])
	add("HALO1", gfx.ImgHalo[1])
	add("BOLT0", gfx.ImgBolt[0])
	add("BOLT1", gfx.ImgBolt[1])
	add("TOK0", gfx.ImgToken[0])
	add("TOK1", gfx.ImgToken[1])
	add("TOK2", gfx.ImgToken[2])
	add("TOK3", gfx.ImgToken[3])
	add("HAIKU", gfx.ImgModel[0][0])
	add("SONNET", gfx.ImgModel[1][0])
	add("FABLE", gfx.ImgModel[2][0])
	add("OPUS", gfx.ImgModel[3][0])
	add("LIFE", gfx.ImgHeart)
	add("LOCK", gfx.ImgLock)
	add("STAR", gfx.ImgStarIcon)
	add("CAGE", gfx.ImgCage)
	save("art_entities.png", sheet("ENEMIES / ITEMS", 11, 18, 26, 4, en))

	var pipes []labeled
	pipes = append(pipes, labeled{"PIPE0", gfx.ImgDataPipe[0]}, labeled{"PIPE1", gfx.ImgDataPipe[1]}, labeled{"LOGO", gfx.ImgLogo})
	save("art_misc.png", sheet("MISC", 3, 64, 64, 3, pipes))

	for i, ts := range gfx.Tilesets {
		var tl []labeled
		tl = append(tl,
			labeled{"SURFL", ts.Surf[0]}, labeled{"SURFM", ts.Surf[1]}, labeled{"SURFR", ts.Surf[2]},
			labeled{"SOIL", ts.Soil}, labeled{"BRICK", ts.Brick}, labeled{"BLOCK", ts.BrickSolid},
			labeled{"Q0", ts.Question[0]}, labeled{"Q3", ts.Question[3]},
			labeled{"PTL", ts.Pipe[0]}, labeled{"PTR", ts.Pipe[1]}, labeled{"PBL", ts.Pipe[2]}, labeled{"PBR", ts.Pipe[3]},
			labeled{"SPIKE", ts.Spike}, labeled{"LIQ", ts.Liquid[0]}, labeled{"LIQTOP", ts.LiquidTop[0]},
			labeled{"PLAT", ts.Platform}, labeled{"GIRDER", ts.Girder}, labeled{"BELT", ts.Belt[0]},
			labeled{"DECO0", ts.Deco[0]}, labeled{"DECO1", ts.Deco[1]}, labeled{"LADDER", ts.Ladder}, labeled{"CLOUD", ts.Cloud},
		)
		save(fmt.Sprintf("art_tiles_w%d.png", i+1), sheet(ts.Theme.Name, 11, 16, 16, 5, tl))
	}

	var bs []labeled
	for i := 0; i < gfx.BossCount; i++ {
		bs = append(bs, labeled{gfx.BossNames[i], gfx.Bosses[i].Frames[0]})
	}
	save("art_bosses.png", sheet("SYNDICATE", 6, 64, 64, 3, bs))

	// backgrounds: render each world's full backdrop
	for i, bg := range gfx.Backgrounds {
		scene := ebiten.NewImage(480, 270)
		for k := 0; k < 40; k++ {
			bg.Update()
		}
		bg.Draw(scene, 120, 0)
		gfx.TextShadow(scene, gfx.Themes[i].Name, 6, 6, gfx.ColCream, gfx.ColInkDeep, 1)
		save(fmt.Sprintf("art_bg_w%d.png", i+1), scene)
	}
}

func main() {
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}
	ebiten.SetWindowSize(320, 240)
	ebiten.SetWindowTitle("artdump")
	if err := ebiten.RunGame(&dumper{}); err != nil && err != ebiten.Termination {
		fmt.Println(err)
	}
}
