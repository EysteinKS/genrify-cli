package cli

import (
	"fmt"

	"genrify/internal/helpers"
	"genrify/internal/spotify"
)

// Exported wrappers for backwards compatibility.
var (
	JoinArtistNames      = helpers.JoinArtistNames
	NormalizeTrackURI    = helpers.NormalizeTrackURI
	NormalizePlaylistID  = helpers.NormalizePlaylistID
	FilterPlaylistsByName = helpers.FilterPlaylistsByName
)

// Local helper functions.
func joinArtistNames(artists []spotify.Artist) string {
	return helpers.JoinArtistNames(artists)
}

func normalizeTrackURI(s string) (string, error) {
	return helpers.NormalizeTrackURI(s)
}

func normalizePlaylistID(s string) (string, error) {
	return helpers.NormalizePlaylistID(s)
}

func filterPlaylistsByName(playlists []spotify.SimplifiedPlaylist, filter string) []spotify.SimplifiedPlaylist {
	return helpers.FilterPlaylistsByName(playlists, filter)
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func formatPlaylistRow(p spotify.SimplifiedPlaylist) string {
	// Keep output stable between interactive and command mode.
	return fmt.Sprintf("%s\t%s\t%d\t%s", p.ID, p.Name, p.Tracks.Total, p.Owner.ID)
}

func formatTrackRow(t spotify.FullTrack) string {
	return fmt.Sprintf("%s\t%s\t%s", t.URI, t.Name, joinArtistNames(t.Artists))
}
