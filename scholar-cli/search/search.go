package search

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"scholar-cli/paper"
)

type Source interface {
	Name() string
	Search(ctx context.Context, query string, limit int) ([]paper.Paper, error)
}

type SearchResult struct {
	Papers  []paper.Paper    `json:"papers"`
	Total   int              `json:"total"`
	Sources []SourceSummary  `json:"sources"`
}

type SourceSummary struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Error string `json:"error,omitempty"`
}

var defaultHTTPClient = &http.Client{Timeout: 30 * time.Second}

func allEnglishSources() []Source {
	return []Source{
		NewOpenAlex(),
		NewSemanticScholar(),
		NewCrossRef(),
		NewArXiv(),
		NewPubMed(),
		NewDBLP(),
		NewBioRxiv(),
	}
}

func SearchEnglish(ctx context.Context, query string, limit int, sourceNames []string) (*SearchResult, error) {
	all := allEnglishSources()

	var sources []Source
	if len(sourceNames) > 0 {
		nameMap := make(map[string]Source)
		for _, s := range all {
			nameMap[strings.ToLower(s.Name())] = s
		}
		for _, name := range sourceNames {
			s, ok := nameMap[strings.ToLower(strings.TrimSpace(name))]
			if !ok {
				return nil, fmt.Errorf("unknown source: %s", name)
			}
			sources = append(sources, s)
		}
	} else {
		sources = all
	}

	return SearchAll(ctx, sources, query, limit)
}

func SearchAll(ctx context.Context, sources []Source, query string, limit int) (*SearchResult, error) {
	type sourceResult struct {
		name   string
		papers []paper.Paper
		err    error
	}

	results := make([]sourceResult, len(sources))
	var wg sync.WaitGroup

	for i, src := range sources {
		wg.Add(1)
		go func(idx int, s Source) {
			defer wg.Done()
			papers, err := s.Search(ctx, query, limit)
			results[idx] = sourceResult{name: s.Name(), papers: papers, err: err}
		}(i, src)
	}

	wg.Wait()

	var allPapers []paper.Paper
	var summaries []SourceSummary

	for _, r := range results {
		summary := SourceSummary{Name: r.name, Count: len(r.papers)}
		if r.err != nil {
			summary.Error = r.err.Error()
		}
		summaries = append(summaries, summary)
		allPapers = append(allPapers, r.papers...)
	}

	deduped := paper.Deduplicate(allPapers)

	return &SearchResult{
		Papers:  deduped,
		Total:   len(deduped),
		Sources: summaries,
	}, nil
}
