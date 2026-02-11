package spotify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"genrify/internal/auth"
	"genrify/internal/testutil"
)

func newTestClient(t *testing.T, baseURL string, store *testutil.MemStore, refresher Refresher) *Client {
	t.Helper()
	m := NewTokenManager(store, 0, refresher)
	c, err := New(
		WithBaseURL(baseURL),
		WithHTTPClient(&http.Client{Timeout: 5 * time.Second}),
		WithTokenManager(m),
	)
	if err != nil {
		t.Fatalf("New client: %v", err)
	}
	return c
}

func TestClient_ListCurrentUserPlaylists_Paginates(t *testing.T) {
	store := testutil.NewMemStore(auth.Token{AccessToken: "ok", RefreshToken: "r", ExpiresAt: time.Now().Add(10 * time.Minute)})

	mux := http.NewServeMux()
	mux.HandleFunc("/me/playlists", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer ok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		_ = limit

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		if offset == 0 {
			_ = enc.Encode(map[string]any{
				"items": []map[string]any{{"id": "p1", "name": "One", "owner": map[string]any{"id": "u"}, "tracks": map[string]any{"total": 1}}},
				"next":  "http://example/next",
			})
			return
		}
		if offset > 0 {
			_ = enc.Encode(map[string]any{
				"items": []map[string]any{{"id": "p2", "name": "Two", "owner": map[string]any{"id": "u"}, "tracks": map[string]any{"total": 2}}},
				"next":  "",
			})
			return
		}
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv.URL, store, func(ctx context.Context, refreshToken string) (auth.Token, error) {
		return auth.Token{}, nil
	})

	items, err := c.ListCurrentUserPlaylists(context.Background(), 0)
	if err != nil {
		t.Fatalf("ListCurrentUserPlaylists: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items want 2", len(items))
	}
	if items[0].ID != "p1" || items[1].ID != "p2" {
		t.Fatalf("unexpected ids: %#v", []string{items[0].ID, items[1].ID})
	}
}

