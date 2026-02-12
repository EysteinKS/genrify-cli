package helpers

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"genrify/internal/spotify"
)

var (
	openTrackURLRe       = regexp.MustCompile(`(?i)^https?://open\.spotify\.com/track/([A-Za-z0-9]+)(?:\?.*)?$`)
	openPlaylistURLRe    = regexp.MustCompile(`(?i)^https?://open\.spotify\.com/playlist/([A-Za-z0-9]+)(?:\?.*)?$`)
	spotifyPlaylistURIRe = regexp.MustCompile(`(?i)^spotify:playlist:([A-Za-z0-9]+)$`)
)

// JoinArtistNames joins artist names into a comma-separated string.
func JoinArtistNames(artists []spotify.Artist) string {
	n := make([]string, 0, len(artists))
	for _, a := range artists {
		if a.Name != "" {
			n = append(n, a.Name)
		}
	}
	return strings.Join(n, ", ")
}

// NormalizeTrackURI converts a track ID, URI, or URL to a Spotify track URI.
func NormalizeTrackURI(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("empty track value")
	}
	if strings.HasPrefix(strings.ToLower(s), "spotify:track:") {
		return s, nil
	}
	if m := openTrackURLRe.FindStringSubmatch(s); len(m) == 2 {
		return "spotify:track:" + m[1], nil
	}
	if u, err := url.Parse(s); err == nil && u.Scheme != "" {
		return "", fmt.Errorf("unsupported track url: %s", s)
	}
	// Treat as raw track id.
	return "spotify:track:" + s, nil
}

// NormalizePlaylistID converts a playlist ID, URI, or URL to a Spotify playlist ID.
func NormalizePlaylistID(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("empty playlist id")
	}
	if m := spotifyPlaylistURIRe.FindStringSubmatch(s); len(m) == 2 {
		return m[1], nil
	}
	if m := openPlaylistURLRe.FindStringSubmatch(s); len(m) == 2 {
		return m[1], nil
	}
	if u, err := url.Parse(s); err == nil && u.Scheme != "" {
		return "", fmt.Errorf("unsupported playlist url: %s", s)
	}
	// Treat as raw playlist id.
	return s, nil
}

// FilterPlaylistsByName filters playlists by name (case-insensitive substring match).
func FilterPlaylistsByName(playlists []spotify.SimplifiedPlaylist, filter string) []spotify.SimplifiedPlaylist {
	want := strings.ToLower(strings.TrimSpace(filter))
	if want == "" {
		return playlists
	}
	out := make([]spotify.SimplifiedPlaylist, 0, len(playlists))
	for _, p := range playlists {
		if strings.Contains(strings.ToLower(p.Name), want) {
			out = append(out, p)
		}
	}
	return out
}
