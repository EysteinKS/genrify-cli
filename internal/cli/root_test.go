package cli

import (
	"testing"
)

func TestNewRoot(t *testing.T) {
	cmd, root := NewRoot()

	if cmd == nil {
		t.Fatal("expected non-nil command")
	}
	if root == nil {
		t.Fatal("expected non-nil root state")
	}

	if cmd.Use != "genrify" {
		t.Errorf("expected Use='genrify', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if !cmd.SilenceUsage {
		t.Error("expected SilenceUsage to be true")
	}

	// Check that subcommands are registered
	expectedCommands := []string{"version", "login", "start", "playlists"}
	for _, expected := range expectedCommands {
		found := false
		for _, sub := range cmd.Commands() {
			if sub.Use == expected || sub.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q not found", expected)
		}
	}
}

func TestRootPersistentPreRunE(t *testing.T) {
	// This test requires SPOTIFY_CLIENT_ID env var
	// We'll skip if not set to avoid breaking CI
	t.Skip("Skipping test that requires environment configuration")

	// Alternative: test with mocked config
	// This would require refactoring to inject config loader
}
