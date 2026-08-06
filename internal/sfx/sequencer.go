package sfx

import (
	"encoding/binary"
	"math"
	"strings"
	"sync"
)

// A tiny tracker. A Track is a handful of monophonic channels plus a drum
// lane; each lane is a list of bars, and each bar is 16 space-separated
// tokens (one sixteenth note apiece):
//
//	"c4"  play middle-C-octave-4      "-"  hold the previous note
//	"."   rest (release the note)     "^"  hard note-off
//
// The drum lane uses one letter per hit and packs several into one token:
// k kick, s snare, h closed hat, H open hat, c crash, t tom, r rim.
// "." and "-" are both rests there.
//
// Lanes loop independently at their own length, so a one-bar drum pattern
// happily repeats under an eight-bar melody.

type Chan struct {
	Wave    Wave
	Vol     float64
	Env     Env
	Pan     float64
	Bars    []string
	Arp     []int   // semitone offsets; turns each note into a rolling chord
	ArpRate float64 // arpeggio cycles per second
	Vib     float64 // vibrato depth (fraction of pitch)
	VibRate float64
	Octave  int // transpose, in octaves
	Gate    float64
}

type Track struct {
	Name    string
	BPM     float64
	Chans   []Chan
	Drums   []string
	DrumVol float64
	Delay   float64 // echo wet mix, 0 disables
	DelayT  float64 // echo time in seconds
	Cutoff  float64 // master low-pass cutoff, Hz
	Swing   float64 // 0..0.4, delays every other sixteenth
	Once    bool    // fanfares play through and stop
}

// ---------------------------------------------------------------------------
// Stream
// ---------------------------------------------------------------------------

type chanState struct {
	cfg    Chan
	tokens []string
	v      voice
}

// Stream renders a Track to interleaved 32-bit float stereo on demand.
// It satisfies io.Reader, which is what Ebitengine's audio player wants.
type Stream struct {
	mu sync.Mutex

	track   Track
	chans   []chanState
	drums   []string
	drumV   [6]voice
	drumIdx int

	step        int
	sampPerStep float64
	sampLeft    float64
	totalSteps  int
	finished    bool

	delay *delayLine
	lpL   *onePole
	lpR   *onePole

	volume float64
	fade   float64 // 0..1 multiplier ramp target
	fadeTo float64
}

func splitBars(bars []string) []string {
	var out []string
	for _, b := range bars {
		out = append(out, strings.Fields(b)...)
	}
	return out
}

func NewStream(t Track) *Stream {
	s := &Stream{track: t, volume: 1, fade: 1, fadeTo: 1}
	if t.BPM <= 0 {
		t.BPM = 120
	}
	s.sampPerStep = SampleRate * 60 / (t.BPM * 4)
	for _, c := range t.Chans {
		cs := chanState{cfg: c, tokens: splitBars(c.Bars)}
		cs.v.pan = c.Pan
		cs.v.lfsr = 0x7FFF
		s.chans = append(s.chans, cs)
		if n := len(cs.tokens); n > s.totalSteps {
			s.totalSteps = n
		}
	}
	s.drums = splitBars(t.Drums)
	if n := len(s.drums); n > s.totalSteps {
		s.totalSteps = n
	}
	for i := range s.drumV {
		s.drumV[i].lfsr = uint16(0x7FFF - i*37)
	}
	dt := t.DelayT
	if dt <= 0 {
		dt = 0.24
	}
	mix := t.Delay
	s.delay = newDelay(dt, 0.34, mix)
	cut := t.Cutoff
	if cut <= 0 {
		cut = 9000
	}
	s.lpL, s.lpR = newLowpass(cut), newLowpass(cut)
	s.step = -1 // so the first advance lands on step 0
	return s
}

// SetVolume adjusts the stream's output level (used by the music mixer).
func (s *Stream) SetVolume(v float64) {
	s.mu.Lock()
	s.volume = v
	s.mu.Unlock()
}

// FadeOut ramps the stream down; the mixer drops the player once silent.
func (s *Stream) FadeOut() {
	s.mu.Lock()
	s.fadeTo = 0
	s.mu.Unlock()
}

func (s *Stream) Done() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.finished || (s.fadeTo == 0 && s.fade <= 0.001)
}

func (s *Stream) advance() {
	s.step++
	if s.totalSteps == 0 {
		return
	}
	if s.track.Once && s.step >= s.totalSteps {
		for i := range s.chans {
			s.chans[i].v.noteOff()
		}
		s.finished = true
		return
	}

	for i := range s.chans {
		cs := &s.chans[i]
		if len(cs.tokens) == 0 {
			continue
		}
		tok := cs.tokens[s.step%len(cs.tokens)]
		switch tok {
		case "-":
			// hold
		case ".", "^":
			cs.v.noteOff()
		default:
			f, ok := noteFreq(tok)
			if !ok {
				cs.v.noteOff()
				break
			}
			f *= math.Pow(2, float64(cs.cfg.Octave))
			env := cs.cfg.Env
			if env.A == 0 && env.D == 0 && env.S == 0 && env.R == 0 {
				env = Env{A: 0.004, D: 0.08, S: 0.62, R: 0.09}
			}
			cs.v.noteOn(f, cs.cfg.Vol, cs.cfg.Wave, env)
			cs.v.arp = cs.cfg.Arp
			cs.v.arpRate = cs.cfg.ArpRate
			if cs.v.arpRate == 0 {
				cs.v.arpRate = 18
			}
			cs.v.vibDepth = cs.cfg.Vib
			cs.v.vibRate = cs.cfg.VibRate
			if cs.v.vibRate == 0 {
				cs.v.vibRate = 5.5
			}
			cs.v.sweep = 0
		}
	}

	if len(s.drums) > 0 {
		tok := s.drums[s.step%len(s.drums)]
		for i := 0; i < len(tok); i++ {
			s.triggerDrum(tok[i])
		}
	}
}

