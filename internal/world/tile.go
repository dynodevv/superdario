package world

// Tile kinds. The map is a flat grid of these plus a parallel state array
// used for bump animations and spent question blocks.
type Tile uint8

const (
	TEmpty     Tile = iota
	TSurf           // ground top (grass / panel / crust)
	TSoil           // ground interior
	TBrick          // breakable when Dario is big; bumpable when small
	TBrickCoin      // brick that pays out coins when bumped
	TBlock          // indestructible hard block
	TQPower         // ? block: Compute Cluster, or Constitution Scroll if already big
	TQCoin          // ? block: one token
	TQ1Up           // ? block: extra life
	TQHalo          // ? block: Safety Halo
	TQUsed          // spent ? block
	THidden         // invisible until bumped from below, then becomes a block
	TPipeTL
	TPipeTR
	TPipeBL
	TPipeBR
	TSpike     // hazard, points up
	TSpikeDown // hazard, points down (ceiling)
	TLiquid    // water or lava, depending on the world theme
	TLiquidTop
	TPlatform // one-way: solid only from above
	TCloud    // one-way, decorative cloud
	TGirder   // solid structural beam
	TBelt     // solid, pushes Dario along
	TLadder   // climbable, never solid
	TDeco0
	TDeco1
	TDeco2
	TDeco3
	TileCount
)

// Solid reports whether a tile blocks movement from every direction.
func (t Tile) Solid() bool {
	switch t {
	case TSurf, TSoil, TBrick, TBrickCoin, TBlock,
		TQPower, TQCoin, TQ1Up, TQHalo, TQUsed,
		TPipeTL, TPipeTR, TPipeBL, TPipeBR,
		TGirder, TBelt:
		return true
	}
	return false
}

// OneWay tiles are solid only when falling onto them from above.
func (t Tile) OneWay() bool { return t == TPlatform || t == TCloud }

// Bumpable tiles react to being hit from below.
func (t Tile) Bumpable() bool {
	switch t {
	case TBrick, TBrickCoin, TQPower, TQCoin, TQ1Up, TQHalo, THidden:
		return true
	}
	return false
}

// Question reports the ? blocks (spent ones excluded).
func (t Tile) Question() bool {
	switch t {
	case TQPower, TQCoin, TQ1Up, TQHalo:
		return true
	}
	return false
}

// Hazard tiles hurt on contact.
func (t Tile) Hazard() bool { return t == TSpike || t == TSpikeDown }

// Liquid tiles change the physics (swimming) or kill (lava).
func (t Tile) Liquid() bool { return t == TLiquid || t == TLiquidTop }

// Climbable tiles let Dario go up and down freely.
func (t Tile) Climbable() bool { return t == TLadder }

// Decorative tiles are drawn but never interact.
func (t Tile) Decor() bool { return t >= TDeco0 && t <= TDeco3 }

// SeeThrough reports tiles whose art does not fill its whole cell — spikes
// have gaps between the points, a platform is a thin plank, kelp is a few
// strands. Underwater those gaps must have water painted behind them or they
// show the raw backdrop and the seabed looks like it is sitting in a void.
func (t Tile) SeeThrough() bool {
	switch t {
	case TSpike, TSpikeDown, TPlatform, TCloud, TLadder,
		TDeco0, TDeco1, TDeco2, TDeco3:
		return true
	}
	return false
}

// runeToTile is the level-authoring alphabet. Chunks are written as ASCII art
// using these characters; see chunks.go.
var runeToTile = map[byte]Tile{
	' ': TEmpty,
	'.': TEmpty,
	'#': TSurf,
	'X': TSoil,
	'=': TBrick,
	'B': TBrickCoin,
	'?': TQPower,
	'c': TQCoin,
	'1': TQ1Up,
	'*': TQHalo,
	'H': THidden,
	'b': TBlock,
	'[': TPipeTL,
	']': TPipeTR,
	'{': TPipeBL,
	'}': TPipeBR,
	'^': TSpike,
	'"': TSpikeDown,
	'~': TLiquidTop,
	'_': TLiquid,
	'-': TPlatform,
	'C': TCloud,
	'G': TGirder,
	'>': TBelt,
	'|': TLadder,
	'v': TDeco0,
	'w': TDeco1,
	'x': TDeco2,
	'y': TDeco3,
}

// Characters that place an entity rather than a tile. The tile underneath
// becomes empty.
const (
	SpawnPlayer     = '@'
	SpawnToken      = 'o'
	SpawnDistill    = 'g'
	SpawnCrawler    = 'k'
	SpawnSpike      = 's'
	SpawnFlyer      = 'f'
	SpawnLeech      = 'l'
	SpawnPlant      = 'p'
	SpawnSpring     = 'S'
	SpawnPlatH      = 'A'
	SpawnPlatV      = 'V'
	SpawnCheckpoint = '!'
	SpawnGoal       = 'F'
	SpawnCompute    = 'M'
	SpawnFireBar    = '&'
	SpawnFalling    = '%'
	SpawnBossGate   = 'Z'
)

func isSpawn(c byte) bool {
	switch c {
	case SpawnPlayer, SpawnToken, SpawnDistill, SpawnCrawler, SpawnSpike,
		SpawnFlyer, SpawnLeech, SpawnPlant, SpawnSpring, SpawnPlatH,
		SpawnPlatV, SpawnCheckpoint, SpawnGoal, SpawnCompute, SpawnFireBar,
		SpawnFalling, SpawnBossGate:
		return true
	}
	return false
}
