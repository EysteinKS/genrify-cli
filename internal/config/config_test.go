package config

import (
	"path/filepath"
	"os"
	"reflect"
	"testing"
)

func withTempConfigDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	t.Cleanup(func() {
		os.Unsetenv("XDG_CONFIG_HOME")
	})
	return dir
}

func TestLoad_Success(t *testing.T) {
	withTempConfigDir(t)
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
	withTempConfigDir(t)
	os.Unsetenv("SPOTIFY_CLIENT_ID")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.SpotifyClientID != "" {
		t.Fatalf("expected empty client id when not configured, got %q", cfg.SpotifyClientID)
	}
}

func TestLoad_CustomRedirectURI(t *testing.T) {
	withTempConfigDir(t)
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
	withTempConfigDir(t)
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
	withTempConfigDir(t)
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
	withTempConfigDir(t)
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

func TestLoad_UsesConfigFileWhenEnvMissing(t *testing.T) {
	base := withTempConfigDir(t)

	os.Unsetenv("SPOTIFY_CLIENT_ID")

	path, err := Path()
	if err != nil {
		t.Fatalf("Path() failed: %v", err)
	}
	if filepath.Dir(path) != filepath.Join(base, "genrify") {
		t.Fatalf("expected config dir under temp, got %s", path)
	}

	_, err = Save(Config{SpotifyClientID: "from-file"})
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.SpotifyClientID != "from-file" {
		t.Fatalf("expected client id from file, got %q", cfg.SpotifyClientID)
	}
}

func TestLoad_EnvOverridesConfigFile(t *testing.T) {
	withTempConfigDir(t)

	_, err := Save(Config{SpotifyClientID: "from-file"})
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	os.Setenv("SPOTIFY_CLIENT_ID", "from-env")
	defer os.Unsetenv("SPOTIFY_CLIENT_ID")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.SpotifyClientID != "from-env" {
		t.Fatalf("expected env override, got %q", cfg.SpotifyClientID)
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
