package auth

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestNewPKCE(t *testing.T) {
	p, err := newPKCE()
	if err != nil {
		t.Fatalf("newPKCE() failed: %v", err)
	}

	if p.Verifier == "" {
		t.Error("expected non-empty verifier")
	}
	if p.Challenge == "" {
		t.Error("expected non-empty challenge")
	}

	// Verifier should be base64 URL-safe encoded (no padding)
	if strings.Contains(p.Verifier, "=") {
		t.Error("verifier should not contain padding")
	}

	// Challenge should be base64 URL-safe encoded (no padding)
	if strings.Contains(p.Challenge, "=") {
		t.Error("challenge should not contain padding")
	}

	// Decode verifier to check length
	decoded, err := base64.RawURLEncoding.DecodeString(p.Verifier)
	if err != nil {
		t.Errorf("verifier is not valid base64: %v", err)
	}
	if len(decoded) != 64 {
		t.Errorf("expected verifier to be 64 bytes, got %d", len(decoded))
	}

	// Challenge should be different from verifier (it's a hash)
	if p.Challenge == p.Verifier {
		t.Error("challenge should be different from verifier")
	}
}

func TestNewPKCE_Uniqueness(t *testing.T) {
	// Generate multiple PKCEs and ensure they're unique
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		p, err := newPKCE()
		if err != nil {
			t.Fatalf("newPKCE() iteration %d failed: %v", i, err)
		}
		if seen[p.Verifier] {
			t.Errorf("duplicate verifier generated: %s", p.Verifier)
		}
		seen[p.Verifier] = true
	}
}

func TestRandomURLSafe(t *testing.T) {
	s, err := randomURLSafe(32)
	if err != nil {
		t.Fatalf("randomURLSafe(32) failed: %v", err)
	}

	if s == "" {
		t.Error("expected non-empty string")
	}

	// Should be base64 URL-safe encoded (no padding)
	if strings.Contains(s, "=") {
		t.Error("result should not contain padding")
	}
	if strings.Contains(s, "+") || strings.Contains(s, "/") {
		t.Error("result should use URL-safe encoding (no + or /)")
	}

	// Decode to verify length
	decoded, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		t.Errorf("result is not valid base64: %v", err)
	}
	if len(decoded) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(decoded))
	}
}

func TestRandomURLSafe_DifferentLengths(t *testing.T) {
	tests := []int{16, 32, 64, 128}
	for _, n := range tests {
		s, err := randomURLSafe(n)
		if err != nil {
			t.Errorf("randomURLSafe(%d) failed: %v", n, err)
			continue
		}

		decoded, err := base64.RawURLEncoding.DecodeString(s)
		if err != nil {
			t.Errorf("randomURLSafe(%d) result is not valid base64: %v", n, err)
			continue
		}
		if len(decoded) != n {
			t.Errorf("randomURLSafe(%d): expected %d bytes, got %d", n, n, len(decoded))
		}
	}
}

func TestRandomURLSafe_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		s, err := randomURLSafe(16)
		if err != nil {
			t.Fatalf("randomURLSafe(16) iteration %d failed: %v", i, err)
		}
		if seen[s] {
			t.Errorf("duplicate value generated: %s", s)
		}
		seen[s] = true
	}
}
