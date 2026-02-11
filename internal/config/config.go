package config

import (
	"fmt"
	"os"
	"strings"

	"genrify/internal/buildinfo"
)

type Config struct {
	SpotifyClientID  string
	SpotifyRedirect  string
	SpotifyScopes    []string
	SpotifyTLSCert   string
	SpotifyTLSKey    string
	UserAgent        string
	TokenCacheAppKey string
}

func Load() (Config, error) {
	clientID := strings.TrimSpace(os.Getenv("SPOTIFY_CLIENT_ID"))
	if clientID == "" {
		return Config{}, fmt.Errorf("SPOTIFY_CLIENT_ID is required")
	}

	redirect := strings.TrimSpace(os.Getenv("SPOTIFY_REDIRECT_URI"))
	if redirect == "" {
		redirect = "http://localhost:8888/callback"
	}

	scopes := strings.TrimSpace(os.Getenv("SPOTIFY_SCOPES"))
	var scopeList []string
	if scopes == "" {
		scopeList = []string{
			"playlist-read-private",
			"playlist-read-collaborative",
			"playlist-modify-private",
			"playlist-modify-public",
		}
	} else {
		scopeList = splitScopes(scopes)
	}

	certFile := strings.TrimSpace(os.Getenv("SPOTIFY_TLS_CERT_FILE"))
	keyFile := strings.TrimSpace(os.Getenv("SPOTIFY_TLS_KEY_FILE"))

	return Config{
		SpotifyClientID:  clientID,
		SpotifyRedirect:  redirect,
		SpotifyScopes:    scopeList,
		SpotifyTLSCert:   certFile,
		SpotifyTLSKey:    keyFile,
		UserAgent:        buildinfo.UserAgent,
		TokenCacheAppKey: buildinfo.AppName,
	}, nil
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
