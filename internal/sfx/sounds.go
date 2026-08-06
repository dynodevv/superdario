package sfx

import "math"

// Sound effects are rendered offline into small PCM buffers at startup.

type SoundID int

const (
	SndJump SoundID = iota
	SndJumpBig
	SndStomp
	SndToken
	SndBump
	SndBreak
	SndPowerUp
	SndPowerDown
	SndDie
	SndShoot
	SndKick
	SndOneUp
	SndPause
	SndMenuMove
	SndMenuPick
	SndBossHit
	SndBossDie
	SndPipe
	SndSplash
	SndCheckpoint
	SndBlip
	SndUnlock
	SndRescue
	SndSpring
	SndSwim
	SndCount
)

// tiny offline oscillator
type osc struct {
	ph   float64
	lfsr uint16
	nAcc float64
	nVal float64
}

func (o *osc) adv(f float64) {
	o.ph += f * dt
	if o.ph >= 1 {
		o.ph -= math.Floor(o.ph)
	}
}

func (o *osc) pulse(f, duty float64) float64 {
	o.adv(f)
	if o.ph < duty {
		return 1
	}
	return -1
}

func (o *osc) sine(f float64) float64 {
	o.adv(f)
	return math.Sin(o.ph * 2 * math.Pi)
}

func (o *osc) tri(f float64) float64 {
	o.adv(f)
	return 4*math.Abs(o.ph-0.5) - 1
}

func (o *osc) saw(f float64) float64 {
	o.adv(f)
	return 2*o.ph - 1
}

func (o *osc) noise(f float64) float64 {
	if o.lfsr == 0 {
		o.lfsr = 0x7FFF
	}
	o.nAcc += f * dt
	for o.nAcc >= 1 {
		o.nAcc--
		bit := (o.lfsr ^ (o.lfsr >> 1)) & 1
		o.lfsr = (o.lfsr >> 1) | (bit << 14)
		if o.lfsr&1 == 1 {
			o.nVal = 1
		} else {
			o.nVal = -1
		}
	}
	return o.nVal
}

func decay(t, rate float64) float64 { return math.Exp(-t * rate) }

// stepFreq walks a frequency through a list of (time, freq) points — the
// classic way arcade blips get their shape.
func stepFreq(t float64, pts ...[2]float64) float64 {
	if len(pts) == 0 {
		return 440
	}
	if t <= pts[0][0] {
		return pts[0][1]
	}
	for i := 0; i < len(pts)-1; i++ {
		a, b := pts[i], pts[i+1]
		if t >= a[0] && t < b[0] {
			f := (t - a[0]) / (b[0] - a[0])
			return a[1] + (b[1]-a[1])*f
		}
	}
	return pts[len(pts)-1][1]
}

func mono(dur float64, gen func(t float64) float64) []byte {
	return renderBytes(dur, func(i int, t float64) (float64, float64) {
		v := gen(t)
		return v, v
	})
}

