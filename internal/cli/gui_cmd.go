//go:build !nogui

package cli

import (
	"github.com/spf13/cobra"

	"genrify/internal/config"
	"genrify/internal/gui"
)

// newGUICmd creates a command that launches the GTK GUI.
func newGUICmd(root *Root) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gui",
		Short: "Launch graphical user interface",
		RunE: func(cmd *cobra.Command, args []string) error {
			return gui.Run(root.Cfg, gui.Options{
				DoLogin: root.doLogin,
				NewSpotifyClient: func(cfg config.Config) (gui.SpotifyClient, error) {
					return root.newSpotifyClient(cfg)
				},
				LoadConfig: root.loadConfig,
				SaveConfig: root.saveConfig,
			})
		},
	}
	return cmd
}

// defaultCommand returns the GUI command as the default when no subcommand is specified.
func defaultCommand(root *Root) *cobra.Command {
	return newGUICmd(root)
}
