package auth

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestExchangeCode_Success(t *testing.T) {
	// This test actually hits the real Spotify API
	// For a proper test, we'd need to refactor exchangeCode to accept a custom HTTP client or URL
	// For now, we'll test the error cases that don't require network calls
	t.Skip("exchangeCode needs refactoring to accept custom HTTP client for testing")
}

func TestExchangeCode_MissingAccessToken(t *testing.T) {
	// This would require refactoring exchangeCode to inject the HTTP client
	t.Skip("exchangeCode needs refactoring to accept custom HTTP client for testing")
}

func TestRefresh_Success(t *testing.T) {
	// This would require refactoring Refresh to inject the HTTP client
	t.Skip("Refresh needs refactoring to accept custom HTTP client for testing")
}

func TestRefresh_MissingRefreshToken(t *testing.T) {
	cfg := OAuthConfig{ClientID: "test"}
	_, err := Refresh(context.Background(), cfg, "")
	if err == nil || err.Error() != "missing refresh token" {
		t.Errorf("expected 'missing refresh token' error, got: %v", err)
	}
}

func TestBuildAuthorizeURL(t *testing.T) {
	cfg := OAuthConfig{
		ClientID: "test-client-id",
		Scopes:   []string{"user-read-private", "playlist-modify-public"},
	}

	redirectURI := "http://localhost:8888/callback"
	state := "test-state"
	codeChallenge := "test-challenge"

	result := buildAuthorizeURL(cfg, redirectURI, state, codeChallenge)

	// Parse the result
	u, err := url.Parse(result)
	if err != nil {
		t.Fatalf("failed to parse result URL: %v", err)
	}

	// Check base URL
	if u.Scheme != "https" {
		t.Errorf("expected scheme https, got %s", u.Scheme)
	}
	if u.Host != "accounts.spotify.com" {
		t.Errorf("expected host accounts.spotify.com, got %s", u.Host)
	}
	if u.Path != "/authorize" {
		t.Errorf("expected path /authorize, got %s", u.Path)
	}

	// Check query parameters
	q := u.Query()
	if q.Get("client_id") != "test-client-id" {
		t.Errorf("expected client_id=test-client-id, got %s", q.Get("client_id"))
	}
	if q.Get("response_type") != "code" {
		t.Errorf("expected response_type=code, got %s", q.Get("response_type"))
	}
	if q.Get("redirect_uri") != redirectURI {
		t.Errorf("expected redirect_uri=%s, got %s", redirectURI, q.Get("redirect_uri"))
	}
	if q.Get("state") != state {
		t.Errorf("expected state=%s, got %s", state, q.Get("state"))
	}
	if q.Get("code_challenge_method") != "S256" {
		t.Errorf("expected code_challenge_method=S256, got %s", q.Get("code_challenge_method"))
	}
	if q.Get("code_challenge") != codeChallenge {
		t.Errorf("expected code_challenge=%s, got %s", codeChallenge, q.Get("code_challenge"))
	}
	expectedScope := "user-read-private playlist-modify-public"
	if q.Get("scope") != expectedScope {
		t.Errorf("expected scope=%s, got %s", expectedScope, q.Get("scope"))
	}
}

func TestBuildAuthorizeURL_NoScopes(t *testing.T) {
	cfg := OAuthConfig{
		ClientID: "test-client-id",
		Scopes:   []string{},
	}

	result := buildAuthorizeURL(cfg, "http://localhost:8888", "state", "challenge")
	u, _ := url.Parse(result)
	q := u.Query()

	if q.Get("scope") != "" {
		t.Errorf("expected no scope parameter, got %s", q.Get("scope"))
	}
}

func TestIsPortZero(t *testing.T) {
	tests := []struct {
		hostport string
		want     bool
	}{
		{"localhost:0", true},
		{"127.0.0.1:0", true},
		{"localhost:8888", false},
		{"127.0.0.1:8080", false},
		{"localhost", false},
		{"example.com:443", false},
		{"[::1]:0", true},
		{"[::1]:8080", false},
	}

	for _, tt := range tests {
		t.Run(tt.hostport, func(t *testing.T) {
			if got := isPortZero(tt.hostport); got != tt.want {
				t.Errorf("isPortZero(%q) = %v, want %v", tt.hostport, got, tt.want)
			}
		})
	}
}

func TestLoginPKCE_RequiresClientID(t *testing.T) {
	cfg := OAuthConfig{
		RedirectURI: "http://localhost:8888",
	}
	_, err := LoginPKCE(context.Background(), cfg)
	if err == nil || !strings.Contains(err.Error(), "client id is required") {
		t.Errorf("expected 'client id is required' error, got: %v", err)
	}
}

func TestLoginPKCE_RequiresRedirectURI(t *testing.T) {
	cfg := OAuthConfig{
		ClientID: "test",
	}
	_, err := LoginPKCE(context.Background(), cfg)
	if err == nil || !strings.Contains(err.Error(), "redirect uri is required") {
		t.Errorf("expected 'redirect uri is required' error, got: %v", err)
	}
}

func TestLoginPKCE_ValidatesRedirectScheme(t *testing.T) {
	cfg := OAuthConfig{
		ClientID:    "test",
		RedirectURI: "ftp://localhost:8888",
	}
	_, err := LoginPKCE(context.Background(), cfg)
	if err == nil || !strings.Contains(err.Error(), "scheme must be http or https") {
		t.Errorf("expected scheme validation error, got: %v", err)
	}
}

func TestLoginPKCE_ValidatesRedirectHost(t *testing.T) {
	cfg := OAuthConfig{
		ClientID:    "test",
		RedirectURI: "http:///callback",
	}
	_, err := LoginPKCE(context.Background(), cfg)
	if err == nil || !strings.Contains(err.Error(), "must include host") {
		t.Errorf("expected host validation error, got: %v", err)
	}
}

func TestLoginPKCE_HTTPSRequiresCerts(t *testing.T) {
	cfg := OAuthConfig{
		ClientID:    "test",
		RedirectURI: "https://localhost:8443",
	}
	_, err := LoginPKCE(context.Background(), cfg)
	if err == nil || !strings.Contains(err.Error(), "requires SPOTIFY_TLS_CERT_FILE") {
		t.Errorf("expected TLS cert requirement error, got: %v", err)
	}
}

func TestStartCallbackServer(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "OK")
	})

	// Start server with port 0 to get a random available port
	listener, server, err := startCallbackServer(mux, "http", "localhost:0", "", "")
	if err != nil {
		t.Fatalf("startCallbackServer failed: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
		_ = listener.Close()
	}()

	// Make a request to the server
	addr := listener.Addr().String()
	resp, err := http.Get("http://" + addr)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "OK" {
		t.Errorf("expected body 'OK', got %q", string(body))
	}
}
