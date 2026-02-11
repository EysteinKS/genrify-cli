package config

import (
	"os"
	"reflect"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	// Set required env var
	os.Setenv("SPOTIFY_CLIENT_ID", "test-client-id")
	defer os.Unsetenv("SPOTIFY_CLIENT_ID")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.SpotifyClientID != "test-client-id" {
		t.Errorf("expected SpotifyClientID='test-client-id', got %q", cfg.SpotifyClientID)
	}

	// Check defaults
	if cfg.SpotifyRedirect != "http://localhost:8888/callback" {
		t.Errorf("expected default redirect, got %q", cfg.SpotifyRedirect)
	}

	expectedScopes := []string{
		"playlist-read-private",
		"playlist-read-collaborative",
		"playlist-modify-private",
		"playlist-modify-public",
	}
	if !reflect.DeepEqual(cfg.SpotifyScopes, expectedScopes) {
		t.Errorf("expected default scopes %v, got %v", expectedScopes, cfg.SpotifyScopes)
	}

	if cfg.UserAgent == "" {
		t.Error("expected non-empty UserAgent")
	}

	if cfg.TokenCacheAppKey == "" {
		t.Error("expected non-empty TokenCacheAppKey")
	}
}

func TestLoad_MissingClientID(t *testing.T) {
	os.Unsetenv("SPOTIFY_CLIENT_ID")
	_, err := Load()
	if err == nil || err.Error() != "SPOTIFY_CLIENT_ID is required" {
		t.Errorf("expected 'SPOTIFY_CLIENT_ID is required' error, got: %v", err)
	}
}

func TestLoad_CustomRedirectURI(t *testing.T) {
	os.Setenv("SPOTIFY_CLIENT_ID", "test-client-id")
	os.Setenv("SPOTIFY_REDIRECT_URI", "http://localhost:9999/auth")
	defer func() {
		os.Unsetenv("SPOTIFY_CLIENT_ID")
		os.Unsetenv("SPOTIFY_REDIRECT_URI")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.SpotifyRedirect != "http://localhost:9999/auth" {
		t.Errorf("expected custom redirect, got %q", cfg.SpotifyRedirect)
	}
}

func TestLoad_CustomScopes(t *testing.T) {
	os.Setenv("SPOTIFY_CLIENT_ID", "test-client-id")
	os.Setenv("SPOTIFY_SCOPES", "user-read-private playlist-modify-public")
	defer func() {
		os.Unsetenv("SPOTIFY_CLIENT_ID")
		os.Unsetenv("SPOTIFY_SCOPES")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	expectedScopes := []string{"user-read-private", "playlist-modify-public"}
	if !reflect.DeepEqual(cfg.SpotifyScopes, expectedScopes) {
		t.Errorf("expected scopes %v, got %v", expectedScopes, cfg.SpotifyScopes)
	}
}

func TestLoad_TLSFiles(t *testing.T) {
	os.Setenv("SPOTIFY_CLIENT_ID", "test-client-id")
	os.Setenv("SPOTIFY_TLS_CERT_FILE", "/path/to/cert.pem")
	os.Setenv("SPOTIFY_TLS_KEY_FILE", "/path/to/key.pem")
	defer func() {
		os.Unsetenv("SPOTIFY_CLIENT_ID")
		os.Unsetenv("SPOTIFY_TLS_CERT_FILE")
		os.Unsetenv("SPOTIFY_TLS_KEY_FILE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.SpotifyTLSCert != "/path/to/cert.pem" {
		t.Errorf("expected TLSCert='/path/to/cert.pem', got %q", cfg.SpotifyTLSCert)
	}
	if cfg.SpotifyTLSKey != "/path/to/key.pem" {
		t.Errorf("expected TLSKey='/path/to/key.pem', got %q", cfg.SpotifyTLSKey)
	}
}

func TestLoad_TrimsWhitespace(t *testing.T) {
	os.Setenv("SPOTIFY_CLIENT_ID", "  test-client-id  ")
	os.Setenv("SPOTIFY_REDIRECT_URI", "  http://localhost:8888  ")
	defer func() {
		os.Unsetenv("SPOTIFY_CLIENT_ID")
		os.Unsetenv("SPOTIFY_REDIRECT_URI")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.SpotifyClientID != "test-client-id" {
		t.Errorf("expected trimmed client ID, got %q", cfg.SpotifyClientID)
	}
	if cfg.SpotifyRedirect != "http://localhost:8888" {
		t.Errorf("expected trimmed redirect URI, got %q", cfg.SpotifyRedirect)
	}
}

func TestSplitScopes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "space separated",
			input: "scope1 scope2 scope3",
			want:  []string{"scope1", "scope2", "scope3"},
		},
		{
			name:  "comma separated",
			input: "scope1,scope2,scope3",
			want:  []string{"scope1", "scope2", "scope3"},
		},
		{
			name:  "mixed separators",
			input: "scope1, scope2 scope3,scope4",
			want:  []string{"scope1", "scope2", "scope3", "scope4"},
		},
		{
			name:  "extra whitespace",
			input: "  scope1  ,  scope2   scope3  ",
			want:  []string{"scope1", "scope2", "scope3"},
		},
		{
			name:  "single scope",
			input: "scope1",
			want:  []string{"scope1"},
		},
		{
			name:  "empty string",
			input: "",
			want:  []string{},
		},
		{
			name:  "only whitespace",
			input: "   ",
			want:  []string{},
		},
		{
			name:  "trailing comma",
			input: "scope1,scope2,",
			want:  []string{"scope1", "scope2"},
		},
		{
			name:  "leading comma",
			input: ",scope1,scope2",
			want:  []string{"scope1", "scope2"},
		},
		{
			name:  "multiple separators",
			input: "scope1,,  scope2",
			want:  []string{"scope1", "scope2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitScopes(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitScopes(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
