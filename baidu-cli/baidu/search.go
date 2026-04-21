package baidu

import (
	"encoding/json"
	"fmt"
	"net/url"

	"baidu-cli/browser"
)

// Result is one item in the baidu search result list.
type Result struct {
	Rank     int    `json:"rank"`
	ID       string `json:"id"`
	Tpl      string `json:"tpl"`      // baidu's internal template name — useful for filtering
	Title    string `json:"title"`
	URL      string `json:"url"`
	Abstract string `json:"abstract"`
	Source   string `json:"source"`
}

// extractorJS is the verified evaluate call from Phase 3 archaeology.
// Reads the SSR'd DOM directly — no XHR involved.
const extractorJS = `(() => {
  const items = document.querySelectorAll(".result.c-container, .result-op.c-container");
  return Array.from(items).map((el, i) => {
    const titleEl = el.querySelector("h3 a") || el.querySelector(".t a");
    const absEl = el.querySelector("[class*=summary-text]")
      || el.querySelector("[class*=abstract]")
      || el.querySelector(".c-abstract")
      || el.querySelector("[class*=paragraph]");
    const sourceEl = el.querySelector("[class*=source-text]")
      || el.querySelector("[class*=source]");
    return {
      rank: i + 1,
      id: el.id || "",
      tpl: el.getAttribute("tpl") || "",
      title: titleEl ? titleEl.innerText.trim() : "",
      url: el.getAttribute("mu") || (titleEl && titleEl.href) || "",
      abstract: absEl ? absEl.innerText.trim().replace(/\s+/g, " ").slice(0, 400) : "",
      source: sourceEl ? sourceEl.innerText.trim().replace(/\s+/g, " ").slice(0, 80) : ""
    };
  });
})()`

// Search navigates to the baidu SERP for the given query, then extracts
// results from the SSR'd DOM. limit is the requested number of results
// (passed as the rn= query param; baidu may return more or fewer).
func Search(client *browser.Client, query string, limit int, includeAll bool) ([]Result, error) {
	if query == "" {
		return nil, fmt.Errorf("query is empty")
	}
	if limit <= 0 {
		limit = 10
	}

	serpURL := fmt.Sprintf("https://www.baidu.com/s?wd=%s&rn=%d",
		url.QueryEscape(query), limit)

	if err := client.Navigate(serpURL); err != nil {
		return nil, fmt.Errorf("navigate to SERP: %w", err)
	}

	raw, err := client.Evaluate(extractorJS)
	if err != nil {
		return nil, fmt.Errorf("extract results: %w", err)
	}

	var results []Result
	if err := json.Unmarshal(raw, &results); err != nil {
		return nil, fmt.Errorf("parse extractor output: %w", err)
	}

	if !includeAll {
		results = filterOrganic(results)
	}
	if len(results) > limit {
		results = results[:limit]
	}
	// Re-rank after filtering so consumers see 1..N contiguous.
	for i := range results {
		results[i].Rank = i + 1
	}
	return results, nil
}
