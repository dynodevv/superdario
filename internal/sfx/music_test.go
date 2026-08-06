package sfx

import (
	"encoding/binary"
	"math"
	"strings"
	"testing"
)

// Every bar must be exactly sixteen sixteenth-notes, and every non-rest token
// must be a note the synth understands. A typo here would silently mangle a
// melody, so it is worth a test.
func TestBarsWellFormed(t *testing.T) {
	checkLane := func(name string, bars []string) {
		for i, bar := range bars {
			toks := strings.Fields(bar)
			if len(toks) != 16 {
				t.Errorf("%s bar %d: got %d tokens, want 16 (%q)", name, i+1, len(toks), bar)
			}
			for _, tok := range toks {
				if tok == "." || tok == "-" || tok == "^" {
					continue
				}
				if _, ok := noteFreq(tok); !ok {
					t.Errorf("%s bar %d: unparseable note %q", name, i+1, tok)
				}
			}
		}
	}
	for id, tr := range tracks {
		for ci, c := range tr.Chans {
			checkLane(nameOf(id)+" chan"+string(rune('0'+ci)), c.Bars)
		}
		for i, bar := range tr.Drums {
			toks := strings.Fields(bar)
			if len(toks) != 16 {
				t.Errorf("%s drums bar %d: got %d tokens, want 16", nameOf(id), i+1, len(toks))
			}
			for _, tok := range toks {
				if tok == "." || tok == "-" {
					continue
				}
				for k := 0; k < len(tok); k++ {
					if !strings.ContainsRune("kshHctr", rune(tok[k])) {
						t.Errorf("%s drums bar %d: unknown drum %q", nameOf(id), i+1, string(tok[k]))
					}
				}
			}
		}
	}
}

func nameOf(id TrackID) string {
	if tr, ok := tracks[id]; ok {
		return tr.Name
	}
	return "?"
}

// Render a second of each track and make sure it is audible, finite and not
// pinned against the clipper.
func TestTracksProduceAudio(t *testing.T) {
	for id, tr := range tracks {
		s := NewStream(tr)
		const frames = SampleRate * 2
		buf := make([]byte, frames*8)
		if _, err := s.Read(buf); err != nil {
			t.Fatalf("%s: read: %v", tr.Name, err)
		}
		var sum, peak float64
		for i := 0; i < frames; i++ {
			l := float64(math.Float32frombits(binary.LittleEndian.Uint32(buf[i*8:])))
			if math.IsNaN(l) || math.IsInf(l, 0) {
				t.Fatalf("%s: non-finite sample at %d", tr.Name, i)
			}
			sum += l * l
			if a := math.Abs(l); a > peak {
				peak = a
			}
		}
		rms := math.Sqrt(sum / frames)
		if rms < 0.005 {
			t.Errorf("%s (id %d): effectively silent, rms=%.4f", tr.Name, id, rms)
		}
		if peak > 0.99 {
			t.Errorf("%s: peak %.3f is clipping", tr.Name, peak)
		}
		t.Logf("%-26s rms=%.3f peak=%.3f", tr.Name, rms, peak)
	}
}

func TestSoundsProduceAudio(t *testing.T) {
	snds := buildSounds()
	for i, data := range snds {
		if len(data) == 0 {
			t.Errorf("sound %d is empty", i)
			continue
		}
		n := len(data) / 8
		var sum, peak float64
		for k := 0; k < n; k++ {
			v := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[k*8:])))
			if math.IsNaN(v) || math.IsInf(v, 0) {
				t.Fatalf("sound %d: non-finite sample", i)
			}
			sum += v * v
			if a := math.Abs(v); a > peak {
				peak = a
			}
		}
		rms := math.Sqrt(sum / float64(n))
		if rms < 0.004 {
			t.Errorf("sound %d: effectively silent (rms=%.4f)", i, rms)
		}
		if peak > 0.99 {
			t.Errorf("sound %d: peak %.3f is clipping", i, peak)
		}
	}
}

func TestNoteFreq(t *testing.T) {
	cases := []struct {
		tok  string
		want float64
	}{
		{"a4", 440}, {"c4", 261.626}, {"a5", 880}, {"c#4", 277.183},
		{"bb3", 233.082}, {"e2", 82.407}, {"g#1", 51.913},
	}
	for _, c := range cases {
		got, ok := noteFreq(c.tok)
		if !ok {
			t.Fatalf("%q failed to parse", c.tok)
		}
		if math.Abs(got-c.want) > 0.01 {
			t.Errorf("%q = %.3f, want %.3f", c.tok, got, c.want)
		}
	}
	for _, bad := range []string{"h4", "", "c", "x#9", "4"} {
		if _, ok := noteFreq(bad); ok {
			t.Errorf("%q parsed but should not have", bad)
		}
	}
}
