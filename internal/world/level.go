package world

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"superdario/internal/core"
	"superdario/internal/gfx"
)

const TS = core.TileSize

// Spawn records something the chunk asked us to place as an entity.
type Spawn struct {
	Kind byte
	X, Y float64 // pixel position of the tile's top-left corner
	TX   int
	TY   int
}

// Level is a built, playable stage.
type Level struct {
	Def  LevelDef
	W, H int // in tiles

	tiles  []Tile
	bump   []int8 // >0 while a block is animating after a hit
	hidden []bool // revealed hidden blocks
	Spawns []Spawn

	Theme *gfx.Tileset
	BG    *gfx.Background

	StartX, StartY float64
	GoalX, GoalY   float64
	HasGoal        bool
	BossGateX      float64
	HasBossGate    bool
}

// BuildLevel assembles a level from its chunk list.
func BuildLevel(def LevelDef) (*Level, error) {
	l := &Level{Def: def, H: core.LevelH}

	// measure
	for _, name := range def.Chunks {
		c, ok := chunks[name]
		if !ok {
			return nil, fmt.Errorf("level %d-%d references unknown chunk %q", def.World+1, def.Index+1, name)
		}
		if err := validateChunk(name, c); err != nil {
			return nil, err
		}
		l.W += c.Width()
	}
	if l.W == 0 {
		return nil, fmt.Errorf("level %d-%d has no chunks", def.World+1, def.Index+1)
	}

	l.tiles = make([]Tile, l.W*l.H)
	l.bump = make([]int8, l.W*l.H)
	l.hidden = make([]bool, l.W*l.H)

	x0 := 0
	for _, name := range def.Chunks {
		c := chunks[name]
		rows := c.Rows
		top := l.H - len(rows) // bottom-align
		// Enclosed stages (caves, the fortress) need the space above a chunk
		// filled in, or the chunk's ceiling row floats with sky above it.
		if def.Ceiling && top > 0 {
			for rx := 0; rx < c.Width(); rx++ {
				fill := TSoil
				if t, ok := runeToTile[rows[0][rx]]; ok && t.Solid() {
					fill = t
				}
				for ty := 0; ty < top; ty++ {
					l.tiles[ty*l.W+(x0+rx)] = fill
				}
			}
		}
		for ry, row := range rows {
			ty := top + ry
			if ty < 0 || ty >= l.H {
				continue
			}
			for rx := 0; rx < len(row); rx++ {
				ch := row[rx]
				tx := x0 + rx
				if isSpawn(ch) {
					l.addSpawn(ch, tx, ty)
					continue
				}
				if t, ok := runeToTile[ch]; ok {
					l.tiles[ty*l.W+tx] = t
				}
			}
		}
		x0 += c.Width()
	}

	if l.StartX == 0 && l.StartY == 0 {
		// no '@' in the chunk list: drop Dario on the first solid ground
		l.StartX, l.StartY = float64(2*TS), float64((l.H-4)*TS)
	}
	return l, nil
}

// AttachArt binds the level to its generated tileset and backdrop. It is
// separate from BuildLevel so levels can be built and validated without a
// graphics context (which is what the tests do).
func (l *Level) AttachArt() {
	if len(gfx.Tilesets) == 0 {
		return
	}
	w := l.Def.World
	if w < 0 || w >= len(gfx.Tilesets) {
		w = 0
	}
	l.Theme = gfx.Tilesets[w]
	l.BG = gfx.Backgrounds[w]
}

func validateChunk(name string, c Chunk) error {
	if len(c.Rows) == 0 {
		return fmt.Errorf("chunk %q is empty", name)
	}
	w := len(c.Rows[0])
	for i, r := range c.Rows {
		if len(r) != w {
			return fmt.Errorf("chunk %q row %d is %d wide, expected %d", name, i, len(r), w)
		}
	}
	if len(c.Rows) > core.LevelH {
		return fmt.Errorf("chunk %q is %d rows tall, max is %d", name, len(c.Rows), core.LevelH)
	}
	return nil
}

func (l *Level) addSpawn(kind byte, tx, ty int) {
	x, y := float64(tx*TS), float64(ty*TS)
	switch kind {
	case SpawnPlayer:
		l.StartX, l.StartY = x, y
		return
	case SpawnGoal:
		l.GoalX, l.GoalY = x, y
		l.HasGoal = true
		return
	case SpawnBossGate:
		l.BossGateX = x
		l.HasBossGate = true
		return
	}
	l.Spawns = append(l.Spawns, Spawn{Kind: kind, X: x, Y: y, TX: tx, TY: ty})
}

// --- tile access ------------------------------------------------------------

// At returns the tile at a tile coordinate. Off the sides of the map reads as
// solid so Dario cannot walk out of the level; above and below read as empty.
func (l *Level) At(tx, ty int) Tile {
	if tx < 0 || tx >= l.W {
		return TBlock
	}
	if ty < 0 || ty >= l.H {
		return TEmpty
	}
	t := l.tiles[ty*l.W+tx]
	if t == THidden && !l.hidden[ty*l.W+tx] {
		return TEmpty // invisible and intangible until revealed
	}
	return t
}

