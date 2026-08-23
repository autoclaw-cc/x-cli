package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"unsplash-cli/browser"
	"unsplash-cli/output"
)

const (
	sessionName = "unsplash"
	groupTitle  = "unsplash-cli"
)

var rootCmd = &cobra.Command{
	Use:   "unsplash-cli",
	Short: "Search and download Unsplash photos via the kimi-webbridge daemon",
	Long: `unsplash-cli drives a real Chrome tab through the kimi-webbridge daemon to
read unsplash.com, then pulls image files straight from Unsplash's CDN.

unsplash.com sits behind Anubis, a proof-of-work bot check: its /napi/* JSON API
answers 401 to plain HTTP clients, so search results are read from the
server-rendered search page in a real browser instead. Image bytes are a
different story — the CDN has no bot check, so downloads skip the browser entirely.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the CLI. Cobra's own error printing is silenced so that
// argument and usage errors come out in the same {ok:false, error} envelope as
// everything else — a caller parsing stdout should never get a bare
// "Error: accepts 1 arg(s)" line it can't unmarshal.
func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		output.Error("bad_args", err.Error())
	}
	return err
}

// newBrowser returns a daemon client, failing fast (and in the JSON contract)
// when the daemon or the Chrome extension isn't there.
func newBrowser() *browser.Client {
	client := withActivateMode(browser.NewClient(sessionName, groupTitle))

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
	return client
}

// fail emits the error contract and exits non-zero.
func fail(code, msg string) {
	output.Error(code, msg)
	os.Exit(1)
}
