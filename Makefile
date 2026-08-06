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

.PHONY: all run test vet fmt clean windows linux darwin darwin-universal check

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

# --- macOS (needs cgo + Xcode command line tools; build on macOS) ----------
darwin:
	mkdir -p $(DIST)/darwin
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
		go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST)/darwin/SuperDario-arm64 $(PKG)
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
		go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST)/darwin/SuperDario-amd64 $(PKG)

# A universal binary plus a double-clickable .app bundle.
darwin-universal: darwin
	lipo -create -output $(DIST)/darwin/SuperDario \
		$(DIST)/darwin/SuperDario-arm64 $(DIST)/darwin/SuperDario-amd64
	./scripts/make-app-bundle.sh $(DIST)/darwin/SuperDario $(DIST)/darwin/SuperDario.app $(VERSION)
