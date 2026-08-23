package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"unsplash-cli/output"
	"unsplash-cli/unsplash"
)

var (
	validOrientations = map[string]bool{"landscape": true, "portrait": true, "squarish": true}
	validOrders       = map[string]bool{"relevant": true, "latest": true}
	validColors       = map[string]bool{
		"black_and_white": true, "black": true, "white": true, "yellow": true,
		"orange": true, "red": true, "purple": true, "magenta": true,
		"green": true, "teal": true, "blue": true,
	}
)

func init() {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search Unsplash photos and return structured results",
		Long: `Reads the server-rendered search page at unsplash.com/s/photos/<query>.

Unsplash renders 20 results server-side and loads the rest 20 at a time — first
via a "Load more" button, then by infinite scroll. The CLI drives both, so
--limit can go well past 20 (200 results takes about 15 seconds). If Unsplash
runs out of photos first, the response says so via "truncated" and "note".

Unsplash+ (subscription) assets are excluded by default; pass --include-plus to
keep them.

Examples:
  unsplash-cli search ocean
  unsplash-cli search "misty forest" --limit 100 --orientation landscape
  unsplash-cli search sunset --color orange --order latest
  unsplash-cli search cat --include-plus
`,
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			query := strings.Join(args, " ")

			opts := unsplash.SearchOptions{Query: query}
			opts.Limit, _ = c.Flags().GetInt("limit")
			opts.Orientation, _ = c.Flags().GetString("orientation")
			opts.Color, _ = c.Flags().GetString("color")
			opts.OrderBy, _ = c.Flags().GetString("order")
			opts.IncludePlus, _ = c.Flags().GetBool("include-plus")

			if opts.Orientation != "" && !validOrientations[opts.Orientation] {
				fail("bad_flag", fmt.Sprintf("--orientation must be landscape, portrait or squarish (got %q)", opts.Orientation))
			}
			if opts.Color != "" && !validColors[opts.Color] {
				fail("bad_flag", fmt.Sprintf("--color %q is not one of Unsplash's filters (black_and_white, black, white, yellow, orange, red, purple, magenta, green, teal, blue)", opts.Color))
			}
			if opts.OrderBy != "" && !validOrders[opts.OrderBy] {
				fail("bad_flag", fmt.Sprintf("--order must be relevant or latest (got %q)", opts.OrderBy))
			}

			client := newBrowser()
			res, err := unsplash.Search(client, opts)
			if err != nil {
				fail("search_failed", err.Error())
			}

			data := map[string]any{
				"query":      query,
				"count":      len(res.Photos),
				"requested":  opts.Limit,
				"truncated":  res.Truncated,
				"search_url": res.SearchURL,
				"results":    res.Photos,
			}
			if res.Note != "" {
				data["note"] = res.Note
			}
			output.Success(data)
			return nil
		},
	}

	cmd.Flags().IntP("limit", "n", 20, "max results to return (paginates in batches of 20)")
	cmd.Flags().String("orientation", "", "landscape | portrait | squarish")
	cmd.Flags().String("color", "", "black_and_white | black | white | yellow | orange | red | purple | magenta | green | teal | blue")
	cmd.Flags().String("order", "relevant", "relevant | latest")
	cmd.Flags().Bool("include-plus", false, "include Unsplash+ subscription assets (excluded by default)")
	rootCmd.AddCommand(cmd)
}