func buildSounds() [][]byte {
	s := make([][]byte, SndCount)

	s[SndJump] = mono(0.18, func() func(float64) float64 {
		var o osc
		return func(t float64) float64 {
			f := stepFreq(t, [2]float64{0, 330}, [2]float64{0.10, 880}, [2]float64{0.18, 900})
			return o.pulse(f, 0.5) * decay(t, 12) * 0.42
		}
	}())

	s[SndJumpBig] = mono(0.22, func() func(float64) float64 {
		var o osc
		return func(t float64) float64 {
			f := stepFreq(t, [2]float64{0, 230}, [2]float64{0.13, 700}, [2]float64{0.22, 720})
			return o.pulse(f, 0.25)*decay(t, 9)*0.34 + o.tri(f*0.5)*decay(t, 11)*0.18
		}
	}())

	s[SndStomp] = mono(0.16, func() func(float64) float64 {
		var o, n osc
		return func(t float64) float64 {
			f := stepFreq(t, [2]float64{0, 620}, [2]float64{0.08, 120})
			return o.pulse(f, 0.5)*decay(t, 18)*0.36 + n.noise(3200)*decay(t, 34)*0.22
		}
	}())

	s[SndToken] = mono(0.30, func() func(float64) float64 {
		var o osc
		return func(t float64) float64 {
			f := 988.0 // B5
			if t > 0.045 {
				f = 1319 // E6
			}
			return o.pulse(f, 0.25) * decay(math.Max(0, t-0.04), 9) * 0.30
		}
	}())

	s[SndBump] = mono(0.09, func() func(float64) float64 {
		var o osc
		return func(t float64) float64 {
			f := stepFreq(t, [2]float64{0, 200}, [2]float64{0.09, 90})
			return o.tri(f) * decay(t, 26) * 0.5
		}
	}())

	s[SndBreak] = mono(0.30, func() func(float64) float64 {
		var n, o osc
		return func(t float64) float64 {
			return n.noise(stepFreq(t, [2]float64{0, 9000}, [2]float64{0.3, 2200}))*decay(t, 11)*0.35 +
				o.saw(stepFreq(t, [2]float64{0, 400}, [2]float64{0.25, 80}))*decay(t, 16)*0.18
		}
	}())

	s[SndPowerUp] = mono(0.62, func() func(float64) float64 {
		var o, h osc
		// a rising arpeggio: C E G C E G C
		notes := []float64{523, 659, 784, 1047, 1319, 1568, 2093}
		return func(t float64) float64 {
			i := int(t / 0.075)
			if i >= len(notes) {
				i = len(notes) - 1
			}
			f := notes[i]
			e := decay(math.Mod(t, 0.075), 14)
			return (o.pulse(f, 0.5)*0.26 + h.pulse(f*2, 0.25)*0.10) * e * math.Min(1, (0.62-t)*6)
		}
	}())

	s[SndPowerDown] = mono(0.55, func() func(float64) float64 {
		var o osc
		notes := []float64{1047, 880, 740, 587, 494, 392}
		return func(t float64) float64 {
			i := int(t / 0.09)
			if i >= len(notes) {
				i = len(notes) - 1
			}
			return o.pulse(notes[i], 0.25) * decay(math.Mod(t, 0.09), 12) * 0.30
		}
	}())

	s[SndDie] = mono(1.10, func() func(float64) float64 {
		var o, b osc
		return func(t float64) float64 {
			var f float64
			switch {
			case t < 0.18:
				f = stepFreq(t, [2]float64{0, 660}, [2]float64{0.18, 440})
			case t < 0.34:
				f = 330
			default:
				f = stepFreq(t, [2]float64{0.34, 700}, [2]float64{1.10, 90})
			}
			e := 0.34
			if t > 0.9 {
				e *= (1.10 - t) / 0.2
			}
			return o.pulse(f, 0.5)*e + b.tri(f/2)*e*0.4
		}
	}())

	s[SndShoot] = mono(0.14, func() func(float64) float64 {
		var o, n osc
		return func(t float64) float64 {
			f := stepFreq(t, [2]float64{0, 1400}, [2]float64{0.14, 300})
			return o.pulse(f, 0.125)*decay(t, 20)*0.28 + n.noise(6000)*decay(t, 40)*0.12
		}
	}())

	s[SndKick] = mono(0.20, func() func(float64) float64 {
		var o, n osc
		return func(t float64) float64 {
			f := stepFreq(t, [2]float64{0, 900}, [2]float64{0.05, 260}, [2]float64{0.2, 200})
			return o.pulse(f, 0.25)*decay(t, 14)*0.34 + n.noise(4500)*decay(t, 30)*0.16
		}
	}())

	s[SndOneUp] = mono(0.70, func() func(float64) float64 {
		var o, h osc
		notes := []float64{659, 784, 1047, 1319, 1568, 2093}
		return func(t float64) float64 {
			i := int(t / 0.10)
			if i >= len(notes) {
				i = len(notes) - 1
			}
			f := notes[i]
			return (o.pulse(f, 0.5)*0.24 + h.pulse(f*1.5, 0.25)*0.10) * decay(math.Mod(t, 0.10), 9)
		}
	}())

	s[SndPause] = mono(0.16, func() func(float64) float64 {
		var o osc
		return func(t float64) float64 {
			f := 660.0
			if t > 0.08 {
				f = 440
			}
			return o.pulse(f, 0.5) * decay(math.Mod(t, 0.08), 16) * 0.28
		}
	}())

	s[SndMenuMove] = mono(0.07, func() func(float64) float64 {
		var o osc
		return func(t float64) float64 { return o.pulse(880, 0.25) * decay(t, 34) * 0.22 }
	}())

	s[SndMenuPick] = mono(0.24, func() func(float64) float64 {
		var o osc
		return func(t float64) float64 {
			f := 880.0
			if t > 0.06 {
				f = 1319
			}
			return o.pulse(f, 0.5) * decay(math.Mod(t, 0.06), 11) * 0.28
		}
	}())

	s[SndBossHit] = mono(0.32, func() func(float64) float64 {
		var o, n osc
		return func(t float64) float64 {
			f := stepFreq(t, [2]float64{0, 300}, [2]float64{0.32, 70})
			return o.saw(f)*decay(t, 9)*0.32 + n.noise(2600)*decay(t, 15)*0.24
		}
	}())

	s[SndBossDie] = mono(1.60, func() func(float64) float64 {
		var n, o osc
		return func(t float64) float64 {
			burst := 0.0
			// four staggered explosions
			for _, s := range []float64{0, 0.28, 0.62, 1.0} {
				if t >= s {
					burst += n.noise(stepFreq(t-s, [2]float64{0, 8000}, [2]float64{0.4, 900})) * decay(t-s, 6) * 0.16
				}
			}
			return burst + o.saw(stepFreq(t, [2]float64{0, 220}, [2]float64{1.6, 40}))*decay(t, 2.2)*0.16
		}
	}())

	s[SndPipe] = mono(0.42, func() func(float64) float64 {
		var o osc
		return func(t float64) float64 {
			f := stepFreq(t, [2]float64{0, 700}, [2]float64{0.42, 120})
			return o.pulse(f, 0.5) * decay(t, 5) * 0.28
		}
	}())

	s[SndSplash] = mono(0.36, func() func(float64) float64 {
		var n osc
		return func(t float64) float64 {
			return n.noise(stepFreq(t, [2]float64{0, 5200}, [2]float64{0.36, 1400})) * decay(t, 9) * 0.30
		}
	}())

	s[SndCheckpoint] = mono(0.55, func() func(float64) float64 {
		var o, h osc
		notes := []float64{523, 784, 1047}
		return func(t float64) float64 {
			i := int(t / 0.13)
			if i >= len(notes) {
				i = len(notes) - 1
			}
			f := notes[i]
			return (o.pulse(f, 0.5)*0.22 + h.tri(f*2)*0.12) * decay(math.Mod(t, 0.13), 7)
		}
	}())

	s[SndBlip] = mono(0.035, func() func(float64) float64 {
		var o osc
		return func(t float64) float64 { return o.pulse(1200, 0.125) * decay(t, 60) * 0.14 }
	}())

	s[SndUnlock] = mono(0.60, func() func(float64) float64 {
		var o, h osc
		notes := []float64{392, 523, 659, 784, 1047}
		return func(t float64) float64 {
			i := int(t / 0.11)
			if i >= len(notes) {
				i = len(notes) - 1
			}
			f := notes[i]
			return (o.pulse(f, 0.25)*0.22 + h.pulse(f*2, 0.5)*0.09) * decay(math.Mod(t, 0.11), 8)
		}
	}())

	s[SndRescue] = mono(1.00, func() func(float64) float64 {
		var o, h, b osc
		notes := []float64{523, 659, 784, 1047, 1319}
		return func(t float64) float64 {
			i := int(t / 0.14)
			if i >= len(notes) {
				i = len(notes) - 1
			}
			f := notes[i]
			shimmer := h.sine(f*3) * 0.06 * decay(math.Mod(t, 0.14), 5)
			return o.pulse(f, 0.5)*0.22*decay(math.Mod(t, 0.14), 5) + shimmer + b.tri(f/2)*0.12*decay(t, 2)
		}
	}())

	s[SndSpring] = mono(0.26, func() func(float64) float64 {
		var o osc
		return func(t float64) float64 {
			f := stepFreq(t, [2]float64{0, 200}, [2]float64{0.20, 1500})
			return o.tri(f) * decay(t, 7) * 0.34
		}
	}())

	s[SndSwim] = mono(0.20, func() func(float64) float64 {
		var n, o osc
		return func(t float64) float64 {
			return n.noise(stepFreq(t, [2]float64{0, 2600}, [2]float64{0.2, 900}))*decay(t, 14)*0.16 +
				o.sine(stepFreq(t, [2]float64{0, 300}, [2]float64{0.2, 620}))*decay(t, 12)*0.14
		}
	}())

	return s
}
