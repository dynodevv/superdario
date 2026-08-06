package sfx

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// Public audio API. Everything degrades gracefully: if the platform has no
// working audio device the game still runs, just silently.

const sfxPlayersPerSound = 4

type slot struct {
	id TrackID
	st *Stream
	pl *audio.Player
}

var (
	mu       sync.Mutex
	ctx      *audio.Context
	ready    bool
	disabled bool
	sndData  [][]byte
	sndPool  [][]*audio.Player
	sndNext  []int
	cur      slot
	fading   []slot
	musicVol = 0.7
	sfxVol   = 0.85
	stack    []TrackID // for temporary overlays such as the Safety Halo
)

// Disable turns audio off for the whole process. It must be called before
// Init. Development tools that run the engine uncapped use this — without it,
// thousands of frames per second would fire thousands of overlapping sounds.
func Disable() {
	mu.Lock()
	defer mu.Unlock()
	disabled = true
}

// Init creates the audio context and renders every sound effect.
func Init() {
	mu.Lock()
	defer mu.Unlock()
	if ready || disabled {
		return
	}
	defer func() {
		// A missing or busy audio device should never take the game down.
		if r := recover(); r != nil {
			ready = false
		}
	}()
	ctx = audio.NewContext(SampleRate)
	sndData = buildSounds()
	sndPool = make([][]*audio.Player, len(sndData))
	sndNext = make([]int, len(sndData))
	for i, data := range sndData {
		if len(data) == 0 {
			continue
		}
		for j := 0; j < sfxPlayersPerSound; j++ {
			p := ctx.NewPlayerF32FromBytes(data)
			sndPool[i] = append(sndPool[i], p)
		}
	}
	ready = true
}

func Ready() bool {
	mu.Lock()
	defer mu.Unlock()
	return ready
}

// SetMusicVolume / SetSFXVolume take 0..1.
func SetMusicVolume(v float64) {
	mu.Lock()
	defer mu.Unlock()
	musicVol = clamp01(v)
	if cur.pl != nil {
		cur.pl.SetVolume(musicVol)
	}
}

func SetSFXVolume(v float64) {
	mu.Lock()
	defer mu.Unlock()
	sfxVol = clamp01(v)
}

func MusicVolume() float64 {
	mu.Lock()
	defer mu.Unlock()
	return musicVol
}

func SFXVolume() float64 {
	mu.Lock()
	defer mu.Unlock()
	return sfxVol
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// Play fires a sound effect, rotating through that sound's player pool so
// rapid repeats overlap instead of cutting each other off.
func Play(id SoundID) {
	mu.Lock()
	defer mu.Unlock()
	if !ready || int(id) >= len(sndPool) || len(sndPool[id]) == 0 || sfxVol <= 0 {
		return
	}
	i := sndNext[id]
	sndNext[id] = (i + 1) % len(sndPool[id])
	p := sndPool[id][i]
	if err := p.Rewind(); err != nil {
		return
	}
	p.SetVolume(sfxVol)
	p.Play()
}

// PlayMusic cross-fades to a track. Calling it with the already-playing track
// is a no-op, so scenes can call it unconditionally every frame.
func PlayMusic(id TrackID) {
	mu.Lock()
	defer mu.Unlock()
	playMusicLocked(id)
}

func playMusicLocked(id TrackID) {
	if !ready {
		cur.id = id
		return
	}
	if cur.id == id && cur.pl != nil && cur.pl.IsPlaying() {
		return
	}
	t, ok := tracks[id]
	if !ok {
		stopMusicLocked()
		cur.id = MusNone
		return
	}
	// retire the current track
	if cur.pl != nil {
		cur.st.FadeOut()
		fading = append(fading, cur)
	}
	st := NewStream(t)
	pl, err := ctx.NewPlayerF32(st)
	if err != nil {
		cur = slot{id: MusNone}
		return
	}
	pl.SetVolume(musicVol)
	// A large buffer would add latency to music switches; keep it modest.
	pl.Play()
	cur = slot{id: id, st: st, pl: pl}
}

// StopMusic fades out whatever is playing.
func StopMusic() {
	mu.Lock()
	defer mu.Unlock()
	stopMusicLocked()
}

func stopMusicLocked() {
	if cur.pl != nil {
		cur.st.FadeOut()
		fading = append(fading, cur)
	}
	cur = slot{id: MusNone}
}

// PushMusic swaps in a temporary track (the Safety Halo theme) remembering
// what was playing so PopMusic can restore it.
func PushMusic(id TrackID) {
	mu.Lock()
	defer mu.Unlock()
	stack = append(stack, cur.id)
	playMusicLocked(id)
}

// PopMusic returns to the track that was playing before the last PushMusic.
func PopMusic() {
	mu.Lock()
	defer mu.Unlock()
	if len(stack) == 0 {
		return
	}
	prev := stack[len(stack)-1]
	stack = stack[:len(stack)-1]
	playMusicLocked(prev)
}

// Current reports the track that is playing.
func Current() TrackID {
	mu.Lock()
	defer mu.Unlock()
	return cur.id
}

// MusicFinished is true once a one-shot track (a fanfare) has played through.
func MusicFinished() bool {
	mu.Lock()
	defer mu.Unlock()
	if cur.st == nil {
		return true
	}
	return cur.st.Done()
}

// PauseMusic / ResumeMusic are used by the pause menu.
func PauseMusic() {
	mu.Lock()
	defer mu.Unlock()
	if cur.pl != nil {
		cur.pl.Pause()
	}
}

func ResumeMusic() {
	mu.Lock()
	defer mu.Unlock()
	if cur.pl != nil && !cur.pl.IsPlaying() {
		cur.pl.Play()
	}
}

// Duck lowers the music temporarily (used under fanfares and cutscene text).
func Duck(f float64) {
	mu.Lock()
	defer mu.Unlock()
	if cur.pl != nil {
		cur.pl.SetVolume(musicVol * clamp01(f))
	}
}

// Update must be called once per frame; it reaps finished players.
func Update() {
	mu.Lock()
	defer mu.Unlock()
	if !ready {
		return
	}
	keep := fading[:0]
	for _, s := range fading {
		if s.st.Done() {
			s.pl.Pause()
			_ = s.pl.Close()
			continue
		}
		keep = append(keep, s)
	}
	fading = keep
}
