//go:build js

package save

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"syscall/js"
)

// Browsers have no filesystem, so the save document lives in localStorage.
// The path keeps the same shape as the native one and is shown verbatim in
// Options; storageKey strips the URI scheme before touching localStorage.
const webSavePath = "localStorage://SuperDario/save.json"

// webStore is a JS object whose methods wrap localStorage in try/catch. A
// browser can deny storage entirely (private mode, sandboxed iframe, some
// file:// configurations); that must become an error the game can report,
// not an exception that kills the wasm program.
var webStore js.Value

func init() {
	// Function.New avoids depending on any helper injected by index.html, so
	// the game still runs when the wasm is embedded somewhere else.
	factory := js.Global().Get("Function").New(`
		return {
			get: function (key) {
				try {
					var v = localStorage.getItem(key);
					if (v === null) return [false, "", ""];
					return [true, v, ""];
				} catch (e) {
					return [false, "", String(e)];
				}
			},
			set: function (key, value) {
				try {
					localStorage.setItem(key, value);
					return "";
				} catch (e) {
					return String(e);
				}
			},
			rename: function (from, to) {
				try {
					var v = localStorage.getItem(from);
					if (v === null) return "not found";
					localStorage.setItem(to, v);
					localStorage.removeItem(from);
					return "";
				} catch (e) {
					return String(e);
				}
			}
		};
	`)
	webStore = factory.Invoke()
}

func saveDir() (string, error) {
	return "localStorage://SuperDario", nil
}

func saveFilePath() (string, error) {
	return webSavePath, nil
}

func storageKey(p string) string {
	return "superdario:" + strings.TrimPrefix(p, "localStorage://")
}

func storageError(msg string) error {
	return fmt.Errorf("browser storage: %s", msg)
}

func readSaveFile(p string) ([]byte, error) {
	res := webStore.Call("get", storageKey(p))
	if res.Length() < 3 {
		return nil, errors.New("browser storage: unexpected result")
	}
	if msg := res.Index(2).String(); msg != "" {
		return nil, storageError(msg)
	}
	if !res.Index(0).Bool() {
		return nil, fs.ErrNotExist
	}
	return []byte(res.Index(1).String()), nil
}

func writeSaveFile(p string, data []byte) error {
	if msg := webStore.Call("set", storageKey(p), string(data)).String(); msg != "" {
		return storageError(msg)
	}
	return nil
}

func renameSaveFile(from, to string) error {
	if msg := webStore.Call("rename", storageKey(from), storageKey(to)).String(); msg != "" {
		return storageError(msg)
	}
	return nil
}
