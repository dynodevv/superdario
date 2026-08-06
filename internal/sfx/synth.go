// Package sfx is a small chiptune synthesizer and sequencer. All music and
// sound effects are generated at runtime — the game ships no audio files.
package sfx

import (
	"math"
	"strings"
)

const SampleRate = 48000

// ---------------------------------------------------------------------------
// Oscillators
// ---------------------------------------------------------------------------

type Wave uint8

const (
	WavePulse12 Wave = iota // thin, reedy — leads
	WavePulse25             // classic NES square
	WavePulse50             // hollow square — harmony
	WaveTri                 // soft — bass
	WaveSaw                 // buzzy — brass-ish
	WaveSine                // pure — kicks, pads
	WaveNoise               // percussion / effects
)

func (w Wave) duty() float64 {
	switch w {
	case WavePulse12:
		return 0.125
	case WavePulse25:
		return 0.25
	default:
		return 0.5
	}
}

// ---------------------------------------------------------------------------
// Envelope
// ---------------------------------------------------------------------------

type Env struct {
	A, D, S, R float64 // attack/decay/release in seconds, sustain as 0..1
}

type envStage uint8

const (
	envIdle envStage = iota
	envAttack
	envDecay
	envSustain
	envRelease
)

// ---------------------------------------------------------------------------
// Voice
// ---------------------------------------------------------------------------

type voice struct {
	wave  Wave
	phase float64
	freq  float64
	vol   float64
	pan   float64 // -1 left .. +1 right

	env      Env
	stage    envStage
	level    float64
	stageT   float64
	relLevel float64

	// expression
	vibDepth, vibRate, vibPhase float64
	sweep                       float64 // multiplicative pitch drift per second
	arp                         []int   // semitone offsets cycled while held
	arpRate                     float64 // cycles per second
	arpT                        float64

	// noise generator state (15-bit LFSR, as on the NES)
	lfsr     uint16
	noiseAcc float64
	noiseVal float64

	baseFreq float64
	active   bool
}

func (v *voice) noteOn(freq, vol float64, w Wave, e Env) {
	v.freq = freq
	v.baseFreq = freq
	v.vol = vol
	v.wave = w
	v.env = e
	v.stage = envAttack
	v.stageT = 0
	v.active = true
	v.arpT = 0
	if v.lfsr == 0 {
		v.lfsr = 0x7FFF
	}
	// A fresh phase on every note makes attacks click; keep the running phase
	// for tonal waves and only reset percussive ones.
	if w == WaveNoise || w == WaveSine {
		v.phase = 0
	}
}

func (v *voice) noteOff() {
	if v.stage != envIdle && v.stage != envRelease {
		v.relLevel = v.level
		v.stage = envRelease
		v.stageT = 0
	}
}

func (v *voice) silence() {
	v.stage = envIdle
	v.level = 0
	v.active = false
}

const dt = 1.0 / SampleRate

