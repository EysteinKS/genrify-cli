package spotify

import (
	"context"
	"testing"
	"time"

	"genrify/internal/auth"
	"genrify/internal/testutil"
)

func TestTokenManager_AccessToken_NoRefreshWhenValid(t *testing.T) {
	store := testutil.NewMemStore(auth.Token{AccessToken: "ok", RefreshToken: "r", ExpiresAt: time.Now().Add(10 * time.Minute)})
	refreshCalls := 0
	m := NewTokenManager(store, 60*time.Second, func(ctx context.Context, refreshToken string) (auth.Token, error) {
		refreshCalls++
		return auth.Token{AccessToken: "new", RefreshToken: refreshToken, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil
	})

	tok, err := m.AccessToken(context.Background())
	if err != nil {
		t.Fatalf("AccessToken error: %v", err)
	}
	if tok != "ok" {
		t.Fatalf("got %q want %q", tok, "ok")
	}
	if refreshCalls != 0 {
		t.Fatalf("expected no refresh calls, got %d", refreshCalls)
	}
}

func TestTokenManager_AccessToken_RefreshesWhenExpired(t *testing.T) {
	store := testutil.NewMemStore(auth.Token{AccessToken: "old", RefreshToken: "r", ExpiresAt: time.Now().Add(-time.Minute)})
	refreshCalls := 0
	m := NewTokenManager(store, 0, func(ctx context.Context, refreshToken string) (auth.Token, error) {
		refreshCalls++
		return auth.Token{AccessToken: "new", RefreshToken: refreshToken, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil
	})

	tok, err := m.AccessToken(context.Background())
	if err != nil {
		t.Fatalf("AccessToken error: %v", err)
	}
	if tok != "new" {
		t.Fatalf("got %q want %q", tok, "new")
	}
	if refreshCalls != 1 {
		t.Fatalf("expected 1 refresh call, got %d", refreshCalls)
	}
}
