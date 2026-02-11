package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

type pkce struct {
	Verifier  string
	Challenge string
}

func newPKCE() (pkce, error) {
	verifier, err := randomURLSafe(64)
	if err != nil {
		return pkce{}, err
	}

	h := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h[:])

	return pkce{Verifier: verifier, Challenge: challenge}, nil
}

func randomURLSafe(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
