package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"scholar-cli/browser"
	"scholar-cli/paper"
)

type CNKI struct {
	client *browser.Client
}

func NewCNKI(client *browser.Client) *CNKI {
	return &CNKI{client: client}
}

func (c *CNKI) Name() string { return "cnki" }

func (c *CNKI) Search(ctx context.Context, query string, limit int) ([]paper.Paper, error) {
	if limit <= 0 {
		limit = 20
	}

	u := fmt.Sprintf("https://kns.cnki.net/kns8s/search?classid=WD0FTY92&kw=%s&korder=SU",
		url.QueryEscape(query))

	if err := c.client.NavigateNewTab(u); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	time.Sleep(5 * time.Second)

	js := fmt.Sprintf(`(function(){
		// Check login status
		var loginEl = document.querySelector(".loginbar a[href*='login']");
		if (loginEl && loginEl.textContent.indexOf("登录") > -1) {
			return JSON.stringify({"error": "not_logged_in", "message": "Not logged in to CNKI. Please log in at https://www.cnki.net and retry."});
		}

		var rows = document.querySelectorAll(".result-table-list tbody tr");
		if (rows.length === 0) {
			// Try alternate layout
			rows = document.querySelectorAll("#gridTable .result-table-list tr");
		}

		var papers = [];
		for (var i = 0; i < Math.min(rows.length, %d); i++) {
			var row = rows[i];
			var titleEl = row.querySelector(".name a");
			var authorEls = row.querySelectorAll(".author a");
			var sourceEl = row.querySelector(".source a");
			var dateEl = row.querySelector(".date");
			var citeEl = row.querySelector(".quote .route, .quote a");

			var title = titleEl ? titleEl.textContent.trim() : "";
			if (!title) continue;

			var authors = [];
			for (var j = 0; j < authorEls.length; j++) {
				authors.push({name: authorEls[j].textContent.trim()});
			}

			var venue = sourceEl ? sourceEl.textContent.trim() : "";
			var dateStr = dateEl ? dateEl.textContent.trim() : "";
			var year = 0;
			var yearMatch = dateStr.match(/\d{4}/);
			if (yearMatch) year = parseInt(yearMatch[0]);

			var citations = 0;
			if (citeEl) {
				var citeText = citeEl.textContent.trim();
				if (citeText) citations = parseInt(citeText) || 0;
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

	raw, err := c.client.EvaluateJSON(js)
	if err != nil {
		return nil, fmt.Errorf("evaluate: %w", err)
	}

	var check struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(raw, &check); err == nil && check.Error != "" {
		return nil, fmt.Errorf("%s: %s", check.Error, check.Message)
	}

	var data struct {
		Papers []struct {
			Title    string `json:"title"`
			Authors  []struct {
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
			Source:    "cnki",
			URLs:      map[string]string{"cnki": d.URL},
			IDs:       make(map[string]string),
		}
		papers = append(papers, p)
	}

	return papers, nil
}
