//go:build !nogui

package gui

import (
	"context"

	"genrify/internal/spotify"
)

// SpotifyClient defines the Spotify API operations needed by the GUI.
type SpotifyClient interface {
	GetMe(ctx context.Context) (spotify.User, error)
	ListCurrentUserPlaylists(ctx context.Context, max int) ([]spotify.SimplifiedPlaylist, error)
	ListPlaylistTracks(ctx context.Context, playlistID string, max int) ([]spotify.FullTrack, error)
	CreatePlaylist(ctx context.Context, userID, name, description string, public bool) (spotify.SimplifiedPlaylist, error)
	AddTracksToPlaylist(ctx context.Context, playlistID string, uris []string) (string, error)
	GetPlaylist(ctx context.Context, playlistID string) (spotify.SimplifiedPlaylist, error)
	DeletePlaylist(ctx context.Context, playlistID string) error
}
