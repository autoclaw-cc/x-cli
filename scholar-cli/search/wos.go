package search

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"scholar-cli/browser"
	"scholar-cli/paper"
)

type WoS struct {
	client *browser.Client
}

func NewWoS(client *browser.Client) *WoS {
	return &WoS{client: client}
}

func (w *WoS) Name() string { return "wos" }

func (w *WoS) Search(ctx context.Context, query string, limit int) ([]paper.Paper, error) {
	if limit <= 0 {
		limit = 10
	}

	u := fmt.Sprintf("https://www.webofscience.com/wos/alldb/basic-search")

	if err := w.client.NavigateNewTab(u); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	time.Sleep(5 * time.Second)

	// Check if we need to log in
	checkJS := `(function(){
		var url = window.location.href;
		if (url.indexOf("login") > -1 || url.indexOf("shibboleth") > -1 || url.indexOf("saml") > -1) {
			return JSON.stringify({"error": "not_logged_in", "message": "Not logged in to Web of Science. Please log in via your institution and retry."});
		}
		return JSON.stringify({"status": "ok", "url": url});
	})()`

	raw, err := w.client.EvaluateJSON(checkJS)
	if err != nil {
		return nil, fmt.Errorf("check login: %w", err)
	}

	var check struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(raw, &check); err == nil && check.Error != "" {
		return nil, fmt.Errorf("%s: %s", check.Error, check.Message)
	}

	// Type query into search box and submit
	searchJS := fmt.Sprintf(`(function(){
		var input = document.querySelector("input[name='search-main-box']") ||
					document.querySelector("input.search-main-box") ||
					document.querySelector("#search-main-box") ||
					document.querySelector("input[placeholder*='topic']") ||
					document.querySelector("input[placeholder*='Topic']") ||
					document.querySelector("input[placeholder*='主题']");
		if (!input) {
			return JSON.stringify({"error": "no_search_box", "message": "Could not find search input. Page: " + document.title});
		}

		// Clear and type
		var nativeInputValueSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
		nativeInputValueSetter.call(input, %s);
		input.dispatchEvent(new Event('input', {bubbles: true}));
		input.dispatchEvent(new Event('change', {bubbles: true}));

		// Click search button
		setTimeout(function(){
			var btn = document.querySelector("button.search-button") ||
					  document.querySelector("button[type='submit']") ||
					  document.querySelector(".search-button");
			if (btn) btn.click();
		}, 500);

		return JSON.stringify({"status": "searching"});
	})()`, jsonString(query))

	raw, err = w.client.EvaluateJSON(searchJS)
	if err != nil {
		return nil, fmt.Errorf("search input: %w", err)
	}

	if err := json.Unmarshal(raw, &check); err == nil && check.Error != "" {
		return nil, fmt.Errorf("%s: %s", check.Error, check.Message)
	}

	// Wait for results
	time.Sleep(8 * time.Second)

	// Extract results
	extractJS := fmt.Sprintf(`(function(){
		// Check for results
		var records = document.querySelectorAll("app-record, .search-results-item, app-records-list app-record-body");
		if (records.length === 0) {
			// Try summary records
			records = document.querySelectorAll("[data-ta='summary-record-title-link']");
			if (records.length === 0) {
				return JSON.stringify({"error": "no_results", "message": "No results found. URL: " + window.location.href + " Title: " + document.title});
			}
		}

		var papers = [];
		var items = document.querySelectorAll("app-record");
		if (items.length === 0) {
			items = document.querySelectorAll(".search-results-item");
		}

		for (var i = 0; i < Math.min(items.length, %d); i++) {
			var el = items[i];
			var titleEl = el.querySelector("[data-ta='summary-record-title-link'] a, .title a, a.title");
			var authorEl = el.querySelector(".authors, [data-ta='summary-record-author']");
			var sourceEl = el.querySelector(".source, .journal, [data-ta='summary-record-source']");
			var yearEl = el.querySelector(".year, [data-ta='summary-record-year']");
			var citeEl = el.querySelector("[data-ta='summary-record-cited-count'], .cited-count");

			var title = titleEl ? titleEl.textContent.trim() : "";
			if (!title) continue;

			var authorText = authorEl ? authorEl.textContent.trim() : "";
			var authors = [];
			if (authorText) {
				var names = authorText.split(";");
				for (var j = 0; j < names.length; j++) {
					var n = names[j].replace(/…/g, "").trim();
					if (n && n.length < 50) authors.push({name: n});
				}
			}

			var venue = sourceEl ? sourceEl.textContent.trim() : "";
			var yearText = yearEl ? yearEl.textContent.trim() : "";
			var year = 0;
			var ym = yearText.match(/\d{4}/);
			if (ym) year = parseInt(ym[0]);

			var citations = 0;
			if (citeEl) {
				var cm = citeEl.textContent.match(/\d+/);
				if (cm) citations = parseInt(cm[0]);
			}

			var link = titleEl ? titleEl.href : "";

			papers.push({
				title: title,
				authors: authors,
				venue: venue,
				year: year,
				citations: citations,
				url: link
			});
		}

		return JSON.stringify({papers: papers});
	})()`, limit)

	raw, err = w.client.EvaluateJSON(extractJS)
	if err != nil {
		return nil, fmt.Errorf("extract: %w", err)
	}

	if err := json.Unmarshal(raw, &check); err == nil && check.Error != "" {
		return nil, fmt.Errorf("%s: %s", check.Error, check.Message)
	}

	var data struct {
		Papers []struct {
			Title     string `json:"title"`
			Authors   []struct {
				Name string `json:"name"`
			} `json:"authors"`
			Venue     string `json:"venue"`
			Year      int    `json:"year"`
			Citations int    `json:"citations"`
			URL       string `json:"url"`
		} `json:"papers"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	var papers []paper.Paper
	for _, d := range data.Papers {
		var authors []paper.Author
		for _, a := range d.Authors {
			authors = append(authors, paper.Author{Name: a.Name})
		}

		p := paper.Paper{
			Title:     d.Title,
			Authors:   authors,
			Venue:     d.Venue,
			Year:      d.Year,
			Citations: d.Citations,
			Source:    "wos",
			URLs:      map[string]string{"wos": d.URL},
			IDs:       make(map[string]string),
		}
		papers = append(papers, p)
	}

	return papers, nil
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func CheckWoSLogin(client *browser.Client) (bool, error) {
	if err := client.NavigateNewTab("https://www.webofscience.com/wos/alldb/basic-search"); err != nil {
		return false, fmt.Errorf("navigate: %w", err)
	}

	time.Sleep(5 * time.Second)

	js := `(function(){
		var url = window.location.href;
		var isLogin = url.indexOf("login") === -1 && url.indexOf("shibboleth") === -1 && url.indexOf("saml") === -1;
		var hasSearch = !!document.querySelector("input[name='search-main-box'], input.search-main-box, #search-main-box");
		return JSON.stringify({logged_in: isLogin && hasSearch, url: url});
	})()`

	raw, err := client.EvaluateJSON(js)
	if err != nil {
		return false, err
	}

	var status struct {
		LoggedIn bool   `json:"logged_in"`
		URL      string `json:"url"`
	}
	if err := json.Unmarshal(raw, &status); err != nil {
		return false, err
	}

	return status.LoggedIn, nil
}

