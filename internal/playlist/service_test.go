package playlist

import (
	"context"
	"errors"
	"strings"
	"testing"

	"genrify/internal/spotify"
)

type fakeClient struct {
	me        spotify.User
	playlists []spotify.SimplifiedPlaylist
	tracks    map[string][]spotify.FullTrack

	created spotify.SimplifiedPlaylist
	deleted []string

	deleteErr error
}

func (f *fakeClient) GetMe(ctx context.Context) (spotify.User, error) {
	return f.me, nil
}

func (f *fakeClient) ListCurrentUserPlaylists(ctx context.Context, max int) ([]spotify.SimplifiedPlaylist, error) {
	return f.playlists, nil
}

func (f *fakeClient) ListPlaylistTracks(ctx context.Context, playlistID string, max int) ([]spotify.FullTrack, error) {
	return f.tracks[playlistID], nil
}

func (f *fakeClient) CreatePlaylist(ctx context.Context, userID, name, description string, public bool) (spotify.SimplifiedPlaylist, error) {
	f.created = spotify.SimplifiedPlaylist{ID: "new1", Name: name, Description: description, Public: public}
	f.created.Owner.ID = userID
	if f.tracks == nil {
		f.tracks = map[string][]spotify.FullTrack{}
	}
	f.tracks[f.created.ID] = nil
	return f.created, nil
}

func (f *fakeClient) AddTracksToPlaylist(ctx context.Context, playlistID string, uris []string) (string, error) {
	for _, u := range uris {
		f.tracks[playlistID] = append(f.tracks[playlistID], spotify.FullTrack{URI: u})
	}
	return "snap", nil
}

func (f *fakeClient) GetPlaylist(ctx context.Context, playlistID string) (spotify.SimplifiedPlaylist, error) {
	if playlistID == f.created.ID {
		return f.created, nil
	}
	for _, p := range f.playlists {
		if p.ID == playlistID {
			return p, nil
		}
	}
	return spotify.SimplifiedPlaylist{}, nil
}

func (f *fakeClient) DeletePlaylist(ctx context.Context, playlistID string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deleted = append(f.deleted, playlistID)
	return nil
}

func TestFindPlaylistsByPattern(t *testing.T) {
	c := &fakeClient{
		me: spotify.User{ID: "me"},
		playlists: []spotify.SimplifiedPlaylist{
			{ID: "p1", Name: "Workout 2024"},
			{ID: "p2", Name: "Chill"},
			{ID: "p3", Name: "Workout - Legs"},
		},
	}
	svc := NewService(c)

	got, err := svc.FindPlaylistsByPattern(context.Background(), "Workout.*")
	if err != nil {
		t.Fatalf("FindPlaylistsByPattern: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d playlists want 2", len(got))
	}
}

func TestFindPlaylistsByPattern_NoMatches(t *testing.T) {
	c := &fakeClient{me: spotify.User{ID: "me"}, playlists: []spotify.SimplifiedPlaylist{{ID: "p1", Name: "Chill"}}}
	svc := NewService(c)

	_, err := svc.FindPlaylistsByPattern(context.Background(), "Workout.*")
	if !errors.Is(err, ErrNoPlaylistsMatched) {
		t.Fatalf("expected ErrNoPlaylistsMatched, got %v", err)
	}
}

func TestMergePlaylists_DeduplicatesAndVerifies(t *testing.T) {
	c := &fakeClient{
		me:        spotify.User{ID: "me"},
		playlists: []spotify.SimplifiedPlaylist{{ID: "a", Name: "Workout A"}, {ID: "b", Name: "Workout B"}},
		tracks: map[string][]spotify.FullTrack{
			"a": {{URI: "spotify:track:1"}, {URI: "spotify:track:2"}, {URI: "spotify:track:1"}},
			"b": {{URI: "spotify:track:2"}, {URI: "spotify:track:3"}},
		},
	}
	svc := NewService(c)

	res, err := svc.MergePlaylists(context.Background(), []string{"a", "b"}, "All Workouts", MergeOptions{Deduplicate: true, Public: false, Description: "d"})
	if err != nil {
		t.Fatalf("MergePlaylists: %v", err)
	}
	if res.NewPlaylistID != "new1" {
		t.Fatalf("got new id %q want %q", res.NewPlaylistID, "new1")
	}
	if res.TrackCount != 3 {
		t.Fatalf("got track count %d want 3", res.TrackCount)
	}
	if res.DuplicatesRemoved != 2 {
		t.Fatalf("got duplicates removed %d want 2", res.DuplicatesRemoved)
	}
	if !res.Verified {
		t.Fatalf("expected verified")
	}

	gotURIs := make([]string, 0, len(c.tracks["new1"]))
	for _, tr := range c.tracks["new1"] {
		gotURIs = append(gotURIs, tr.URI)
	}
	want := "spotify:track:1,spotify:track:2,spotify:track:3"
	if strings.Join(gotURIs, ",") != want {
		t.Fatalf("unexpected uris: %q want %q", strings.Join(gotURIs, ","), want)
	}
}

func TestVerifyPlaylistContents_Missing(t *testing.T) {
	c := &fakeClient{tracks: map[string][]spotify.FullTrack{"p": {{URI: "spotify:track:1"}}}}
	svc := NewService(c)

	ok, missing, err := svc.VerifyPlaylistContents(context.Background(), "p", []string{"spotify:track:1", "spotify:track:2"})
	if err != nil {
		t.Fatalf("VerifyPlaylistContents: %v", err)
	}
	if ok {
		t.Fatalf("expected not ok")
	}
	if len(missing) != 1 || missing[0] != "spotify:track:2" {
		t.Fatalf("unexpected missing: %#v", missing)
	}
}

func TestDeletePlaylists_PermissionDenied(t *testing.T) {
	c := &fakeClient{deleteErr: spotify.APIError{Status: 403, Message: "forbidden"}}
	svc := NewService(c)

	err := svc.DeletePlaylists(context.Background(), []string{"p1"})
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected permission denied error, got %v", err)
	}
}
