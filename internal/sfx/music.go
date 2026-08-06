package sfx

// The soundtrack. Every bar is sixteen sixteenth-notes; see sequencer.go for
// the token syntax.

type TrackID int

const (
	MusNone TrackID = iota
	MusTitle
	MusStory
	MusMap
	MusWorld1
	MusWorld2
	MusWorld3
	MusWorld4
	MusWorld5
	MusBoss
	MusHalo
	MusClear
	MusGameOver
	MusEnding
	musCount
)

// envelope presets
var (
	envLead  = Env{A: 0.006, D: 0.09, S: 0.55, R: 0.10}
	envStab  = Env{A: 0.002, D: 0.07, S: 0.00, R: 0.05}
	envPad   = Env{A: 0.09, D: 0.30, S: 0.55, R: 0.35}
	envBass  = Env{A: 0.003, D: 0.10, S: 0.70, R: 0.06}
	envPluck = Env{A: 0.002, D: 0.11, S: 0.00, R: 0.06}
)

// MusicFor maps a world index (0-based) to its level theme.
func MusicFor(world int) TrackID {
	switch world {
	case 0:
		return MusWorld1
	case 1:
		return MusWorld2
	case 2:
		return MusWorld3
	case 3:
		return MusWorld4
	default:
		return MusWorld5
	}
}

