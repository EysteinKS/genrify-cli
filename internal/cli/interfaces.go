package cli

import (
	"context"

	"genrify/internal/spotify"
)

// SpotifyClient defines the subset of spotify.Client methods used by CLI commands.
// This interface enables testing without a real Spotify API connection.
type SpotifyClient interface {
	GetMe(ctx context.Context) (spotify.User, error)
	ListCurrentUserPlaylists(ctx context.Context, max int) ([]spotify.SimplifiedPlaylist, error)
	ListPlaylistTracks(ctx context.Context, playlistID string, max int) ([]spotify.FullTrack, error)
	CreatePlaylist(ctx context.Context, userID, name, description string, public bool) (spotify.SimplifiedPlaylist, error)
	AddTracksToPlaylist(ctx context.Context, playlistID string, uris []string) (string, error)
	GetPlaylist(ctx context.Context, playlistID string) (spotify.SimplifiedPlaylist, error)
	DeletePlaylist(ctx context.Context, playlistID string) error
}

// Prompter abstracts interactive prompts for testability.
// The default implementation uses promptui; tests can provide a mock.
type Prompter interface {
	// PromptString asks the user for a text input with optional default value.
	PromptString(label, defaultValue string) (string, error)

	// PromptInt asks the user for an integer input with optional default value.
	PromptInt(label string, defaultValue int) (int, error)

	// PromptSelect asks the user to choose from a list of options.
	PromptSelect(label string, items []string) (int, string, error)
}
