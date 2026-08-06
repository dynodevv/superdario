// Package save persists progress and settings to the user's config directory.
// The layout is deliberately boring JSON so a save survives game updates.
package save

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	FileVersion = 1
	NumSlots    = 3
	NumWorlds   = 5
	NumLevels   = 3 // per world
)

// Slot is one save file's worth of progress.
type Slot struct {
	Used     bool      `json:"used"`
	Created  time.Time `json:"created"`
	Played   time.Time `json:"lastPlayed"`
	Seconds  int       `json:"playSeconds"`
	World    int       `json:"world"` // furthest world unlocked, 0-based
	Level    int       `json:"level"` // furthest level unlocked within it
	Lives    int       `json:"lives"`
	Score    int       `json:"score"`
	Tokens   int       `json:"tokens"`
	Deaths   int       `json:"deaths"`
	Rescued  [4]bool   `json:"rescued"` // Haiku, Sonnet, Fable, Opus
	Finished bool      `json:"finished"`
	// Cleared[world][level] — used to draw ticks on the world map.
	Cleared [NumWorlds][NumLevels]bool `json:"cleared"`
	// BestTime[world][level] in ticks; 0 means "never cleared".
	BestTime [NumWorlds][NumLevels]int `json:"bestTime"`
	// Power state Dario carries between levels (0 small, 1 scaled, 2 aligned).
	Power int `json:"power"`
}

// NewSlot returns a fresh game.
func NewSlot() Slot {
	return Slot{
		Used:    true,
		Created: time.Now(),
		Played:  time.Now(),
		Lives:   3,
		World:   0,
		Level:   0,
	}
}

// Settings are shared across all slots.
type Settings struct {
	MusicVolume float64 `json:"musicVolume"`
	SFXVolume   float64 `json:"sfxVolume"`
	Fullscreen  bool    `json:"fullscreen"`
	WindowScale int     `json:"windowScale"`
	ShowTimer   bool    `json:"showTimer"`
	Scanlines   bool    `json:"scanlines"`

	// Control bindings, stored by name (action -> list of key/button names)
	// so a save survives an engine update that renumbers the constants.
	// Empty means "use the defaults".
	KeyBinds map[string][]string `json:"keyBinds,omitempty"`
	PadBinds map[string][]string `json:"padBinds,omitempty"`
}

func DefaultSettings() Settings {
	return Settings{
		// Conservative on first launch; both are adjustable in Options.
		MusicVolume: 0.55,
		SFXVolume:   0.70,
		WindowScale: 3,
		ShowTimer:   true,
		Scanlines:   true,
	}
}

// File is the whole on-disk document.
type File struct {
	Version  int            `json:"version"`
	Settings Settings       `json:"settings"`
	Slots    [NumSlots]Slot `json:"slots"`
	Extra    map[string]any `json:"extra,omitempty"` // forward compatibility

	path    string     `json:"-"`
	mu      sync.Mutex `json:"-"`
	lastErr error      `json:"-"`
}

// Dir returns the directory the save lives in, creating it if needed.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		// Fall back to the executable's folder — better a portable save than
		// no save at all.
		exe, eerr := os.Executable()
		if eerr != nil {
			return "", err
		}
		base = filepath.Dir(exe)
	}
	dir := filepath.Join(base, "SuperDario")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// Path is the full save-file path.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "save.json"), nil
}

// Load reads the save file from the user's config directory, returning a
// usable default document if none exists or if the existing one is unreadable.
func Load() *File {
	p, err := Path()
	if err != nil {
		return &File{Version: FileVersion, Settings: DefaultSettings()}
	}
	return LoadPath(p)
}

// LoadPath reads a save from a specific location.
func LoadPath(p string) *File {
	f := &File{Version: FileVersion, Settings: DefaultSettings(), path: p}

	data, err := os.ReadFile(p)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			f.lastErr = fmt.Errorf("reading save: %w", err)
		}
		return f
	}
	var in File
	if err := json.Unmarshal(data, &in); err != nil {
		// Keep the corrupt file around rather than silently overwriting it.
		_ = os.Rename(p, p+".corrupt")
		f.lastErr = fmt.Errorf("save file was unreadable and has been set aside: %w", err)
		return f
	}
	f.Version = in.Version
	f.Settings = in.Settings
	f.Slots = in.Slots
	f.Extra = in.Extra
	f.migrate()
	return f
}

