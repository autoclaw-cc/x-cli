package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"bing-cli/bing"
	"bing-cli/browser"
	"bing-cli/output"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bing-cli",
	Short: "Bing search CLI powered by browser automation",
	Long: `bing-cli automates Bing Search by running inside the user's real browser
session via kimi-webbridge (http://127.0.0.1:10086). All commands emit JSON on
stdout: {"ok":true,"data":...} on success, {"ok":false,"error":{...}} on failure.
Every command exits non-zero on failure.

No API key required. Results come from real Bing pages, avoiding bot detection.
Use --market to switch between Chinese and international results.

Examples:
  bing-cli search "rust programming"
  bing-cli search "牛仔裤" --market cn
  bing-cli search "AI agents" --market us --count 5`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// ---- search ----
	searchCmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search Bing and return structured results",
		Long: `Navigates to bing.com and extracts the organic result list.

Examples:
  bing-cli search "claude code"
  bing-cli search "英格兰 阿根廷 世界杯" --count 5
  bing-cli search "牛仔裤面料" --market cn --raw
`,
		Args: cobra.MinimumNArgs(1),
		Run:  runSearch,
	}
	searchCmd.Flags().StringP("market", "m", "", "Market: cn (China), us (International), or empty (auto)")
	searchCmd.Flags().IntP("count", "n", 10, "Number of results (max 50)")
	searchCmd.Flags().IntP("offset", "o", 0, "Result offset for pagination")
	searchCmd.Flags().BoolP("raw", "r", false, "Output raw JSON array only (for agent consumption)")
	searchCmd.SilenceUsage = true
	searchCmd.SilenceErrors = true
	rootCmd.AddCommand(searchCmd)

	// ---- result ----
	resultCmd := &cobra.Command{
		Use:   "result <url>",
		Short: "Fetch a page and extract title, description, and text",
		Long: `Navigates to a URL and extracts structured page content.

Examples:
  bing-cli result "https://example.com/article"
`,
		Args: cobra.ExactArgs(1),
		Run:  runResult,
	}
	resultCmd.Flags().BoolP("raw", "r", false, "Output raw JSON only (for agent consumption)")
	resultCmd.SilenceUsage = true
	resultCmd.SilenceErrors = true
	rootCmd.AddCommand(resultCmd)
}

// daemonCheck verifies kimi-webbridge is reachable and the extension is
// connected. Exits with a clear diagnostic on failure.
func daemonCheck(client *browser.Client) {
	st, err := client.Status()
	if err != nil {
		output.Error("daemon_unreachable", err.Error())
		os.Exit(1)
	}
	if !st.Running {
		output.Error("daemon_not_running",
			"kimi-webbridge daemon is not running (open the Kimi Desktop App)")
		os.Exit(1)
	}
	if !st.ExtensionConnected {
		output.Error("extension_not_connected",
			"Chrome WebBridge extension is not connected (see https://www.kimi.com/features/webbridge)")
		os.Exit(1)
	}
}

func runSearch(cmd *cobra.Command, args []string) {
	query := strings.Join(args, " ")
	market, _ := cmd.Flags().GetString("market")
	count, _ := cmd.Flags().GetInt("count")
	offset, _ := cmd.Flags().GetInt("offset")
	raw, _ := cmd.Flags().GetBool("raw")

	if count < 1 {
		count = 1
	}
	if count > 50 {
		count = 50
	}

	client := browser.NewClient("bing-cli")
	defer client.Close()

	daemonCheck(client)

	results, err := bing.Search(client, bing.SearchOptions{
		Query:  query,
		Market: market,
		Count:  count,
		Offset: offset,
	})
	if err != nil {
		var consentErr bing.ErrConsentRequired
		if errors.As(err, &consentErr) {
			output.Error("consent_required", err.Error())
			os.Exit(1)
		}
		if strings.Contains(err.Error(), "daemon unreachable") {
			output.Error("daemon_unreachable", err.Error())
			os.Exit(1)
		}
		output.Error("search_failed", err.Error())
		os.Exit(1)
	}

	if results == nil {
		results = []bing.SearchResult{}
	}

	if raw {
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		enc.Encode(results)
	} else {
		output.Success(map[string]any{
			"query":   query,
			"count":   len(results),
			"results": results,
		})
		fmt.Fprint(os.Stderr, bing.FormatResults(results, query))
	}
}

func runResult(cmd *cobra.Command, args []string) {
	targetURL := args[0]
	raw, _ := cmd.Flags().GetBool("raw")

	client := browser.NewClient("bing-cli")
	defer client.Close()

	daemonCheck(client)

	page, err := bing.FetchResult(client, targetURL)
	if err != nil {
		output.Error("result_failed", err.Error())
		os.Exit(1)
	}

	if raw {
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		enc.Encode(page)
	} else {
		output.Success(page)
	}
}
