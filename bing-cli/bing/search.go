package bing

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"bing-cli/browser"
)

const baseURL = "https://www.bing.com/search"

// ErrConsentRequired is returned when Bing serves a consent/privacy
// interstitial. The user must accept it once in Chrome, then retry.
type ErrConsentRequired struct{}

func (e ErrConsentRequired) Error() string {
	return "Bing served a consent interstitial. Accept once in Chrome, then retry."
}

// bingRedirectRe matches the real target URL embedded in Bing's /ck/a redirect page.
var bingRedirectRe = regexp.MustCompile(`https?://[a-zA-Z0-9][a-zA-Z0-9._/-]*[a-zA-Z0-9/]`)

// SearchResult is one item in the Bing search result list.
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
	Source  string `json:"source"`
}

// SearchOptions bundles all search parameters.
type SearchOptions struct {
	Query  string
	Market string // "cn", "us", "" (auto)
	Count  int
	Offset int
}

// searchExtractJS is the merged extractor:
//   - Async polling (8 s deadline) — google-cli pattern
//   - Consent interstitial detection — google-cli pattern
//   - Bing selectors (#b_results > li.b_algo) — v2
//   - bingHref capture for /ck/a redirect resolution — acc
//   - URL dedup — google-cli pattern
//   - Snippet cleaning (whitespace + trailing "Read more") — both
const searchExtractJS = `(async () => {
  const deadline = Date.now() + 8000;
  let results = [];

  while (Date.now() < deadline) {
    if (location.host.startsWith('consent.')) {
      return JSON.stringify({consent: true, items: []});
    }

    const items = document.querySelectorAll("#b_results > li.b_algo");
    if (items.length > 0) {
      const seen = new Set();
      results = Array.from(items).map(el => {
        const h2 = el.querySelector("h2 a");
        const cite = el.querySelector(".b_attribution cite, cite");
        const p = el.querySelector(".b_caption p");
        if (!h2) return null;
        const displayURL = cite ? cite.textContent.trim() : "";
        if (!displayURL || seen.has(displayURL)) return null;
        seen.add(displayURL);
        return {
          title:    h2.textContent.trim(),
          url:      displayURL,
          snippet:  (p ? p.textContent.trim().replace(/\s+/g, " ") : "")
                      .replace(/\s*Read more\.?$/, "")
                      .slice(0, 300),
          source:   cite ? cite.textContent.trim() : "",
          bingHref: h2.getAttribute("href") || ""
        };
      }).filter(Boolean);
      if (results.length > 0) break;
    }
    await new Promise(r => setTimeout(r, 500));
  }

  return JSON.stringify({consent: false, items: results});
})()`

// resolveBingURL follows Bing's /ck/a redirect and extracts the real target
// URL from the HTML response body. Falls back to the original URL on failure.
func resolveBingURL(bingURL string) string {
	if !strings.Contains(bingURL, "/ck/a") || !strings.Contains(bingURL, "u=") {
		return bingURL
	}
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(bingURL)
	if err != nil {
		return bingURL
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return bingURL
	}
	body, _ := io.ReadAll(resp.Body)
	matches := bingRedirectRe.FindAllString(string(body), -1)
	for _, m := range matches {
		if !strings.Contains(m, "bing.com/ck") {
			return m
		}
	}
	return bingURL
}

// Search navigates to the Bing SERP and extracts organic results.
func Search(client *browser.Client, opts SearchOptions) ([]SearchResult, error) {
	params := url.Values{}
	params.Set("q", opts.Query)
	if opts.Count > 0 {
		params.Set("count", fmt.Sprintf("%d", opts.Count))
	} else {
		params.Set("count", "10")
	}
	if opts.Offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", opts.Offset))
	}
	if opts.Market != "" {
		switch opts.Market {
		case "cn", "zh-CN", "zh":
			params.Set("cc", "cn")
			params.Set("setlang", "zh-Hans")
		case "us", "en-US", "en":
			params.Set("cc", "us")
			params.Set("setlang", "en")
		default:
			params.Set("cc", opts.Market)
		}
	}

	searchURL := baseURL + "?" + params.Encode()

	if err := client.Navigate(searchURL); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	// Extract results with CDP retry.
	var rawPayload struct {
		Consent bool `json:"consent"`
		Items   []struct {
			Title    string `json:"title"`
			URL      string `json:"url"`
			Snippet  string `json:"snippet"`
			Source   string `json:"source"`
			BingHref string `json:"bingHref"`
		} `json:"items"`
	}

	if err := evaluateWithRetry(client, searchExtractJS, &rawPayload); err != nil {
		return nil, fmt.Errorf("extract: %w", err)
	}

	if rawPayload.Consent {
		return nil, ErrConsentRequired{}
	}

	results := make([]SearchResult, 0, len(rawPayload.Items))
	for _, item := range rawPayload.Items {
		finalURL := item.URL
		// Resolve Bing's encrypted /ck/a redirect to the real URL.
		if item.BingHref != "" && strings.Contains(item.BingHref, "/ck/a") {
			if resolved := resolveBingURL(item.BingHref); resolved != item.BingHref {
				finalURL = resolved
			}
		}
		results = append(results, SearchResult{
			Title:   item.Title,
			URL:     finalURL,
			Snippet: item.Snippet,
			Source:  item.Source,
		})
	}

	return results, nil
}

// FormatResults returns a human-readable string for terminal display (stderr).
func FormatResults(results []SearchResult, query string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Bing Search: %s\n", query))
	sb.WriteString(strings.Repeat("─", 80) + "\n")
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("\n%d. %s\n", i+1, r.Title))
		sb.WriteString(fmt.Sprintf("   %s\n", r.URL))
		if r.Snippet != "" {
			sb.WriteString(fmt.Sprintf("   %s\n", truncate(r.Snippet, 200)))
		}
	}
	return sb.String()
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
