package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewPlaylistsCmd(t *testing.T) {
	root := &Root{}
	cmd := newPlaylistsCmd(root)

	if cmd.Use != "playlists" {
		t.Errorf("expected Use='playlists', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	// Check that subcommands are registered
	expectedCommands := []string{"list", "tracks", "create", "add", "merge"}
	for _, expected := range expectedCommands {
		found := false
		for _, sub := range cmd.Commands() {
			if sub.Use == expected || strings.HasPrefix(sub.Use, expected+" ") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q not found", expected)
		}
	}
}

func TestNewPlaylistsListCmd(t *testing.T) {
	root := &Root{}
	cmd := newPlaylistsListCmd(root)

	if cmd.Use != "list" {
		t.Errorf("expected Use='list', got %q", cmd.Use)
	}

	// Check flags exist
	filterFlag := cmd.Flags().Lookup("filter")
	if filterFlag == nil {
		t.Error("expected 'filter' flag to exist")
	}

	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag == nil {
		t.Error("expected 'limit' flag to exist")
	}
}

func TestNewPlaylistsTracksCmd(t *testing.T) {
	root := &Root{}
	cmd := newPlaylistsTracksCmd(root)

	if !strings.HasPrefix(cmd.Use, "tracks") {
		t.Errorf("expected Use to start with 'tracks', got %q", cmd.Use)
	}

	// Check flags exist
	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag == nil {
		t.Error("expected 'limit' flag to exist")
	}

	urisFlag := cmd.Flags().Lookup("uris")
	if urisFlag == nil {
		t.Error("expected 'uris' flag to exist")
	}
}

func TestNewPlaylistsCreateCmd(t *testing.T) {
	root := &Root{}
	cmd := newPlaylistsCreateCmd(root)

	if cmd.Use != "create" {
		t.Errorf("expected Use='create', got %q", cmd.Use)
	}

	// Check flags exist
	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Error("expected 'name' flag to exist")
	}

	descFlag := cmd.Flags().Lookup("description")
	if descFlag == nil {
		t.Error("expected 'description' flag to exist")
	}

	publicFlag := cmd.Flags().Lookup("public")
	if publicFlag == nil {
		t.Error("expected 'public' flag to exist")
	}
}

func TestNewPlaylistsAddCmd(t *testing.T) {
	root := &Root{}
	cmd := newPlaylistsAddCmd(root)

	if !strings.HasPrefix(cmd.Use, "add") {
		t.Errorf("expected Use to start with 'add', got %q", cmd.Use)
	}
}

// TestCommandValidation tests that commands require correct number of args
func TestCommandValidation(t *testing.T) {
	tests := []struct {
		name        string
		cmdFunc     func(*Root) *cobra.Command
		args        []string
		expectError bool
	}{
		{
			name:        "tracks command requires playlist id",
			cmdFunc:     newPlaylistsTracksCmd,
			args:        []string{},
			expectError: true,
		},
		{
			name:        "tracks command accepts one arg",
			cmdFunc:     newPlaylistsTracksCmd,
			args:        []string{"playlist123"},
			expectError: false, // Will fail at execution but not validation
		},
		{
			name:        "add command requires at least 2 args",
			cmdFunc:     newPlaylistsAddCmd,
			args:        []string{"playlist123"},
			expectError: true,
		},
		{
			name:        "add command accepts 2+ args",
			cmdFunc:     newPlaylistsAddCmd,
			args:        []string{"playlist123", "track1"},
			expectError: false, // Will fail at execution but not validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &Root{}
			cmd := tt.cmdFunc(root)
			cmd.SetArgs(tt.args)

			// We're only testing Args validation, not execution
			// So we check if ValidateArgs returns an error
			err := cmd.ValidateArgs(tt.args)
			if tt.expectError && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no validation error, got: %v", err)
			}
		})
	}
}
