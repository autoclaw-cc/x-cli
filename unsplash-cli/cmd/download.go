package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"unsplash-cli/browser"
	"unsplash-cli/output"
	"unsplash-cli/unsplash"
)

var validFormats = map[string]bool{"jpg": true, "png": true, "webp": true}

func init() {
	cmd := &cobra.Command{
		Use:   "download <id|url> [id|url ...]",
		Short: "Download Unsplash photos to disk",
		Long: `Downloads one or more photos. Each argument may be:

  - a bare photo id            uxhmO7BumvA
  - a photo page URL           https://unsplash.com/photos/body-of-water-under-sky-6ArTTluciuA
  - a CDN image URL            https://images.unsplash.com/photo-1518837695005-2083093ee35b

The CDN form needs no browser at all, so piping the image_url out of
"unsplash-cli search" is the fastest path. Bare ids and page URLs cost one
browser page load each to resolve.

Omit --width to get the original file; pass one to have Unsplash's imgix
resize it (never upscaled).

Examples:
  unsplash-cli download 6ArTTluciuA
  unsplash-cli download 6ArTTluciuA --out ~/Pictures/unsplash --width 1920
  unsplash-cli search ocean -n 5 | jq -r '.data.results[].image_url' | xargs unsplash-cli download --out ./shots
`,
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			dir, _ := c.Flags().GetString("out")
			width, _ := c.Flags().GetInt("width")
			quality, _ := c.Flags().GetInt("quality")
			format, _ := c.Flags().GetString("format")
			force, _ := c.Flags().GetBool("force")

			if format != "" && !validFormats[format] {
				fail("bad_flag", fmt.Sprintf("--format must be jpg, png or webp (got %q)", format))
			}
			if quality < 0 || quality > 100 {
				fail("bad_flag", fmt.Sprintf("--quality must be 0-100 (got %d)", quality))
			}

			size := unsplash.SizeSpec{Width: width, Quality: quality, Format: format}

			// The browser is opened lazily: a run made entirely of CDN URLs
			// should not need Chrome at all.
			var client *browser.Client
			ensureBrowser := func() *browser.Client {
				if client == nil {
					client = newBrowser()
				}
				return client
			}

			var (
				results []map[string]any
				failed  []map[string]any
			)

			for _, ref := range args {
				imageURL, id, err := unsplash.ParseRef(ref)
				if err != nil {
					failed = append(failed, map[string]any{"ref": ref, "error": err.Error()})
					continue
				}

				slug := ""
				if imageURL == "" {
					meta, err := unsplash.FetchMeta(ensureBrowser(), id)
					if err != nil {
						failed = append(failed, map[string]any{"ref": ref, "error": err.Error()})
						continue
					}
					imageURL, slug = meta.ImageURL, meta.Slug
				}
				if id == "" {
					// A CDN URL carries no photo id; fall back to the imgix
					// asset name so the file is still traceable to its source.
					id = unsplash.AssetName(imageURL)
				}

				ext := format
				if ext == "" {
					ext = "jpg"
				}
				res, err := unsplash.Download(
					unsplash.BuildImageURL(imageURL, size),
					dir,
					unsplash.Filename(id, slug, ext),
					force,
				)
				if err != nil {
					failed = append(failed, map[string]any{"ref": ref, "error": err.Error()})
					continue
				}
				results = append(results, map[string]any{
					"ref":     ref,
					"id":      id,
					"path":    res.Path,
					"bytes":   res.Bytes,
					"skipped": res.Skipped,
					"source":  res.SourceURL,
				})
			}

			data := map[string]any{
				"dir":        dir,
				"downloaded": len(results),
				"failed":     len(failed),
				"results":    results,
			}
			if len(failed) > 0 {
				data["errors"] = failed
			}

			// Any failure is a non-zero exit, but successful downloads are
			// still reported — a partial batch shouldn't be silently discarded.
			if len(failed) > 0 && len(results) == 0 {
				output.Error("download_failed", fmt.Sprintf("all %d download(s) failed: %v", len(failed), failed[0]["error"]))
				os.Exit(1)
			}
			output.Success(data)
			if len(failed) > 0 {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().StringP("out", "o", ".", "directory to write files into")
	cmd.Flags().IntP("width", "w", 0, "resize to this width in px (0 = original resolution)")
	cmd.Flags().IntP("quality", "q", 0, "JPEG/WebP quality 1-100 (0 = Unsplash default)")
	cmd.Flags().String("format", "", "jpg | png | webp (default: source format, saved as .jpg)")
	cmd.Flags().BoolP("force", "f", false, "re-download even if the file already exists")
	rootCmd.AddCommand(cmd)
}
