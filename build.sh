#!/usr/bin/env bash
# Build Super Dario on macOS or Linux.
#
# Ebitengine needs cgo on both platforms, so run this on the machine you are
# building for. The Windows executable is pure Go and cross-compiles from here
# too (./build.sh windows). The web target (./build.sh web) is pure Go as well
# and produces one self-contained HTML file.
set -euo pipefail

VERSION="${VERSION:-1.0.0}"
PKG="./cmd/superdario"
LDFLAGS="-s -w -X main.Version=${VERSION}"
# -H=windowsgui stops a console window flashing on double-click
LDFLAGS_WIN="-s -w -H=windowsgui -X main.Version=${VERSION}"
DIST="dist"

cd "$(dirname "$0")"

have() { command -v "$1" >/dev/null 2>&1; }

if ! have go; then
  echo "Go was not found on PATH. Install it from https://go.dev/dl/" >&2
  exit 1
fi

echo "Super Dario ${VERSION}"
go version

# --- dependency hints -------------------------------------------------------
check_linux_deps() {
  local missing=()
  if have pkg-config; then
    for p in gl x11 xrandr xcursor xinerama xi alsa; do
      pkg-config --exists "$p" 2>/dev/null || missing+=("$p")
    done
  fi
  if [ ${#missing[@]} -gt 0 ]; then
    cat >&2 <<EOF

Missing development headers: ${missing[*]}

Ebitengine links against OpenGL, X11 and ALSA. Install them first:

  Debian/Ubuntu:
    sudo apt install libgl1-mesa-dev xorg-dev libasound2-dev

  Fedora:
    sudo dnf install mesa-libGL-devel libX11-devel libXrandr-devel \\
                     libXcursor-devel libXinerama-devel libXi-devel alsa-lib-devel

  Arch:
    sudo pacman -S mesa libx11 libxrandr libxcursor libxinerama libxi alsa-lib

EOF
    exit 1
  fi
}

build_windows() {
  echo "== windows/amd64 =="
  mkdir -p "${DIST}/windows-amd64"
  CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
    go build -trimpath -ldflags "${LDFLAGS_WIN}" -o "${DIST}/windows-amd64/SuperDario.exe" "${PKG}"
  echo "  -> ${DIST}/windows-amd64/SuperDario.exe"
}

build_linux() {
  check_linux_deps
  local arch="${1:-$(go env GOARCH)}"
  echo "== linux/${arch} =="
  mkdir -p "${DIST}/linux-${arch}"
  CGO_ENABLED=1 GOOS=linux GOARCH="${arch}" \
    go build -trimpath -ldflags "${LDFLAGS}" -o "${DIST}/linux-${arch}/SuperDario" "${PKG}"
  chmod +x "${DIST}/linux-${arch}/SuperDario"
  echo "  -> ${DIST}/linux-${arch}/SuperDario"
}

build_web() {
  # Pure Go: no cgo, no platform headers, works from any host. buildweb runs
  # on the host, so clear any GOOS/GOARCH the caller exported for other
  # targets (buildweb sets js/wasm for the child build itself).
  echo "== web (single HTML) =="
  env -u GOOS -u GOARCH -u CGO_ENABLED go run ./cmd/buildweb -version "${VERSION}"
}

build_darwin() {
  # Apple Silicon only — Intel Macs are not a supported target.
  echo "== darwin/arm64 =="
  mkdir -p "${DIST}/darwin"
  CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
    go build -trimpath -ldflags "${LDFLAGS}" -o "${DIST}/darwin/SuperDario" "${PKG}"
  chmod +x "${DIST}/darwin/SuperDario"

  ./scripts/make-app-bundle.sh "${DIST}/darwin/SuperDario" "${DIST}/darwin/SuperDario.app" "${VERSION}"
  echo "  -> ${DIST}/darwin/SuperDario.app"
  cat <<'EOF'

Note: the .app is unsigned. On first launch macOS will refuse to open it.
Right-click the app and choose Open, then confirm — or run:

  xattr -dr com.apple.quarantine dist/darwin/SuperDario.app

EOF
}

targets=("$@")
if [ ${#targets[@]} -eq 0 ]; then
  case "$(go env GOOS)" in
    darwin) targets=(test darwin) ;;
    linux)  targets=(test linux) ;;
    *)      targets=(test windows) ;;
  esac
fi

for t in "${targets[@]}"; do
  case "$t" in
    test)    echo "== vet & test =="; go vet ./...; go test ./... ;;
    windows) build_windows ;;
    linux)   build_linux ;;
    darwin|mac|macos) build_darwin ;;
    web)     build_web ;;
    all)     build_windows; [ "$(go env GOOS)" = darwin ] && build_darwin || build_linux ;;
    *)       echo "unknown target: $t" >&2; exit 1 ;;
  esac
done

echo "Done."
