package cmd

import (
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"google-cli/browser"
	"google-cli/google"
	"google-cli/output"
)

func init() {
	searchCmd := &cobra.Command{
		Use:           "search <query>",
		Short:         "Search Google and return structured results",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				output.Error("missing_args", "search requires a query: google-cli search <query>")
				os.Exit(1)
			}
			query := strings.Join(args, " ")
			limit, _ := cmd.Flags().GetInt("limit")
			hl, _ := cmd.Flags().GetString("hl")

			client := browser.NewClient("google-cli")
			results, err := google.FetchSearch(client, query, limit, hl)
			if err != nil {
				var consent google.ErrConsentRequired
				if errors.As(err, &consent) {
					output.Error("consent_required", err.Error())
				} else if strings.Contains(err.Error(), "daemon unreachable") {
					output.Error("daemon_unreachable", err.Error())
				} else {
					output.Error("search_failed", err.Error())
				}
				os.Exit(1)
			}
			if len(results) == 0 {
				output.Error("no_results", "google returned no parseable results (selectors may have drifted — see ARCHAEOLOGY.md)")
				os.Exit(1)
			}
			output.Success(results)
		},
	}
	searchCmd.Flags().Int("limit", 10, "maximum number of results to return")
	searchCmd.Flags().String("hl", "en", "Google UI language (en/zh-CN/etc.) — affects DOM stability")
	rootCmd.AddCommand(searchCmd)
}
