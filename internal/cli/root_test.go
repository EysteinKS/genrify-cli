package cli

import (
	"errors"
	"strings"
	"testing"

	"genrify/internal/config"
	"github.com/spf13/cobra"
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
	cmd, root := NewRoot()
	root.Prompter = &fakePrompter{values: map[string]string{
		"Spotify Client ID":                      "client-id",
		"Spotify Redirect URI":                   "http://localhost:8888/callback",
		"Spotify scopes (space/comma separated)": "playlist-read-private",
		"TLS cert file (for https redirect)":     "",
		"TLS key file (for https redirect)":      "",
	}}
	root.loadConfig = func() (config.Config, error) { return config.Default(), nil }
	var saved config.Config
	root.saveConfig = func(c config.Config) (string, error) { saved = c; return "/tmp/config.json", nil }

	buf := &strings.Builder{}
	cmd.SetOut(buf)

	if err := cmd.PersistentPreRunE(cmd, nil); err != nil {
		t.Fatalf("PersistentPreRunE failed: %v", err)
	}
	if root.Cfg.SpotifyClientID != "client-id" {
		t.Fatalf("expected config to be set on root")
	}
	if saved.SpotifyClientID != "client-id" {
		t.Fatalf("expected config to be saved")
	}
}

func TestRootPersistentPreRunE_SkipsVersion(t *testing.T) {
	cmd, root := NewRoot()
	root.loadConfig = func() (config.Config, error) { return config.Config{}, errors.New("should not load") }

	version := &cobra.Command{Use: "version", Run: func(cmd *cobra.Command, args []string) {}}
	version.SetOut(&strings.Builder{})

	if err := cmd.PersistentPreRunE(version, nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

type fakePrompter struct {
	values map[string]string
}

func (p *fakePrompter) PromptString(label, defaultValue string) (string, error) {
	if v, ok := p.values[label]; ok {
		return v, nil
	}
	return defaultValue, nil
}

func (p *fakePrompter) PromptInt(label string, defaultValue int) (int, error) {
	return defaultValue, nil
}
func (p *fakePrompter) PromptSelect(label string, items []string) (int, string, error) {
	if len(items) == 0 {
		return 0, "", nil
	}
	return 0, items[0], nil
}
