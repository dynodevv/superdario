<#
.SYNOPSIS
    Builds Super Dario.

.DESCRIPTION
    Windows builds are pure Go and need no C toolchain, so this script can
    produce the Windows executable anywhere.

    macOS and Linux builds need cgo (Ebitengine binds to Cocoa/Metal and to
    X11/ALSA respectively), which means they have to be built on those
    platforms — or in CI. Pass -Targets linux only if you have a working
    Linux cross toolchain; otherwise use build.sh on the target machine, or
    push and let .github/workflows/build.yml do it.

.EXAMPLE
    .\build.ps1
    .\build.ps1 -Version 1.1.0 -Targets windows,linux
#>
param(
    [string]   $Version = "1.0.0",
    [string[]] $Targets = @("windows"),
    [switch]   $SkipTests
)

$ErrorActionPreference = "Stop"

# Go is not always on PATH on a fresh Windows box.
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    $candidates = @(
        "C:\Program Files\Go\bin",
        "$env:LOCALAPPDATA\Programs\Go\bin",
        "C:\Go\bin"
    )
    foreach ($c in $candidates) {
        if (Test-Path (Join-Path $c "go.exe")) { $env:PATH = "$c;$env:PATH"; break }
    }
}
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "Go was not found. Install it from https://go.dev/dl/ and re-run."
}

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

Write-Host "Super Dario $Version" -ForegroundColor Cyan
Write-Host "go: $((go version))"

if (-not $SkipTests) {
    Write-Host "`n== vet ==" -ForegroundColor Yellow
    go vet ./...
    Write-Host "`n== test ==" -ForegroundColor Yellow
    go test ./...
}

$ldflags    = "-s -w -X main.Version=$Version"
# -H=windowsgui: no console window flashes when the exe is double-clicked
$ldflagsWin = "-s -w -H=windowsgui -X main.Version=$Version"
$pkg     = "./cmd/superdario"

foreach ($t in $Targets) {
    switch ($t) {
        "windows" {
            $out = "dist\windows-amd64"
            New-Item -ItemType Directory -Force $out | Out-Null
            Write-Host "`n== building windows/amd64 ==" -ForegroundColor Yellow
            $env:CGO_ENABLED = "0"; $env:GOOS = "windows"; $env:GOARCH = "amd64"
            go build -trimpath -ldflags $ldflagsWin -o "$out\SuperDario.exe" $pkg
            $exe = Get-Item "$out\SuperDario.exe"
            Write-Host ("  {0}  {1:N1} MB" -f $exe.FullName, ($exe.Length / 1MB)) -ForegroundColor Green
        }
        "linux" {
            $out = "dist\linux-amd64"
            New-Item -ItemType Directory -Force $out | Out-Null
            Write-Host "`n== building linux/amd64 (needs a cgo cross toolchain) ==" -ForegroundColor Yellow
            $env:CGO_ENABLED = "1"; $env:GOOS = "linux"; $env:GOARCH = "amd64"
            go build -trimpath -ldflags $ldflags -o "$out\SuperDario" $pkg
        }
        "darwin" {
            Write-Host "`nmacOS builds cannot be produced from Windows: Ebitengine needs cgo" -ForegroundColor Red
            Write-Host "against the macOS SDK. Run ./build.sh on a Mac, or use the GitHub"  -ForegroundColor Red
            Write-Host "Actions workflow in .github/workflows/build.yml."                    -ForegroundColor Red
        }
        default { Write-Host "unknown target: $t" -ForegroundColor Red }
    }
}

Remove-Item Env:CGO_ENABLED, Env:GOOS, Env:GOARCH -ErrorAction SilentlyContinue
Write-Host "`nDone." -ForegroundColor Cyan
