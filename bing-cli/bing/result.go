package bing

import (
	"fmt"

	"bing-cli/browser"
)

// PageResult holds the extracted content from a fetched page.
type PageResult struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Text        string `json:"text"`
}

// resultExtractJS extracts page metadata and body text.
// For JS-heavy pages where text hasn't rendered yet, the caller should
// retry once after a 1.5 s wait if text length < 50.
const resultExtractJS = `(() => {
  const meta = document.querySelector('meta[name="description"]');
  const og   = document.querySelector('meta[property="og:description"]');
  return JSON.stringify({
    url:         location.href,
    title:       document.title,
    description: meta ? meta.getAttribute('content') : (og ? og.getAttribute('content') : ''),
    text:        (document.body ? document.body.innerText : '').slice(0, 5000)
  });
})()`

// FetchResult navigates to url and extracts title, description, and text.
// JS-heavy pages get a single retry after a 1.5 s settle delay.
func FetchResult(client *browser.Client, targetURL string) (*PageResult, error) {
	if err := client.Navigate(targetURL); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	var page PageResult
	if err := evaluateWithRetry(client, resultExtractJS, &page); err != nil {
		return nil, fmt.Errorf("extract: %w", err)
	}

	// JS-heavy pages: if body text is suspiciously short, wait and retry once.
	if len(page.Text) < 50 {
		// Use a raw Evaluate call so we can sleep inside JS; the daemon
		// doesn't support time.Sleep in Go, so we re-evaluate after a
		// short settle via a simple JS delay.
		settleJS := `(async () => {
		  await new Promise(r => setTimeout(r, 2000));
		  return JSON.stringify({
		    url:         location.href,
		    title:       document.title,
		    description: (() => {
		      const m = document.querySelector('meta[name="description"]');
		      if (m) return m.getAttribute('content');
		      const og = document.querySelector('meta[property="og:description"]');
		      return og ? og.getAttribute('content') : '';
		    })(),
		    text: (document.body ? document.body.innerText : '').slice(0, 5000)
		  });
		})()`
		var retry PageResult
		if err := client.EvaluateJSON(settleJS, &retry); err != nil {
			// Fall back to the original (possibly short) result.
			return &page, nil
		}
		if len(retry.Text) > len(page.Text) {
			page = retry
		}
	}

	return &page, nil
}
