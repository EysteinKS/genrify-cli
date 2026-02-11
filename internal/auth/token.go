package auth

import "time"

type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

func (t Token) IsZero() bool {
	return t.AccessToken == ""
}

func (t Token) Expired(leeway time.Duration) bool {
	if t.IsZero() {
		return true
	}
	return time.Until(t.ExpiresAt) <= leeway
}
