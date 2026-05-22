package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"scholar-cli/browser"
	"scholar-cli/paper"
)

type GoogleScholar struct {
	client *browser.Client
}

func NewGoogleScholar(client *browser.Client) *GoogleScholar {
	return &GoogleScholar{client: client}
}

func (g *GoogleScholar) Name() string { return "google" }

func (g *GoogleScholar) Search(ctx context.Context, query string, limit int) ([]paper.Paper, error) {
	if limit <= 0 {
		limit = 10
	}
	num := limit
	if num > 20 {
		num = 20
	}

	u := fmt.Sprintf("https://scholar.google.com/scholar?q=%s&hl=en&num=%d",
		url.QueryEscape(query), num)

	if err := g.client.NavigateNewTab(u); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	time.Sleep(4 * time.Second)

	js := `(function(){
		// Check for CAPTCHA
		if (window.location.href.indexOf("/sorry/") > -1 || window.location.href.indexOf("google.com/sorry") > -1) {
			return JSON.stringify({"error": "captcha", "message": "Google Scholar CAPTCHA detected. Please solve it in the browser and retry."});
		}

		var items = document.querySelectorAll(".gs_r.gs_or.gs_scl");
		if (items.length === 0) {
			// Try alternate selectors
			items = document.querySelectorAll(".gs_ri");
			if (items.length === 0) {
				return JSON.stringify({"error": "no_results", "message": "No results found. Page title: " + document.title});
			}
		}

		var results = [];
		for (var i = 0; i < items.length; i++) {
			var el = items[i];
			var titleEl = el.querySelector(".gs_rt a") || el.querySelector("h3 a");
			var metaEl = el.querySelector(".gs_a");
			var snippetEl = el.querySelector(".gs_rs");

			var citeCount = 0;
			var citeLinks = el.querySelectorAll(".gs_fl a, .gs_flb a");
			for (var j = 0; j < citeLinks.length; j++) {
				var t = citeLinks[j].textContent;
				if (t.indexOf("Cited by") > -1 || t.indexOf("被引用") > -1) {
					var m = t.match(/\d+/);
					if (m) citeCount = parseInt(m[0]);
				}
			}

			// Parse meta line: "Author1, Author2 - Venue, Year - Publisher"
			var meta = metaEl ? metaEl.textContent : "";
			var authors = [];
			var venue = "";
			var year = 0;

			var parts = meta.split(" - ");
			if (parts.length >= 1) {
				var authorStr = parts[0].trim();
				var authorNames = authorStr.split(",");
				for (var k = 0; k < authorNames.length; k++) {
					var name = authorNames[k].replace(/…/g, "").trim();
					if (name && name.length < 40) {
						authors.push({name: name});
					}
				}
			}
			if (parts.length >= 2) {
				var venuePart = parts[1].trim();
				var yearMatch = venuePart.match(/\b(19|20)\d{2}\b/);
				if (yearMatch) {
					year = parseInt(yearMatch[0]);
					venue = venuePart.replace(/,?\s*\d{4}\s*$/, "").trim();
				} else {
					venue = venuePart;
				}
			}

			// Get link to article
			var link = titleEl ? titleEl.href : "";
			var title = titleEl ? titleEl.textContent : "";

			// Try to find PDF link
			var pdfUrl = "";
			var pdfEl = el.querySelector(".gs_or_ggsm a, .gs_ggsd a");
			if (pdfEl) {
				pdfUrl = pdfEl.href || "";
			}

			results.push({
				title: title,
				authors: authors,
				venue: venue,
				year: year,
				citations: citeCount,
				url: link,
				pdf_url: pdfUrl,
				snippet: snippetEl ? snippetEl.textContent.trim() : ""
			});
		}
		return JSON.stringify({papers: results});
	})()`

	raw, err := g.client.EvaluateJSON(js)
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
			Venue    string `json:"venue"`
			Year     int    `json:"year"`
			Citations int   `json:"citations"`
			URL      string `json:"url"`
			PDFURL   string `json:"pdf_url"`
			Snippet  string `json:"snippet"`
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

		// Try to extract DOI from URL
		doi := ""
		if strings.Contains(d.URL, "doi.org/") {
			idx := strings.Index(d.URL, "doi.org/")
			doi = d.URL[idx+8:]
		}

		ids := make(map[string]string)
		if doi != "" {
			ids["doi"] = doi
		}

		p := paper.Paper{
			Title:      d.Title,
			Authors:    authors,
			Abstract:   d.Snippet,
			Year:       d.Year,
			DOI:        doi,
			Venue:      d.Venue,
			Citations:  d.Citations,
			PDFURL:     d.PDFURL,
			Source:     "google",
			URLs:       map[string]string{"google_scholar": d.URL},
			IDs:        ids,
		}
		papers = append(papers, p)
	}

	return papers, nil
}
