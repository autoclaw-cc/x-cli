package cmd

import (
	"reflect"
	"testing"

	"scholar-cli/paper"
	"scholar-cli/search"
)

func TestEnglishSearchDataWorkspaceIsFlat(t *testing.T) {
	result := &search.SearchResult{
		Papers:  []paper.Paper{{Title: "Example"}},
		Total:   1,
		Sources: []search.SourceSummary{{Name: "arxiv", Count: 1}},
	}
	want := map[string]any{
		"papers":       result.Papers,
		"total":        1,
		"sources":      result.Sources,
		"workspace":    "/tmp/papers",
		"papers_added": 1,
		"total_stored": 2,
	}

	if got := englishSearchData(result, "/tmp/papers", 1, 2); !reflect.DeepEqual(got, want) {
		t.Fatalf("englishSearchData() = %#v, want %#v", got, want)
	}
}
