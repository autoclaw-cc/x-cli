package paper

import (
	"regexp"
	"strings"
	"unicode"
)

type Author struct {
	Name        string `json:"name"`
	Affiliation string `json:"affiliation,omitempty"`
}

type Paper struct {
	Title      string            `json:"title"`
	Authors    []Author          `json:"authors"`
	Abstract   string            `json:"abstract,omitempty"`
	Year       int               `json:"year,omitempty"`
	DOI        string            `json:"doi,omitempty"`
	Venue      string            `json:"venue,omitempty"`
	Volume     string            `json:"volume,omitempty"`
	Issue      string            `json:"issue,omitempty"`
	Pages      string            `json:"pages,omitempty"`
	Citations  int               `json:"citations,omitempty"`
	References int               `json:"references,omitempty"`
	OpenAccess bool              `json:"open_access,omitempty"`
	PDFURL     string            `json:"pdf_url,omitempty"`
	Source     string            `json:"source"`
	Sources    []string          `json:"sources,omitempty"`
	URLs       map[string]string `json:"urls,omitempty"`
	IDs        map[string]string `json:"identifiers,omitempty"`
}

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9\s]`)
var multiSpace = regexp.MustCompile(`\s+`)

func NormTitle(title string) string {
	s := strings.ToLower(title)
	s = nonAlphaNum.ReplaceAllString(s, "")
	s = multiSpace.ReplaceAllString(strings.TrimSpace(s), " ")
	return s
}

func firstAuthorLastName(authors []Author) string {
	if len(authors) == 0 {
		return ""
	}
	name := authors[0].Name
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]
	return strings.ToLower(strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, last))
}

func (p *Paper) DedupeKey() string {
	if p.DOI != "" {
		return "doi:" + strings.ToLower(p.DOI)
	}
	norm := NormTitle(p.Title)
	if norm == "" {
		return ""
	}
	author := firstAuthorLastName(p.Authors)
	if author != "" {
		return "ta:" + norm + "|" + author
	}
	if len(norm) > 20 {
		return "t:" + norm
	}
	return ""
}

func (p *Paper) MergeFrom(other *Paper) {
	if p.Abstract == "" && other.Abstract != "" {
		p.Abstract = other.Abstract
	}
	if p.Year == 0 && other.Year != 0 {
		p.Year = other.Year
	}
	if p.DOI == "" && other.DOI != "" {
		p.DOI = other.DOI
	}
	if p.Venue == "" && other.Venue != "" {
		p.Venue = other.Venue
	}
	if p.Volume == "" && other.Volume != "" {
		p.Volume = other.Volume
	}
	if p.Issue == "" && other.Issue != "" {
		p.Issue = other.Issue
	}
	if p.Pages == "" && other.Pages != "" {
		p.Pages = other.Pages
	}
	if p.Citations == 0 && other.Citations != 0 {
		p.Citations = other.Citations
	}
	if p.References == 0 && other.References != 0 {
		p.References = other.References
	}
	if !p.OpenAccess && other.OpenAccess {
		p.OpenAccess = other.OpenAccess
	}
	if p.PDFURL == "" && other.PDFURL != "" {
		p.PDFURL = other.PDFURL
	}
	if len(p.Authors) == 0 && len(other.Authors) > 0 {
		p.Authors = other.Authors
	}

	if p.URLs == nil {
		p.URLs = make(map[string]string)
	}
	for k, v := range other.URLs {
		if _, exists := p.URLs[k]; !exists {
			p.URLs[k] = v
		}
	}

	if p.IDs == nil {
		p.IDs = make(map[string]string)
	}
	for k, v := range other.IDs {
		if _, exists := p.IDs[k]; !exists {
			p.IDs[k] = v
		}
	}

	found := false
	for _, s := range p.Sources {
		if s == other.Source {
			found = true
			break
		}
	}
	if !found && other.Source != "" {
		p.Sources = append(p.Sources, other.Source)
	}
}

func Deduplicate(papers []Paper) []Paper {
	seen := make(map[string]int)
	var result []Paper

	for i := range papers {
		p := &papers[i]
		if p.Sources == nil {
			p.Sources = []string{}
		}
		if p.Source != "" {
			found := false
			for _, s := range p.Sources {
				if s == p.Source {
					found = true
					break
				}
			}
			if !found {
				p.Sources = append(p.Sources, p.Source)
			}
		}

		key := p.DedupeKey()
		if key == "" {
			result = append(result, *p)
			continue
		}
		if idx, exists := seen[key]; exists {
			result[idx].MergeFrom(p)
		} else {
			seen[key] = len(result)
			result = append(result, *p)
		}
	}

	// Second pass: check identifier overlap for remaining papers
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if shareIdentifier(&result[i], &result[j]) {
				result[i].MergeFrom(&result[j])
				result = append(result[:j], result[j+1:]...)
				j--
			}
		}
	}

	return result
}

func shareIdentifier(a, b *Paper) bool {
	for k, v := range a.IDs {
		if v == "" {
			continue
		}
		if bv, ok := b.IDs[k]; ok && strings.EqualFold(v, bv) {
			return true
		}
	}
	return false
}
