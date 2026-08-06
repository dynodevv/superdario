//go:build windows

package main

import (
	"os"
	"syscall"
)

// The Windows executable is linked with -H=windowsgui so that double-clicking
// it does not flash a console window. That also means the process starts with
// no standard output at all, so the informational flags would print into the
// void.
//
// attachConsole borrows the parent terminal's console when the game was
// launched from one, and rebinds os.Stdout/os.Stderr to it. Launched from
// Explorer there is no parent console, the call fails harmlessly, and the game
// carries on silently.
func attachConsole() {
	const attachParentProcess = ^uintptr(0) // (DWORD)-1

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	attach := kernel32.NewProc("AttachConsole")
	if r, _, _ := attach.Call(attachParentProcess); r == 0 {
		return // no parent console; nothing to attach to
	}

	if h, err := syscall.Open("CONOUT$", syscall.O_RDWR, 0); err == nil {
		os.Stdout = os.NewFile(uintptr(h), "CONOUT$")
		os.Stderr = os.NewFile(uintptr(h), "CONOUT$")
	}
}
