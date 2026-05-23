package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"scholar-cli/paper"
)

type DBLP struct{}

func NewDBLP() *DBLP { return &DBLP{} }
func (d *DBLP) Name() string { return "dblp" }

func (d *DBLP) Search(ctx context.Context, query string, limit int) ([]paper.Paper, error) {
	if limit <= 0 {
		limit = 10
	}
	u := fmt.Sprintf("https://dblp.org/search/publ/api?q=%s&format=json&h=%d",
		url.QueryEscape(query), limit)

	req, err := newRequest(ctx, "GET", u)
	if err != nil {
		return nil, err
	}

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dblp request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("dblp returned status %d", resp.StatusCode)
	}

	var data struct {
		Result struct {
			Hits struct {
				Hit []struct {
					Info struct {
						Title   string `json:"title"`
						Year    string `json:"year"`
						Venue   string `json:"venue"`
						Type    string `json:"type"`
						DOI     string `json:"doi"`
						URL     string `json:"url"`
						Authors struct {
							Author json.RawMessage `json:"author"`
						} `json:"authors"`
					} `json:"info"`
				} `json:"hit"`
			} `json:"hits"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("dblp decode: %w", err)
	}

	var papers []paper.Paper
	for _, hit := range data.Result.Hits.Hit {
		info := hit.Info

		authors := parseDblpAuthors(info.Authors.Author)

		year := 0
		if info.Year != "" {
			fmt.Sscanf(info.Year, "%d", &year)
		}

		doi := info.DOI
		if strings.HasPrefix(doi, "https://doi.org/") {
			doi = doi[len("https://doi.org/"):]
		}

		ids := make(map[string]string)
		if doi != "" {
			ids["doi"] = doi
		}
		dblpKey := info.URL
		if strings.HasPrefix(dblpKey, "https://dblp.org/rec/") {
			dblpKey = dblpKey[len("https://dblp.org/rec/"):]
		}
		if dblpKey != "" {
			ids["dblp_key"] = dblpKey
		}

		title := info.Title
		title = strings.TrimSuffix(title, ".")

		p := paper.Paper{
			Title:   title,
			Authors: authors,
			Year:    year,
			DOI:     doi,
			Venue:   info.Venue,
			Source:  "dblp",
			URLs:    map[string]string{"dblp": info.URL},
			IDs:     ids,
		}
		papers = append(papers, p)
	}

	return papers, nil
}

func parseDblpAuthors(raw json.RawMessage) []paper.Author {
	if len(raw) == 0 {
		return nil
	}

	// DBLP returns either a single object or an array
	var single struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &single); err == nil && single.Text != "" {
		return []paper.Author{{Name: single.Text}}
	}

	var multi []struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &multi); err == nil {
		var authors []paper.Author
		for _, a := range multi {
			if a.Text != "" {
				authors = append(authors, paper.Author{Name: a.Text})
			}
		}
		return authors
	}

	return nil
}
