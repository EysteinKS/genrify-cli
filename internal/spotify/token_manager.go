package spotify

import (
	"context"
	"errors"
	"sync"
	"time"

	"genrify/internal/auth"
)

type TokenStore interface {
	Load() (auth.Token, error)
	Save(auth.Token) error
}

type Refresher func(ctx context.Context, refreshToken string) (auth.Token, error)

type TokenManager struct {
	store     TokenStore
	leeway    time.Duration
	refresher Refresher

	mu sync.Mutex
}

func NewTokenManager(store TokenStore, leeway time.Duration, refresher Refresher) *TokenManager {
	return &TokenManager{store: store, leeway: leeway, refresher: refresher}
}

func (m *TokenManager) AccessToken(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, err := m.store.Load()
	if err != nil {
		return "", err
	}
	if t.IsZero() {
		return "", errors.New("not logged in (missing token); run genrify login")
	}
	if !t.Expired(m.leeway) {
		return t.AccessToken, nil
	}

	if t.RefreshToken == "" {
		return "", errors.New("access token expired and no refresh token present; run genrify login")
	}
	if m.refresher == nil {
		return "", errors.New("access token expired and no refresher configured")
	}

	nt, err := m.refresher(ctx, t.RefreshToken)
	if err != nil {
		return "", err
	}
	if err := m.store.Save(nt); err != nil {
		return "", err
	}
	return nt.AccessToken, nil
}

// ForceRefresh refreshes regardless of expiry. Used for retry on 401.
func (m *TokenManager) ForceRefresh(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, err := m.store.Load()
	if err != nil {
		return "", err
	}
	if t.IsZero() {
		return "", errors.New("not logged in (missing token); run genrify login")
	}
	if t.RefreshToken == "" {
		return "", errors.New("missing refresh token; run genrify login")
	}
	if m.refresher == nil {
		return "", errors.New("no refresher configured")
	}

	nt, err := m.refresher(ctx, t.RefreshToken)
	if err != nil {
		return "", err
	}
	if err := m.store.Save(nt); err != nil {
		return "", err
	}
	return nt.AccessToken, nil
}
