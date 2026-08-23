package cmd

import (
	"os"

	"anjuke-cli/browser"
	"anjuke-cli/output"
)

// activateFlag backs the persistent --activate flag.
//
// A background Chrome tab lays nothing out, so clicks, infinite scroll and lazy
// images silently do nothing in it. The default renders the tab without taking
// the user's screen; this flag is the escape hatch either way.
// See browser.Activate.
var activateFlag string

func init() {
	rootCmd.PersistentFlags().StringVar(&activateFlag, "activate",
		string(browser.DefaultActivateMode), browser.ActivateModeHelp)
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
