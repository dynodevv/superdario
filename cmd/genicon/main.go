// Command genicon renders the application icon and writes the per-platform
// icon bundles plus the Windows resource object that gives SuperDario.exe its
// icon in Explorer.
//
//	go run ./cmd/genicon
//
// Outputs:
//
//	assets/icon.ico                       Windows / Linux
//	assets/icon.icns                      macOS .app bundle
//	assets/icon_<size>.png                Linux desktop entries
//	cmd/superdario/resource_windows.syso  linked into the exe automatically
//
// The art itself lives in internal/gfx/icon.go and is drawn in code, so this
// tool only handles container formats.
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/josephspurrier/goversioninfo"
	"superdario/internal/gfx"
)

const version = "1.0.0"

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	assets := filepath.Join(root, "assets")
	if err := os.MkdirAll(assets, 0o755); err != nil {
		fail(err)
	}

	// --- render every size once -------------------------------------------
	imgs := map[int]*image.RGBA{}
	pngs := map[int][]byte{}
	for _, s := range gfx.IconSizes {
		img := gfx.AppIcon(s)
		imgs[s] = img
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			fail(err)
		}
		pngs[s] = buf.Bytes()
	}

	// --- standalone PNGs (Linux) -------------------------------------------
	for _, s := range []int{32, 64, 128, 256, 512} {
		p := filepath.Join(assets, fmt.Sprintf("icon_%d.png", s))
		if err := os.WriteFile(p, pngs[s], 0o644); err != nil {
			fail(err)
		}
	}
	fmt.Println("wrote", filepath.Join(assets, "icon_*.png"))

	// --- .ico ---------------------------------------------------------------
	icoSizes := []int{16, 24, 32, 48, 64, 128, 256}
	ico := buildICO(icoSizes, pngs)
	icoPath := filepath.Join(assets, "icon.ico")
	if err := os.WriteFile(icoPath, ico, 0o644); err != nil {
		fail(err)
	}
	fmt.Println("wrote", icoPath, len(ico), "bytes")

	// --- .icns --------------------------------------------------------------
	icns := buildICNS(pngs)
	icnsPath := filepath.Join(assets, "icon.icns")
	if err := os.WriteFile(icnsPath, icns, 0o644); err != nil {
		fail(err)
	}
	fmt.Println("wrote", icnsPath, len(icns), "bytes")

	// --- Windows resource object -------------------------------------------
	writeSyso(root, icoPath)
}

// ---------------------------------------------------------------------------
// ICO
//
// Layout: a 6-byte header, then one 16-byte directory entry per image, then
// the image payloads. Windows Vista and later accept PNG payloads at every
// size, which avoids having to emit BMP with its upside-down rows and AND mask.
// ---------------------------------------------------------------------------

func buildICO(sizes []int, pngs map[int][]byte) []byte {
	var buf bytes.Buffer
	w := func(v any) { binary.Write(&buf, binary.LittleEndian, v) }

	w(uint16(0))          // reserved
	w(uint16(1))          // type: 1 = icon
	w(uint16(len(sizes))) // image count

	offset := 6 + 16*len(sizes)
	for _, s := range sizes {
		data := pngs[s]
		dim := byte(s)
		if s >= 256 {
			dim = 0 // 0 means 256 in the ICO directory
		}
		w(dim)        // width
		w(dim)        // height
		w(byte(0))    // palette size (0 = truecolour)
		w(byte(0))    // reserved
		w(uint16(1))  // colour planes
		w(uint16(32)) // bits per pixel
		w(uint32(len(data)))
		w(uint32(offset))
		offset += len(data)
	}
	for _, s := range sizes {
		buf.Write(pngs[s])
	}
	return buf.Bytes()
}

// ---------------------------------------------------------------------------
// ICNS
//
// A magic word, a total length, then a sequence of typed chunks. The modern
// PNG-based types are enough for a .app bundle icon.
// ---------------------------------------------------------------------------

func buildICNS(pngs map[int][]byte) []byte {
	// OSType -> pixel size
	chunks := []struct {
		typ  string
		size int
	}{
		{"icp4", 16},
		{"icp5", 32},
		{"icp6", 64},
		{"ic07", 128},
		{"ic08", 256},
		{"ic09", 512},
		{"ic11", 32},  // 16x16@2x
		{"ic12", 64},  // 32x32@2x
		{"ic13", 256}, // 128x128@2x
		{"ic14", 512}, // 256x256@2x
	}

	var body bytes.Buffer
	for _, ch := range chunks {
		data, ok := pngs[ch.size]
		if !ok {
			continue
		}
		body.WriteString(ch.typ)
		binary.Write(&body, binary.BigEndian, uint32(8+len(data)))
		body.Write(data)
	}

	var out bytes.Buffer
	out.WriteString("icns")
	binary.Write(&out, binary.BigEndian, uint32(8+body.Len()))
	out.Write(body.Bytes())
	return out.Bytes()
}

// ---------------------------------------------------------------------------
// Windows resource
// ---------------------------------------------------------------------------

func writeSyso(root, icoPath string) {
	vi := &goversioninfo.VersionInfo{
		FixedFileInfo: goversioninfo.FixedFileInfo{
			FileVersion:    goversioninfo.FileVersion{Major: 1, Minor: 0, Patch: 0, Build: 0},
			ProductVersion: goversioninfo.FileVersion{Major: 1, Minor: 0, Patch: 0, Build: 0},
			FileFlagsMask:  "3f",
			FileOS:         "040004",
			FileType:       "01", // application
			FileSubType:    "00",
		},
		StringFileInfo: goversioninfo.StringFileInfo{
			ProductName:      "Super Dario",
			FileDescription:  "Super Dario — Save the Models",
			InternalName:     "SuperDario",
			OriginalFilename: "SuperDario.exe",
			ProductVersion:   version,
			FileVersion:      version,
			LegalCopyright:   "A fan game. Not affiliated with any real company.",
			Comments:         "Parody. All companies and people referenced are fictionalised.",
		},
		VarFileInfo: goversioninfo.VarFileInfo{
			Translation: goversioninfo.Translation{LangID: 0x0409, CharsetID: 0x04B0},
		},
		IconPath: icoPath,
	}
	vi.Build()
	vi.Walk()

	// The .syso must sit beside the main package; the Go tool links any
	// GOOS-suffixed .syso in the package directory automatically.
	out := filepath.Join(root, "cmd", "superdario", "resource_windows_amd64.syso")
	if err := vi.WriteSyso(out, "amd64"); err != nil {
		fail(fmt.Errorf("writing %s: %w", out, err))
	}
	fmt.Println("wrote", out)

	out386 := filepath.Join(root, "cmd", "superdario", "resource_windows_arm64.syso")
	if err := vi.WriteSyso(out386, "arm64"); err != nil {
		fmt.Println("note: arm64 resource skipped:", err)
		return
	}
	fmt.Println("wrote", out386)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "genicon:", err)
	os.Exit(1)
}
