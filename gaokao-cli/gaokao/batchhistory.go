package gaokao

import (
	"fmt"
	"sort"
)

type BatchHistoryEntry struct {
	Year        string `json:"year"`
	Description string `json:"description"`
}

type ProvinceBatchHistory struct {
	ProvinceID   string              `json:"province_id"`
	ProvinceName string              `json:"province_name"`
	History      []BatchHistoryEntry `json:"history"`
}

// FetchBatchHistory fetches batch merge/reform history for provinces.
// If provinceID is empty, returns all provinces that have history.
// If provinceID is specified, returns only that province's history.
func FetchBatchHistory(provinceID string) ([]ProvinceBatchHistory, error) {
	url := fmt.Sprintf("%s/province/batch_merge.json", baseURL)

	// data is {provinceID: {year: description}}
	var data map[string]map[string]string
	if err := FetchData(url, &data); err != nil {
		return nil, fmt.Errorf("fetch batch history: %w", err)
	}

	var results []ProvinceBatchHistory

	if provinceID != "" {
		// Single province
		yearMap, ok := data[provinceID]
		if !ok || len(yearMap) == 0 {
			name := provinceMap[provinceID]
			if name == "" {
				name = provinceID
			}
			return []ProvinceBatchHistory{{
				ProvinceID:   provinceID,
				ProvinceName: name,
				History:      []BatchHistoryEntry{},
			}}, nil
		}
		results = append(results, ProvinceBatchHistory{
			ProvinceID:   provinceID,
			ProvinceName: provinceMap[provinceID],
			History:      buildHistory(yearMap),
		})
	} else {
		// All provinces with history, in standard order
		for _, id := range provinceOrder {
			yearMap, ok := data[id]
			if !ok || len(yearMap) == 0 {
				continue
			}
			results = append(results, ProvinceBatchHistory{
				ProvinceID:   id,
				ProvinceName: provinceMap[id],
				History:      buildHistory(yearMap),
			})
		}
	}

	return results, nil
}

// buildHistory converts a year->description map into a sorted slice.
func buildHistory(yearMap map[string]string) []BatchHistoryEntry {
	entries := make([]BatchHistoryEntry, 0, len(yearMap))
	for year, desc := range yearMap {
		entries = append(entries, BatchHistoryEntry{Year: year, Description: desc})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Year < entries[j].Year
	})
	return entries
}
