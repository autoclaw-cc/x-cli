package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"scholar-cli/paper"
)

type SemanticScholar struct{}

func NewSemanticScholar() *SemanticScholar { return &SemanticScholar{} }
func (s *SemanticScholar) Name() string    { return "semantic" }

func (s *SemanticScholar) Search(ctx context.Context, query string, limit int) ([]paper.Paper, error) {
	if limit <= 0 {
		limit = 10
	}
	u := fmt.Sprintf(
		"https://api.semanticscholar.org/graph/v1/paper/search?query=%s&limit=%d&fields=title,authors,abstract,year,externalIds,citationCount,referenceCount,url,openAccessPdf,venue,publicationVenue",
		url.QueryEscape(query), limit)

	resp, err := s2GET(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("semantic scholar request: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		Data []struct {
			PaperID    string `json:"paperId"`
			Title      string `json:"title"`
			Abstract   string `json:"abstract"`
			Year       int    `json:"year"`
			Venue      string `json:"venue"`
			URL        string `json:"url"`
			CitationCount  int `json:"citationCount"`
			ReferenceCount int `json:"referenceCount"`
			Authors    []struct {
				AuthorID string `json:"authorId"`
				Name     string `json:"name"`
			} `json:"authors"`
			ExternalIds struct {
				DOI     string `json:"DOI"`
				ArXivID string `json:"ArXiv"`
				PubMedID string `json:"PubMed"`
				PMCID   string `json:"PubMedCentral"`
			} `json:"externalIds"`
			OpenAccessPdf *struct {
				URL    string `json:"url"`
				Status string `json:"status"`
			} `json:"openAccessPdf"`
			PublicationVenue *struct {
				Name string `json:"name"`
			} `json:"publicationVenue"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("semantic scholar decode: %w", err)
	}

	var papers []paper.Paper
	for _, d := range data.Data {
		var authors []paper.Author
		for _, a := range d.Authors {
			authors = append(authors, paper.Author{Name: a.Name})
		}

		venue := d.Venue
		if venue == "" && d.PublicationVenue != nil {
			venue = d.PublicationVenue.Name
		}

		pdfURL := ""
		isOA := false
		if d.OpenAccessPdf != nil {
			pdfURL = d.OpenAccessPdf.URL
			isOA = true
		}

		ids := map[string]string{"s2_id": d.PaperID}
		if d.ExternalIds.DOI != "" {
			ids["doi"] = d.ExternalIds.DOI
		}
		if d.ExternalIds.ArXivID != "" {
			ids["arxiv_id"] = d.ExternalIds.ArXivID
		}
		if d.ExternalIds.PubMedID != "" {
			ids["pmid"] = d.ExternalIds.PubMedID
		}

		p := paper.Paper{
			Title:      d.Title,
			Authors:    authors,
			Abstract:   d.Abstract,
			Year:       d.Year,
			DOI:        d.ExternalIds.DOI,
			Venue:      venue,
			Citations:  d.CitationCount,
			References: d.ReferenceCount,
			OpenAccess: isOA,
			PDFURL:     pdfURL,
			Source:     "semantic",
			URLs:       map[string]string{"semantic_scholar": d.URL},
			IDs:        ids,
		}
		papers = append(papers, p)
	}

	return papers, nil
}
