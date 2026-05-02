package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"chatgpt-image-cli/browser"
	"chatgpt-image-cli/chatgpt"
	"chatgpt-image-cli/output"
)

const sessionName = "chatgpt-image-cli"

var rootCmd = &cobra.Command{
	Use:   "chatgpt-image-cli",
	Short: "Generate images on chatgpt.com/images via your logged-in Chrome session",
	Long: `chatgpt-image-cli drives your real Chrome (via kimi-webbridge) to submit a
prompt to ChatGPT's image tool and save the generated PNG locally.

Requires: kimi-webbridge daemon running, Chrome extension connected, and
you must already be signed in to chatgpt.com in that Chrome.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(newGenerateCmd())
}

func newGenerateCmd() *cobra.Command {
	var (
		outDir     string
		timeoutSec int
	)
	c := &cobra.Command{
		Use:     "generate <prompt>",
		Aliases: []string{"gen"},
		Short:   "Generate an image from a prompt and save the PNG",
		Example: `  chatgpt-image-cli generate "a red apple on a wooden table"
  chatgpt-image-cli generate "夕阳下的富士山" -o ./images
  chatgpt-image-cli gen "a cat in a space suit" --timeout 180`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				output.Error("invalid_args", fmt.Sprintf("generate requires exactly one <prompt> argument (got %d)", len(args)))
				os.Exit(1)
			}
			prompt := args[0]
			client := browser.NewClient(sessionName)

			st, err := client.Status()
			if err != nil {
				output.Error("daemon_unreachable", err.Error())
				os.Exit(1)
			}
			if !st.Running {
				output.Error("daemon_not_running", "kimi-webbridge daemon is not running (see https://kimi.com/features/webbridge)")
				os.Exit(1)
			}
			if !st.ExtensionConnected {
				output.Error("extension_not_connected", "Chrome WebBridge extension is not connected")
				os.Exit(1)
			}

			res, err := chatgpt.Gen(client, chatgpt.Options{
				Prompt:  prompt,
				OutDir:  outDir,
				Timeout: time.Duration(timeoutSec) * time.Second,
			})
			if err != nil {
				output.Error("generate_failed", err.Error())
				os.Exit(1)
			}
			output.Success(res)
		},
	}
	c.Flags().StringVarP(&outDir, "out", "o", ".", "output directory for the saved PNG")
	c.Flags().IntVar(&timeoutSec, "timeout", 180, "max seconds to wait for image generation")
	c.SilenceUsage = true
	c.SilenceErrors = true
	return c
}
