package cli

import (
	"reflect"
	"testing"

	"genrify/internal/spotify"
)

func TestJoinArtistNames(t *testing.T) {
	tests := []struct {
		name    string
		artists []spotify.Artist
		want    string
	}{
		{
			name:    "empty",
			artists: []spotify.Artist{},
			want:    "",
		},
		{
			name:    "single artist",
			artists: []spotify.Artist{{ID: "a1", Name: "Artist One"}},
			want:    "Artist One",
		},
		{
			name: "multiple artists",
			artists: []spotify.Artist{
				{ID: "a1", Name: "Artist One"},
				{ID: "a2", Name: "Artist Two"},
				{ID: "a3", Name: "Artist Three"},
			},
			want: "Artist One, Artist Two, Artist Three",
		},
		{
			name: "skip empty names",
			artists: []spotify.Artist{
				{ID: "a1", Name: "Artist One"},
				{ID: "a2", Name: ""},
				{ID: "a3", Name: "Artist Three"},
			},
			want: "Artist One, Artist Three",
		},
		{
			name: "all empty names",
			artists: []spotify.Artist{
				{ID: "a1", Name: ""},
				{ID: "a2", Name: ""},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinArtistNames(tt.artists)
			if got != tt.want {
				t.Errorf("joinArtistNames() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeTrackURI(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "spotify uri",
			input: "spotify:track:6rqhFgbbKwnb9MLmUQDhG6",
			want:  "spotify:track:6rqhFgbbKwnb9MLmUQDhG6",
		},
		{
			name:  "spotify uri mixed case",
			input: "Spotify:Track:6rqhFgbbKwnb9MLmUQDhG6",
			want:  "Spotify:Track:6rqhFgbbKwnb9MLmUQDhG6",
		},
		{
			name:  "open.spotify.com url",
			input: "https://open.spotify.com/track/6rqhFgbbKwnb9MLmUQDhG6",
			want:  "spotify:track:6rqhFgbbKwnb9MLmUQDhG6",
		},
		{
			name:  "open.spotify.com url with query params",
			input: "https://open.spotify.com/track/6rqhFgbbKwnb9MLmUQDhG6?si=abc123",
			want:  "spotify:track:6rqhFgbbKwnb9MLmUQDhG6",
		},
		{
			name:  "http url",
			input: "http://open.spotify.com/track/6rqhFgbbKwnb9MLmUQDhG6",
			want:  "spotify:track:6rqhFgbbKwnb9MLmUQDhG6",
		},
		{
			name:  "raw track id",
			input: "6rqhFgbbKwnb9MLmUQDhG6",
			want:  "spotify:track:6rqhFgbbKwnb9MLmUQDhG6",
		},
		{
			name:  "whitespace trimmed",
			input: "  6rqhFgbbKwnb9MLmUQDhG6  ",
			want:  "spotify:track:6rqhFgbbKwnb9MLmUQDhG6",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "unsupported url",
			input:   "https://example.com/track/123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeTrackURI(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("normalizeTrackURI() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("normalizeTrackURI() unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("normalizeTrackURI() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizePlaylistID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "spotify uri",
			input: "spotify:playlist:37i9dQZF1DXcBWIGoYBM5M",
			want:  "37i9dQZF1DXcBWIGoYBM5M",
		},
		{
			name:  "spotify uri mixed case",
			input: "Spotify:Playlist:37i9dQZF1DXcBWIGoYBM5M",
			want:  "37i9dQZF1DXcBWIGoYBM5M",
		},
		{
			name:  "open.spotify.com url",
			input: "https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M",
			want:  "37i9dQZF1DXcBWIGoYBM5M",
		},
		{
			name:  "open.spotify.com url with query params",
			input: "https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M?si=abc123",
			want:  "37i9dQZF1DXcBWIGoYBM5M",
		},
		{
			name:  "http url",
			input: "http://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M",
			want:  "37i9dQZF1DXcBWIGoYBM5M",
		},
		{
			name:  "raw playlist id",
			input: "37i9dQZF1DXcBWIGoYBM5M",
			want:  "37i9dQZF1DXcBWIGoYBM5M",
		},
		{
			name:  "whitespace trimmed",
			input: "  37i9dQZF1DXcBWIGoYBM5M  ",
			want:  "37i9dQZF1DXcBWIGoYBM5M",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "unsupported url",
			input:   "https://example.com/playlist/123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizePlaylistID(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("normalizePlaylistID() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("normalizePlaylistID() unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("normalizePlaylistID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		max   int
		want  string
	}{
		{
			name:  "no truncation needed",
			input: "short",
			max:   10,
			want:  "short",
		},
		{
			name:  "exact length",
			input: "exact",
			max:   5,
			want:  "exact",
		},
		{
			name:  "needs truncation",
			input: "this is a very long string",
			max:   10,
			want:  "this is...",
		},
		{
			name:  "very short max",
			input: "hello",
			max:   3,
			want:  "hel",
		},
		{
			name:  "max = 1",
			input: "hello",
			max:   1,
			want:  "h",
		},
		{
			name:  "max = 0",
			input: "hello",
			max:   0,
			want:  "",
		},
		{
			name:  "negative max",
			input: "hello",
			max:   -1,
			want:  "",
		},
		{
			name:  "empty string",
			input: "",
			max:   10,
			want:  "",
		},
		{
			name:  "unicode truncation",
			input: "Hello 世界",
			max:   8,
			want:  "Hello...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.max)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
			}
		})
	}
}

func TestFilterPlaylistsByName(t *testing.T) {
	playlists := []spotify.SimplifiedPlaylist{
		{ID: "p1", Name: "Rock Classics"},
		{ID: "p2", Name: "Jazz Favorites"},
		{ID: "p3", Name: "Classical Music"},
		{ID: "p4", Name: "Rock and Roll"},
	}

	tests := []struct {
		name   string
		filter string
		want   []string
	}{
		{
			name:   "no filter",
			filter: "",
			want:   []string{"p1", "p2", "p3", "p4"},
		},
		{
			name:   "whitespace filter",
			filter: "   ",
			want:   []string{"p1", "p2", "p3", "p4"},
		},
		{
			name:   "case insensitive match",
			filter: "rock",
			want:   []string{"p1", "p4"},
		},
		{
			name:   "exact match",
			filter: "Jazz Favorites",
			want:   []string{"p2"},
		},
		{
			name:   "partial match",
			filter: "class",
			want:   []string{"p1", "p3"},
		},
		{
			name:   "no match",
			filter: "pop",
			want:   []string{},
		},
		{
			name:   "case insensitive uppercase",
			filter: "ROCK",
			want:   []string{"p1", "p4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterPlaylistsByName(playlists, tt.filter)
			gotIDs := make([]string, len(got))
			for i, p := range got {
				gotIDs[i] = p.ID
			}
			if !reflect.DeepEqual(gotIDs, tt.want) {
				t.Errorf("filterPlaylistsByName() = %v, want %v", gotIDs, tt.want)
			}
		})
	}
}

func TestFormatPlaylistRow(t *testing.T) {
	p := spotify.SimplifiedPlaylist{
		ID:   "abc123",
		Name: "My Playlist",
		Owner: spotify.User{
			ID: "user123",
		},
	}
	p.Tracks.Total = 42

	want := "abc123\tMy Playlist\t42\tuser123"
	got := formatPlaylistRow(p)
	if got != want {
		t.Errorf("formatPlaylistRow() = %q, want %q", got, want)
	}
}

func TestFormatTrackRow(t *testing.T) {
	track := spotify.FullTrack{
		URI:  "spotify:track:123",
		Name: "Test Song",
		Artists: []spotify.Artist{
			{ID: "a1", Name: "Artist 1"},
			{ID: "a2", Name: "Artist 2"},
		},
	}

	want := "spotify:track:123\tTest Song\tArtist 1, Artist 2"
	got := formatTrackRow(track)
	if got != want {
		t.Errorf("formatTrackRow() = %q, want %q", got, want)
	}
}
