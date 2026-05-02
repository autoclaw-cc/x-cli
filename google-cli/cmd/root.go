package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "google-cli",
	Short: "Google Search CLI backed by the kimi-webbridge browser daemon",
	Long: `google-cli automates Google Search by running inside the user's real browser
session via kimi-webbridge (http://127.0.0.1:10086). All commands emit JSON on
stdout: {"ok":true,"data":...} on success, {"ok":false,"error":{...}} on failure.
Every command exits non-zero on failure.`,
}

func Execute() error {
	return rootCmd.Execute()
}
