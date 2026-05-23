package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"scholar-cli/paper"
)

type CrossRef struct{}

func NewCrossRef() *CrossRef { return &CrossRef{} }
func (c *CrossRef) Name() string { return "crossref" }

func (c *CrossRef) Search(ctx context.Context, query string, limit int) ([]paper.Paper, error) {
	if limit <= 0 {
		limit = 10
	}
	u := fmt.Sprintf("https://api.crossref.org/works?query=%s&rows=%d&mailto=scholar-cli@example.com",
		url.QueryEscape(query), limit)

	req, err := newRequest(ctx, "GET", u)
	if err != nil {
		return nil, err
	}

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("crossref request: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		Message struct {
			Items []struct {
				DOI            string     `json:"DOI"`
				Title          []string   `json:"title"`
				Abstract       string     `json:"abstract"`
				ContainerTitle []string   `json:"container-title"`
				Volume         string     `json:"volume"`
				Issue          string     `json:"issue"`
				Page           string     `json:"page"`
				IsReferencedByCount int   `json:"is-referenced-by-count"`
				ReferencesCount     int   `json:"references-count"`
				Author []struct {
					Given  string `json:"given"`
					Family string `json:"family"`
					Affiliation []struct {
						Name string `json:"name"`
					} `json:"affiliation"`
				} `json:"author"`
				Published struct {
					DateParts [][]int `json:"date-parts"`
				} `json:"published"`
				Link []struct {
					URL         string `json:"URL"`
					ContentType string `json:"content-type"`
				} `json:"link"`
			} `json:"items"`
		} `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("crossref decode: %w", err)
	}

	var papers []paper.Paper
	for _, item := range data.Message.Items {
		title := ""
		if len(item.Title) > 0 {
			title = item.Title[0]
		}

		var authors []paper.Author
		for _, a := range item.Author {
			name := strings.TrimSpace(a.Given + " " + a.Family)
			aff := ""
			if len(a.Affiliation) > 0 {
				aff = a.Affiliation[0].Name
			}
			authors = append(authors, paper.Author{Name: name, Affiliation: aff})
		}

		year := 0
		if len(item.Published.DateParts) > 0 && len(item.Published.DateParts[0]) > 0 {
			year = item.Published.DateParts[0][0]
		}

		venue := ""
		if len(item.ContainerTitle) > 0 {
			venue = item.ContainerTitle[0]
		}

		abstract := item.Abstract
		abstract = strings.ReplaceAll(abstract, "<jats:p>", "")
		abstract = strings.ReplaceAll(abstract, "</jats:p>", "")
		abstract = strings.ReplaceAll(abstract, "<jats:title>Abstract</jats:title>", "")

		pdfURL := ""
		for _, link := range item.Link {
			if strings.Contains(link.ContentType, "pdf") {
				pdfURL = link.URL
				break
			}
		}

		p := paper.Paper{
			Title:      title,
			Authors:    authors,
			Abstract:   abstract,
			Year:       year,
			DOI:        item.DOI,
			Venue:      venue,
			Volume:     item.Volume,
			Issue:      item.Issue,
			Pages:      item.Page,
			Citations:  item.IsReferencedByCount,
			References: item.ReferencesCount,
			PDFURL:     pdfURL,
			Source:     "crossref",
			URLs:       map[string]string{"crossref": "https://doi.org/" + item.DOI},
			IDs:        map[string]string{"doi": item.DOI},
		}
		papers = append(papers, p)
	}

	return papers, nil
}
