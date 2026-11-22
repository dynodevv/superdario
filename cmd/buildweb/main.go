// Command buildweb builds Super Dario for the browser as a single HTML file.
//
// The Go WebAssembly binary and Go's wasm_exec.js runtime are embedded in the
// page, so the result runs from one file with no server, no sidecar assets and
// no network access.
//
// Run it from the repository root:
//
//	go run ./cmd/buildweb
//	go run ./cmd/buildweb -version 1.1.0 -out dist/web/SuperDario.html
//
// The web build stores saves in browser localStorage rather than on disk; see
// internal/save/storage_web.go.
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	version := flag.String("version", "1.0.0", "version stamped into the build")
	out := flag.String("out", filepath.Join("dist", "web", "SuperDario.html"), "output HTML file")
	flag.Parse()

	if err := run(*version, *out); err != nil {
		fmt.Fprintln(os.Stderr, "buildweb:", err)
		os.Exit(1)
	}
}

func run(version, outPath string) error {
	if strings.ContainsAny(version, " \t\r\n\"") {
		return errors.New("version must not contain whitespace or quotes")
	}

	const (
		templatePath = "web/index.html"
		packagePath  = "./cmd/superdario"
	)
	if _, err := os.Stat(templatePath); err != nil {
		return fmt.Errorf("reading %s: %w (run this from the repository root)", templatePath, err)
	}

	wasmExecPath, err := findWasmExec()
	if err != nil {
		return err
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		// Running via `go run` means go is normally on PATH; fall back to the
		// toolchain that compiled this program.
		goBin = filepath.Join(runtime.GOROOT(), "bin", "go")
		if _, statErr := os.Stat(goBin); statErr != nil {
			return errors.New("the go command was not found on PATH")
		}
	}

	tmpDir, err := os.MkdirTemp("", "superdario-web-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	wasmPath := filepath.Join(tmpDir, "superdario.wasm")

	fmt.Printf("Super Dario %s — building js/wasm…\n", version)
	args := []string{
		"build",
		"-trimpath",
		"-buildvcs=false",
		"-ldflags", "-s -w -X main.Version=" + version,
		"-o", wasmPath,
		packagePath,
	}
	cmd := exec.Command(goBin, args...)
	cmd.Env = withEnv(os.Environ(), "GOOS=js", "GOARCH=wasm", "CGO_ENABLED=0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build: %w", err)
	}

	wasm, err := os.ReadFile(wasmPath)
	if err != nil {
		return err
	}
	var gz bytes.Buffer
	zw, err := gzip.NewWriterLevel(&gz, gzip.BestCompression)
	if err != nil {
		return err
	}
	if _, err := zw.Write(wasm); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}

	wasmExec, err := os.ReadFile(wasmExecPath)
	if err != nil {
		return err
	}
	if bytes.Contains(bytes.ToLower(wasmExec), []byte("</script")) {
		return fmt.Errorf("%s cannot be inlined: it contains a closing script tag", wasmExecPath)
	}
	tmpl, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}

	// Fail loudly if the template and this tool have drifted apart, rather
	// than silently producing a page with an unexpanded placeholder.
	for _, token := range []string{"__WASM_EXEC_JS__", "__WASM_GZIP_BASE64__", "__VERSION__"} {
		if n := strings.Count(string(tmpl), token); n != 1 {
			return fmt.Errorf("%s: expected exactly one %s placeholder, found %d", templatePath, token, n)
		}
	}

	page := string(tmpl)
	page = strings.Replace(page, "__VERSION__", html.EscapeString(version), 1)
	page = strings.Replace(page, "__WASM_EXEC_JS__", string(wasmExec), 1)
	page = strings.Replace(page, "__WASM_GZIP_BASE64__", base64.StdEncoding.EncodeToString(gz.Bytes()), 1)

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(outPath, []byte(page), 0o644); err != nil {
		return err
	}

	abs, err := filepath.Abs(outPath)
	if err != nil {
		abs = outPath
	}
	fmt.Printf("  wasm   %s (%s gzip)\n", humanBytes(len(wasm)), humanBytes(gz.Len()))
	fmt.Printf("  page   %s (%s, self-contained)\n", abs, humanBytes(len(page)))
	return nil
}

// findWasmExec locates the wasm_exec.js that matches the Go toolchain. Go
// moved it from misc/wasm to lib/wasm in Go 1.24.
func findWasmExec() (string, error) {
	goroot := runtime.GOROOT()
	candidates := []string{
		filepath.Join(goroot, "lib", "wasm", "wasm_exec.js"),
		filepath.Join(goroot, "misc", "wasm", "wasm_exec.js"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("wasm_exec.js not found under %s; is this a complete Go installation?", goroot)
}

// withEnv replaces any existing values for the given keys, so a caller's
// GOOS/GOARCH (for example while cross-compiling) cannot leak into the
// browser build.
func withEnv(env []string, kv ...string) []string {
	out := make([]string, 0, len(env)+len(kv))
	for _, entry := range env {
		keep := true
		for i := 0; i < len(kv); i += 2 {
			if strings.HasPrefix(entry, kv[i]+"=") {
				keep = false
				break
			}
		}
		if keep {
			out = append(out, entry)
		}
	}
	return append(out, kv...)
}

func humanBytes(n int) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := int64(n) / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGT"[exp])
}
