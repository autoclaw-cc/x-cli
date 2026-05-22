package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"scholar-cli/paper"
)

type BioRxiv struct{}

func NewBioRxiv() *BioRxiv { return &BioRxiv{} }
func (b *BioRxiv) Name() string { return "biorxiv" }

// Search uses CrossRef with DOI prefix filter 10.1101 to find bioRxiv/medRxiv papers.
// bioRxiv's own API only supports date-range browsing, not keyword search.
func (b *BioRxiv) Search(ctx context.Context, query string, limit int) ([]paper.Paper, error) {
	if limit <= 0 {
		limit = 10
	}
	u := fmt.Sprintf(
		"https://api.crossref.org/works?query=%s&filter=prefix:10.1101&rows=%d&sort=relevance&order=desc&mailto=scholar-cli@example.com",
		url.QueryEscape(query), limit)

	req, err := newRequest(ctx, "GET", u)
	if err != nil {
		return nil, err
	}

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("biorxiv search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("biorxiv search returned status %d", resp.StatusCode)
	}

	var data struct {
		Message struct {
			Items []struct {
				DOI            string   `json:"DOI"`
				Title          []string `json:"title"`
				Abstract       string   `json:"abstract"`
				SubType        string   `json:"subtype"`
				ContainerTitle []string `json:"container-title"`
				Author         []struct {
					Given  string `json:"given"`
					Family string `json:"family"`
				} `json:"author"`
				Posted struct {
					DateParts [][]int `json:"date-parts"`
				} `json:"posted"`
				Published struct {
					DateParts [][]int `json:"date-parts"`
				} `json:"published"`
				Created struct {
					DateParts [][]int `json:"date-parts"`
				} `json:"created"`
				Institution []struct {
					Name string `json:"name"`
				} `json:"institution"`
				GroupTitle string `json:"group-title"`
			} `json:"items"`
		} `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("biorxiv decode: %w", err)
	}

	var papers []paper.Paper
	for _, item := range data.Message.Items {
		title := ""
		if len(item.Title) > 0 {
			title = stripHTMLTags(item.Title[0])
		}

		var authors []paper.Author
		for _, a := range item.Author {
			name := strings.TrimSpace(a.Given + " " + a.Family)
			authors = append(authors, paper.Author{Name: name})
		}

		year := 0
		for _, dp := range [][][]int{item.Posted.DateParts, item.Published.DateParts, item.Created.DateParts} {
			if len(dp) > 0 && len(dp[0]) > 0 && dp[0][0] > 0 {
				year = dp[0][0]
				break
			}
		}

		// Determine if bioRxiv or medRxiv
		server := "biorxiv"
		venue := "bioRxiv"
		if len(item.Institution) > 0 {
			instName := strings.ToLower(item.Institution[0].Name)
			if strings.Contains(instName, "medrxiv") {
				server = "medrxiv"
				venue = "medRxiv"
			}
		}

		category := item.GroupTitle

		abstract := item.Abstract
		abstract = strings.ReplaceAll(abstract, "<jats:p>", "")
		abstract = strings.ReplaceAll(abstract, "</jats:p>", "")
		abstract = strings.ReplaceAll(abstract, "<jats:title>Abstract</jats:title>", "")
		abstract = strings.ReplaceAll(abstract, "<jats:title>", "")
		abstract = strings.ReplaceAll(abstract, "</jats:title>", "")
		abstract = strings.TrimSpace(abstract)

		pdfURL := fmt.Sprintf("https://www.%s.org/content/%sv1.full.pdf", server, item.DOI)

		ids := map[string]string{"doi": item.DOI}
		urls := map[string]string{
			server: fmt.Sprintf("https://doi.org/%s", item.DOI),
		}

		p := paper.Paper{
			Title:      title,
			Authors:    authors,
			Abstract:   abstract,
			Year:       year,
			DOI:        item.DOI,
			Venue:      venue,
			OpenAccess: true,
			PDFURL:     pdfURL,
			Source:     server,
			URLs:       urls,
			IDs:        ids,
		}
		if category != "" {
			p.Venue = venue + " — " + category
		}
		papers = append(papers, p)
	}

	return papers, nil
}
