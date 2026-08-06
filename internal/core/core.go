// Package core holds tuning constants and small shared types used across the game.
package core

const (
	// Internal render resolution. Everything is drawn at this size and then
	// scaled to the window, which keeps the pixel art crisp at any size.
	ScreenW = 480
	ScreenH = 270

	TileSize = 16
	LevelH   = 17 // tiles; 17*16 = 272px, just taller than the screen

	TPS = 60
)

// Player physics. Units are pixels and ticks (1/60s).
const (
	Gravity      = 0.34
	GravityLow   = 0.22 // world 4, low gravity
	MaxFall      = 8.0
	WaterGravity = 0.10
	WaterMaxFall = 1.6

	WalkAccel = 0.22
	RunAccel  = 0.31
	MaxWalk   = 1.70
	MaxRun    = 2.95
	Friction  = 0.17
	AirDrag   = 0.06
	SkidDecel = 0.34

	JumpVel      = -6.55
	JumpRunBonus = -0.85 // extra impulse when at full speed
	JumpCut      = 0.42  // velocity multiplier when jump is released early
	SwimStroke   = -1.55
	CoyoteTicks  = 6
	BufferTicks  = 7
	StompBounce  = -4.6
	StompHold    = -6.2
)

// PowerState is Dario's current upgrade tier.
type PowerState int

const (
	PowerSmall  PowerState = iota // base
	PowerScaled                   // "Compute Cluster" — bigger, takes an extra hit
	PowerAlign                    // "Constitution Scroll" — throws alignment bolts
)

func (p PowerState) Big() bool { return p != PowerSmall }

// Facing directions.
const (
	Left  = -1
	Right = 1
)
