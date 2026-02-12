//go:build nogui

package cli

import "github.com/spf13/cobra"

// defaultCommand returns nil when GUI is disabled, causing the CLI to fall back to start command.
func defaultCommand(root *Root) *cobra.Command {
	return nil
}
