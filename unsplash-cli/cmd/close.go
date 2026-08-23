package cmd

import (
	"github.com/spf13/cobra"

	"unsplash-cli/output"
)

func init() {
	cmd := &cobra.Command{
		Use:   "close",
		Short: "Close the Chrome tab group this CLI opened",
		Long: `Closes every tab in the "unsplash-cli" session.

The search commands deliberately reuse one tab rather than opening a new one
per run, and leaving it open keeps the Anubis bot-check cookie warm for the
next command. Run this when you're done and want your browser tidy.`,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			client := newBrowser()
			if err := client.CloseSession(); err != nil {
				fail("close_failed", err.Error())
			}
			output.Success(map[string]any{"closed": true, "session": sessionName})
			return nil
		},
	}
	rootCmd.AddCommand(cmd)
}
