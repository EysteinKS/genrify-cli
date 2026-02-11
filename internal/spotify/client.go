package spotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.spotify.com/v1/"

type Client struct {
	baseURL   *url.URL
	http      *http.Client
	userAgent string
	tokens    *TokenManager
}

type Option func(*Client)

func WithBaseURL(raw string) Option {
	return func(c *Client) {
		u, err := url.Parse(raw)
		if err == nil {
			c.baseURL = u
		}
	}
}

func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.http = h
		}
	}
}

func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.userAgent = ua
	}
}

func WithTokenManager(m *TokenManager) Option {
	return func(c *Client) {
		c.tokens = m
	}
}

func New(opts ...Option) (*Client, error) {
	u, _ := url.Parse(defaultBaseURL)
	c := &Client{
		baseURL: u,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.baseURL == nil {
		return nil, fmt.Errorf("baseURL is required")
	}
	if c.http == nil {
		return nil, fmt.Errorf("http client is required")
	}
	if c.tokens == nil {
		return nil, fmt.Errorf("token manager is required")
	}
	return c, nil
}

func (c *Client) GetMe(ctx context.Context) (User, error) {
	var me User
	if err := c.doJSON(ctx, http.MethodGet, "/me", nil, nil, &me); err != nil {
		return User{}, err
	}
	return me, nil
}

func (c *Client) ListCurrentUserPlaylists(ctx context.Context, max int) ([]SimplifiedPlaylist, error) {
	const pageSize = 50
	items, err := collectPaged(ctx, pageSize, max, func(ctx context.Context, limit, offset int) (paging[SimplifiedPlaylist], error) {
		q := url.Values{}
		q.Set("limit", fmt.Sprintf("%d", limit))
		q.Set("offset", fmt.Sprintf("%d", offset))
		var p paging[SimplifiedPlaylist]
		err := c.doJSON(ctx, http.MethodGet, "/me/playlists", q, nil, &p)
		return p, err
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (c *Client) ListPlaylistTracks(ctx context.Context, playlistID string, max int) ([]FullTrack, error) {
	const pageSize = 100
	playlistID = strings.TrimSpace(playlistID)
	if playlistID == "" {
		return nil, fmt.Errorf("playlist id is required")
	}

	endpoint := "/playlists/" + url.PathEscape(playlistID) + "/tracks"
	items, err := collectPaged(ctx, pageSize, max, func(ctx context.Context, limit, offset int) (paging[playlistTrackItem], error) {
		q := url.Values{}
		q.Set("limit", fmt.Sprintf("%d", limit))
		q.Set("offset", fmt.Sprintf("%d", offset))
		var p paging[playlistTrackItem]
		err := c.doJSON(ctx, http.MethodGet, endpoint, q, nil, &p)
		return p, err
	})
	if err != nil {
		return nil, err
	}

	tracks := make([]FullTrack, 0, len(items))
	for _, it := range items {
		// Some items can have null tracks; ignore zero values.
		if it.Track.URI != "" {
			tracks = append(tracks, it.Track)
		}
	}
	return tracks, nil
}

func (c *Client) CreatePlaylist(ctx context.Context, userID, name, description string, public bool) (SimplifiedPlaylist, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return SimplifiedPlaylist{}, fmt.Errorf("name is required")
	}

	body := map[string]any{
		"name":        name,
		"public":      public,
		"description": description,
	}

	var pl SimplifiedPlaylist
	// Spotify now supports creating playlists for the current user via /me.
	// Using /me avoids user-id mismatch errors.
	p := "/me/playlists"
	if err := c.doJSON(ctx, http.MethodPost, p, nil, body, &pl); err != nil {
		return SimplifiedPlaylist{}, err
	}
	return pl, nil
}

func (c *Client) AddTracksToPlaylist(ctx context.Context, playlistID string, uris []string) (string, error) {
	playlistID = strings.TrimSpace(playlistID)
	if playlistID == "" {
		return "", fmt.Errorf("playlist id is required")
	}
	clean := make([]string, 0, len(uris))
	for _, u := range uris {
		u = strings.TrimSpace(u)
		if u != "" {
			clean = append(clean, u)
		}
	}
	if len(clean) == 0 {
		return "", fmt.Errorf("at least one track uri is required")
	}

	p := "/playlists/" + url.PathEscape(playlistID) + "/tracks"
	var lastSnapshot string
	for i := 0; i < len(clean); i += 100 {
		end := i + 100
		if end > len(clean) {
			end = len(clean)
		}
		body := map[string]any{"uris": clean[i:end]}
		var resp snapshotResponse
		if err := c.doJSON(ctx, http.MethodPost, p, nil, body, &resp); err != nil {
			return "", err
		}
		lastSnapshot = resp.SnapshotID
	}
	return lastSnapshot, nil
}

func (c *Client) GetPlaylist(ctx context.Context, playlistID string) (SimplifiedPlaylist, error) {
	playlistID = strings.TrimSpace(playlistID)
	if playlistID == "" {
		return SimplifiedPlaylist{}, fmt.Errorf("playlist id is required")
	}
	var pl SimplifiedPlaylist
	endpoint := "/playlists/" + url.PathEscape(playlistID)
	if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, nil, &pl); err != nil {
		return SimplifiedPlaylist{}, err
	}
	return pl, nil
}

// DeletePlaylist removes the current user's follows for a playlist.
// For playlists the user owns, this effectively deletes the playlist.
// Uses: DELETE /playlists/{playlist_id}/followers
func (c *Client) DeletePlaylist(ctx context.Context, playlistID string) error {
	playlistID = strings.TrimSpace(playlistID)
	if playlistID == "" {
		return fmt.Errorf("playlist id is required")
	}
	endpoint := "/playlists/" + url.PathEscape(playlistID) + "/followers"
	return c.doJSON(ctx, http.MethodDelete, endpoint, nil, nil, nil)
}

// RemoveTracksFromPlaylist removes track URIs from a playlist.
// Uses: DELETE /playlists/{playlist_id}/tracks
func (c *Client) RemoveTracksFromPlaylist(ctx context.Context, playlistID string, uris []string) (string, error) {
	playlistID = strings.TrimSpace(playlistID)
	if playlistID == "" {
		return "", fmt.Errorf("playlist id is required")
	}
	tracks := make([]map[string]string, 0, len(uris))
	for _, u := range uris {
		u = strings.TrimSpace(u)
		if u != "" {
			tracks = append(tracks, map[string]string{"uri": u})
		}
	}
	if len(tracks) == 0 {
		return "", fmt.Errorf("at least one track uri is required")
	}

	body := map[string]any{"tracks": tracks}
	var resp snapshotResponse
	endpoint := "/playlists/" + url.PathEscape(playlistID) + "/tracks"
	if err := c.doJSON(ctx, http.MethodDelete, endpoint, nil, body, &resp); err != nil {
		return "", err
	}
	return resp.SnapshotID, nil
}

func (c *Client) doJSON(ctx context.Context, method, p string, query url.Values, body any, out any) error {
	return c.doJSONWithRetry(ctx, method, p, query, body, out, true)
}

func (c *Client) doJSONWithRetry(ctx context.Context, method, p string, query url.Values, body any, out any, allowRetry bool) error {
	u := *c.baseURL
	u.Path = path.Join(c.baseURL.Path, p)
	u.RawQuery = ""
	if query != nil {
		u.RawQuery = query.Encode()
	}

	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyBytes = b
	}

	refreshed := false
	rateRetries := 0
	for {
		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
		if err != nil {
			return err
		}

		accessToken, err := c.tokens.AccessToken(ctx)
		if err != nil {
			return err
		}

		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Accept", "application/json")
		if c.userAgent != "" {
			req.Header.Set("User-Agent", c.userAgent)
		}
		if bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized && allowRetry && !refreshed {
			// Try refresh + retry once.
			if _, err := c.tokens.ForceRefresh(ctx); err == nil {
				refreshed = true
				continue
			}
		}

		if resp.StatusCode == http.StatusTooManyRequests && allowRetry && rateRetries < 5 {
			wait := retryAfterDuration(resp.Header.Get("Retry-After"), rateRetries)
			rateRetries++
			if err := sleepWithContext(ctx, wait); err != nil {
				return err
			}
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return decodeAPIError(b, resp.StatusCode)
		}
		if out == nil {
			return nil
		}
		if len(b) == 0 {
			return nil
		}
		return json.Unmarshal(b, out)
	}
}

func retryAfterDuration(headerVal string, attempt int) time.Duration {
	if headerVal != "" {
		if secs, err := strconv.Atoi(strings.TrimSpace(headerVal)); err == nil && secs >= 0 {
			if secs == 0 {
				return 0
			}
			return time.Duration(secs) * time.Second
		}
	}
	// Exponential backoff with a small cap.
	base := 250 * time.Millisecond
	d := base * time.Duration(1<<attempt)
	if d > 5*time.Second {
		return 5 * time.Second
	}
	return d
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
