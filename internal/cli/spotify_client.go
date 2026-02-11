package cli

import (
	"context"
	"fmt"
	"net/http"

	"genrify/internal/auth"
	"genrify/internal/config"
	"genrify/internal/spotify"
)

func newSpotifyClient(cfg config.Config) (*spotify.Client, error) {
	store, err := auth.NewStore(cfg.TokenCacheAppKey)
	if err != nil {
		return nil, err
	}

	refresher := func(ctx context.Context, refreshToken string) (auth.Token, error) {
		return auth.Refresh(ctx, auth.OAuthConfig{
			ClientID:  cfg.SpotifyClientID,
			UserAgent: cfg.UserAgent,
		}, refreshToken)
	}

	// Refresh when we're within 60s of expiry.
	m := spotify.NewTokenManager(store, TokenRefreshLeeway, refresher)

	c, err := spotify.New(
		spotify.WithHTTPClient(&http.Client{Timeout: DefaultHTTPTimeout}),
		spotify.WithUserAgent(cfg.UserAgent),
		spotify.WithTokenManager(m),
	)
	if err != nil {
		return nil, fmt.Errorf("create spotify client: %w", err)
	}
	return c, nil
}