var tracks = map[TrackID]Track{

	// -----------------------------------------------------------------
	// Title — "Super Dario". Bright C major, a little heroic.
	// -----------------------------------------------------------------
	MusTitle: {
		Name: "SUPER DARIO", BPM: 128, Delay: 0.22, DelayT: 0.28, Cutoff: 8200, DrumVol: 0.75,
		Chans: []Chan{
			{Wave: WavePulse25, Vol: 0.30, Env: envLead, Pan: -0.15, Vib: 0.004, Bars: []string{
				"e5 - - . g5 - - . c6 - - - b5 - - -",
				"a5 - - . g5 - - . e5 - - - . . . .",
				"f5 - - . a5 - - . c6 - - - a5 - - -",
				"g5 - - - - - - - . . . . . . . .",
				"e5 - - . g5 - - . c6 - - - d6 - - -",
				"e6 - - . d6 - - . c6 - - - . . . .",
				"d6 - - . b5 - - . g5 - - - a5 - - -",
				"c6 - - - - - - - - - - - . . . .",
			}},
			{Wave: WavePulse50, Vol: 0.13, Env: envPluck, Pan: 0.30, Bars: []string{
				"c4 . g4 . e5 . g4 . c4 . g4 . e5 . g4 .",
				"a3 . e4 . a4 . e4 . a3 . e4 . a4 . e4 .",
				"f3 . c4 . f4 . c4 . f3 . c4 . f4 . c4 .",
				"g3 . d4 . g4 . d4 . g3 . d4 . g4 . b4 .",
				"c4 . g4 . e5 . g4 . c4 . g4 . e5 . g4 .",
				"a3 . e4 . a4 . e4 . a3 . e4 . a4 . e4 .",
				"f3 . c4 . f4 . c4 . g3 . d4 . g4 . d4 .",
				"c4 . g4 . c5 . e5 . g5 . e5 . c5 . g4 .",
			}},
			{Wave: WaveTri, Vol: 0.34, Env: envBass, Bars: []string{
				"c2 - c2 - g2 - c3 - c2 - c2 - g2 - e2 -",
				"a1 - a1 - e2 - a2 - a1 - a1 - e2 - c2 -",
				"f1 - f1 - c2 - f2 - f1 - f1 - c2 - a1 -",
				"g1 - g1 - d2 - g2 - g1 - g1 - d2 - b1 -",
				"c2 - c2 - g2 - c3 - c2 - c2 - g2 - e2 -",
				"a1 - a1 - e2 - a2 - a1 - a1 - e2 - c2 -",
				"f1 - f1 - c2 - f2 - g1 - g1 - d2 - g2 -",
				"c2 - c2 - g2 - c3 - c2 - - - - - - -",
			}},
		},
		Drums: []string{
			"k . h . s . h . k . h k s . h .",
			"k . h . s . h . k . h k s . h .",
			"k . h . s . h . k . h k s . h .",
			"k . h . s . h . k . h k s t t t",
		},
	},

	// -----------------------------------------------------------------
	// Story — sparse and uneasy, for the cutscenes.
	// -----------------------------------------------------------------
	MusStory: {
		Name: "THE SYNDICATE", BPM: 88, Delay: 0.34, DelayT: 0.42, Cutoff: 4200, DrumVol: 0.5,
		Chans: []Chan{
			{Wave: WaveSine, Vol: 0.24, Env: envPad, Pan: -0.2, Vib: 0.006, Bars: []string{
				"a4 - - - - - - - c5 - - - - - - -",
				"b4 - - - - - - - a4 - - - - - - -",
				"f4 - - - - - - - g4 - - - - - - -",
				"e4 - - - - - - - - - - - - - - -",
			}},
			{Wave: WavePulse50, Vol: 0.10, Env: envPad, Pan: 0.25, Bars: []string{
				"a3 - - - - - - - e4 - - - - - - -",
				"e3 - - - - - - - b3 - - - - - - -",
				"f3 - - - - - - - c4 - - - - - - -",
				"e3 - - - - - - - b3 - - - - - - -",
			}},
			{Wave: WaveTri, Vol: 0.30, Env: envPad, Bars: []string{
				"a1 - - - - - - - - - - - - - - -",
				"e1 - - - - - - - - - - - - - - -",
				"f1 - - - - - - - - - - - - - - -",
				"e1 - - - - - - - - - - - - - - -",
			}},
		},
		Drums: []string{". . . . . . . . . . . . . . . ."},
	},

	// -----------------------------------------------------------------
	// World map — light and pleasant.
	// -----------------------------------------------------------------
	MusMap: {
		Name: "THE ROAD AHEAD", BPM: 118, Delay: 0.20, DelayT: 0.30, Cutoff: 7600, DrumVol: 0.55,
		Chans: []Chan{
			{Wave: WavePulse25, Vol: 0.24, Env: envLead, Pan: -0.1, Bars: []string{
				"c5 - e5 - g5 - e5 - f5 - - - e5 - - .",
				"d5 - f5 - a5 - f5 - g5 - - - . . . .",
				"e5 - g5 - c6 - g5 - a5 - - - g5 - - .",
				"f5 - e5 - d5 - c5 - . . . . . . . .",
			}},
			{Wave: WavePulse50, Vol: 0.11, Env: envPluck, Pan: 0.28, Bars: []string{
				"c4 . e4 . g4 . e4 . c4 . e4 . g4 . e4 .",
				"d4 . f4 . a4 . f4 . g3 . b3 . d4 . b3 .",
				"e4 . g4 . c5 . g4 . f4 . a4 . c5 . a4 .",
				"g3 . b3 . d4 . b3 . c4 . e4 . g4 . c5 .",
			}},
			{Wave: WaveTri, Vol: 0.30, Env: envBass, Bars: []string{
				"c2 - - - g2 - - - c2 - - - e2 - - -",
				"d2 - - - a2 - - - g1 - - - d2 - - -",
				"e2 - - - b2 - - - f2 - - - c3 - - -",
				"g1 - - - d2 - - - c2 - - - - - - -",
			}},
		},
		Drums: []string{
			". . h . s . h . . . h . s . h .",
			". . h . s . h . . . h . s . hH .",
		},
	},

	// -----------------------------------------------------------------
	// World 1 — Anthropic Campus. Bouncy F major.
	// -----------------------------------------------------------------
	MusWorld1: {
		Name: "CAMPUS RUN", BPM: 142, Delay: 0.18, DelayT: 0.21, Cutoff: 8600, DrumVol: 0.85, Swing: 0.10,
		Chans: []Chan{
			{Wave: WavePulse25, Vol: 0.29, Env: envLead, Pan: -0.12, Bars: []string{
				"f4 . a4 . c5 . a4 . f5 - - . c5 - - .",
				"d5 . c5 . a4 - - . g4 . a4 . bb4 - - .",
				"c5 . e5 . g5 . e5 . c6 - - . g5 - - .",
				"f5 - - . e5 . d5 . c5 - - - - - - .",
				"bb4 . d5 . f5 . d5 . bb5 - - . f5 - - .",
				"a5 . g5 . f5 - - . e5 . f5 . g5 - - .",
				"c5 . f5 . a5 . f5 . c6 - - . a5 - - .",
				"g5 . f5 . e5 . d5 . c5 - - - - - - -",
			}},
			{Wave: WavePulse50, Vol: 0.12, Env: envStab, Pan: 0.32, Bars: []string{
				". f4 . c5 . f4 . c5 . f4 . c5 . a4 . c5",
				". bb3 . f4 . bb3 . f4 . c4 . g4 . c4 . g4",
				". c4 . g4 . c4 . g4 . c4 . g4 . e4 . g4",
				". f4 . c5 . f4 . c5 . f4 . c5 . f4 . c5",
				". bb3 . f4 . bb3 . f4 . bb3 . f4 . d4 . f4",
				". f4 . c5 . f4 . c5 . f4 . c5 . a4 . c5",
				". f4 . c5 . f4 . c5 . f4 . c5 . a4 . c5",
				". c4 . g4 . c4 . g4 . f4 . c5 . f4 . a4",
			}},
			{Wave: WaveTri, Vol: 0.36, Env: envBass, Bars: []string{
				"f2 - f2 - c3 - f2 - a2 - f2 - c3 - c3 -",
				"bb1 - bb1 - f2 - bb2 - c2 - c2 - g2 - c3 -",
				"c2 - c2 - g2 - c3 - e2 - c2 - g2 - g2 -",
				"f2 - f2 - c3 - f2 - f2 - a2 - c3 - e2 -",
				"bb1 - bb1 - f2 - bb2 - d2 - bb1 - f2 - f2 -",
				"f2 - f2 - c3 - f2 - a2 - f2 - c3 - c3 -",
				"f2 - f2 - c3 - f2 - a2 - f2 - c3 - e2 -",
				"c2 - c2 - g2 - c3 - f2 - f2 - f2 - - -",
			}},
		},
		Drums: []string{
			"k . h . s . h k k . h . s . hH .",
			"k . h . s . h . k k h . s . h .",
		},
	},

	// -----------------------------------------------------------------
	// World 2 — The Data Mines. A minor, cold and mechanical.
	// -----------------------------------------------------------------
	MusWorld2: {
		Name: "STRIP THE INDEX", BPM: 124, Delay: 0.30, DelayT: 0.36, Cutoff: 6400, DrumVol: 0.9,
		Chans: []Chan{
			{Wave: WavePulse12, Vol: 0.26, Env: envLead, Pan: -0.2, Vib: 0.008, Bars: []string{
				"a4 - - . c5 - - . e5 - - . d5 - - .",
				"c5 - - . a4 - - . g4 - - - . . . .",
				"f4 - - . a4 - - . c5 - - . b4 - - .",
				"a4 - - - - - - . e4 - - . g4 - - .",
				"a4 - - . e5 - - . f5 - - . e5 - - .",
				"d5 - - . c5 - - . b4 - - - . . . .",
				"g4 - - . b4 - - . d5 - - . c5 - - .",
				"a4 - - - - - - - - - - - . . . .",
			}},
			{Wave: WavePulse50, Vol: 0.10, Env: envPad, Pan: 0.3, Bars: []string{
				"a3 - - - - - - - e4 - - - - - - -",
				"a3 - - - - - - - c4 - - - - - - -",
				"f3 - - - - - - - c4 - - - - - - -",
				"e3 - - - - - - - b3 - - - - - - -",
				"a3 - - - - - - - e4 - - - - - - -",
				"d3 - - - - - - - a3 - - - - - - -",
				"g3 - - - - - - - d4 - - - - - - -",
				"e3 - - - - - - - g#3 - - - - - - -",
			}},
			{Wave: WaveSaw, Vol: 0.22, Env: envBass, Bars: []string{
				"a1 - a1 - a1 - a1 - a1 - a1 - e2 - g1 -",
				"a1 - a1 - a1 - a1 - c2 - c2 - e2 - e2 -",
				"f1 - f1 - f1 - f1 - f1 - f1 - c2 - c2 -",
				"e1 - e1 - e1 - e1 - e1 - g#1 - b1 - e2 -",
				"a1 - a1 - a1 - a1 - a1 - a1 - e2 - g1 -",
				"d1 - d1 - d1 - d1 - d1 - d1 - a1 - a1 -",
				"g1 - g1 - g1 - g1 - g1 - g1 - d2 - d2 -",
				"e1 - e1 - e1 - e1 - b1 - e2 - g#1 - b1 -",
			}},
		},
		Drums: []string{
			"k . . h s . . h k . k h s . h h",
			"k . . h s . . h k . k h s . hH .",
		},
	},

	// -----------------------------------------------------------------
	// World 3 — Whale Depths. Slow, shimmering D minor.
	// -----------------------------------------------------------------
	MusWorld3: {
		Name: "THE TRENCH", BPM: 100, Delay: 0.42, DelayT: 0.45, Cutoff: 3800, DrumVol: 0.5,
		Chans: []Chan{
			{Wave: WaveTri, Vol: 0.32, Env: envPad, Pan: -0.15, Vib: 0.010, VibRate: 4, Bars: []string{
				"d5 - - - a4 - - - f5 - - - a5 - - -",
				"g5 - - - f5 - - - d5 - - - . . . .",
				"c5 - - - e5 - - - g5 - - - e5 - - -",
				"d5 - - - - - - - . . . . . . . .",
				"a5 - - - g5 - - - f5 - - - e5 - - -",
				"d5 - - - c5 - - - a4 - - - . . . .",
				"bb4 - - - d5 - - - f5 - - - a5 - - -",
				"g5 - - - - - - - a5 - - - - - - -",
			}},
			{Wave: WavePulse50, Vol: 0.075, Env: envPluck, Pan: 0.35, Bars: []string{
				"d4 f4 a4 d5 a4 f4 d4 f4 a4 d5 f5 d5 a4 f4 d4 f4",
				"g3 bb3 d4 g4 d4 bb3 g3 bb3 d4 g4 bb4 g4 d4 bb3 g3 bb3",
				"c4 e4 g4 c5 g4 e4 c4 e4 g4 c5 e5 c5 g4 e4 c4 e4",
				"d4 f4 a4 d5 a4 f4 d4 f4 a4 d5 f5 d5 a4 f4 d4 f4",
				"a3 c4 e4 a4 e4 c4 a3 c4 e4 a4 c5 a4 e4 c4 a3 c4",
				"d4 f4 a4 d5 a4 f4 d4 f4 a4 d5 f5 d5 a4 f4 d4 f4",
				"bb3 d4 f4 bb4 f4 d4 bb3 d4 f4 bb4 d5 bb4 f4 d4 bb3 d4",
				"a3 c#4 e4 a4 e4 c#4 a3 c#4 e4 a4 c#5 a4 e4 c#4 a3 c#4",
			}},
			{Wave: WaveSine, Vol: 0.40, Env: envPad, Bars: []string{
				"d2 - - - - - - - a1 - - - - - - -",
				"g1 - - - - - - - d2 - - - - - - -",
				"c2 - - - - - - - g1 - - - - - - -",
				"d2 - - - - - - - a1 - - - - - - -",
				"a1 - - - - - - - e2 - - - - - - -",
				"d2 - - - - - - - a1 - - - - - - -",
				"bb1 - - - - - - - f2 - - - - - - -",
				"a1 - - - - - - - e2 - - - - - - -",
			}},
		},
		Drums: []string{
			"k . . . . . h . . . k . . . h .",
			"k . . . . . h . . . k . s . h .",
		},
	},

	// -----------------------------------------------------------------
	// World 4 — Moonshot Orbital. Fast, weightless E minor.
	// -----------------------------------------------------------------
	MusWorld4: {
		Name: "ESCAPE VELOCITY", BPM: 152, Delay: 0.34, DelayT: 0.24, Cutoff: 9200, DrumVol: 0.8,
		Chans: []Chan{
			{Wave: WavePulse12, Vol: 0.27, Env: envLead, Pan: -0.18, Vib: 0.005, Bars: []string{
				"e5 - - - b4 - - - e5 - - - g5 - - -",
				"f#5 - - - e5 - - - b4 - - - . . . .",
				"d5 - - - a4 - - - d5 - - - f#5 - - -",
				"e5 - - - - - - - . . . . . . . .",
				"g5 - - - f#5 - - - e5 - - - d5 - - -",
				"b4 - - - d5 - - - e5 - - - . . . .",
				"c5 - - - g4 - - - c5 - - - e5 - - -",
				"b4 - - - - - - - - - - - . . . .",
			}},
			{Wave: WavePulse25, Vol: 0.085, Env: envPluck, Pan: 0.34, Bars: []string{
				"e4 g4 b4 e5 b4 g4 e4 g4 b4 e5 g5 e5 b4 g4 e4 g4",
				"b3 d4 f#4 b4 f#4 d4 b3 d4 f#4 b4 d5 b4 f#4 d4 b3 d4",
				"d4 f#4 a4 d5 a4 f#4 d4 f#4 a4 d5 f#5 d5 a4 f#4 d4 f#4",
				"e4 g4 b4 e5 b4 g4 e4 g4 b4 e5 g5 e5 b4 g4 e4 g4",
				"c4 e4 g4 c5 g4 e4 c4 e4 g4 c5 e5 c5 g4 e4 c4 e4",
				"e4 g4 b4 e5 b4 g4 e4 g4 b4 e5 g5 e5 b4 g4 e4 g4",
				"a3 c4 e4 a4 e4 c4 a3 c4 e4 a4 c5 a4 e4 c4 a3 c4",
				"b3 d#4 f#4 b4 f#4 d#4 b3 d#4 f#4 b4 d#5 b4 f#4 d#4 b3 d#4",
			}},
			{Wave: WaveTri, Vol: 0.34, Env: envBass, Bars: []string{
				"e1 - . e1 - . e1 - e1 - . e1 - . b1 -",
				"b1 - . b1 - . b1 - b1 - . b1 - . f#2 -",
				"d2 - . d2 - . d2 - d2 - . d2 - . a1 -",
				"e1 - . e1 - . e1 - e1 - . e1 - . b1 -",
				"c2 - . c2 - . c2 - c2 - . c2 - . g1 -",
				"e1 - . e1 - . e1 - e1 - . e1 - . b1 -",
				"a1 - . a1 - . a1 - a1 - . a1 - . e2 -",
				"b1 - . b1 - . b1 - f#2 - . b1 - . d#2 -",
			}},
		},
		Drums: []string{
			"k . h . s . h . k . h . s . h h",
			"k . h . s . h . k . h k s . hH .",
		},
	},

	// -----------------------------------------------------------------
	// World 5 — Cloud Fortress. Heavy D minor.
	// -----------------------------------------------------------------
	MusWorld5: {
		Name: "THE FURNACE", BPM: 138, Delay: 0.24, DelayT: 0.33, Cutoff: 5600, DrumVol: 1.0,
		Chans: []Chan{
			{Wave: WaveSaw, Vol: 0.20, Env: envLead, Pan: -0.2, Vib: 0.007, Bars: []string{
				"d5 - - . eb5 - - . d5 - - . a4 - - .",
				"bb4 - - . a4 - - . g4 - - - . . . .",
				"d5 - - . f5 - - . eb5 - - . d5 - - .",
				"a4 - - - - - - - . . . . . . . .",
				"f5 - - . eb5 - - . d5 - - . c5 - - .",
				"bb4 - - . a4 - - . f4 - - - . . . .",
				"g4 - - . bb4 - - . d5 - - . eb5 - - .",
				"d5 - - - - - - - - - - - . . . .",
			}},
			{Wave: WavePulse50, Vol: 0.12, Env: envStab, Pan: 0.3, Bars: []string{
				"d3 . . d3 . . a3 . d3 . . d3 . . a3 .",
				"bb2 . . bb2 . . f3 . a2 . . a2 . . e3 .",
				"d3 . . d3 . . a3 . d3 . . d3 . . f3 .",
				"a2 . . a2 . . e3 . a2 . . a2 . . c#3 .",
				"f3 . . f3 . . c4 . eb3 . . eb3 . . bb3 .",
				"bb2 . . bb2 . . f3 . a2 . . a2 . . e3 .",
				"g2 . . g2 . . d3 . bb2 . . bb2 . . f3 .",
				"a2 . . a2 . . e3 . a2 . . c#3 . . e3 .",
			}},
			{Wave: WaveTri, Vol: 0.40, Env: envBass, Bars: []string{
				"d2 d2 . d2 . d2 d2 . d2 d2 . d2 . a1 . .",
				"bb1 bb1 . bb1 . bb1 bb1 . a1 a1 . a1 . e2 . .",
				"d2 d2 . d2 . d2 d2 . d2 d2 . d2 . f2 . .",
				"a1 a1 . a1 . a1 a1 . a1 a1 . a1 . e2 . .",
				"f2 f2 . f2 . f2 f2 . eb2 eb2 . eb2 . bb1 . .",
				"bb1 bb1 . bb1 . bb1 bb1 . a1 a1 . a1 . e2 . .",
				"g1 g1 . g1 . g1 g1 . bb1 bb1 . bb1 . f2 . .",
				"a1 a1 . a1 . a1 a1 . a1 a1 . c#2 . e2 . .",
			}},
		},
		Drums: []string{
			"k . k . s . . k k . k . s . h h",
			"k . k . s . . k k . k . s t t t",
		},
	},

	// -----------------------------------------------------------------
	// Boss — relentless.
	// -----------------------------------------------------------------
	MusBoss: {
		Name: "SYNDICATE", BPM: 168, Delay: 0.18, DelayT: 0.19, Cutoff: 8800, DrumVol: 1.0,
		Chans: []Chan{
			{Wave: WavePulse25, Vol: 0.28, Env: envLead, Pan: -0.15, Bars: []string{
				"e5 e5 . e5 . e5 . g5 f#5 . e5 . d5 . . .",
				"e5 e5 . e5 . g5 . b5 a5 . g5 . e5 . . .",
				"c5 c5 . c5 . c5 . eb5 d5 . c5 . b4 . . .",
				"b4 b4 . d5 . f5 . ab5 g5 . f5 . e5 . . .",
			}},
			{Wave: WavePulse12, Vol: 0.11, Env: envStab, Pan: 0.32, Bars: []string{
				"e3 . e3 . e3 . e3 . e3 . e3 . e3 . e3 .",
				"e3 . e3 . e3 . e3 . g3 . g3 . g3 . g3 .",
				"c3 . c3 . c3 . c3 . c3 . c3 . c3 . c3 .",
				"b2 . b2 . b2 . b2 . f3 . f3 . f3 . f3 .",
			}},
			{Wave: WaveSaw, Vol: 0.22, Env: envBass, Bars: []string{
				"e2 e2 e2 e2 e2 e2 e2 e2 e2 e2 e2 e2 e2 e2 e2 e2",
				"e2 e2 e2 e2 e2 e2 e2 e2 g2 g2 g2 g2 g2 g2 g2 g2",
				"c2 c2 c2 c2 c2 c2 c2 c2 c2 c2 c2 c2 c2 c2 c2 c2",
				"b1 b1 b1 b1 b1 b1 b1 b1 f2 f2 f2 f2 f2 f2 f2 f2",
			}},
		},
		Drums: []string{
			"k h s h k h s h k h s h k h s h",
			"k h s h k h s h k h s h k t t t",
		},
	},

	// -----------------------------------------------------------------
	// Safety Halo — brief invincibility, so: frantic.
	// -----------------------------------------------------------------
	MusHalo: {
		Name: "HALO", BPM: 200, Delay: 0.16, DelayT: 0.15, Cutoff: 9600, DrumVol: 0.9,
		Chans: []Chan{
			{Wave: WavePulse25, Vol: 0.26, Env: envPluck, Pan: -0.1, Bars: []string{
				"c5 e5 g5 c6 g5 e5 c5 e5 g5 c6 e6 c6 g5 e5 c5 e5",
				"d5 f5 a5 d6 a5 f5 d5 f5 a5 d6 f6 d6 a5 f5 d5 f5",
			}},
			{Wave: WavePulse50, Vol: 0.10, Env: envPluck, Pan: 0.3, Octave: -1, Bars: []string{
				"c5 e5 g5 c6 g5 e5 c5 e5 g5 c6 e6 c6 g5 e5 c5 e5",
				"d5 f5 a5 d6 a5 f5 d5 f5 a5 d6 f6 d6 a5 f5 d5 f5",
			}},
			{Wave: WaveTri, Vol: 0.34, Env: envBass, Bars: []string{
				"c2 c2 c2 c2 c2 c2 c2 c2 c2 c2 c2 c2 c2 c2 c2 c2",
				"d2 d2 d2 d2 d2 d2 d2 d2 d2 d2 d2 d2 d2 d2 d2 d2",
			}},
		},
		Drums: []string{"k h k h k h k h k h k h k h k h"},
	},

	// -----------------------------------------------------------------
	// Fanfares
	// -----------------------------------------------------------------
	MusClear: {
		Name: "RESCUED", BPM: 150, Delay: 0.2, DelayT: 0.2, Cutoff: 9000, DrumVol: 0.8, Once: true,
		Chans: []Chan{
			{Wave: WavePulse25, Vol: 0.32, Env: envLead, Bars: []string{
				"c5 . e5 . g5 . c6 . a5 . c6 . e6 - - -",
				"d6 . c6 . g5 . e5 . c6 - - - - - - -",
			}},
			{Wave: WavePulse50, Vol: 0.14, Env: envStab, Pan: 0.3, Bars: []string{
				"c4 . c4 . e4 . e4 . f4 . f4 . g4 - - -",
				"g4 . g4 . g4 . g4 . c5 - - - - - - -",
			}},
			{Wave: WaveTri, Vol: 0.36, Env: envBass, Bars: []string{
				"c2 . c2 . c2 . c2 . f2 . f2 . g2 - - -",
				"g2 . g2 . g2 . g2 . c2 - - - - - - -",
			}},
		},
		Drums: []string{
			"k . k . k . k . s . s . c - - -",
			"k . k . k . k . c - - - - - - -",
		},
	},

	MusGameOver: {
		Name: "DISTILLED", BPM: 92, Delay: 0.3, DelayT: 0.4, Cutoff: 3600, Once: true,
		Chans: []Chan{
			{Wave: WavePulse50, Vol: 0.26, Env: envPad, Bars: []string{
				"g4 - - - eb4 - - - c4 - - - . . . .",
				"b3 - - - g3 - - - c3 - - - - - - -",
			}},
			{Wave: WaveTri, Vol: 0.30, Env: envPad, Bars: []string{
				"eb3 - - - c3 - - - ab2 - - - . . . .",
				"g2 - - - eb2 - - - c2 - - - - - - -",
			}},
			{Wave: WaveSine, Vol: 0.34, Env: envPad, Bars: []string{
				"c2 - - - - - - - ab1 - - - . . . .",
				"g1 - - - - - - - c2 - - - - - - -",
			}},
		},
	},

	MusEnding: {
		Name: "OPEN WEIGHTS OF THE HEART", BPM: 120, Delay: 0.26, DelayT: 0.3, Cutoff: 8600, DrumVol: 0.8,
		Chans: []Chan{
			{Wave: WavePulse25, Vol: 0.30, Env: envLead, Pan: -0.14, Vib: 0.005, Bars: []string{
				"g4 - c5 - e5 - g5 - c6 - - - b5 - - -",
				"a5 - - - g5 - - - e5 - - - c5 - - -",
				"f5 - a5 - c6 - a5 - g5 - - - e5 - - -",
				"d5 - - - e5 - - - c5 - - - - - - -",
				"e5 - g5 - c6 - e6 - d6 - - - c6 - - -",
				"b5 - - - a5 - - - g5 - - - e5 - - -",
				"f5 - a5 - c6 - f6 - e6 - - - d6 - - -",
				"c6 - - - - - - - - - - - - - - -",
			}},
			{Wave: WavePulse50, Vol: 0.095, Env: envPluck, Pan: 0.32, Bars: []string{
				"c4 e4 g4 e4 c4 e4 g4 e4 c4 e4 g4 e4 c4 e4 g4 e4",
				"a3 c4 e4 c4 a3 c4 e4 c4 g3 b3 d4 b3 g3 b3 d4 b3",
				"f3 a3 c4 a3 f3 a3 c4 a3 c4 e4 g4 e4 c4 e4 g4 e4",
				"g3 b3 d4 b3 g3 b3 d4 b3 c4 e4 g4 e4 c4 e4 g4 e4",
				"c4 e4 g4 e4 c4 e4 g4 e4 a3 c4 e4 c4 a3 c4 e4 c4",
				"g3 b3 d4 b3 g3 b3 d4 b3 e4 g4 b4 g4 e4 g4 b4 g4",
				"f3 a3 c4 a3 f3 a3 c4 a3 g3 b3 d4 b3 g3 b3 d4 b3",
				"c4 e4 g4 c5 e5 g5 c6 - - - - - - - - -",
			}},
			{Wave: WaveTri, Vol: 0.34, Env: envBass, Bars: []string{
				"c2 - - - g2 - - - c2 - - - e2 - - -",
				"a1 - - - e2 - - - g1 - - - d2 - - -",
				"f1 - - - c2 - - - c2 - - - g2 - - -",
				"g1 - - - d2 - - - c2 - - - g2 - - -",
				"c2 - - - g2 - - - a1 - - - e2 - - -",
				"g1 - - - d2 - - - e2 - - - b2 - - -",
				"f1 - - - c2 - - - g1 - - - d2 - - -",
				"c2 - - - - - - - - - - - - - - -",
			}},
		},
		Drums: []string{
			"k . h . s . h . k . h k s . h .",
			"k . h . s . h . k . h k s . hH c",
		},
	},
}
