package testutil

import (
	"sync"

	"genrify/internal/auth"
)

// MemStore is an in-memory token store for testing.
type MemStore struct {
	mu sync.Mutex
	t  auth.Token
}

// NewMemStore creates a new in-memory token store.
func NewMemStore(initial auth.Token) *MemStore {
	return &MemStore{t: initial}
}

// Load returns the stored token.
func (s *MemStore) Load() (auth.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.t, nil
}

// Save saves the token.
func (s *MemStore) Save(t auth.Token) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.t = t
	return nil
}
