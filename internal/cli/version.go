package cli

import "github.com/spf13/cobra"

import "genrify/internal/buildinfo"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(buildinfo.AppName + " " + buildinfo.Version)
		},
		// Skip config loading for version command.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
