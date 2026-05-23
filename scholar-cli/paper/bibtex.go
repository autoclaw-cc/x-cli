package paper

import (
	"fmt"
	"regexp"
	"strings"
)

var nonAlpha = regexp.MustCompile(`[^a-zA-Z0-9]`)

func (p *Paper) BibTeXKey() string {
	author := "unknown"
	if len(p.Authors) > 0 {
		parts := strings.Fields(p.Authors[0].Name)
		if len(parts) > 0 {
			author = nonAlpha.ReplaceAllString(parts[len(parts)-1], "")
		}
	}
	year := "0000"
	if p.Year > 0 {
		year = fmt.Sprintf("%d", p.Year)
	}
	titleWord := "untitled"
	for _, word := range strings.Fields(p.Title) {
		clean := nonAlpha.ReplaceAllString(word, "")
		lower := strings.ToLower(clean)
		if len(clean) > 3 && lower != "the" && lower != "and" && lower != "for" && lower != "with" {
			titleWord = strings.ToLower(clean)
			break
		}
	}
	return strings.ToLower(author) + year + titleWord
}

func (p *Paper) BibTeXType() string {
	venueLower := strings.ToLower(p.Venue)
	if strings.Contains(venueLower, "conference") ||
		strings.Contains(venueLower, "proceedings") ||
		strings.Contains(venueLower, "symposium") ||
		strings.Contains(venueLower, "workshop") {
		return "inproceedings"
	}
	if strings.Contains(venueLower, "journal") ||
		p.Volume != "" ||
		p.Issue != "" {
		return "article"
	}
	if p.IDs["arxiv_id"] != "" {
		return "misc"
	}
	return "article"
}

func (p *Paper) ToBibTeX() string {
	entryType := p.BibTeXType()
	key := p.BibTeXKey()

	var b strings.Builder
	fmt.Fprintf(&b, "@%s{%s,\n", entryType, key)

	writeField := func(name, value string) {
		if value != "" {
			escaped := strings.NewReplacer(
				"&", "\\&",
				"%", "\\%",
				"#", "\\#",
				"_", "\\_",
			).Replace(value)
			fmt.Fprintf(&b, "  %s = {%s},\n", name, escaped)
		}
	}

	writeField("title", p.Title)

	if len(p.Authors) > 0 {
		var names []string
		for _, a := range p.Authors {
			names = append(names, a.Name)
		}
		writeField("author", strings.Join(names, " and "))
	}

	if p.Year > 0 {
		fmt.Fprintf(&b, "  year = {%d},\n", p.Year)
	}

	if entryType == "inproceedings" {
		writeField("booktitle", p.Venue)
	} else {
		writeField("journal", p.Venue)
	}

	writeField("volume", p.Volume)
	writeField("number", p.Issue)
	writeField("pages", p.Pages)
	writeField("doi", p.DOI)
	writeField("abstract", truncate(p.Abstract, 500))

	if p.PDFURL != "" {
		writeField("url", p.PDFURL)
	} else if p.DOI != "" {
		writeField("url", "https://doi.org/"+p.DOI)
	}

	if arxivID, ok := p.IDs["arxiv_id"]; ok {
		writeField("eprint", arxivID)
		writeField("archiveprefix", "arXiv")
	}

	b.WriteString("}\n")
	return b.String()
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

func ExportBibTeX(papers []Paper) string {
	var b strings.Builder
	seen := make(map[string]bool)
	for i := range papers {
		key := papers[i].BibTeXKey()
		if seen[key] {
			key = fmt.Sprintf("%s_%c", key, rune('a'+i%26))
		}
		seen[key] = true
		b.WriteString(papers[i].ToBibTeX())
		b.WriteString("\n")
	}
	return b.String()
}

