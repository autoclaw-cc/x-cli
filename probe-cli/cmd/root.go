package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"probe-cli/browser"
	"probe-cli/output"
	"probe-cli/probe"
)

var (
	sessionName string
	quickMode   bool
)

var rootCmd = &cobra.Command{
	Use:   "probe-cli <url>",
	Short: "Auto-detect site auth, API endpoints, and DOM elements for x-cli development",
	Long: `probe-cli automates the Site Archaeology protocol from agent-cli-creator.

It connects to the kimi-webbridge daemon, navigates to the target URL,
and auto-detects:
  - Authentication mechanism (localStorage, cookie, sessionStorage, CSRF)
  - API endpoints (XHR/Fetch with auth header analysis)
  - DOM elements (forms, inputs, buttons)
  - Recommended CLI pattern (dom-scrape, api-reverse, form-submit, async-poll)

Output is a JSON site profile ready for CLI implementation.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		targetURL := args[0]

		// Use URL hostname as session name if not specified
		sess := sessionName
		if sess == "" {
			sess = "probe"
		}

		client := browser.NewClient(sess)

		// Check daemon is running
		if _, err := client.Status(); err != nil {
			output.Error("daemon_unreachable", fmt.Sprintf(
				"kimi-webbridge daemon not running at %s. Start it first.",
				browser.DefaultDaemonURL,
			))
			os.Exit(1)
		}

		var profile *probe.SiteProfile
		var err error

		if quickMode {
			profile, err = probe.RunQuick(client, targetURL)
		} else {
			profile, err = probe.Run(client, targetURL)
		}

		if err != nil {
			output.Error("probe_failed", err.Error())
			os.Exit(1)
		}

		output.Success(profile)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&sessionName, "session", "s", "", "kimi-webbridge session name (default: probe)")
	rootCmd.Flags().BoolVarP(&quickMode, "quick", "q", false, "skip network capture (DOM + auth only)")
}
