package cli

import (
	"context"
	"errors"
	"io"
	"testing"

	"genrify/internal/config"
	"genrify/internal/spotify"
)

type fakeSpotifyClient struct {
	loggedIn bool
	getMeN   int
}

func (c *fakeSpotifyClient) GetMe(ctx context.Context) (spotify.User, error) {
	c.getMeN++
	if !c.loggedIn {
		return spotify.User{}, errors.New("not logged in (missing token); run genrify login")
	}
	return spotify.User{ID: "me"}, nil
}

func (c *fakeSpotifyClient) ListCurrentUserPlaylists(ctx context.Context, max int) ([]spotify.SimplifiedPlaylist, error) {
	return nil, nil
}

func (c *fakeSpotifyClient) ListPlaylistTracks(ctx context.Context, playlistID string, max int) ([]spotify.FullTrack, error) {
	return nil, nil
}

func (c *fakeSpotifyClient) CreatePlaylist(ctx context.Context, userID, name, description string, public bool) (spotify.SimplifiedPlaylist, error) {
	return spotify.SimplifiedPlaylist{}, nil
}

func (c *fakeSpotifyClient) AddTracksToPlaylist(ctx context.Context, playlistID string, uris []string) (string, error) {
	return "", nil
}

func (c *fakeSpotifyClient) GetPlaylist(ctx context.Context, playlistID string) (spotify.SimplifiedPlaylist, error) {
	return spotify.SimplifiedPlaylist{}, nil
}

func (c *fakeSpotifyClient) DeletePlaylist(ctx context.Context, playlistID string) error {
	return nil
}

func TestStart_AutoLoginWhenNotLoggedIn(t *testing.T) {
	rootCmd, root := NewRoot()

	cfg := config.Default()
	cfg.SpotifyClientID = "client-id"
	root.loadConfig = func() (config.Config, error) { return cfg, nil }
	root.saveConfig = func(c config.Config) (string, error) { return "", nil }

	fakeClient := &fakeSpotifyClient{}
	root.newSpotifyClient = func(cfg config.Config) (SpotifyClient, error) {
		return fakeClient, nil
	}

	var loginCalled bool
	root.doLogin = func(ctx context.Context, cfg config.Config) (string, error) {
		loginCalled = true
		fakeClient.loggedIn = true
		return "/tmp/token", nil
	}

	var loopCalled bool
	root.runInteractiveLoop = func(ctx context.Context, client SpotifyClient, prompter Prompter) error {
		loopCalled = true
		return nil
	}

	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	rootCmd.SetArgs([]string{"start"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !loginCalled {
		t.Fatalf("expected login to be called")
	}
	if !loopCalled {
		t.Fatalf("expected interactive loop to be called")
	}
	if fakeClient.getMeN < 2 {
		t.Fatalf("expected GetMe to be called at least twice, got %d", fakeClient.getMeN)
	}
}
