package save

import (
	"os"
	"path/filepath"
	"testing"
)

func tempFile(t *testing.T) *File {
	t.Helper()
	f := &File{Version: FileVersion, Settings: DefaultSettings()}
	f.path = filepath.Join(t.TempDir(), "save.json")
	return f
}

func TestSaveLoadRoundTrip(t *testing.T) {
	f := tempFile(t)
	f.Slots[1] = NewSlot()
	f.Slots[1].Score = 12345
	f.Slots[1].Tokens = 99
	f.Slots[1].Rescued[2] = true
	f.Slots[1].Unlock(0, 0)
	f.Settings.MusicVolume = 0.42
	if err := f.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	g := LoadPath(f.path)
	if err := g.Err(); err != nil {
		t.Fatalf("load reported: %v", err)
	}
	if g.Slots[1].Score != 12345 || g.Slots[1].Tokens != 99 {
		t.Errorf("slot did not round-trip: %+v", g.Slots[1])
	}
	if !g.Slots[1].Rescued[2] {
		t.Error("rescued flag lost")
	}
	if g.Settings.MusicVolume != 0.42 {
		t.Errorf("music volume = %v, want 0.42", g.Settings.MusicVolume)
	}
	if !g.Slots[1].Cleared[0][0] {
		t.Error("cleared flag lost")
	}
}

func TestLoadMissingFileIsClean(t *testing.T) {
	f := LoadPath(filepath.Join(t.TempDir(), "nope.json"))
	if err := f.Err(); err != nil {
		t.Errorf("a missing save should not be an error, got %v", err)
	}
	if f.Slots[0].Used {
		t.Error("missing save should produce empty slots")
	}
}

func TestLoadCorruptFileIsSetAside(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "save.json")
	if err := os.WriteFile(p, []byte("{not json at all"), 0o644); err != nil {
		t.Fatal(err)
	}
	f := LoadPath(p)
	if f.Err() == nil {
		t.Error("corrupt save should report an error")
	}
	if _, err := os.Stat(p + ".corrupt"); err != nil {
		t.Errorf("corrupt save should be preserved alongside: %v", err)
	}
	if f.Settings.MusicVolume != DefaultSettings().MusicVolume {
		t.Error("corrupt save should fall back to default settings")
	}
}

func TestUnlockProgression(t *testing.T) {
	var s Slot
	if s.IsUnlocked(0, 1) {
		t.Error("level 1-2 should start locked")
	}
	s.Unlock(0, 0)
	if !s.IsUnlocked(0, 1) {
		t.Error("clearing 1-1 should unlock 1-2")
	}
	s.Unlock(0, 1)
	s.Unlock(0, 2)
	if !s.IsUnlocked(1, 0) {
		t.Error("clearing the last level of world 1 should unlock world 2")
	}
	// Going back and re-clearing an early level must not roll the frontier back.
	s.Unlock(0, 0)
	if s.World != 1 || s.Level != 0 {
		t.Errorf("frontier moved backwards: world=%d level=%d", s.World, s.Level)
	}
	// Clearing the very last level finishes the game.
	s.World, s.Level = NumWorlds-1, NumLevels-1
	s.Unlock(NumWorlds-1, NumLevels-1)
	if !s.Finished {
		t.Error("clearing the final level should set Finished")
	}
}

func TestUnlockRejectsOutOfRange(t *testing.T) {
	var s Slot
	s.Unlock(-1, 0)
	s.Unlock(99, 0)
	s.Unlock(0, 99)
	if s.World != 0 || s.Level != 0 || s.Finished {
		t.Errorf("out-of-range unlock changed state: %+v", s)
	}
	if s.IsUnlocked(99, 99) {
		t.Error("out-of-range level reported unlocked")
	}
}

func TestRecordTimeKeepsBest(t *testing.T) {
	var s Slot
	if !s.RecordTime(0, 0, 3600) {
		t.Error("first time should be recorded")
	}
	if s.RecordTime(0, 0, 4000) {
		t.Error("slower time should not replace the best")
	}
	if !s.RecordTime(0, 0, 3000) {
		t.Error("faster time should replace the best")
	}
	if s.BestTime[0][0] != 3000 {
		t.Errorf("best = %d, want 3000", s.BestTime[0][0])
	}
}

func TestMigrateClampsGarbage(t *testing.T) {
	f := &File{}
	f.Settings.MusicVolume = 99
	f.Settings.SFXVolume = -5
	f.Settings.WindowScale = 0
	f.Slots[0].World = 40
	f.Slots[0].Level = -3
	f.Slots[0].Lives = -2
	f.Slots[0].Power = 7
	f.migrate()
	s := f.Slots[0]
	d := DefaultSettings()
	if f.Settings.MusicVolume != d.MusicVolume || f.Settings.SFXVolume != d.SFXVolume || f.Settings.WindowScale != d.WindowScale {
		t.Errorf("settings not clamped to defaults: %+v", f.Settings)
	}
	if s.World != NumWorlds-1 || s.Level != 0 || s.Lives != 0 || s.Power != 0 {
		t.Errorf("slot not clamped: %+v", s)
	}
}

func TestCompletion(t *testing.T) {
	var s Slot
	if s.Completion() != 0 {
		t.Error("fresh slot should be 0% complete")
	}
	for w := 0; w < NumWorlds; w++ {
		for l := 0; l < NumLevels; l++ {
			s.Cleared[w][l] = true
		}
	}
	if s.Completion() != 1 {
		t.Errorf("all cleared should be 100%%, got %v", s.Completion())
	}
}

func TestSaveIsAtomic(t *testing.T) {
	f := tempFile(t)
	f.Slots[0] = NewSlot()
	f.Slots[0].Score = 1
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}
	f.Slots[0].Score = 2
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}
	// No stray temp files should survive a successful save.
	entries, err := os.ReadDir(filepath.Dir(f.path))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("leftover temp file: %s", e.Name())
		}
	}
}
