package google

import (
	"fmt"
	"net/url"
	"strings"

	"google-cli/browser"
)

type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// searchExtractJS waits for Google's results to mount, then extracts title/url/snippet.
// Returns a JSON string for EvaluateJSON compatibility.
const searchExtractJS = `
(async () => {
  const deadline = Date.now() + 8000;
  while (Date.now() < deadline) {
    if (location.host.startsWith('consent.')) break;
    if (document.querySelector('div#search div[data-hveid] h3')) break;
    await new Promise(r => setTimeout(r, 150));
  }
  const seen = new Set();
  const items = Array.from(document.querySelectorAll('div#search div[data-hveid]'))
    .filter(el => el.querySelector('h3') && el.querySelector('a[href]'))
    .map(el => {
      const a = el.querySelector('a[href]');
      const h = el.querySelector('h3');
      return {
        title: h.innerText,
        url: a.href,
        snippet: (el.querySelector('[data-sncf]')?.innerText || '').replace(/\s*Read more\s*$/, '')
      };
    })
    .filter(r => { if (seen.has(r.url)) return false; seen.add(r.url); return true; });
  return JSON.stringify({ consent: location.host.startsWith('consent.'), items });
})()
`

type ErrConsentRequired struct{}

func (ErrConsentRequired) Error() string {
	return "google served a consent interstitial; accept it once in Chrome and retry"
}

// FetchSearch runs a Google Search for `query` and returns up to `limit` results.
// `hl` is Google's UI language (e.g. "en") — keep it stable to avoid DOM drift.
func FetchSearch(client *browser.Client, query string, limit int, hl string) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if hl == "" {
		hl = "en"
	}

	// Over-fetch: Google pads result slots with ads / PAA cards.
	requestN := max(limit*2, 10)

	searchURL := fmt.Sprintf(
		"https://www.google.com/search?q=%s&hl=%s&num=%d",
		url.QueryEscape(query), url.QueryEscape(hl), requestN,
	)
	if err := client.Navigate(searchURL); err != nil {
		return nil, err
	}

	var payload struct {
		Consent bool           `json:"consent"`
		Items   []SearchResult `json:"items"`
	}
	if err := evaluateWithRetry(client, searchExtractJS, &payload); err != nil {
		return nil, err
	}
	if payload.Consent {
		return nil, ErrConsentRequired{}
	}

	// Drop entries with empty URLs — they're not actionable links.
	filtered := payload.Items[:0]
	for _, item := range payload.Items {
		if strings.TrimSpace(item.URL) != "" {
			filtered = append(filtered, item)
		}
	}
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered, nil
}
