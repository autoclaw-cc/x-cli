package browser

import (
	"fmt"
	"os"
)

// ActivateMode decides what Navigate does about the fact that the tab this CLI
// drives is one nobody is looking at. See Client.Activate for why that matters.
type ActivateMode string

const (
	// ActivateAuto emulates an active, focused tab without disturbing whatever
	// the user is actually looking at. The default, and almost always right.
	ActivateAuto ActivateMode = "auto"
	// ActivateFront really does raise the tab, stealing the user's screen. For
	// pages that auto emulation cannot convince — ones that check
	// document.hasFocus(), or need real compositing.
	ActivateFront ActivateMode = "front"
	// ActivateOff skips activation entirely, saving a round trip per navigate.
	// Fine when only the server-rendered first screen is read.
	//
	// It means "this command does not activate", not "force the tab back to
	// the background": the CDP overrides stick to the tab, so one an earlier
	// command already woke up stays awake until it is closed. Close the
	// session first if you need a guaranteed cold tab.
	ActivateOff ActivateMode = "off"
)

// DefaultActivateMode is what this CLI uses when nothing overrides it.
//
// DefaultActivateMode is what this CLI uses when nothing overrides it.
//
// Activating is the default across this repo, and ActivateFront is what
// "activated" means: the tab is really raised, so it really renders. Emulation
// (ActivateAuto) is a stand-in that works today but is one Chrome change away
// from quietly not working, and a CLI the user just typed is allowed to take
// the screen for the seconds it runs.
//
// A CLI that never needs a rendered page — one that only reads the
// server-rendered first screen and never clicks, scrolls, or waits on lazy
// content — can set this to ActivateOff and stay out of the user's way. Say why
// in a comment when you do; the failure mode of getting it wrong is invisible.
const DefaultActivateMode = ActivateFront

// ActivateEnv overrides the default mode process-wide, so a CLI that exposes no
// flag for it can still be steered.
const ActivateEnv = "WEBBRIDGE_ACTIVATE"

// ParseActivateMode validates a mode name. The empty string means the default.
func ParseActivateMode(s string) (ActivateMode, error) {
	switch ActivateMode(s) {
	case ActivateAuto, ActivateFront, ActivateOff:
		return ActivateMode(s), nil
	case "":
		return ActivateAuto, nil
	}
	return "", fmt.Errorf("unknown activate mode %q (want auto, front or off)", s)
}

// defaultActivateMode reads ActivateEnv. A typo there shouldn't be fatal, and
// shouldn't be silent either.
func defaultActivateMode() ActivateMode {
	mode, err := ParseActivateMode(os.Getenv(ActivateEnv))
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %s ignored: %v\n", ActivateEnv, err)
		return DefaultActivateMode
	}
	return mode
}

// SetActivateMode overrides the mode taken from the environment.
func (c *Client) SetActivateMode(m ActivateMode) { c.activate = m }

// CDP is the raw chrome.debugger passthrough — the escape hatch for anything
// the daemon's higher-level actions don't cover.
func (c *Client) CDP(method string, params map[string]any) error {
	if params == nil {
		params = map[string]any{}
	}
	_, err := c.Call("cdp", map[string]any{"method": method, "params": params})
	return err
}

// Activate makes the session tab behave as if the user were looking at it.
//
// This matters more than it sounds. A background tab in Chrome reports
// document.visibilityState "hidden", skips layout entirely, and never fires the
// IntersectionObserver callbacks that drive lazy loading and infinite scroll.
// Anything that depends on an element having a box — clicking a button,
// scrolling to load more results, waiting for a lazy image — silently does
// nothing, and the result is indistinguishable from the site refusing to
// cooperate. It is not: it is a tab nobody is watching.
//
// Navigate calls this, so callers get it for free. Failures are ignored on
// purpose: an older Chrome missing one of these methods should degrade to
// first-screen-only, not turn every command into an error.
func (c *Client) Activate() {
	switch c.activate {
	case ActivateOff:
		return
	case ActivateFront:
		_ = c.CDP("Page.bringToFront", nil)
	default:
		_ = c.CDP("Page.setWebLifecycleState", map[string]any{"state": "active"})
		_ = c.CDP("Emulation.setFocusEmulationEnabled", map[string]any{"enabled": true})
	}
}
