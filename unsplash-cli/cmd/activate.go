package cmd

import (
	"os"

	"unsplash-cli/browser"
	"unsplash-cli/output"
)

// activateFlag backs the persistent --activate flag.
//
// A background Chrome tab lays nothing out, so clicks, infinite scroll and lazy
// images silently do nothing in it — which is exactly what made Unsplash's
// pagination look like a server-side cap. The default really raises the tab;
// use this flag for the quieter modes. See browser.Activate.
var activateFlag string

func init() {
	rootCmd.PersistentFlags().StringVar(&activateFlag, "activate", string(browser.DefaultActivateMode),
		"how to make the Chrome tab render: auto (emulate an active tab, no focus stolen) | front (really raise the tab) | off (leave it backgrounded)")
}

// withActivateMode applies an explicit --activate to a fresh client. Leaving the
// flag alone keeps whatever WEBBRIDGE_ACTIVATE said.
func withActivateMode(c *browser.Client) *browser.Client {
	if !rootCmd.PersistentFlags().Changed("activate") {
		return c
	}
	mode, err := browser.ParseActivateMode(activateFlag)
	if err != nil {
		output.Error("bad_flag", err.Error())
		os.Exit(1)
	}
	c.SetActivateMode(mode)
	return c
}
