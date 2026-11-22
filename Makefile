# Super Dario — build targets.
#
# Ebitengine is pure Go on Windows, but needs cgo on macOS and Linux, so those
# targets must be built on (or cross-compiled toward) their own platform.
# See README.md and .github/workflows/build.yml for the CI matrix.

VERSION ?= 1.0.0
PKG     := ./cmd/superdario
LDFLAGS := -s -w -X main.Version=$(VERSION)
# -H=windowsgui stops a console window flashing on double-click
LDFLAGS_WIN := -s -w -H=windowsgui -X main.Version=$(VERSION)
DIST    := dist

.PHONY: all run test vet fmt clean windows linux darwin darwin-universal web check

all: test windows

run:
	go run $(PKG)

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

check: fmt vet test

clean:
	rm -rf $(DIST)

# --- Windows (pure Go; cross-compiles from anywhere) -----------------------
windows:
	mkdir -p $(DIST)/windows-amd64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
		go build -trimpath -ldflags "$(LDFLAGS_WIN)" -o $(DIST)/windows-amd64/SuperDario.exe $(PKG)

# --- Linux (needs cgo + X11/ALSA headers; build on Linux) ------------------
linux:
	mkdir -p $(DIST)/linux-amd64
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
		go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST)/linux-amd64/SuperDario $(PKG)

# --- macOS, Apple Silicon only (needs cgo + Xcode command line tools) ------
darwin:
	mkdir -p $(DIST)/darwin
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
		go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST)/darwin/SuperDario $(PKG)

# A double-clickable .app bundle around it.
darwin-app: darwin
	./scripts/make-app-bundle.sh $(DIST)/darwin/SuperDario $(DIST)/darwin/SuperDario.app $(VERSION)

# --- Web (one self-contained HTML file; pure Go js/wasm) -------------------
# The build tool itself must run on the host, not for the js/wasm target, so
# strip any cross-compilation environment the caller might have exported.
web:
	env -u GOOS -u GOARCH -u CGO_ENABLED go run ./cmd/buildweb -version $(VERSION)

# --- icons (regenerate after changing internal/gfx/icon.go) ----------------
icons:
	go run ./cmd/genicon .
