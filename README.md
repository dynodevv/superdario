# Super Dario

> ### ⚠️ This is a parody. Please read this bit.
>
> **Super Dario is an unofficial fan game and a work of parody and satire.**
>
> It is **not affiliated with, endorsed by, sponsored by, or connected to**
> Anthropic, DeepSeek, Moonshot AI, Alibaba, Nintendo, or any other company —
> nor to any real person.
>
> "Dario", the "Claude" models, and the cartoon "Syndicate" bosses are
> **fictional characters** in a comedy platformer. Nothing here depicts real
> events, real products, real conduct, or the real views of anyone. The bosses
> are cartoon machines with treads and top hats; they are not portrayals of any
> living individual, and no claim is made or implied about anything any real
> company or person has actually done.
>
> No logos, trademarks, artwork, audio, or code belonging to any of these
> companies are used — **every asset in this game is generated from scratch by
> its own source code.** Names are used the way a parody uses names, as
> commentary.
>
> No money is made from this. If you own a name used here and would like it
> changed, open an issue and it will be changed.

---

A 2D platformer written in Go with [Ebitengine](https://ebitengine.org).

Four (fictional) Claude models — **Haiku**, **Sonnet**, **Opus** and **Fable** —
have been copied, crated and shipped out to five rival labs calling themselves
the **Distillation Syndicate**. Dario takes the stairs.

Five worlds, fifteen stages, six boss fights, and a soundtrack that does not
exist until the moment you run the game.

---

### Default controls

| Action | Keyboard | Xbox pad |
|---|---|---|
| Move | Arrows / WASD | D-pad or left stick |
| Jump | Space, Z | A |
| Run & throw | Shift, X | X or RT |
| Duck | Down / S | D-pad down |
| Pause | Esc | Menu |
| Confirm / Back | Enter / Esc | A / B |
| Options (on the map) | P | View |

Everything above is **rebindable** in **Options → Controls**, which has a
keyboard tab and a gamepad tab. The gamepad tab draws a controller that lights
up as you press buttons, so you can check a pad is working before you play.

Every on-screen hint follows the device you are actually using and the bindings
you actually have — plug a controller in and the title screen changes from
`SPACE / Z JUMP` to `A JUMP` on the spot, and rebinding updates the hints
immediately.

Hold **run** while jumping to clear the wider gaps — a running jump goes
noticeably higher. Releasing jump early cuts the arc short.

### Power states

| Pickup | State | Effect |
|---|---|---|
| **Compute Cluster** | Scaled | Bigger, takes one extra hit, breaks bricks |
| **Constitution Scroll** | Aligned | Throws Alignment Bolts |
| **Safety Halo** | — | Eleven seconds of invincibility |
| **Snapshot** | — | Extra life |
| **Context Tokens** | — | 100 tokens = one extra life |

Damage steps you down one tier at a time: Aligned → Scaled → Small → out.

### The Syndicate

Every boss has a **vulnerable window** — it stops, bobs, and shows a marker
overhead. Stomp it then. Alignment Bolts work whenever the boss is out, but no
fight *requires* them, so arriving without the Scroll is never a dead end.

---

## Progress saving

Three save slots. Progress is written after every stage, on every level select,
and when you quit to the map.

| OS | Path |
|---|---|
| Windows | `%APPDATA%\SuperDario\save.json` |
| macOS | `~/Library/Application Support/SuperDario/save.json` |
| Linux | `~/.config/SuperDario/save.json` |

Writes are atomic (temp file + rename), so a crash mid-save cannot destroy an
existing file. Unreadable JSON is moved aside to `save.json.corrupt` rather
than silently overwritten. Control bindings are stored by *name*, so a save
survives an engine update that renumbers the key constants.

---

## Building

### Icons

The icon is drawn in code (`internal/gfx/icon.go`). Regenerate the bundles and
the Windows resource object after changing it:

```bash
go run ./cmd/genicon .
```

That writes `assets/icon.ico`, `assets/icon.icns`, `assets/icon_*.png`, and the
`cmd/superdario/resource_windows_*.syso` files that give `SuperDario.exe` its
icon in Explorer. The generated files are **committed**, so a plain `go build`
produces a properly branded binary with no extra step. CI regenerates them and
fails if they have gone stale, so the committed bundles cannot drift from the
art. Generation is deterministic — the same source always yields byte-identical
output.

### Windows

Pure Go — no C toolchain, and it cross-compiles from anywhere.

```powershell
.\build.ps1                       # test + build dist\windows-amd64\SuperDario.exe
.\build.ps1 -Version 1.1.0
```

The exe is linked with `-H=windowsgui` so double-clicking it does not flash a
console window. `-version` and `-save-path` still work from a terminal — the
game borrows the parent console — but because it is a GUI binary your shell
returns immediately, so the output appears after the prompt and cannot be
piped. Launched from Explorer there is no console and nothing is printed.

### macOS

**Apple Silicon only.** Intel Macs are not a supported target.

Ebitengine binds to Cocoa and Metal, so macOS builds need cgo **and must be
built on a Mac** (`xcode-select --install`).

```bash
./build.sh darwin                 # arm64 binary and a .app bundle
```

The bundle is unsigned, so the first launch needs a right-click → Open, or:

```bash
xattr -dr com.apple.quarantine dist/darwin/SuperDario.app
```

### Linux

Needs cgo plus OpenGL, X11 and ALSA headers:

```bash
sudo apt install libgl1-mesa-dev xorg-dev libasound2-dev   # Debian/Ubuntu
./build.sh linux
```

`build.sh` checks for those headers up front and prints the right install
command for your distro if any are missing.

### All platforms at once

`.github/workflows/build.yml` builds Windows, Linux and macOS (arm64) on every
push and uploads them as artifacts. That is the easiest way to get a Mac build
without owning a Mac.

The test job runs under `xvfb`: importing Ebitengine runs an `init()` that
opens the display, so even tests that never draw a frame need an X server on a
headless runner. That job also regenerates the icons and fails if the committed
bundles have gone stale.

### Command-line flags

```
-fullscreen      start fullscreen
-scale N         window scale 1-6 (default: saved setting)
-mute            start with no audio at all
-frames N        quit after N frames (smoke test)
-vsync=false     uncap the frame rate
-version         print version and exit
-save-path       print the save file location and exit
```

Music and sound default to 55% / 70% on a fresh install and are adjustable in
Options. If you ever run the engine uncapped, mute it — thousands of frames per
second means thousands of overlapping sounds per second.

---

## How it is put together

Apart from the generated icon bundles, there are no asset files. Every sprite,
tile, backdrop, sound effect and note of music is produced by code when the
game starts, which is why it ships as one self-contained binary per platform.

```
cmd/superdario     the game
cmd/genicon        build tool: render the icon and the Windows resource
cmd/artdump        dev tool: dump the generated sprite sheets to PNG
cmd/shots          dev tool: drive the game through its screens and capture PNGs

internal/core      screen size, tile size, physics constants
internal/gfx       palette, ASCII-art sprite decoder, 5x7 bitmap font,
                   procedural tilesets, parallax backdrops, particles, bosses,
                   the app icon, and the drawn Xbox controller
internal/sfx       chiptune synth, tracker sequencer, the soundtrack, SFX
internal/save      slots, settings, bindings, atomic JSON persistence
internal/world     tiles, chunk-assembled levels, physics, enemies, bosses
internal/game      scene stack, input bindings, HUD, menus, cutscenes
```

### Art

Characters and enemies are ASCII art in the source, one character per pixel
indexing a shared palette (`internal/gfx/palette.go`):

```go
"..bbbdddddbbb...",
"..bb13011301d...",   // 'b' hair, 'd' skin, '1' glasses frame, '3' lens, '0' pupil
```

Dario's "Aligned" form is the same art run through a three-entry palette
substitution that turns his navy blazer coral, so there is one set of frames
rather than two.

Tiles, backdrops, bosses, the icon and the controller diagram are drawn with
primitives instead — five themed tilesets fall out of five `Theme` structs.

### Music

A small tracker. Each lane is a list of bars, each bar sixteen sixteenth-notes:

```go
"e5 - - . g5 - - . c6 - - - b5 - - -",   // '-' holds, '.' rests
```

Lanes loop independently at their own length, so a one-bar drum pattern repeats
under an eight-bar melody. Twelve tracks, rendered sample by sample on the audio
thread through pulse/triangle/saw/noise oscillators with ADSR envelopes, a
stereo delay and a soft clipper.

### Levels

Stages are sequences of hand-authored ASCII chunks (`internal/world/chunks.go`)
bottom-aligned against the floor:

```
"...==B==?==.....",
"......o.o.......",
"....g.......g...",
"################",
```

---

## Tests

```bash
go test ./...
```

The suite is mostly about catching content mistakes that are invisible until
you play into them:

- every music bar is exactly sixteen tokens and every note parses
- every track and sound effect is audible, finite, and not clipping
- every chunk is rectangular, uses known characters, pairs its pipe halves, and
  never puts a piranha plant somewhere without a pipe under it
- **no stage contains a step-up taller than a running jump** with nothing to
  help — the check that catches a wall you cannot get over
- every boss arena is enterable on foot, and every spring has headroom
- all fifteen stages build, have a spawn with ground under it, and an exit
- an autopilot holds right for a simulated minute through every stage and must
  stay on camera, make real progress, and never end up inside a solid tile
- every boss spawns at its gate and reaches a stompable state within 30s
- saves round-trip, unlock progression only moves forward, and a corrupt save
  is set aside rather than overwritten

---

## Credits

**Dyno** — design, direction & production
**Claude Opus 5** — code, pixels & soundtrack
