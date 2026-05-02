package google

import (
	"fmt"
	"net/url"

	"google-cli/browser"
)

type Page struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Text        string `json:"text"`
}

// Extraction JS — waits for DOMContentLoaded-ish state and for non-empty body,
// then reads title/description/text. See ARCHAEOLOGY.md §"Feature B".
const pageExtractJS = `
(async () => {
  const deadline = Date.now() + 8000;
  while (Date.now() < deadline) {
    if (document.readyState !== 'loading' && document.body && document.body.innerText.length > 50) break;
    await new Promise(r => setTimeout(r, 150));
  }
  return JSON.stringify({
    url: location.href,
    title: document.title,
    description: document.querySelector('meta[name="description"]')?.content
              || document.querySelector('meta[property="og:description"]')?.content
              || '',
    text: (document.body?.innerText || '').slice(0, 5000)
  });
})()
`

// FetchResult loads `pageURL` in the browser and returns title/description/text.
func FetchResult(client *browser.Client, pageURL string) (*Page, error) {
	u, err := url.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("invalid_url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("invalid_url: scheme must be http or https, got %q", u.Scheme)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("invalid_url: missing host")
	}

	if err := client.Navigate(pageURL); err != nil {
		return nil, err
	}

	var page Page
	if err := evaluateWithRetry(client, pageExtractJS, &page); err != nil {
		return nil, err
	}
	if len(page.Text) < 50 {
		return nil, fmt.Errorf("empty_content: extracted text length %d; page may be JS-heavy or blocked", len(page.Text))
	}
	return &page, nil
}
