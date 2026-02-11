package playlist

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"genrify/internal/spotify"
)

var (
	// ErrNoPlaylistsMatched indicates no playlists matched the provided pattern.
	ErrNoPlaylistsMatched = errors.New("no playlists matched pattern")
)

// Client defines the subset of Spotify operations needed by the playlist service.
type Client interface {
	GetMe(ctx context.Context) (spotify.User, error)
	ListCurrentUserPlaylists(ctx context.Context, max int) ([]spotify.SimplifiedPlaylist, error)
	ListPlaylistTracks(ctx context.Context, playlistID string, max int) ([]spotify.FullTrack, error)
	CreatePlaylist(ctx context.Context, userID, name, description string, public bool) (spotify.SimplifiedPlaylist, error)
	AddTracksToPlaylist(ctx context.Context, playlistID string, uris []string) (string, error)
	GetPlaylist(ctx context.Context, playlistID string) (spotify.SimplifiedPlaylist, error)
	DeletePlaylist(ctx context.Context, playlistID string) error
}

type Service struct {
	c Client
}

func NewService(c Client) *Service {
	return &Service{c: c}
}

func (s *Service) FindPlaylistsByPattern(ctx context.Context, pattern string) ([]spotify.SimplifiedPlaylist, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return nil, fmt.Errorf("pattern is required")
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern: %w", err)
	}

	pls, err := s.c.ListCurrentUserPlaylists(ctx, 0)
	if err != nil {
		return nil, fmt.Errorf("list playlists: %w", err)
	}

	out := make([]spotify.SimplifiedPlaylist, 0)
	for _, p := range pls {
		if re.MatchString(p.Name) {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil, ErrNoPlaylistsMatched
	}
	return out, nil
}

func (s *Service) MergePlaylists(ctx context.Context, sourceIDs []string, targetName string, opts MergeOptions) (*MergeResult, error) {
	targetName = strings.TrimSpace(targetName)
	if targetName == "" {
		return nil, fmt.Errorf("target name is required")
	}
	if len(sourceIDs) == 0 {
		return nil, fmt.Errorf("at least one source playlist is required")
	}

	// Collect tracks first so we can fail early without creating anything.
	uris := make([]string, 0)
	for _, id := range sourceIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		tracks, err := s.c.ListPlaylistTracks(ctx, id, 0)
		if err != nil {
			return nil, fmt.Errorf("list tracks for %s: %w", id, err)
		}
		for _, t := range tracks {
			if t.URI != "" {
				uris = append(uris, t.URI)
			}
		}
	}

	duplicatesRemoved := 0
	if opts.Deduplicate {
		var kept []string
		kept, duplicatesRemoved = deduplicate(uris)
		uris = kept
	}

	pl, err := s.c.CreatePlaylist(ctx, "", targetName, opts.Description, opts.Public)
	if err != nil {
		return nil, fmt.Errorf("create playlist: %w", err)
	}

	// Best-effort rollback if we fail after creating the playlist.
	rollback := func() {
		_ = s.c.DeletePlaylist(ctx, pl.ID)
	}

	if len(uris) > 0 {
		if _, err := s.c.AddTracksToPlaylist(ctx, pl.ID, uris); err != nil {
			rollback()
			return nil, fmt.Errorf("add tracks: %w", err)
		}
	}

	ok, missing, err := s.VerifyPlaylistContents(ctx, pl.ID, uris)
	if err != nil {
		rollback()
		return nil, err
	}

	return &MergeResult{
		NewPlaylistID:     pl.ID,
		TrackCount:        len(uris),
		DuplicatesRemoved: duplicatesRemoved,
		Verified:          ok,
		MissingURIs:       missing,
	}, nil
}

func (s *Service) VerifyPlaylistContents(ctx context.Context, playlistID string, expectedURIs []string) (bool, []string, error) {
	playlistID = strings.TrimSpace(playlistID)
	if playlistID == "" {
		return false, nil, fmt.Errorf("playlist id is required")
	}

	expected := make(map[string]struct{}, len(expectedURIs))
	for _, u := range expectedURIs {
		u = strings.TrimSpace(u)
		if u != "" {
			expected[u] = struct{}{}
		}
	}
	if len(expected) == 0 {
		return true, nil, nil
	}

	// Spotify can be eventually consistent; retry a couple times.
	for attempt := 0; attempt < 3; attempt++ {
		tracks, err := s.c.ListPlaylistTracks(ctx, playlistID, 0)
		if err != nil {
			return false, nil, fmt.Errorf("list playlist tracks: %w", err)
		}
		seen := make(map[string]struct{}, len(tracks))
		for _, t := range tracks {
			if t.URI != "" {
				seen[t.URI] = struct{}{}
			}
		}

		missing := make([]string, 0)
		for u := range expected {
			if _, ok := seen[u]; !ok {
				missing = append(missing, u)
			}
		}
		if len(missing) == 0 {
			return true, nil, nil
		}
		if attempt < 2 {
			if err := sleepContext(ctx, 200*time.Millisecond); err != nil {
				return false, nil, err
			}
			continue
		}
		return false, missing, nil
	}

	return false, nil, nil
}

func (s *Service) DeletePlaylists(ctx context.Context, playlistIDs []string) error {
	for _, id := range playlistIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if err := s.c.DeletePlaylist(ctx, id); err != nil {
			var apiErr spotify.APIError
			if errors.As(err, &apiErr) && apiErr.Status == 403 {
				return fmt.Errorf("delete playlist %s: permission denied", id)
			}
			return fmt.Errorf("delete playlist %s: %w", id, err)
		}
	}
	return nil
}

func deduplicate(in []string) ([]string, int) {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	dupes := 0
	for _, u := range in {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		if _, ok := seen[u]; ok {
			dupes++
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out, dupes
}

func sleepContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
