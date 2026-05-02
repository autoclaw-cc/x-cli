package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"nanobanana-cli/browser"
	"nanobanana-cli/nanobanana"
	"nanobanana-cli/output"
)

const sessionName = "nanobanana-cli"

var rootCmd = &cobra.Command{
	Use:   "nanobanana-cli",
	Short: "Generate images via Google Gemini (Nano Banana) and save full + thumbnail",
	Long: `nanobanana-cli drives your real Chrome session (via kimi-webbridge) to send a
prompt to Gemini, intercepts the "Download full-size" response chain, and saves
the real high-res PNG (currently 2816×1536) plus a locally-scaled thumbnail.

Requires: Kimi Desktop App running, Chrome extension connected, Gemini logged in.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(newGenCmd())
}

func newGenCmd() *cobra.Command {
	var (
		outDir     string
		thumbWidth int
		timeoutSec int
	)
	c := &cobra.Command{
		Use:   "gen <prompt>",
		Short: "Generate an image from a prompt, save full + thumbnail",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				output.Error("invalid_args", fmt.Sprintf("gen requires exactly one <prompt> argument (got %d)", len(args)))
				os.Exit(1)
			}
			prompt := args[0]
			client := browser.NewClient(sessionName)

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

			res, err := nanobanana.Gen(client, nanobanana.Options{
				Prompt:     prompt,
				OutDir:     outDir,
				ThumbWidth: thumbWidth,
				Timeout:    time.Duration(timeoutSec) * time.Second,
			})
			if err != nil {
				output.Error("gen_failed", err.Error())
				os.Exit(1)
			}
			output.Success(res)
		},
	}
	c.Flags().StringVarP(&outDir, "out", "o", ".", "output directory for *-full.png and *-thumb.png")
	c.Flags().IntVar(&thumbWidth, "thumb-width", 256, "thumbnail width in px (height preserves aspect ratio)")
	c.Flags().IntVar(&timeoutSec, "timeout", 300, "max seconds to wait for image generation (Nano Banana 2 + 'thinking' mode can exceed 2 minutes)")
	// Silence cobra's own error prints; we already emit structured JSON on stderr paths.
	c.SilenceUsage = true
	c.SilenceErrors = true
	return c
}
