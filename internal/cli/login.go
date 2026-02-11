package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"genrify/internal/auth"
)

func newLoginCmd(root *Root) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to Spotify (OAuth PKCE)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), LoginTimeout)
			defer cancel()

			store, err := auth.NewStore(root.Cfg.TokenCacheAppKey)
			if err != nil {
				return fmt.Errorf("create token store: %w", err)
			}

			res, err := auth.LoginPKCE(ctx, auth.OAuthConfig{
				ClientID:    root.Cfg.SpotifyClientID,
				RedirectURI: root.Cfg.SpotifyRedirect,
				Scopes:      root.Cfg.SpotifyScopes,
				UserAgent:   root.Cfg.UserAgent,
				TLSCertFile: root.Cfg.SpotifyTLSCert,
				TLSKeyFile:  root.Cfg.SpotifyTLSKey,
			})
			if err != nil {
				return fmt.Errorf("oauth login: %w", err)
			}
			if err := store.Save(res.Token); err != nil {
				return fmt.Errorf("save token: %w", err)
			}

			cmd.Println("Logged in successfully.")
			cmd.Println(fmt.Sprintf("Token cache: %s", store.Path()))
			return nil
		},
	}
	return cmd
}
