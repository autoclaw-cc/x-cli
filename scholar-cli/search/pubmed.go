package search

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"scholar-cli/paper"
)

type PubMed struct{}

func NewPubMed() *PubMed { return &PubMed{} }
func (p *PubMed) Name() string { return "pubmed" }

func (p *PubMed) Search(ctx context.Context, query string, limit int) ([]paper.Paper, error) {
	if limit <= 0 {
		limit = 10
	}

	// Step 1: search for IDs (tool and email params required for NCBI polite access)
	searchURL := fmt.Sprintf(
		"https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi?db=pubmed&term=%s&retmax=%d&retmode=json&tool=scholar-cli&email=scholar-cli@example.com",
		url.QueryEscape(query), limit)

	req, err := newRequest(ctx, "GET", searchURL)
	if err != nil {
		return nil, err
	}
	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pubmed search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("pubmed search returned status %d (may be rate-limited)", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "json") && !strings.Contains(ct, "javascript") {
		return nil, fmt.Errorf("pubmed returned non-JSON response (content-type: %s), likely blocked by NCBI", ct)
	}

	var searchResult struct {
		ESearchResult struct {
			IDList []string `json:"idlist"`
		} `json:"esearchresult"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("pubmed search decode: %w", err)
	}

	ids := searchResult.ESearchResult.IDList
	if len(ids) == 0 {
		return nil, nil
	}

	// Step 2: fetch details
	fetchURL := fmt.Sprintf(
		"https://eutils.ncbi.nlm.nih.gov/entrez/eutils/efetch.fcgi?db=pubmed&id=%s&retmode=xml&tool=scholar-cli&email=scholar-cli@example.com",
		strings.Join(ids, ","))

	req2, err := newRequest(ctx, "GET", fetchURL)
	if err != nil {
		return nil, err
	}
	resp2, err := defaultHTTPClient.Do(req2)
	if err != nil {
		return nil, fmt.Errorf("pubmed fetch: %w", err)
	}
	defer resp2.Body.Close()

	var articleSet struct {
		XMLName  xml.Name `xml:"PubmedArticleSet"`
		Articles []struct {
			MedlineCitation struct {
				PMID struct {
					Value string `xml:",chardata"`
				} `xml:"PMID"`
				Article struct {
					ArticleTitle string `xml:"ArticleTitle"`
					Abstract     struct {
						AbstractText []struct {
							Label string `xml:"Label,attr"`
							Text  string `xml:",chardata"`
						} `xml:"AbstractText"`
					} `xml:"Abstract"`
					AuthorList struct {
						Authors []struct {
							LastName    string `xml:"LastName"`
							ForeName    string `xml:"ForeName"`
							Affiliation string `xml:"AffiliationInfo>Affiliation"`
						} `xml:"Author"`
					} `xml:"AuthorList"`
					Journal struct {
						Title   string `xml:"Title"`
						Volume  string `xml:"JournalIssue>Volume"`
						Issue   string `xml:"JournalIssue>Issue"`
						PubDate struct {
							Year string `xml:"Year"`
						} `xml:"JournalIssue>PubDate"`
					} `xml:"Journal"`
					Pagination struct {
						MedlinePgn string `xml:"MedlinePgn"`
					} `xml:"Pagination"`
					ELocationID []struct {
						EIdType string `xml:"EIdType,attr"`
						Value   string `xml:",chardata"`
					} `xml:"ELocationID"`
				} `xml:"Article"`
			} `xml:"MedlineCitation"`
		} `xml:"PubmedArticle"`
	}

	if err := xml.NewDecoder(resp2.Body).Decode(&articleSet); err != nil {
		return nil, fmt.Errorf("pubmed xml decode: %w", err)
	}

	var papers []paper.Paper
	for _, a := range articleSet.Articles {
		mc := a.MedlineCitation
		art := mc.Article

		var authors []paper.Author
		for _, au := range art.AuthorList.Authors {
			name := strings.TrimSpace(au.ForeName + " " + au.LastName)
			authors = append(authors, paper.Author{Name: name, Affiliation: au.Affiliation})
		}

		var abstractParts []string
		for _, at := range art.Abstract.AbstractText {
			if at.Label != "" {
				abstractParts = append(abstractParts, at.Label+": "+at.Text)
			} else {
				abstractParts = append(abstractParts, at.Text)
			}
		}
		abstract := strings.Join(abstractParts, " ")

		year := 0
		if art.Journal.PubDate.Year != "" {
			year, _ = strconv.Atoi(art.Journal.PubDate.Year)
		}

		doi := ""
		for _, eid := range art.ELocationID {
			if eid.EIdType == "doi" {
				doi = eid.Value
				break
			}
		}

		pmid := mc.PMID.Value
		paperIDs := map[string]string{"pmid": pmid}
		if doi != "" {
			paperIDs["doi"] = doi
		}

		pp := paper.Paper{
			Title:    art.ArticleTitle,
			Authors:  authors,
			Abstract: abstract,
			Year:     year,
			DOI:      doi,
			Venue:    art.Journal.Title,
			Volume:   art.Journal.Volume,
			Issue:    art.Journal.Issue,
			Pages:    art.Pagination.MedlinePgn,
			Source:   "pubmed",
			URLs:     map[string]string{"pubmed": "https://pubmed.ncbi.nlm.nih.gov/" + pmid + "/"},
			IDs:      paperIDs,
		}
		papers = append(papers, pp)
	}

	return papers, nil
}
