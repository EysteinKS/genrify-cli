package cli

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

func joinArtistNames(artists []spotify.Artist) string {
	n := make([]string, 0, len(artists))
	for _, a := range artists {
		if a.Name != "" {
			n = append(n, a.Name)
		}
	}
	return strings.Join(n, ", ")
}

func normalizeTrackURI(s string) (string, error) {
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

func normalizePlaylistID(s string) (string, error) {
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

func filterPlaylistsByName(playlists []spotify.SimplifiedPlaylist, filter string) []spotify.SimplifiedPlaylist {
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

func formatPlaylistRow(p spotify.SimplifiedPlaylist) string {
	// Keep output stable between interactive and command mode.
	return fmt.Sprintf("%s\t%s\t%d\t%s", p.ID, p.Name, p.Tracks.Total, p.Owner.ID)
}

func formatTrackRow(t spotify.FullTrack) string {
	return fmt.Sprintf("%s\t%s\t%s", t.URI, t.Name, joinArtistNames(t.Artists))
}
