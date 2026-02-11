package cli

import (
	"errors"
	"fmt"
)

// Common CLI errors that can be checked with errors.Is.

var (
	// ErrNotLoggedIn indicates the user needs to run `genrify login`.
	ErrNotLoggedIn = errors.New("not logged in; run genrify login")

	// ErrInvalidInput indicates user-provided input was malformed.
	ErrInvalidInput = errors.New("invalid input")

	// ErrCancelled indicates the user cancelled an interactive prompt.
	ErrCancelled = errors.New("cancelled")
)

// WrapLoginError wraps errors from the Spotify client/token manager
// to detect "not logged in" conditions and return ErrNotLoggedIn.
func WrapLoginError(err error) error {
	if err == nil {
		return nil
	}
	// Token manager returns text-based errors mentioning "not logged in".
	msg := err.Error()
	if contains(msg, "not logged in") || contains(msg, "missing token") {
		return fmt.Errorf("%w: %v", ErrNotLoggedIn, err)
	}
	return err
}

func contains(s, substr string) bool {
	// Simple substring check; could use strings.Contains but keep it minimal.
	return len(s) >= len(substr) && (s == substr || indexSubstring(s, substr) >= 0)
}

func indexSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
