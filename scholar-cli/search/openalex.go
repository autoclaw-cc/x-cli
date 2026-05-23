package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"scholar-cli/paper"
)

type OpenAlex struct{}

func NewOpenAlex() *OpenAlex { return &OpenAlex{} }
func (o *OpenAlex) Name() string { return "openalex" }

func (o *OpenAlex) Search(ctx context.Context, query string, limit int) ([]paper.Paper, error) {
	if limit <= 0 {
		limit = 10
	}
	u := fmt.Sprintf("https://api.openalex.org/works?search=%s&per_page=%d&mailto=%s",
		url.QueryEscape(query), limit, url.QueryEscape(ContactEmail()))

	req, err := newRequest(ctx, "GET", u)
	if err != nil {
		return nil, err
	}

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openalex request: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		Results []struct {
			Title           string           `json:"title"`
			DOI             string           `json:"doi"`
			PublicationYear int              `json:"publication_year"`
			CitedByCount    int              `json:"cited_by_count"`
			ID              string           `json:"id"`
			AbstractIndex   map[string][]int `json:"abstract_inverted_index"`
			PrimaryLocation struct {
				Source struct {
					DisplayName string `json:"display_name"`
				} `json:"source"`
			} `json:"primary_location"`
			OpenAccess struct {
				IsOA  bool   `json:"is_oa"`
				OAURL string `json:"oa_url"`
			} `json:"open_access"`
			Authorships []struct {
				Author struct {
					DisplayName string `json:"display_name"`
				} `json:"author"`
				Institutions []struct {
					DisplayName string `json:"display_name"`
				} `json:"institutions"`
			} `json:"authorships"`
			Biblio struct {
				Volume    string `json:"volume"`
				Issue     string `json:"issue"`
				FirstPage string `json:"first_page"`
				LastPage  string `json:"last_page"`
			} `json:"biblio"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("openalex decode: %w", err)
	}

	var papers []paper.Paper
	for _, w := range data.Results {
		doi := w.DOI
		if strings.HasPrefix(doi, "https://doi.org/") {
			doi = doi[len("https://doi.org/"):]
		}

		var authors []paper.Author
		for _, a := range w.Authorships {
			aff := ""
			if len(a.Institutions) > 0 {
				aff = a.Institutions[0].DisplayName
			}
			authors = append(authors, paper.Author{Name: a.Author.DisplayName, Affiliation: aff})
		}

		abstract := reconstructAbstract(w.AbstractIndex)

		pages := ""
		if w.Biblio.FirstPage != "" {
			pages = w.Biblio.FirstPage
			if w.Biblio.LastPage != "" {
				pages += "-" + w.Biblio.LastPage
			}
		}

		oaID := w.ID
		if strings.HasPrefix(oaID, "https://openalex.org/") {
			oaID = oaID[len("https://openalex.org/"):]
		}

		p := paper.Paper{
			Title:      w.Title,
			Authors:    authors,
			Abstract:   abstract,
			Year:       w.PublicationYear,
			DOI:        doi,
			Venue:      w.PrimaryLocation.Source.DisplayName,
			Volume:     w.Biblio.Volume,
			Issue:      w.Biblio.Issue,
			Pages:      pages,
			Citations:  w.CitedByCount,
			OpenAccess: w.OpenAccess.IsOA,
			PDFURL:     w.OpenAccess.OAURL,
			Source:     "openalex",
			URLs:       map[string]string{"openalex": w.ID},
			IDs:        map[string]string{"openalex_id": oaID},
		}
		if doi != "" {
			p.IDs["doi"] = doi
		}
		papers = append(papers, p)
	}

	return papers, nil
}

func reconstructAbstract(index map[string][]int) string {
	if len(index) == 0 {
		return ""
	}
	type wordPos struct {
		word string
		pos  int
	}
	var words []wordPos
	for word, positions := range index {
		for _, pos := range positions {
			words = append(words, wordPos{word, pos})
		}
	}
	sort.Slice(words, func(i, j int) bool { return words[i].pos < words[j].pos })

	parts := make([]string, len(words))
	for i, w := range words {
		parts[i] = w.word
	}
	return strings.Join(parts, " ")
}
