package gaokao

import (
	"encoding/json"
	"fmt"
	"sort"
)

// ScoreLineEntry represents one batch line for a province/year/type.
type ScoreLineEntry struct {
	Year         string          `json:"year"`
	Province     string          `json:"province"`
	TypeName     string          `json:"type_name"`
	BatchName    string          `json:"batch_name"`
	Score        string          `json:"score"`
	MajorScore   string          `json:"major_score,omitempty"`
	ScoreSection string          `json:"score_section,omitempty"`
	Diff         json.RawMessage `json:"diff"`
}

// ScoreLineResult is the return value of FetchScoreLine.
type ScoreLineResult struct {
	Year  string           `json:"year"`
	Lines []ScoreLineEntry `json:"lines"`
}

// rawEntry mirrors the JSON entry from the API, using json.RawMessage for diff.
type rawEntry struct {
	Type           string          `json:"type"`
	TypeName       string          `json:"type_name"`
	BatchName      string          `json:"batch_name"`
	Batch          string          `json:"batch"`
	Score          string          `json:"score"`
	MajorScore     string          `json:"major_score"`
	Rank           string          `json:"rank"`
	Year           string          `json:"year"`
	Province       string          `json:"province"`
	ScoreSection   string          `json:"score_section"`
	ConventionBatch bool           `json:"convention_batch"`
	Name           string          `json:"name"`
	Diff           json.RawMessage `json:"diff"`
}

// FetchScoreLine fetches score-line (省控线) data for a province.
// If year is empty, the latest available year is used.
// If typeName is empty, all types for that year are returned.
// If typeName is specified, only entries matching that type_name are returned.
func FetchScoreLine(provinceID, year, typeName string) (*ScoreLineResult, error) {
	url := fmt.Sprintf("%s/proprovince/%s/pro.json", baseURL, provinceID)

	// The response data is: { "2025": { "t_3": [...], ... }, ... }
	var data map[string]map[string][]rawEntry
	if err := FetchData(url, &data); err != nil {
		return nil, fmt.Errorf("fetch score-line: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no score-line data for province %s", provinceID)
	}

	// Determine year
	if year == "" {
		year = latestYear(data)
	}

	yearData, ok := data[year]
	if !ok {
		available := make([]string, 0, len(data))
		for y := range data {
			available = append(available, y)
		}
		sort.Strings(available)
		return nil, fmt.Errorf("no data for year %s, available years: %v", year, available)
	}

	// Collect entries
	var lines []ScoreLineEntry
	for _, entries := range yearData {
		for _, e := range entries {
			if typeName != "" && e.TypeName != typeName {
				continue
			}
			lines = append(lines, ScoreLineEntry{
				Year:         e.Year,
				Province:     e.Province,
				TypeName:     e.TypeName,
				BatchName:    e.BatchName,
				Score:        e.Score,
				MajorScore:   e.MajorScore,
				ScoreSection: e.ScoreSection,
				Diff:         e.Diff,
			})
		}
	}

	if len(lines) == 0 && typeName != "" {
		return nil, fmt.Errorf("no data for type %q in year %s", typeName, year)
	}

	return &ScoreLineResult{
		Year:  year,
		Lines: lines,
	}, nil
}

// latestYear returns the largest year key from the data map.
func latestYear(data map[string]map[string][]rawEntry) string {
	best := ""
	for y := range data {
		if y > best {
			best = y
		}
	}
	return best
}