// RawAt ignores the hidden-block rule.
func (l *Level) RawAt(tx, ty int) Tile {
	if tx < 0 || tx >= l.W || ty < 0 || ty >= l.H {
		return TEmpty
	}
	return l.tiles[ty*l.W+tx]
}

func (l *Level) Set(tx, ty int, t Tile) {
	if tx < 0 || tx >= l.W || ty < 0 || ty >= l.H {
		return
	}
	l.tiles[ty*l.W+tx] = t
}

// SolidAt is the main collision query.
func (l *Level) SolidAt(tx, ty int) bool { return l.At(tx, ty).Solid() }

// TileAtPx converts pixel coordinates to a tile.
func (l *Level) TileAtPx(x, y float64) Tile {
	return l.At(int(math.Floor(x/TS)), int(math.Floor(y/TS)))
}

// PixelWidth / PixelHeight in pixels.
func (l *Level) PixelWidth() float64  { return float64(l.W * TS) }
func (l *Level) PixelHeight() float64 { return float64(l.H * TS) }

// Bump starts the hit animation on a block and returns what it contained.
func (l *Level) Bump(tx, ty int) {
	if tx < 0 || tx >= l.W || ty < 0 || ty >= l.H {
		return
	}
	l.bump[ty*l.W+tx] = 10
}

func (l *Level) Reveal(tx, ty int) {
	if tx < 0 || tx >= l.W || ty < 0 || ty >= l.H {
		return
	}
	l.hidden[ty*l.W+tx] = true
}

// BumpOffset is the vertical displacement of a block mid-animation.
func (l *Level) BumpOffset(tx, ty int) float64 {
	if tx < 0 || tx >= l.W || ty < 0 || ty >= l.H {
		return 0
	}
	b := l.bump[ty*l.W+tx]
	if b <= 0 {
		return 0
	}
	// up quickly, back down
	f := float64(b) / 10
	return -6 * math.Sin(f*math.Pi)
}

// Update ticks block animations.
func (l *Level) Update() {
	for i := range l.bump {
		if l.bump[i] > 0 {
			l.bump[i]--
		}
	}
}

// --- drawing ----------------------------------------------------------------

// Draw paints every visible tile.
func (l *Level) Draw(dst *ebiten.Image, camX, camY float64, tick int) {
	ts := l.Theme
	x0 := int(camX) / TS
	x1 := (int(camX)+core.ScreenW)/TS + 1
	y0 := int(camY) / TS
	y1 := (int(camY)+core.ScreenH)/TS + 1
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > l.W {
		x1 = l.W
	}
	if y1 > l.H {
		y1 = l.H
	}

	qFrame := (tick / 9) % 3
	liqFrame := (tick / 10) % 3
	beltFrame := (tick / 5) % 4

	for ty := y0; ty < y1; ty++ {
		for tx := x0; tx < x1; tx++ {
			t := l.RawAt(tx, ty)
			if t == TEmpty {
				continue
			}
			if t == THidden && !l.hidden[ty*l.W+tx] {
				continue
			}
			px := float64(tx*TS) - camX
			py := float64(ty*TS) - camY + l.BumpOffset(tx, ty)

			var img *ebiten.Image
			switch t {
			case TSurf:
				variant := 1
				if !l.At(tx-1, ty).Solid() {
					variant = 0
				} else if !l.At(tx+1, ty).Solid() {
					variant = 2
				}
				img = ts.Surf[variant]
			case TSoil:
				switch {
				case !l.At(tx-1, ty).Solid():
					img = ts.SoilEdgeL
				case !l.At(tx+1, ty).Solid():
					img = ts.SoilEdgeR
				default:
					img = ts.Soil
				}
			case TBrick, TBrickCoin:
				img = ts.Brick
			case TBlock, THidden:
				img = ts.BrickSolid
			case TQPower, TQCoin, TQ1Up, TQHalo:
				img = ts.Question[qFrame]
			case TQUsed:
				img = ts.Question[3]
			case TPipeTL:
				img = ts.Pipe[0]
			case TPipeTR:
				img = ts.Pipe[1]
			case TPipeBL:
				img = ts.Pipe[2]
			case TPipeBR:
				img = ts.Pipe[3]
			case TSpike:
				img = ts.Spike
			case TSpikeDown:
				img = ts.SpikeUp
			case TLiquid:
				img = ts.Liquid[liqFrame]
			case TLiquidTop:
				img = ts.LiquidTop[liqFrame]
			case TPlatform:
				img = ts.Platform
			case TCloud:
				img = ts.Cloud
			case TGirder:
				img = ts.Girder
			case TBelt:
				img = ts.Belt[beltFrame]
			case TLadder:
				img = ts.Ladder
			case TDeco0:
				img = ts.Deco[0]
			case TDeco1:
				img = ts.Deco[1]
			case TDeco2:
				img = ts.Deco[2]
			case TDeco3:
				img = ts.Deco[3]
			}
			if img == nil {
				continue
			}
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(math.Floor(px), math.Floor(py))
			dst.DrawImage(img, op)
		}
	}
}
