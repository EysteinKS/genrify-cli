package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"genrify/internal/buildinfo"
)

type Config struct {
	SpotifyClientID string   `json:"spotify_client_id,omitempty"`
	SpotifyRedirect string   `json:"spotify_redirect_uri,omitempty"`
	SpotifyScopes   []string `json:"spotify_scopes,omitempty"`
	SpotifyTLSCert  string   `json:"spotify_tls_cert_file,omitempty"`
	SpotifyTLSKey   string   `json:"spotify_tls_key_file,omitempty"`

	UserAgent        string `json:"-"`
	TokenCacheAppKey string `json:"-"`
}


func Default() Config {
	return Config{
		SpotifyRedirect: "http://localhost:8888/callback",
		SpotifyScopes: []string{
			"playlist-read-private",
			"playlist-read-collaborative",
			"playlist-modify-private",
			"playlist-modify-public",
		},
		UserAgent:        buildinfo.UserAgent,
		TokenCacheAppKey: buildinfo.AppName,
	}
}

func Path() (string, error) {
	dir, err := userConfigDir()
	if err != nil {
		return "", fmt.Errorf("get user config dir: %w", err)
	}
	return filepath.Join(dir, buildinfo.AppName, "config.json"), nil
}

func userConfigDir() (string, error) {
	// Prefer XDG when explicitly set (even on macOS) to support
	// predictable config locations and easier testing.
	if v := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); v != "" {
		return v, nil
	}
	return os.UserConfigDir()
}

func Load() (Config, error) {
	cfg := Default()

	path, err := Path()
	if err != nil {
		return Config{}, err
	}
	fileCfg, err := loadFile(path)
	if err != nil {
		return Config{}, err
	}
	merge(&cfg, fileCfg)

	applyEnvOverrides(&cfg)
	normalize(&cfg)
	applyComputed(&cfg)
	return cfg, nil
}

func Save(cfg Config) (string, error) {
	path, err := Path()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}

	copy := cfg
	// Ensure we don't accidentally persist computed-only fields.
	copy.UserAgent = ""
	copy.TokenCacheAppKey = ""

	b, err := json.MarshalIndent(copy, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode config: %w", err)
	}

	// Best-effort to keep file perms private.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return "", fmt.Errorf("write config tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return "", fmt.Errorf("replace config: %w", err)
	}

	return path, nil
}

func IsConfigured(cfg Config) bool {
	return strings.TrimSpace(cfg.SpotifyClientID) != ""
}

func loadFile(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func merge(dst *Config, src *Config) {
	if src == nil {
		return
	}
	if src.SpotifyClientID != "" {
		dst.SpotifyClientID = src.SpotifyClientID
	}
	if src.SpotifyRedirect != "" {
		dst.SpotifyRedirect = src.SpotifyRedirect
	}
	if len(src.SpotifyScopes) > 0 {
		dst.SpotifyScopes = append([]string(nil), src.SpotifyScopes...)
	}
	if src.SpotifyTLSCert != "" {
		dst.SpotifyTLSCert = src.SpotifyTLSCert
	}
	if src.SpotifyTLSKey != "" {
		dst.SpotifyTLSKey = src.SpotifyTLSKey
	}
}

func applyEnvOverrides(cfg *Config) {
	if v := strings.TrimSpace(os.Getenv("SPOTIFY_CLIENT_ID")); v != "" {
		cfg.SpotifyClientID = v
	}
	if v := strings.TrimSpace(os.Getenv("SPOTIFY_REDIRECT_URI")); v != "" {
		cfg.SpotifyRedirect = v
	}
	if v := strings.TrimSpace(os.Getenv("SPOTIFY_SCOPES")); v != "" {
		cfg.SpotifyScopes = splitScopes(v)
	}
	if v := strings.TrimSpace(os.Getenv("SPOTIFY_TLS_CERT_FILE")); v != "" {
		cfg.SpotifyTLSCert = v
	}
	if v := strings.TrimSpace(os.Getenv("SPOTIFY_TLS_KEY_FILE")); v != "" {
		cfg.SpotifyTLSKey = v
	}
}

func normalize(cfg *Config) {
	cfg.SpotifyClientID = strings.TrimSpace(cfg.SpotifyClientID)
	cfg.SpotifyRedirect = strings.TrimSpace(cfg.SpotifyRedirect)
	cfg.SpotifyTLSCert = strings.TrimSpace(cfg.SpotifyTLSCert)
	cfg.SpotifyTLSKey = strings.TrimSpace(cfg.SpotifyTLSKey)

	if cfg.SpotifyRedirect == "" {
		cfg.SpotifyRedirect = Default().SpotifyRedirect
	}
	if len(cfg.SpotifyScopes) == 0 {
		cfg.SpotifyScopes = append([]string(nil), Default().SpotifyScopes...)
	}
}

func applyComputed(cfg *Config) {
	cfg.UserAgent = buildinfo.UserAgent
	cfg.TokenCacheAppKey = buildinfo.AppName
}

func splitScopes(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == ','
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
