package browser

import (
	"fmt"
	"os"
)

// ActivateMode decides what Navigate does about the fact that the tab this CLI
// drives is one nobody is looking at. See Client.Activate for why that matters.
//
// The names describe *what gets activated*, because that is the distinction
// that matters to whoever is at the keyboard: making a tab render is invisible
// to them, raising Chrome takes over their screen.
type ActivateMode string

const (
	// ActivateTab makes the page render — layout, lazy loading, infinite
	// scroll — while leaving the user's foreground application alone. Measured:
	// the frontmost app does not change. The tab is still not the *selected*
	// tab in its window (Chrome offers no way to select it without raising the
	// window), but the page behaves as though it were, which is the part any
	// extractor depends on.
	ActivateTab ActivateMode = "tab"
	// ActivateWindow raises the whole Chrome window with Page.bringToFront.
	// That really does select the tab — and takes the screen away from
	// whatever the user was doing. Measured: the frontmost application becomes
	// Google Chrome. Only for pages ActivateTab cannot convince.
	ActivateWindow ActivateMode = "window"
	// ActivateOff skips activation entirely, saving a round trip per navigate.
	// Fine when only the server-rendered first screen is read.
	//
	// It means "this command does not activate", not "force the tab back to
	// the background": the overrides stick to the tab, so one an earlier
	// command already woke up stays awake until it is closed. Close the
	// session first if you need a guaranteed cold tab.
	ActivateOff ActivateMode = "off"
)

// DefaultActivateMode is what this CLI uses when nothing overrides it.
//
// Activating is the default across this repo, and ActivateTab is what
// "activated" means here: the page renders, the user keeps their screen. A CLI
// that never needs a rendered page — one that only reads the server-rendered
// first screen and never clicks, scrolls, or waits on lazy content — can set
// this to ActivateOff. Say why in a comment; getting it wrong fails invisibly.
const DefaultActivateMode = ActivateTab

// ActivateEnv overrides the default mode process-wide, so a CLI that exposes no
// flag for it can still be steered.
const ActivateEnv = "WEBBRIDGE_ACTIVATE"

// ActivateModeHelp is the shared --activate flag description.
const ActivateModeHelp = "how to make the Chrome tab render: " +
	"tab (render it, leave your screen alone) | " +
	"window (raise the Chrome window — takes over your screen) | " +
	"off (leave it backgrounded)"

// ParseActivateMode validates a mode name. The empty string means the default.
func ParseActivateMode(s string) (ActivateMode, error) {
	switch ActivateMode(s) {
	case ActivateTab, ActivateWindow, ActivateOff:
		return ActivateMode(s), nil
	case "":
		return DefaultActivateMode, nil
	}
	return "", fmt.Errorf("unknown activate mode %q (want tab, window or off)", s)
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
	case ActivateWindow:
		_ = c.CDP("Page.bringToFront", nil)
	default:
		_ = c.CDP("Page.setWebLifecycleState", map[string]any{"state": "active"})
		_ = c.CDP("Emulation.setFocusEmulationEnabled", map[string]any{"enabled": true})
	}
}
