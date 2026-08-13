package search

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"scholar-cli/paper"
)

type ArXiv struct{}

func NewArXiv() *ArXiv { return &ArXiv{} }
func (a *ArXiv) Name() string { return "arxiv" }

func (a *ArXiv) Search(ctx context.Context, query string, limit int) ([]paper.Paper, error) {
	if limit <= 0 {
		limit = 10
	}
	u := fmt.Sprintf("http://export.arxiv.org/api/query?search_query=all:%s&max_results=%d&sortBy=relevance",
		url.QueryEscape(query), limit)
	return fetchArXivFeed(ctx, u)
}

// FetchByArXivID looks a single record up by its bare arXiv ID (e.g.
// "2309.15217"). arXiv's own API imposes no shared rate limit, which makes it a
// dependable fallback for arXiv DOIs — those 404 on CrossRef by design, leaving
// Semantic Scholar as the only other source and its no-key tier as a single
// point of failure.
func FetchByArXivID(ctx context.Context, id string) (*paper.Paper, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("empty arxiv id")
	}
	u := fmt.Sprintf("http://export.arxiv.org/api/query?id_list=%s&max_results=1",
		url.QueryEscape(id))
	papers, err := fetchArXivFeed(ctx, u)
	if err != nil {
		return nil, err
	}
	if len(papers) == 0 {
		return nil, fmt.Errorf("arxiv has no record for %s", id)
	}
	return &papers[0], nil
}

func fetchArXivFeed(ctx context.Context, u string) ([]paper.Paper, error) {
	req, err := newRequest(ctx, "GET", u)
	if err != nil {
		return nil, err
	}

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("arxiv request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("arxiv returned HTTP %d", resp.StatusCode)
	}

	var feed struct {
		XMLName xml.Name `xml:"feed"`
		Entries []struct {
			Title     string `xml:"title"`
			Summary   string `xml:"summary"`
			Published string `xml:"published"`
			ID        string `xml:"id"`
			DOI       string `xml:"doi"`
			Authors   []struct {
				Name        string `xml:"name"`
				Affiliation string `xml:"affiliation"`
			} `xml:"author"`
			Links []struct {
				Href  string `xml:"href,attr"`
				Title string `xml:"title,attr"`
				Type  string `xml:"type,attr"`
				Rel   string `xml:"rel,attr"`
			} `xml:"link"`
			Categories []struct {
				Term string `xml:"term,attr"`
			} `xml:"category"`
		} `xml:"entry"`
	}

	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("arxiv decode: %w", err)
	}

	var papers []paper.Paper
	for _, e := range feed.Entries {
		var authors []paper.Author
		for _, a := range e.Authors {
			authors = append(authors, paper.Author{Name: a.Name, Affiliation: a.Affiliation})
		}

		year := 0
		if len(e.Published) >= 4 {
			fmt.Sscanf(e.Published[:4], "%d", &year)
		}

		arxivID := e.ID
		if idx := strings.LastIndex(arxivID, "/abs/"); idx >= 0 {
			arxivID = arxivID[idx+5:]
		}
		// Strip version suffix
		if idx := strings.LastIndex(arxivID, "v"); idx > 0 {
			if _, err := fmt.Sscanf(arxivID[idx+1:], "%d", new(int)); err == nil {
				arxivID = arxivID[:idx]
			}
		}

		pdfURL := ""
		for _, link := range e.Links {
			if link.Title == "pdf" || link.Type == "application/pdf" {
				pdfURL = link.Href
				break
			}
		}

		abstract := strings.TrimSpace(e.Summary)
		abstract = strings.ReplaceAll(abstract, "\n", " ")

		title := strings.TrimSpace(e.Title)
		title = strings.ReplaceAll(title, "\n", " ")
		title = strings.Join(strings.Fields(title), " ")

		ids := map[string]string{"arxiv_id": arxivID}
		if e.DOI != "" {
			ids["doi"] = e.DOI
		}

		p := paper.Paper{
			Title:      title,
			Authors:    authors,
			Abstract:   abstract,
			Year:       year,
			DOI:        e.DOI,
			OpenAccess: true,
			PDFURL:     pdfURL,
			Source:     "arxiv",
			URLs:       map[string]string{"arxiv": e.ID},
			IDs:        ids,
		}
		papers = append(papers, p)
	}

	return papers, nil
}
