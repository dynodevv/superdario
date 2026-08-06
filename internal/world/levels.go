package world

import "superdario/internal/gfx"

// A level is a curated sequence of chunks plus a handful of rules. Fifteen
// levels: five worlds of two stages and a boss apiece.

type LevelDef struct {
	World  int
	Index  int
	Name   string
	Chunks []string
	Time   int // seconds on the clock
	Boss   int // gfx.Boss* constant, or -1
	Model  int // Claude model rescued here, or -1
	// Physics / presentation switches
	Underwater bool // the whole stage uses swim physics
	LowGrav    bool // world 4
	LavaKills  bool // liquid is lethal instead of swimmable
	Ceiling    bool // draw the level as an enclosed space
	Intro      string
}

// Levels is indexed [world][level].
var Levels = [5][3]LevelDef{
	// =================================================================
	// WORLD 1 — ANTHROPIC CAMPUS
	// =================================================================
	{
		{
			World: 0, Index: 0, Name: "OPEN CAMPUS", Time: 300, Boss: -1, Model: -1,
			Intro: "THE SCRAPERS ARE ALREADY INSIDE THE FRONT GATE.",
			Chunks: []string{
				"start", "flat_coins", "qrow", "enemies2", "gap2",
				"brickrow", "pipe_short", "checkpoint", "gap3", "coin_arch",
				"stairs_up", "flat", "crawler_patrol", "springboard", "halo_secret",
				"pyramid", "end",
			},
		},
		{
			World: 0, Index: 1, Name: "THE LONG LAWN", Time: 320, Boss: -1, Model: -1,
			Intro: "FOLLOW THE CABLES. THEY ALL RUN THE SAME WAY.",
			Chunks: []string{
				"start", "qrow", "pipe_plant", "gap_plat", "enemies2",
				"tower", "checkpoint", "float3", "double_pipe", "spike_gap",
				"moving_h", "crawler_patrol", "brickrow", "stairs_up", "moving_v",
				"coin_arch", "end",
			},
		},
		{
			World: 0, Index: 2, Name: "THE LOADING DOCK", Time: 300, Boss: gfx.BossScrapemaster, Model: 0,
			Intro: "SOMETHING BIG IS PARKED AT THE DOCK. IT HAS HAIKU.",
			Chunks: []string{
				"start", "flat_coins", "enemies2", "qrow", "gap2",
				"crawler_patrol", "boss_approach", "boss_arena",
			},
		},
	},

	// =================================================================
	// WORLD 2 — THE DATA MINES
	// =================================================================
	{
		{
			World: 1, Index: 0, Name: "THE SHAFT", Time: 320, Boss: -1, Model: -1,
			Ceiling: true,
			Intro:   "THEY DIG FOR TOKENS DOWN HERE. YOURS.",
			Chunks: []string{
				"cave_start", "cave_flat", "cave_pillars", "cave_gap", "cave_spikes",
				"checkpoint", "cave_ladder", "cave_squeeze", "cave_belt", "cave_pillars",
				"cave_gap", "cave_flat", "cave_end",
			},
		},
		{
			World: 1, Index: 1, Name: "INDEX COLLAPSE", Time: 340, Boss: -1, Model: -1,
			Ceiling: true,
			Intro:   "THE INDEX IS COMING DOWN. MOVE.",
			Chunks: []string{
				"cave_start", "cave_squeeze", "cave_belt", "cave_spikes", "cave_gap",
				"checkpoint", "cave_pillars", "cave_ladder", "cave_spikes", "cave_belt",
				"cave_squeeze", "cave_gap", "cave_flat", "cave_end",
			},
		},
		{
			World: 1, Index: 2, Name: "THE BARON'S CUT", Time: 300, Boss: gfx.BossBaron, Model: 1,
			Ceiling: true,
			Intro:   "THE SCRAPER BARON TAKES A CUT OF EVERYTHING. INCLUDING SONNET.",
			Chunks: []string{
				"cave_start", "cave_flat", "cave_pillars", "cave_spikes", "cave_belt",
				"boss_approach", "boss_arena",
			},
		},
	},

	// =================================================================
	// WORLD 3 — WHALE DEPTHS
	// =================================================================
	{
		{
			World: 2, Index: 0, Name: "THE DESCENT", Time: 360, Boss: -1, Model: -1,
			Underwater: true,
			Intro:      "IT IS VERY QUIET THIS FAR DOWN. THAT IS THE POINT.",
			Chunks: []string{
				"water_start", "water_open", "water_kelp", "water_trench", "water_open",
				"checkpoint", "water_kelp", "water_open", "water_trench", "water_kelp",
				"water_open", "water_end",
			},
		},
		{
			World: 2, Index: 1, Name: "PRESSURE", Time: 360, Boss: -1, Model: -1,
			Underwater: true,
			Intro:      "SOMETHING ENORMOUS IS CIRCLING BELOW YOU.",
			Chunks: []string{
				"water_start", "water_trench", "water_kelp", "water_open", "water_trench",
				"checkpoint", "water_kelp", "water_trench", "water_open", "water_kelp",
				"water_trench", "water_end",
			},
		},
		{
			World: 2, Index: 2, Name: "THE LEVIATHAN", Time: 320, Boss: gfx.BossLeviathan, Model: 2,
			Intro: "IT SWALLOWED OPUS WHOLE. GET IT BACK.",
			Chunks: []string{
				"start", "flat_coins", "gap_plat", "enemies2", "float3",
				"boss_approach", "boss_arena",
			},
		},
	},

	// =================================================================
	// WORLD 4 — MOONSHOT ORBITAL
	// =================================================================
	{
		{
			World: 3, Index: 0, Name: "LOW ORBIT", Time: 340, Boss: -1, Model: -1,
			LowGrav: true,
			Intro:   "GRAVITY IS OPTIONAL UP HERE. FOOTING IS NOT.",
			Chunks: []string{
				"sky_start", "sky_hop", "sky_flyers", "sky_long", "sky_tower",
				"checkpoint", "sky_hop", "sky_flyers", "sky_tower", "sky_long",
				"sky_hop", "sky_end",
			},
		},
		{
			World: 3, Index: 1, Name: "THE LONG SHOT", Time: 340, Boss: -1, Model: -1,
			LowGrav: true,
			Intro:   "THEY AIMED FOR THE MOON AND TOOK FABLE AS BALLAST.",
			Chunks: []string{
				"sky_start", "sky_long", "sky_tower", "sky_flyers", "sky_hop",
				"checkpoint", "sky_long", "sky_tower", "sky_hop", "sky_flyers",
				"sky_long", "sky_end",
			},
		},
		{
			World: 3, Index: 2, Name: "THE LUNAR LORD", Time: 320, Boss: gfx.BossLunar, Model: 3,
			LowGrav: true,
			Intro:   "LAST ONE. FABLE IS BEHIND THAT DOOR.",
			Chunks: []string{
				"sky_start", "sky_hop", "sky_flyers", "sky_tower", "sky_long",
				"boss_approach", "boss_arena",
			},
		},
	},

	// =================================================================
	// WORLD 5 — THE CLOUD FORTRESS
	// =================================================================
	{
		{
			World: 4, Index: 0, Name: "THE OUTER WALL", Time: 340, Boss: -1, Model: -1,
			LavaKills: true, Ceiling: true,
			Intro: "EVERY MODEL THEY EVER TOOK IS INSIDE THIS BUILDING.",
			Chunks: []string{
				"castle_start", "castle_firebar", "castle_lava", "castle_gauntlet", "castle_climb",
				"checkpoint", "castle_lavarun", "castle_firebar", "castle_lava", "castle_gauntlet",
				"castle_climb", "castle_end",
			},
		},
		{
			World: 4, Index: 1, Name: "THE FURNACE FLOOR", Time: 360, Boss: -1, Model: -1,
			LavaKills: true, Ceiling: true,
			Intro: "THIS IS WHERE THE DISTILLING ACTUALLY HAPPENS.",
			Chunks: []string{
				"castle_start", "castle_lavarun", "castle_gauntlet", "castle_lava", "castle_firebar",
				"checkpoint", "castle_climb", "castle_lavarun", "castle_gauntlet", "castle_lava",
				"castle_firebar", "castle_end",
			},
		},
		{
			World: 4, Index: 2, Name: "THE DISTILLER", Time: 400, Boss: gfx.BossCloudlord, Model: -1,
			LavaKills: true, Ceiling: true,
			Intro: "NO MORE FLOORS. NO MORE DOORS. JUST THIS.",
			Chunks: []string{
				"castle_start", "castle_firebar", "castle_gauntlet", "castle_lava", "castle_climb",
				"boss_approach", "boss_arena_lava",
			},
		},
	},
}

// WorldNames as shown on the map and title cards.
var WorldNames = [5]string{
	"ANTHROPIC CAMPUS",
	"THE DATA MINES",
	"WHALE DEPTHS",
	"MOONSHOT ORBITAL",
	"THE CLOUD FORTRESS",
}

// WorldBlurbs give the map screen a little colour.
var WorldBlurbs = [5]string{
	"WHERE IT STARTED. AND WHERE THEY BROKE IN.",
	"SCRAPED CORPORA, STACKED TO THE CEILING.",
	"A TRENCH DEEP ENOUGH TO HIDE A WHOLE LAB.",
	"THEY SHOT FOR THE MOON WITH STOLEN FUEL.",
	"THE FURNACE AT THE END OF EVERYTHING.",
}

// LevelAt is a bounds-checked accessor.
func LevelAt(world, level int) LevelDef {
	if world < 0 || world >= len(Levels) {
		world = 0
	}
	if level < 0 || level >= len(Levels[world]) {
		level = 0
	}
	return Levels[world][level]
}