// migrate normalises anything an older or hand-edited file might have wrong.
func (f *File) migrate() {
	d := DefaultSettings()
	if f.Settings.MusicVolume < 0 || f.Settings.MusicVolume > 1 {
		f.Settings.MusicVolume = d.MusicVolume
	}
	if f.Settings.SFXVolume < 0 || f.Settings.SFXVolume > 1 {
		f.Settings.SFXVolume = d.SFXVolume
	}
	if f.Settings.WindowScale < 1 || f.Settings.WindowScale > 8 {
		f.Settings.WindowScale = d.WindowScale
	}
	for i := range f.Slots {
		s := &f.Slots[i]
		if s.Lives < 0 {
			s.Lives = 0
		}
		if s.World < 0 {
			s.World = 0
		}
		if s.World >= NumWorlds {
			s.World = NumWorlds - 1
		}
		if s.Level < 0 {
			s.Level = 0
		}
		if s.Level >= NumLevels {
			s.Level = NumLevels - 1
		}
		if s.Power < 0 || s.Power > 2 {
			s.Power = 0
		}
	}
	f.Version = FileVersion
}

// Err returns the last load/save problem, if any. Save failures are
// non-fatal — the game keeps running and the UI reports them.
func (f *File) Err() error { return f.lastErr }

// Save writes the document atomically: a temp file in the same directory,
// then a rename, so an interrupted write cannot destroy an existing save.
func (f *File) Save() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.path == "" {
		p, err := Path()
		if err != nil {
			return err
		}
		f.path = p
	}
	f.Version = FileVersion
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(f.path), "save-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once the rename succeeds

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	// Windows will not rename onto an existing file, so clear the way first.
	_ = os.Remove(f.path)
	if err := os.Rename(tmpName, f.path); err != nil {
		return err
	}
	f.lastErr = nil
	return nil
}

// SaveQuiet writes and records (rather than returns) any failure, for call
// sites in the game loop where there is nothing sensible to do with an error.
func (f *File) SaveQuiet() {
	if err := f.Save(); err != nil {
		f.lastErr = err
	}
}

// --- slot helpers -----------------------------------------------------------

// Unlock records that a level was cleared and opens the next one.
func (s *Slot) Unlock(world, level int) {
	if world < 0 || world >= NumWorlds || level < 0 || level >= NumLevels {
		return
	}
	s.Cleared[world][level] = true
	nw, nl := world, level+1
	if nl >= NumLevels {
		nw, nl = world+1, 0
	}
	if nw >= NumWorlds {
		s.Finished = true
		return
	}
	// only ever move the frontier forward
	if nw > s.World || (nw == s.World && nl > s.Level) {
		s.World, s.Level = nw, nl
	}
}

// IsUnlocked reports whether the player may enter a level.
func (s *Slot) IsUnlocked(world, level int) bool {
	if world < 0 || world >= NumWorlds || level < 0 || level >= NumLevels {
		return false
	}
	if s.Cleared[world][level] {
		return true
	}
	if world < s.World {
		return true
	}
	return world == s.World && level <= s.Level
}

// RecordTime keeps the best (lowest) clear time for a level.
func (s *Slot) RecordTime(world, level, ticks int) bool {
	if world < 0 || world >= NumWorlds || level < 0 || level >= NumLevels || ticks <= 0 {
		return false
	}
	if s.BestTime[world][level] == 0 || ticks < s.BestTime[world][level] {
		s.BestTime[world][level] = ticks
		return true
	}
	return false
}

// Completion is the fraction of levels cleared, for the file-select screen.
func (s *Slot) Completion() float64 {
	n := 0
	for w := 0; w < NumWorlds; w++ {
		for l := 0; l < NumLevels; l++ {
			if s.Cleared[w][l] {
				n++
			}
		}
	}
	return float64(n) / float64(NumWorlds*NumLevels)
}

// PlayTime renders the slot's play time as H:MM.
func (s *Slot) PlayTime() string {
	h := s.Seconds / 3600
	m := (s.Seconds % 3600) / 60
	return fmt.Sprintf("%d:%02d", h, m)
}