func (s *Stream) nextDrumVoice() *voice {
	v := &s.drumV[s.drumIdx]
	s.drumIdx = (s.drumIdx + 1) % len(s.drumV)
	return v
}

func (s *Stream) triggerDrum(k byte) {
	vol := s.track.DrumVol
	if vol == 0 {
		vol = 0.8
	}
	switch k {
	case 'k': // kick
		v := s.nextDrumVoice()
		v.noteOn(170, 0.95*vol, WaveSine, Env{A: 0.001, D: 0.12, S: 0, R: 0.02})
		v.sweep = 1e-7
		v.pan = 0
	case 's': // snare: noise crack plus a little body
		v := s.nextDrumVoice()
		v.noteOn(7200, 0.42*vol, WaveNoise, Env{A: 0.001, D: 0.13, S: 0, R: 0.04})
		v.pan = 0
		b := s.nextDrumVoice()
		b.noteOn(210, 0.28*vol, WaveTri, Env{A: 0.001, D: 0.07, S: 0, R: 0.02})
		b.sweep = 0.02
		b.pan = 0
	case 'h': // closed hat
		v := s.nextDrumVoice()
		v.noteOn(15000, 0.17*vol, WaveNoise, Env{A: 0.0005, D: 0.030, S: 0, R: 0.01})
		v.pan = 0.35
	case 'H': // open hat
		v := s.nextDrumVoice()
		v.noteOn(14000, 0.15*vol, WaveNoise, Env{A: 0.001, D: 0.17, S: 0, R: 0.05})
		v.pan = 0.35
	case 'c': // crash
		v := s.nextDrumVoice()
		v.noteOn(9500, 0.20*vol, WaveNoise, Env{A: 0.002, D: 0.55, S: 0, R: 0.2})
		v.pan = -0.3
	case 't': // tom
		v := s.nextDrumVoice()
		v.noteOn(300, 0.55*vol, WaveSine, Env{A: 0.001, D: 0.16, S: 0, R: 0.04})
		v.sweep = 0.03
		v.pan = -0.25
	case 'r': // rim / click
		v := s.nextDrumVoice()
		v.noteOn(3000, 0.22*vol, WaveNoise, Env{A: 0.0005, D: 0.02, S: 0, R: 0.008})
		v.pan = -0.4
	}
}

// Read renders the next block of audio.
func (s *Stream) Read(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	frames := len(p) / 8
	for i := 0; i < frames; i++ {
		if s.sampLeft <= 0 {
			s.advance()
			step := s.sampPerStep
			// swing: hold odd sixteenths back a little
			if s.track.Swing > 0 {
				if s.step%2 == 0 {
					step *= 1 + s.track.Swing
				} else {
					step *= 1 - s.track.Swing
				}
			}
			s.sampLeft += step
		}
		s.sampLeft--

		var l, r float64
		for c := range s.chans {
			cl, cr := s.chans[c].v.sample()
			l += cl
			r += cr
		}
		for d := range s.drumV {
			dl, dr := s.drumV[d].sample()
			l += dl
			r += dr
		}

		l = s.lpL.process(l)
		r = s.lpR.process(r)
		if s.delay.mix > 0 {
			l, r = s.delay.process(l, r)
		}

		// fade ramp (about 0.5s)
		if s.fade < s.fadeTo {
			s.fade = math.Min(s.fadeTo, s.fade+dt*2)
		} else if s.fade > s.fadeTo {
			s.fade = math.Max(s.fadeTo, s.fade-dt*2)
		}
		g := s.volume * s.fade * 0.5
		l = softClip(l * g)
		r = softClip(r * g)

		binary.LittleEndian.PutUint32(p[i*8:], math.Float32bits(float32(l)))
		binary.LittleEndian.PutUint32(p[i*8+4:], math.Float32bits(float32(r)))
	}
	// pad any partial frame with silence so the player never stutters
	for i := frames * 8; i < len(p); i++ {
		p[i] = 0
	}
	return len(p), nil
}

// ---------------------------------------------------------------------------
// Offline rendering, used for the short sound effects
// ---------------------------------------------------------------------------

// renderBytes runs a per-sample generator for d seconds into a float32 buffer.
func renderBytes(d float64, gen func(i int, t float64) (float64, float64)) []byte {
	n := int(d * SampleRate)
	buf := make([]byte, n*8)
	for i := 0; i < n; i++ {
		l, r := gen(i, float64(i)*dt)
		binary.LittleEndian.PutUint32(buf[i*8:], math.Float32bits(float32(softClip(l))))
		binary.LittleEndian.PutUint32(buf[i*8+4:], math.Float32bits(float32(softClip(r))))
	}
	return buf
}
