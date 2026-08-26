package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"bing-cli/bing"
	"bing-cli/browser"
	"bing-cli/output"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bing-cli",
	Short: "Bing search CLI powered by browser automation",
	Long: `Bing search CLI using real browser sessions via kimi-webbridge.
No API key required — results from real Bing pages, avoiding bot detection.

Default output is a flat JSON array — designed for agent consumption.
Use --human (-H) to also print readable text on stderr.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	searchCmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search Bing, returns JSON array to stdout",
		Args:  cobra.MinimumNArgs(1),
		Run:   runSearch,
	}
	searchCmd.Flags().StringP("market", "m", "us", "Market: us (default, international), cn (China-local)")
	searchCmd.Flags().IntP("count", "n", 10, "Number of results (max 50)")
	searchCmd.Flags().IntP("offset", "o", 0, "Result offset for pagination")
	searchCmd.Flags().BoolP("unique", "u", false, "Title dedup (>80% similar) + max 3 per domain")
	searchCmd.Flags().BoolP("human", "H", false, "Also print human-readable text to stderr")

	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) {
	query := args[0]
	market, _ := cmd.Flags().GetString("market")
	count, _ := cmd.Flags().GetInt("count")
	offset, _ := cmd.Flags().GetInt("offset")
	unique, _ := cmd.Flags().GetBool("unique")
	human, _ := cmd.Flags().GetBool("human")

	if count < 1 {
		count = 1
	}
	if count > 50 {
		count = 50
	}

	client := browser.NewClient("bing")
	defer client.Close()

	results, err := bing.Search(client, bing.SearchOptions{
		Query:  query,
		Market: market,
		Count:  count,
		Offset: offset,
		Unique: unique,
	})
	if err != nil {
		output.Error("search_error", err.Error())
		os.Exit(1)
	}

	// Default: flat JSON array on stdout (optimized for agent consumption)
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(results); err != nil {
		output.Error("encode_error", err.Error())
		os.Exit(1)
	}

	// Optional human-readable output on stderr
	if human {
		fmt.Fprint(os.Stderr, "\n"+bing.FormatResults(results, query))
	}
}
