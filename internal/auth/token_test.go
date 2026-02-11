package auth

import (
	"testing"
	"time"
)

func TestToken_IsZero(t *testing.T) {
	tests := []struct {
		name  string
		token Token
		want  bool
	}{
		{
			name:  "empty token",
			token: Token{},
			want:  true,
		},
		{
			name: "token with only access token",
			token: Token{
				AccessToken: "abc123",
			},
			want: false,
		},
		{
			name: "full token",
			token: Token{
				AccessToken:  "abc123",
				TokenType:    "Bearer",
				Scope:        "user-read-private",
				ExpiresAt:    time.Now().Add(time.Hour),
				RefreshToken: "refresh123",
			},
			want: false,
		},
		{
			name: "token with other fields but no access token",
			token: Token{
				TokenType:    "Bearer",
				RefreshToken: "refresh123",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToken_Expired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name   string
		token  Token
		leeway time.Duration
		want   bool
	}{
		{
			name:   "zero token",
			token:  Token{},
			leeway: 0,
			want:   true,
		},
		{
			name: "expired token",
			token: Token{
				AccessToken: "abc123",
				ExpiresAt:   now.Add(-time.Hour),
			},
			leeway: 0,
			want:   true,
		},
		{
			name: "valid token",
			token: Token{
				AccessToken: "abc123",
				ExpiresAt:   now.Add(time.Hour),
			},
			leeway: 0,
			want:   false,
		},
		{
			name: "token expiring soon without leeway",
			token: Token{
				AccessToken: "abc123",
				ExpiresAt:   now.Add(30 * time.Second),
			},
			leeway: 0,
			want:   false,
		},
		{
			name: "token expiring soon with leeway",
			token: Token{
				AccessToken: "abc123",
				ExpiresAt:   now.Add(30 * time.Second),
			},
			leeway: time.Minute,
			want:   true,
		},
		{
			name: "token not expired but within leeway",
			token: Token{
				AccessToken: "abc123",
				ExpiresAt:   now.Add(45 * time.Second),
			},
			leeway: time.Minute,
			want:   true,
		},
		{
			name: "token well beyond leeway",
			token: Token{
				AccessToken: "abc123",
				ExpiresAt:   now.Add(2 * time.Minute),
			},
			leeway: time.Minute,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.Expired(tt.leeway); got != tt.want {
				t.Errorf("Expired(%v) = %v, want %v (expiresAt=%v, now=%v)",
					tt.leeway, got, tt.want, tt.token.ExpiresAt, now)
			}
		})
	}
}

func TestToken_Expired_EdgeCases(t *testing.T) {
	now := time.Now()

	// Test exact expiry time
	token := Token{
		AccessToken: "abc123",
		ExpiresAt:   now,
	}
	if !token.Expired(0) {
		t.Error("token expiring exactly now should be expired")
	}

	// Test with negative leeway (should still work)
	token = Token{
		AccessToken: "abc123",
		ExpiresAt:   now.Add(time.Hour),
	}
	if token.Expired(-time.Hour) {
		t.Error("token should not be expired with negative leeway")
	}
}
