//go:build !nogui

package gui

import (
	"github.com/gotk3/gotk3/glib"
)

// RunAsync executes a function in a goroutine and calls onDone on the GTK main thread.
// onStart is called immediately on the GTK thread before starting the goroutine.
// fn is executed in a background goroutine.
// onDone is called on the GTK thread with the result of fn.
func RunAsync[T any](onStart func(), fn func() (T, error), onDone func(T, error)) {
	if onStart != nil {
		onStart()
	}

	go func() {
		result, err := fn()
		glib.IdleAdd(func() bool {
			onDone(result, err)
			return false // Remove the idle callback after execution.
		})
	}()
}
