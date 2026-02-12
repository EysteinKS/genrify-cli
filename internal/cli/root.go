package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"genrify/internal/config"
)

type Root struct {
	Cfg      config.Config
	Prompter Prompter

	loadConfig func() (config.Config, error)
	saveConfig func(config.Config) (string, error)

	newSpotifyClient   func(config.Config) (SpotifyClient, error)
	doLogin            func(context.Context, config.Config) (string, error)
	runInteractiveLoop func(context.Context, SpotifyClient, Prompter) error
}

func NewRoot() (*cobra.Command, *Root) {
	rootState := &Root{
		Prompter:   NewPrompter(),
		loadConfig: config.Load,
		saveConfig: config.Save,
		newSpotifyClient: func(cfg config.Config) (SpotifyClient, error) {
			return newSpotifyClient(cfg)
		},
		doLogin:            doLogin,
		runInteractiveLoop: runInteractiveLoop,
	}

	cmd := &cobra.Command{
		Use:          "genrify",
		Short:        "Spotify CLI (playlists)",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Allow `genrify version` without requiring config.
			if cmd.Name() == "version" {
				return nil
			}

			cfg, err := rootState.loadConfig()
			if err != nil {
				return err
			}

			if !config.IsConfigured(cfg) {
				if err := promptForSpotifyConfig(rootState.Prompter, &cfg); err != nil {
					return err
				}
				path, err := rootState.saveConfig(cfg)
				if err != nil {
					return err
				}
				cmd.Println("Saved config to: " + path)
			}

			rootState.Cfg = cfg
			return nil
		},
	}

	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newLoginCmd(rootState))
	startCmd := newStartCmd(rootState)
	cmd.AddCommand(startCmd)
	cmd.AddCommand(newPlaylistsCmd(rootState))

	// Default to `start` when no subcommand is provided.
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		startCmd.SetContext(cmd.Context())
		startCmd.SetOut(cmd.OutOrStdout())
		startCmd.SetErr(cmd.ErrOrStderr())
		if startCmd.RunE == nil {
			return fmt.Errorf("start command missing RunE")
		}
		return startCmd.RunE(startCmd, args)
	}

	return cmd, rootState
}

func promptForSpotifyConfig(p Prompter, cfg *config.Config) error {
	if p == nil {
		return fmt.Errorf("missing prompter")
	}

	clientID, err := p.PromptString("Spotify Client ID", cfg.SpotifyClientID)
	if err != nil {
		return err
	}
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return fmt.Errorf("spotify client id is required")
	}
	cfg.SpotifyClientID = clientID

	redirect, err := p.PromptString("Spotify Redirect URI", cfg.SpotifyRedirect)
	if err != nil {
		return err
	}
	redirect = strings.TrimSpace(redirect)
	if redirect == "" {
		redirect = config.Default().SpotifyRedirect
	}
	cfg.SpotifyRedirect = redirect

	defaultScopes := strings.Join(cfg.SpotifyScopes, " ")
	scopesStr, err := p.PromptString("Spotify scopes (space/comma separated)", defaultScopes)
	if err != nil {
		return err
	}
	scopesStr = strings.TrimSpace(scopesStr)
	if scopesStr == "" {
		cfg.SpotifyScopes = append([]string(nil), config.Default().SpotifyScopes...)
	} else {
		cfg.SpotifyScopes = configScopes(scopesStr)
	}

	if strings.HasPrefix(strings.ToLower(cfg.SpotifyRedirect), "https://") {
		cert, err := p.PromptString("TLS cert file (for https redirect)", cfg.SpotifyTLSCert)
		if err != nil {
			return err
		}
		key, err := p.PromptString("TLS key file (for https redirect)", cfg.SpotifyTLSKey)
		if err != nil {
			return err
		}
		cfg.SpotifyTLSCert = strings.TrimSpace(cert)
		cfg.SpotifyTLSKey = strings.TrimSpace(key)
		if cfg.SpotifyTLSCert == "" || cfg.SpotifyTLSKey == "" {
			return fmt.Errorf("https redirect requires TLS cert and key files")
		}
	}

	return nil
}

func configScopes(s string) []string {
	// Reuse config's scope splitting by going through the env override path.
	// This keeps behavior consistent without exporting internal helpers.
	// Note: config.Load() will normalize defaults when scopes are empty.
	// Here we only want the split.
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == ','
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func Execute() {
	cmd, _ := NewRoot()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
