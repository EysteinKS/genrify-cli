package cli

import "time"

// Default values and constants used across CLI commands.

const (
	// DefaultPlaylistLimit is the default maximum playlists to display.
	DefaultPlaylistLimit = 50

	// DefaultTrackLimit is the default maximum tracks to display.
	DefaultTrackLimit = 100

	// DefaultHTTPTimeout is the default timeout for Spotify API requests.
	DefaultHTTPTimeout = 30 * time.Second

	// LoginTimeout is the maximum time to wait for OAuth login.
	LoginTimeout = 2 * time.Minute

	// TokenRefreshLeeway is how early to refresh tokens before expiry.
	TokenRefreshLeeway = 60 * time.Second
)