func TestClient_RetriesOn401_AfterRefresh(t *testing.T) {
	store := testutil.NewMemStore(auth.Token{AccessToken: "bad", RefreshToken: "r", ExpiresAt: time.Now().Add(10 * time.Minute)})
	refreshCalls := 0

	mux := http.NewServeMux()
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get("Authorization") {
		case "Bearer bad":
			w.WriteHeader(http.StatusUnauthorized)
			return
		case "Bearer good":
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"id":"me","display_name":"Me"}`)
			return
		default:
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv.URL, store, func(ctx context.Context, refreshToken string) (auth.Token, error) {
		refreshCalls++
		return auth.Token{AccessToken: "good", RefreshToken: refreshToken, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil
	})

	me, err := c.GetMe(context.Background())
	if err != nil {
		t.Fatalf("GetMe: %v", err)
	}
	if me.ID != "me" {
		t.Fatalf("got me.ID=%q want %q", me.ID, "me")
	}
	if refreshCalls != 1 {
		t.Fatalf("expected 1 refresh call, got %d", refreshCalls)
	}
}

func TestClient_RetriesOn429_ThenSucceeds(t *testing.T) {
	store := testutil.NewMemStore(auth.Token{AccessToken: "ok", RefreshToken: "r", ExpiresAt: time.Now().Add(10 * time.Minute)})
	calls := 0

	mux := http.NewServeMux()
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		calls++
		if got := r.Header.Get("Authorization"); got != "Bearer ok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if calls <= 2 {
			w.Header().Set("Retry-After", "0")
			testutil.RespondError(w, http.StatusTooManyRequests, "rate limited")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"me","display_name":"Me"}`)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv.URL, store, func(ctx context.Context, refreshToken string) (auth.Token, error) {
		return auth.Token{}, nil
	})

	me, err := c.GetMe(context.Background())
	if err != nil {
		t.Fatalf("GetMe: %v", err)
	}
	if me.ID != "me" {
		t.Fatalf("got me.ID=%q want %q", me.ID, "me")
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestClient_AddTracks_BatchesBy100(t *testing.T) {
	store := testutil.NewMemStore(auth.Token{AccessToken: "ok", RefreshToken: "r", ExpiresAt: time.Now().Add(10 * time.Minute)})

	uris := make([]string, 0, 205)
	for i := 0; i < 205; i++ {
		uris = append(uris, "spotify:track:"+strconv.Itoa(i))
	}

	calls := 0
	seenCounts := []int{}

	mux := http.NewServeMux()
	mux.HandleFunc("/playlists/pl123/tracks", func(w http.ResponseWriter, r *http.Request) {
		calls++
		if got := r.Header.Get("Authorization"); got != "Bearer ok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		b, _ := io.ReadAll(r.Body)
		var req struct {
			URIs []string `json:"uris"`
		}
		if err := json.Unmarshal(b, &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		seenCounts = append(seenCounts, len(req.URIs))
		w.Header().Set("Content-Type", "application/json")
		snap := "s" + strconv.Itoa(calls)
		io.WriteString(w, `{"snapshot_id":"`+snap+`"}`)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv.URL, store, func(ctx context.Context, refreshToken string) (auth.Token, error) {
		return auth.Token{}, nil
	})

	snapshot, err := c.AddTracksToPlaylist(context.Background(), "pl123", uris)
	if err != nil {
		t.Fatalf("AddTracksToPlaylist: %v", err)
	}
	if snapshot != "s3" {
		t.Fatalf("got snapshot %q want %q", snapshot, "s3")
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
	if !(seenCounts[0] == 100 && seenCounts[1] == 100 && seenCounts[2] == 5) {
		t.Fatalf("unexpected batch sizes: %#v", seenCounts)
	}
}

func TestClient_CreatePlaylist(t *testing.T) {
	store := testutil.NewMemStore(auth.Token{AccessToken: "ok", RefreshToken: "r", ExpiresAt: time.Now().Add(10 * time.Minute)})

	mux := http.NewServeMux()
	mux.HandleFunc("/me/playlists", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer ok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		b, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(b, &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"id":          "pl123",
			"name":        req["name"],
			"description": req["description"],
			"public":      req["public"],
			"owner":       map[string]any{"id": "me"},
			"tracks":      map[string]any{"total": 0},
		}
		json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv.URL, store, func(ctx context.Context, refreshToken string) (auth.Token, error) {
		return auth.Token{}, nil
	})

	pl, err := c.CreatePlaylist(context.Background(), "", "My Playlist", "Test description", true)
	if err != nil {
		t.Fatalf("CreatePlaylist: %v", err)
	}
	if pl.ID != "pl123" {
		t.Errorf("got ID %q want %q", pl.ID, "pl123")
	}
	if pl.Name != "My Playlist" {
		t.Errorf("got Name %q want %q", pl.Name, "My Playlist")
	}
	if pl.Description != "Test description" {
		t.Errorf("got Description %q want %q", pl.Description, "Test description")
	}
	if !pl.Public {
		t.Error("expected Public to be true")
	}
}

func TestClient_GetPlaylist(t *testing.T) {
	store := testutil.NewMemStore(auth.Token{AccessToken: "ok", RefreshToken: "r", ExpiresAt: time.Now().Add(10 * time.Minute)})

	mux := http.NewServeMux()
	mux.HandleFunc("/playlists/pl123", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer ok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":   "pl123",
			"name": "Test",
			"owner": map[string]any{
				"id": "me",
			},
			"tracks": map[string]any{"total": 1},
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv.URL, store, func(ctx context.Context, refreshToken string) (auth.Token, error) {
		return auth.Token{}, nil
	})

	pl, err := c.GetPlaylist(context.Background(), "pl123")
	if err != nil {
		t.Fatalf("GetPlaylist: %v", err)
	}
	if pl.ID != "pl123" || pl.Name != "Test" {
		t.Fatalf("unexpected playlist: %#v", pl)
	}
}

func TestClient_DeletePlaylist(t *testing.T) {
	store := testutil.NewMemStore(auth.Token{AccessToken: "ok", RefreshToken: "r", ExpiresAt: time.Now().Add(10 * time.Minute)})

	mux := http.NewServeMux()
	mux.HandleFunc("/playlists/pl123/followers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer ok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv.URL, store, func(ctx context.Context, refreshToken string) (auth.Token, error) {
		return auth.Token{}, nil
	})

	if err := c.DeletePlaylist(context.Background(), "pl123"); err != nil {
		t.Fatalf("DeletePlaylist: %v", err)
	}
}

func TestClient_CreatePlaylist_RequiresName(t *testing.T) {
	store := testutil.NewMemStore(auth.Token{AccessToken: "ok", RefreshToken: "r", ExpiresAt: time.Now().Add(10 * time.Minute)})
	c := newTestClient(t, "http://example.com", store, nil)

	_, err := c.CreatePlaylist(context.Background(), "user", "", "", false)
	if err == nil || err.Error() != "name is required" {
		t.Errorf("expected 'name is required' error, got: %v", err)
	}
}

func TestClient_ListPlaylistTracks(t *testing.T) {
	store := testutil.NewMemStore(auth.Token{AccessToken: "ok", RefreshToken: "r", ExpiresAt: time.Now().Add(10 * time.Minute)})

	mux := http.NewServeMux()
	mux.HandleFunc("/playlists/pl123/tracks", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer ok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		w.Header().Set("Content-Type", "application/json")
		if offset == 0 {
			resp := map[string]any{
				"items": []map[string]any{
					{
						"track": map[string]any{
							"id":   "t1",
							"name": "Track 1",
							"uri":  "spotify:track:t1",
							"artists": []map[string]any{
								{"id": "a1", "name": "Artist 1"},
							},
							"album": map[string]any{"id": "alb1", "name": "Album 1"},
						},
					},
					{
						"track": map[string]any{
							"id":   "t2",
							"name": "Track 2",
							"uri":  "spotify:track:t2",
							"artists": []map[string]any{
								{"id": "a2", "name": "Artist 2"},
							},
							"album": map[string]any{"id": "alb2", "name": "Album 2"},
						},
					},
				},
				"next": "",
			}
			json.NewEncoder(w).Encode(resp)
		}
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv.URL, store, func(ctx context.Context, refreshToken string) (auth.Token, error) {
		return auth.Token{}, nil
	})

	tracks, err := c.ListPlaylistTracks(context.Background(), "pl123", 0)
	if err != nil {
		t.Fatalf("ListPlaylistTracks: %v", err)
	}
	if len(tracks) != 2 {
		t.Fatalf("got %d tracks want 2", len(tracks))
	}
	if tracks[0].URI != "spotify:track:t1" {
		t.Errorf("got URI %q want %q", tracks[0].URI, "spotify:track:t1")
	}
	if tracks[1].Name != "Track 2" {
		t.Errorf("got Name %q want %q", tracks[1].Name, "Track 2")
	}
}

func TestClient_ListPlaylistTracks_RequiresPlaylistID(t *testing.T) {
	store := testutil.NewMemStore(auth.Token{AccessToken: "ok", RefreshToken: "r", ExpiresAt: time.Now().Add(10 * time.Minute)})
	c := newTestClient(t, "http://example.com", store, nil)

	_, err := c.ListPlaylistTracks(context.Background(), "", 0)
	if err == nil || err.Error() != "playlist id is required" {
		t.Errorf("expected 'playlist id is required' error, got: %v", err)
	}
}

func TestClient_ListPlaylistTracks_FiltersNullTracks(t *testing.T) {
	store := testutil.NewMemStore(auth.Token{AccessToken: "ok", RefreshToken: "r", ExpiresAt: time.Now().Add(10 * time.Minute)})

	mux := http.NewServeMux()
	mux.HandleFunc("/playlists/pl123/tracks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"items": []map[string]any{
				{
					"track": map[string]any{
						"id":   "t1",
						"name": "Track 1",
						"uri":  "spotify:track:t1",
						"artists": []map[string]any{
							{"id": "a1", "name": "Artist 1"},
						},
						"album": map[string]any{"id": "alb1", "name": "Album 1"},
					},
				},
				{
					// null track (URI empty)
					"track": map[string]any{
						"id":   "",
						"name": "",
						"uri":  "",
					},
				},
				{
					"track": map[string]any{
						"id":   "t2",
						"name": "Track 2",
						"uri":  "spotify:track:t2",
						"artists": []map[string]any{
							{"id": "a2", "name": "Artist 2"},
						},
						"album": map[string]any{"id": "alb2", "name": "Album 2"},
					},
				},
			},
			"next": "",
		}
		json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv.URL, store, func(ctx context.Context, refreshToken string) (auth.Token, error) {
		return auth.Token{}, nil
	})

	tracks, err := c.ListPlaylistTracks(context.Background(), "pl123", 0)
	if err != nil {
		t.Fatalf("ListPlaylistTracks: %v", err)
	}
	if len(tracks) != 2 {
		t.Fatalf("got %d tracks want 2 (null track should be filtered)", len(tracks))
	}
}

func TestClient_ErrorHandling_404(t *testing.T) {
	store := testutil.NewMemStore(auth.Token{AccessToken: "ok", RefreshToken: "r", ExpiresAt: time.Now().Add(10 * time.Minute)})

	mux := http.NewServeMux()
	mux.HandleFunc("/playlists/notfound/tracks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, `{"error": {"status": 404, "message": "Not found"}}`)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv.URL, store, func(ctx context.Context, refreshToken string) (auth.Token, error) {
		return auth.Token{}, nil
	})

	_, err := c.ListPlaylistTracks(context.Background(), "notfound", 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 404 {
		t.Errorf("expected status 404, got %d", apiErr.Status)
	}
	if apiErr.Message != "Not found" {
		t.Errorf("expected message 'Not found', got %q", apiErr.Message)
	}
}

func TestClient_ErrorHandling_500(t *testing.T) {
	store := testutil.NewMemStore(auth.Token{AccessToken: "ok", RefreshToken: "r", ExpiresAt: time.Now().Add(10 * time.Minute)})

	mux := http.NewServeMux()
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `{"error": {"status": 500, "message": "Internal server error"}}`)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv.URL, store, func(ctx context.Context, refreshToken string) (auth.Token, error) {
		return auth.Token{}, nil
	})

	_, err := c.GetMe(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 500 {
		t.Errorf("expected status 500, got %d", apiErr.Status)
	}
}

func TestClient_ErrorHandling_RateLimiting(t *testing.T) {
	store := testutil.NewMemStore(auth.Token{AccessToken: "ok", RefreshToken: "r", ExpiresAt: time.Now().Add(10 * time.Minute)})

	mux := http.NewServeMux()
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		io.WriteString(w, `{"error": {"status": 429, "message": "Rate limit exceeded"}}`)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv.URL, store, func(ctx context.Context, refreshToken string) (auth.Token, error) {
		return auth.Token{}, nil
	})

	_, err := c.GetMe(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 429 {
		t.Errorf("expected status 429, got %d", apiErr.Status)
	}
}