// sample advances the voice one frame and returns (left, right).
func (v *voice) sample() (float64, float64) {
	if !v.active || v.stage == envIdle {
		return 0, 0
	}

	// --- envelope -------------------------------------------------------
	v.stageT += dt
	switch v.stage {
	case envAttack:
		if v.env.A <= 0 {
			v.level = 1
			v.stage = envDecay
			v.stageT = 0
		} else {
			v.level = v.stageT / v.env.A
			if v.level >= 1 {
				v.level = 1
				v.stage = envDecay
				v.stageT = 0
			}
		}
	case envDecay:
		if v.env.D <= 0 {
			v.level = v.env.S
			v.stage = envSustain
		} else {
			f := v.stageT / v.env.D
			if f >= 1 {
				v.level = v.env.S
				v.stage = envSustain
				v.stageT = 0
			} else {
				v.level = 1 + (v.env.S-1)*f
			}
		}
	case envSustain:
		v.level = v.env.S
	case envRelease:
		if v.env.R <= 0 {
			v.silence()
			return 0, 0
		}
		f := v.stageT / v.env.R
		if f >= 1 {
			v.silence()
			return 0, 0
		}
		v.level = v.relLevel * (1 - f)
	}

	// --- pitch ----------------------------------------------------------
	f := v.freq
	if v.sweep != 0 {
		v.freq *= math.Pow(v.sweep, dt)
		if v.freq < 8 {
			v.freq = 8
		}
	}
	if len(v.arp) > 0 {
		v.arpT += dt
		idx := int(v.arpT*v.arpRate) % len(v.arp)
		f = v.baseFreq * math.Pow(2, float64(v.arp[idx])/12)
	}
	if v.vibDepth > 0 {
		v.vibPhase += v.vibRate * dt
		f *= 1 + v.vibDepth*math.Sin(v.vibPhase*2*math.Pi)
	}

	// --- oscillator -----------------------------------------------------
	var s float64
	if v.wave == WaveNoise {
		// step the LFSR at a rate proportional to the requested frequency
		v.noiseAcc += f * dt
		for v.noiseAcc >= 1 {
			v.noiseAcc -= 1
			bit := (v.lfsr ^ (v.lfsr >> 1)) & 1
			v.lfsr = (v.lfsr >> 1) | (bit << 14)
			if v.lfsr&1 == 1 {
				v.noiseVal = 1
			} else {
				v.noiseVal = -1
			}
		}
		s = v.noiseVal
	} else {
		v.phase += f * dt
		if v.phase >= 1 {
			v.phase -= math.Floor(v.phase)
		}
		switch v.wave {
		case WaveTri:
			s = 4*math.Abs(v.phase-0.5) - 1
		case WaveSaw:
			s = 2*v.phase - 1
		case WaveSine:
			s = math.Sin(v.phase * 2 * math.Pi)
		default: // pulse family
			if v.phase < v.wave.duty() {
				s = 1
			} else {
				s = -1
			}
		}
	}

	a := s * v.level * v.vol
	// equal-power-ish pan
	l := a * math.Sqrt((1-v.pan)/2)
	r := a * math.Sqrt((1+v.pan)/2)
	return l, r
}

// ---------------------------------------------------------------------------
// Note names
// ---------------------------------------------------------------------------

var semitone = map[byte]int{'c': 0, 'd': 2, 'e': 4, 'f': 5, 'g': 7, 'a': 9, 'b': 11}

// noteFreq parses tokens like "c4", "f#3", "bb5" into a frequency in Hz.
func noteFreq(tok string) (float64, bool) {
	if len(tok) < 2 {
		return 0, false
	}
	t := strings.ToLower(tok)
	st, ok := semitone[t[0]]
	if !ok {
		return 0, false
	}
	i := 1
	if i < len(t) && (t[i] == '#' || t[i] == 's') {
		st++
		i++
	} else if i < len(t) && t[i] == 'b' {
		st--
		i++
	}
	if i >= len(t) {
		return 0, false
	}
	oct := int(t[i] - '0')
	if oct < 0 || oct > 9 {
		return 0, false
	}
	// MIDI note number, with A4 (midi 69) = 440 Hz
	midi := (oct+1)*12 + st
	return 440 * math.Pow(2, float64(midi-69)/12), true
}

// ---------------------------------------------------------------------------
// Master effects
// ---------------------------------------------------------------------------

// delayLine is a stereo feedback echo. It gives the chip voices a sense of
// space that a bare oscillator mix never has.
type delayLine struct {
	bufL, bufR []float64
	idx        int
	feedback   float64
	mix        float64
}

func newDelay(seconds, feedback, mix float64) *delayLine {
	n := int(seconds * SampleRate)
	if n < 1 {
		n = 1
	}
	return &delayLine{bufL: make([]float64, n), bufR: make([]float64, n), feedback: feedback, mix: mix}
}

func (d *delayLine) process(l, r float64) (float64, float64) {
	dl := d.bufL[d.idx]
	dr := d.bufR[d.idx]
	// cross-feed the channels slightly for a wider image
	d.bufL[d.idx] = l + dr*d.feedback
	d.bufR[d.idx] = r + dl*d.feedback
	d.idx++
	if d.idx >= len(d.bufL) {
		d.idx = 0
	}
	return l + dl*d.mix, r + dr*d.mix
}

// onePole is a gentle low-pass that takes the harsh edge off square waves.
type onePole struct {
	z float64
	a float64
}

func newLowpass(cutoff float64) *onePole {
	x := math.Exp(-2 * math.Pi * cutoff / SampleRate)
	return &onePole{a: 1 - x}
}

func (p *onePole) process(v float64) float64 {
	p.z += p.a * (v - p.z)
	return p.z
}

// softClip keeps the mix from squaring off when several voices align.
func softClip(v float64) float64 {
	if v > 1.4 {
		v = 1.4
	}
	if v < -1.4 {
		v = -1.4
	}
	return v - (v*v*v)/6.75
}
