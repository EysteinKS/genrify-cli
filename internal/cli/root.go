package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"genrify/internal/config"
)

type Root struct {
	Cfg config.Config
}

func NewRoot() (*cobra.Command, *Root) {
	rootState := &Root{}

	cmd := &cobra.Command{
		Use:          "genrify",
		Short:        "Spotify CLI (playlists)",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			rootState.Cfg = cfg
			return nil
		},
	}

	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newLoginCmd(rootState))
	cmd.AddCommand(newStartCmd(rootState))
	cmd.AddCommand(newPlaylistsCmd(rootState))

	return cmd, rootState
}

func Execute() {
	cmd, _ := NewRoot()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
