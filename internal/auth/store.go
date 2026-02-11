package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Store struct {
	path string
}

func NewStore(appName string) (*Store, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("get user config dir: %w", err)
	}

	base := filepath.Join(dir, appName)
	if err := os.MkdirAll(base, 0o700); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", base, err)
	}

	return &Store{path: filepath.Join(base, "token.json")}, nil
}

func (s *Store) Load() (Token, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return Token{}, nil
		}
		return Token{}, fmt.Errorf("read token cache: %w", err)
	}

	var t Token
	if err := json.Unmarshal(b, &t); err != nil {
		return Token{}, fmt.Errorf("parse token cache: %w", err)
	}
	return t, nil
}

func (s *Store) Save(t Token) error {
	b, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("encode token: %w", err)
	}

	// Best-effort to keep file perms private.
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return fmt.Errorf("write token tmp: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("replace token cache: %w", err)
	}
	return nil
}

func (s *Store) Path() string {
	return s.path
}
