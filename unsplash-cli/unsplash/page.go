package unsplash

import (
	"encoding/json"
	"fmt"
	"time"
)

// Browser is the slice of the kimi-webbridge client this package needs. Taking
// an interface here (rather than *browser.Client) keeps the extraction logic
// testable against a fake tab, with no daemon running.
type Browser interface {
	// Navigate points the tab at url. Implementations are expected to make the
	// tab render — see browser.Client.Activate — since a background tab lays
	// nothing out and every extractor here depends on real element boxes.
	Navigate(url string) error
	Evaluate(code string) (json.RawMessage, error)
}

// Unsplash sits behind Anubis (https://github.com/TecharoHQ/anubis), a
// proof-of-work bot check. A document navigation first lands on
// unsplash.com/.within.website?redir=..., the page's JS solves a SHA-256
// challenge, and only then does the real page render. Nothing in the daemon's
// navigate call waits for that, so every read has to poll the tab until the
// bot check clears and the content we want exists.
//
// This is also why the CLI cannot skip the browser: hitting /napi/* with a
// plain HTTP client — or even with fetch() from inside the page — comes back as
// the 401 challenge, never JSON.

// probeJS reports the tab's readiness for a given CSS selector. Kept as one
// round trip so a poll costs a single daemon call.
const probeJS = `(() => {
  const sel = %q;
  const botCheck = location.href.indexOf("/.within.website") >= 0
    || !!document.getElementById("anubis_challenge");
  return {
    href: location.href,
    title: document.title,
    bot_check: botCheck,
    ready_state: document.readyState,
    count: botCheck ? 0 : document.querySelectorAll(sel).length
  };
})()`

type pageState struct {
	Href       string `json:"href"`
	Title      string `json:"title"`
	BotCheck   bool   `json:"bot_check"`
	ReadyState string `json:"ready_state"`
	Count      int    `json:"count"`
}

// waitPolicy tunes how Load polls a freshly navigated tab.
type waitPolicy struct {
	Timeout  time.Duration // give up after this long
	Interval time.Duration // pause between probes
}

var defaultWait = waitPolicy{Timeout: 30 * time.Second, Interval: 700 * time.Millisecond}

// Load navigates the session tab to url and blocks until selector matches at
// least one element, the Anubis bot check has cleared, or the policy times out.
//
// It returns the final page state so callers can tell "the bot check never cleared"
// (retryable, and worth saying out loud) apart from "the page rendered but had
// nothing matching" (an empty result, not an error).
func Load(client Browser, url, selector string, policy waitPolicy) (*pageState, error) {
	if err := client.Navigate(url); err != nil {
		return nil, fmt.Errorf("navigate to %s: %w", url, err)
	}

	code := fmt.Sprintf(probeJS, selector)
	deadline := time.Now().Add(policy.Timeout)
	var last *pageState

	for {
		raw, err := client.Evaluate(code)
		if err == nil {
			var st pageState
			if err := json.Unmarshal(raw, &st); err == nil {
				last = &st
				if !st.BotCheck && st.Count > 0 {
					return &st, nil
				}
				// Bot check cleared, DOM settled, still nothing: a genuinely empty
				// page (no search hits, or a 404). Report it rather than
				// stalling until the deadline.
				if !st.BotCheck && st.ReadyState == "complete" && st.Count == 0 {
					return &st, nil
				}
			}
		}
		// An evaluate error here is expected while the tab swaps documents
		// mid-challenge ("Inspected target navigated or closed"), so it is not
		// fatal on its own — only the deadline is.
		if time.Now().After(deadline) {
			if last != nil && last.BotCheck {
				return last, fmt.Errorf("anubis bot check did not clear within %s (still on %s) — retry, or load unsplash.com in Chrome once by hand", policy.Timeout, last.Href)
			}
			return last, fmt.Errorf("timed out after %s waiting for %q on %s", policy.Timeout, selector, url)
		}
		time.Sleep(policy.Interval)
	}
}
