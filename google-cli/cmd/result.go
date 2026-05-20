package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"google-cli/browser"
	"google-cli/google"
	"google-cli/output"
)

func init() {
	resultCmd := &cobra.Command{
		Use:           "result <url>",
		Short:         "Fetch a page and extract title, description, and text",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				output.Error("missing_args", "result requires a URL: google-cli result <url>")
				os.Exit(1)
			}
			pageURL := args[0]

			client := browser.NewClient("google-cli")

			// Fail fast if the daemon / extension isn't ready.
			st, err := client.Status()
			if err != nil {
				output.Error("daemon_unreachable", err.Error())
				os.Exit(1)
			}
			if !st.Running {
				output.Error("daemon_not_running", "kimi-webbridge daemon is not running (open the Kimi Desktop App)")
				os.Exit(1)
			}
			if !st.ExtensionConnected {
				output.Error("extension_not_connected", "Chrome WebBridge extension is not connected (see https://www.kimi.com/features/webbridge)")
				os.Exit(1)
			}

			page, err := google.FetchResult(client, pageURL)
			if err != nil {
				msg := err.Error()
				switch {
				case strings.HasPrefix(msg, "invalid_url:"):
					output.Error("invalid_url", msg)
				case strings.HasPrefix(msg, "empty_content:"):
					output.Error("empty_content", msg)
				case strings.Contains(msg, "daemon unreachable"):
					output.Error("daemon_unreachable", msg)
				default:
					output.Error("result_failed", msg)
				}
				os.Exit(1)
			}
			output.Success(page)
		},
	}
	rootCmd.AddCommand(resultCmd)
}
