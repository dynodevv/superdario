//go:build !windows

package main

// Everywhere except Windows the process already owns its standard streams.
func attachConsole() {}
