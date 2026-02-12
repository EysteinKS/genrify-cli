//go:build nogui

package cli

import "github.com/spf13/cobra"

func addGUICmd(cmd *cobra.Command, root *Root) {
}

// defaultCommand returns nil when no default subcommand is desired.
func defaultCommand(root *Root) *cobra.Command {
	return nil
}
