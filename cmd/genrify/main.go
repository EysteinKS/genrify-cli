package main

import (
	"runtime"

	"genrify/internal/cli"
)

func init() {
	// Lock the main goroutine to the OS thread.
	// This is required for GTK on macOS (Quartz backend) which requires
	// all GUI operations to happen on the main thread.
	// For CLI-only builds, this has no negative effect.
	runtime.LockOSThread()
}

func main() {
	cli.Execute()
}
