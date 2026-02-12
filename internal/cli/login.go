package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"genrify/internal/auth"
	"genrify/internal/certs"
	"genrify/internal/config"
)

func doLogin(ctx context.Context, cfg config.Config) (string, error) {
	// Auto-generate certificates if HTTPS redirect is configured but cert files are missing.
	if strings.HasPrefix(strings.ToLower(cfg.SpotifyRedirect), "https://") {
		if cfg.SpotifyTLSCert == "" || cfg.SpotifyTLSKey == "" {
			certPath, keyPath, err := certs.EnsureCerts("", "")
			if err != nil {
				return "", fmt.Errorf("generate certificates: %w", err)
			}
			cfg.SpotifyTLSCert = certPath
			cfg.SpotifyTLSKey = keyPath
		}
	}

	store, err := auth.NewStore(cfg.TokenCacheAppKey)
	if err != nil {
		return "", fmt.Errorf("create token store: %w", err)
	}

	res, err := auth.LoginPKCE(ctx, auth.OAuthConfig{
		ClientID:    cfg.SpotifyClientID,
		RedirectURI: cfg.SpotifyRedirect,
		Scopes:      cfg.SpotifyScopes,
		UserAgent:   cfg.UserAgent,
		TLSCertFile: cfg.SpotifyTLSCert,
		TLSKeyFile:  cfg.SpotifyTLSKey,
	})
	if err != nil {
		return "", fmt.Errorf("oauth login: %w", err)
	}
	if err := store.Save(res.Token); err != nil {
		return "", fmt.Errorf("save token: %w", err)
	}

	return store.Path(), nil
}

func newLoginCmd(root *Root) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to Spotify (OAuth PKCE)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), LoginTimeout)
			defer cancel()

			if root.doLogin == nil {
				return fmt.Errorf("missing login handler")
			}
			path, err := root.doLogin(ctx, root.Cfg)
			if err != nil {
				return err
			}

			cmd.Println("Logged in successfully.")
			cmd.Println(fmt.Sprintf("Token cache: %s", path))
			return nil
		},
	}
	return cmd
}
