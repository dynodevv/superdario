// Command superdario is the game's entry point.
//
// Super Dario — a 2D platformer in which Dario Amodei rescues the Claude
// models from the Distillation Syndicate. Everything (art, music, levels) is
// generated in code, so the game ships as a single self-contained binary.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"superdario/internal/core"
	"superdario/internal/game"
	"superdario/internal/gfx"
	"superdario/internal/save"
	"superdario/internal/sfx"
)

// Version is stamped by the build scripts via -ldflags.
var Version = "dev"

func main() {
	var (
		showVersion = flag.Bool("version", false, "print the version and exit")
		fullscreen  = flag.Bool("fullscreen", false, "start in fullscreen")
		scale       = flag.Int("scale", 0, "window scale factor (1-6); 0 uses the saved setting")
		vsync       = flag.Bool("vsync", true, "wait for vertical sync")
		savePath    = flag.Bool("save-path", false, "print the save file location and exit")
		mute        = flag.Bool("mute", false, "start with no audio at all")
		frames      = flag.Int("frames", 0, "quit after N frames (smoke test)")
	)
	flag.Parse()

	// Any flag that prints needs a console to print to.
	if *showVersion || *savePath {
		attachConsole()
	}

	if *mute {
		sfx.Disable()
	}

	if *showVersion {
		fmt.Printf("Super Dario %s (%s/%s, %s)\n", Version, runtime.GOOS, runtime.GOARCH, runtime.Version())
		return
	}
	if *savePath {
		p, err := save.Path()
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not resolve save path:", err)
			os.Exit(1)
		}
		fmt.Println(p)
		return
	}

	g := game.New()

	sc := g.Save.Settings.WindowScale
	if *scale > 0 {
		sc = *scale
	}
	if sc < 1 || sc > 6 {
		sc = 3
	}

	ebiten.SetWindowTitle("Super Dario — Save the Models")
	// The icon is drawn in code, so it works identically on every platform
	// and needs no file alongside the binary.
	ebiten.SetWindowIcon(gfx.AppIcons())
	ebiten.SetWindowSize(core.ScreenW*sc, core.ScreenH*sc)
	ebiten.SetWindowSizeLimits(core.ScreenW, core.ScreenH, -1, -1)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(core.TPS)
	ebiten.SetVsyncEnabled(*vsync)
	ebiten.SetScreenClearedEveryFrame(true)

	web := runtime.GOOS == "js"
	if web {
		// In a browser the canvas fills the window, so the game is usually
		// scaled by a fractional amount. Nearest-neighbour keeps the pixels
		// crisp instead of blurring them through the box filter.
		ebiten.SetScreenFilterEnabled(false)
	}
	// Browsers only grant fullscreen from a user gesture, so applying a saved
	// fullscreen setting at boot cannot work there. The Options screen can
	// still toggle it during play.
	if (*fullscreen || g.Save.Settings.Fullscreen) && !web {
		ebiten.SetFullscreen(true)
	}

	var runnable ebiten.Game = g
	if *frames > 0 {
		runnable = &limited{Game: g, left: *frames}
	}
	if err := ebiten.RunGame(runnable); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}

// limited wraps the game so -frames can end it after a fixed number of ticks.
// Used to smoke-test that the whole stack starts up on a given machine.
type limited struct {
	ebiten.Game
	left int
}

func (l *limited) Update() error {
	l.left--
	if l.left <= 0 {
		return ebiten.Termination
	}
	return l.Game.Update()
}
