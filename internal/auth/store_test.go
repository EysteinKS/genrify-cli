package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStore_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "testapp", "token.json")

	// Create store manually with temp directory
	if err := os.MkdirAll(filepath.Dir(storePath), 0o700); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	store := &Store{path: storePath}

	// Create a test token
	expiresAt := time.Now().Add(time.Hour).Round(time.Second)
	testToken := Token{
		AccessToken:  "test-access",
		TokenType:    "Bearer",
		Scope:        "user-read-private playlist-modify-public",
		ExpiresAt:    expiresAt,
		RefreshToken: "test-refresh",
	}

	// Save the token
	if err := store.Save(testToken); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		t.Fatal("token file was not created")
	}

	// Load the token back
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify loaded token matches
	if loaded.AccessToken != testToken.AccessToken {
		t.Errorf("AccessToken mismatch: got %q, want %q", loaded.AccessToken, testToken.AccessToken)
	}
	if loaded.TokenType != testToken.TokenType {
		t.Errorf("TokenType mismatch: got %q, want %q", loaded.TokenType, testToken.TokenType)
	}
	if loaded.Scope != testToken.Scope {
		t.Errorf("Scope mismatch: got %q, want %q", loaded.Scope, testToken.Scope)
	}
	if loaded.RefreshToken != testToken.RefreshToken {
		t.Errorf("RefreshToken mismatch: got %q, want %q", loaded.RefreshToken, testToken.RefreshToken)
	}
	// Round to seconds for comparison since JSON doesn't preserve nanoseconds perfectly
	if !loaded.ExpiresAt.Round(time.Second).Equal(testToken.ExpiresAt.Round(time.Second)) {
		t.Errorf("ExpiresAt mismatch: got %v, want %v", loaded.ExpiresAt, testToken.ExpiresAt)
	}
}

func TestStore_Load_FileNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "testapp", "token.json")

	if err := os.MkdirAll(filepath.Dir(storePath), 0o700); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	store := &Store{path: storePath}

	// Load when file doesn't exist should return zero token
	token, err := store.Load()
	if err != nil {
		t.Fatalf("Load() with no file should not error: %v", err)
	}
	if !token.IsZero() {
		t.Error("expected zero token when file doesn't exist")
	}
}

func TestStore_Load_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "testapp", "token.json")

	if err := os.MkdirAll(filepath.Dir(storePath), 0o700); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	store := &Store{path: storePath}

	// Write invalid JSON
	if err := os.WriteFile(storePath, []byte("not valid json"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Load should return error
	_, err := store.Load()
	if err == nil {
		t.Fatal("Load() should return error for invalid JSON")
	}
}

func TestStore_Save_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "testapp", "token.json")

	if err := os.MkdirAll(filepath.Dir(storePath), 0o700); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	store := &Store{path: storePath}

	token := Token{
		AccessToken:  "test",
		ExpiresAt:    time.Now(),
		RefreshToken: "refresh",
	}

	if err := store.Save(token); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file exists and has correct permissions
	info, err := os.Stat(storePath)
	if err != nil {
		t.Fatalf("token file was not created: %v", err)
	}

	// Check file is readable
	data, err := os.ReadFile(storePath)
	if err != nil {
		t.Fatalf("failed to read token file: %v", err)
	}

	// Verify it's valid JSON
	var loaded Token
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("token file is not valid JSON: %v", err)
	}

	// Verify it's formatted (indented)
	if len(data) < 50 { // A minimal indented JSON should be at least this long
		t.Error("expected indented JSON output")
	}

	// File should be private (mode 0600 or similar)
	mode := info.Mode()
	if mode.Perm()&0o077 != 0 {
		t.Errorf("token file has too permissive mode: %v", mode)
	}
}

func TestStore_Save_UpdatesExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "testapp", "token.json")

	if err := os.MkdirAll(filepath.Dir(storePath), 0o700); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	store := &Store{path: storePath}

	// Save first token
	token1 := Token{
		AccessToken:  "first",
		RefreshToken: "refresh1",
		ExpiresAt:    time.Now(),
	}
	if err := store.Save(token1); err != nil {
		t.Fatalf("Save() first token failed: %v", err)
	}

	// Save second token (update)
	token2 := Token{
		AccessToken:  "second",
		RefreshToken: "refresh2",
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	if err := store.Save(token2); err != nil {
		t.Fatalf("Save() second token failed: %v", err)
	}

	// Load and verify it's the second token
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if loaded.AccessToken != "second" {
		t.Errorf("expected token to be updated, got %q", loaded.AccessToken)
	}
}

func TestNewStore(t *testing.T) {
	// This test uses the actual user config directory
	// It's not ideal but tests the real behavior
	store, err := NewStore("genrify-test-" + time.Now().Format("20060102150405"))
	if err != nil {
		t.Fatalf("NewStore() failed: %v", err)
	}

	// Verify path is set
	if store.Path() == "" {
		t.Error("expected non-empty path")
	}

	// Verify directory exists
	dir := filepath.Dir(store.Path())
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("store directory was not created: %s", dir)
	}

	// Clean up
	os.RemoveAll(dir)
}

func TestStore_Path(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "testapp", "token.json")

	if err := os.MkdirAll(filepath.Dir(storePath), 0o700); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	store := &Store{path: storePath}

	if got := store.Path(); got != storePath {
		t.Errorf("Path() = %q, want %q", got, storePath)
	}
}
