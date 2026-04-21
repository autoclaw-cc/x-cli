package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"baidu-cli/baidu"
	"baidu-cli/browser"
	"baidu-cli/output"
)

const sessionName = "baidu"

var rootCmd = &cobra.Command{
	Use:   "baidu-cli",
	Short: "Automate baidu search via the kimi-webbridge daemon",
	Long: `baidu-cli drives a real Chrome tab through the kimi-webbridge daemon to
query baidu.com. Because it uses your actual browser session, there's no
cookie/header dance — the SERP is fetched with full user context.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	searchCmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Run a baidu web search and return structured results",
		Long: `Navigates to https://www.baidu.com/s?wd=<query> and extracts the
result list from the SSR'd DOM.

Examples:
  baidu-cli search "claude code"
  baidu-cli search "天气 北京" --limit 20
  baidu-cli search "大模型" --all
`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			query := strings.Join(args, " ")
			limit, _ := cmd.Flags().GetInt("limit")
			includeAll, _ := cmd.Flags().GetBool("all")

			client := browser.NewClient(sessionName)
			results, err := baidu.Search(client, query, limit, includeAll)
			if err != nil {
				output.Error("search_failed", err.Error())
				os.Exit(1)
			}
			output.Success(map[string]any{
				"query":   query,
				"count":   len(results),
				"results": results,
			})
		},
	}
	searchCmd.Flags().IntP("limit", "n", 10, "max results to return")
	searchCmd.Flags().Bool("all", false, "include aladdin cards / filtered tpls (bypass organic filter)")
	rootCmd.AddCommand(searchCmd)
}
