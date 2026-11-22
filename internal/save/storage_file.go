//go:build !js

package save

import (
	"os"
	"path/filepath"
)

// saveDir is the per-user config directory, created if needed. If the
// platform has no config directory, the executable's folder is used as a
// portable fallback — better a save next to the game than no save at all.
func saveDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
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

// saveFilePath is the full save-file path.
func saveFilePath() (string, error) {
	dir, err := saveDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "save.json"), nil
}

func readSaveFile(p string) ([]byte, error) {
	return os.ReadFile(p)
}

// writeSaveFile writes atomically: a temp file in the same directory, then a
// rename, so an interrupted write cannot destroy an existing save.
func writeSaveFile(p string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(p), "save-*.tmp")
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
	_ = os.Remove(p)
	return os.Rename(tmpName, p)
}

func renameSaveFile(from, to string) error {
	return os.Rename(from, to)
}
