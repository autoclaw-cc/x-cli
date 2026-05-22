package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"scholar-cli/paper"
)

func EnrichByDOI(ctx context.Context, doi string) (*paper.Paper, error) {
	doi = strings.TrimSpace(doi)
	if strings.HasPrefix(doi, "https://doi.org/") {
		doi = doi[len("https://doi.org/"):]
	}

	// Fetch from CrossRef
	crURL := fmt.Sprintf("https://api.crossref.org/works/%s?mailto=scholar-cli@example.com",
		url.PathEscape(doi))

	req, err := newRequest(ctx, "GET", crURL)
	if err != nil {
		return nil, err
	}

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("crossref: %w", err)
	}
	defer resp.Body.Close()

	var crData struct {
		Message struct {
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
		} `json:"message"`
	}

	var p paper.Paper
	p.DOI = doi
	p.Source = "crossref"
	p.IDs = map[string]string{"doi": doi}
	p.URLs = map[string]string{"doi": "https://doi.org/" + doi}

	if err := json.NewDecoder(resp.Body).Decode(&crData); err == nil {
		m := crData.Message
		if len(m.Title) > 0 {
			p.Title = m.Title[0]
		}
		for _, a := range m.Author {
			name := strings.TrimSpace(a.Given + " " + a.Family)
			aff := ""
			if len(a.Affiliation) > 0 {
				aff = a.Affiliation[0].Name
			}
			p.Authors = append(p.Authors, paper.Author{Name: name, Affiliation: aff})
		}
		p.Abstract = m.Abstract
		if len(m.ContainerTitle) > 0 {
			p.Venue = m.ContainerTitle[0]
		}
		p.Volume = m.Volume
		p.Issue = m.Issue
		p.Pages = m.Page
		p.Citations = m.IsReferencedByCount
		p.References = m.ReferencesCount
		if len(m.Published.DateParts) > 0 && len(m.Published.DateParts[0]) > 0 {
			p.Year = m.Published.DateParts[0][0]
		}
	}

	// Enrich from Semantic Scholar
	s2URL := fmt.Sprintf(
		"https://api.semanticscholar.org/graph/v1/paper/DOI:%s?fields=title,authors,abstract,year,citationCount,referenceCount,openAccessPdf,externalIds,venue",
		url.PathEscape(doi))

	req2, err := newRequest(ctx, "GET", s2URL)
	if err == nil {
		resp2, err := defaultHTTPClient.Do(req2)
		if err == nil {
			defer resp2.Body.Close()
			var s2Data struct {
				PaperID    string `json:"paperId"`
				Title      string `json:"title"`
				Abstract   string `json:"abstract"`
				Year       int    `json:"year"`
				Venue      string `json:"venue"`
				CitationCount  int `json:"citationCount"`
				ReferenceCount int `json:"referenceCount"`
				ExternalIds struct {
					ArXivID string `json:"ArXiv"`
				} `json:"externalIds"`
				OpenAccessPdf *struct {
					URL string `json:"url"`
				} `json:"openAccessPdf"`
			}
			if err := json.NewDecoder(resp2.Body).Decode(&s2Data); err == nil {
				if p.Abstract == "" {
					p.Abstract = s2Data.Abstract
				}
				if p.Citations == 0 {
					p.Citations = s2Data.CitationCount
				}
				if p.References == 0 {
					p.References = s2Data.ReferenceCount
				}
				if s2Data.OpenAccessPdf != nil {
					p.PDFURL = s2Data.OpenAccessPdf.URL
					p.OpenAccess = true
				}
				if s2Data.PaperID != "" {
					p.IDs["s2_id"] = s2Data.PaperID
					p.URLs["semantic_scholar"] = "https://www.semanticscholar.org/paper/" + s2Data.PaperID
				}
				if s2Data.ExternalIds.ArXivID != "" {
					p.IDs["arxiv_id"] = s2Data.ExternalIds.ArXivID
				}
			}
		}
	}

	if p.Title == "" {
		return nil, fmt.Errorf("could not find paper with DOI %s", doi)
	}

	p.Sources = []string{"crossref", "semantic"}
	return &p, nil
}
