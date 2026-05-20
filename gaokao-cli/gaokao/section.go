package gaokao

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

// SectionEntry represents one row in the score-section table (一分一段表).
type SectionEntry struct {
	Score            string            `json:"score"`
	Num              string            `json:"num"`                          // 同分人数
	Total            string            `json:"total"`                        // 累计人数
	RankRange        string            `json:"rank_range"`                   // 排名区间
	BatchName        string            `json:"batch_name"`
	ControlScore     string            `json:"controlscore"`
	AppositiveScores []AppositiveScore `json:"appositive_fraction,omitempty"`
}

// AppositiveScore represents a historical equivalent score (同位分).
type AppositiveScore struct {
	Year      int    `json:"year"`
	Score     string `json:"score"`
	RankRange string `json:"rank_range"`
}

// SectionResult is the return value of FetchScoreSection.
type SectionResult struct {
	Year     string         `json:"year"`
	TypeName string         `json:"type_name"`
	Level    string         `json:"level"`
	Entries  []SectionEntry `json:"entries"`
}

// typeInfo holds id and name for a type or level entry.
type typeInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// FetchScoreSection fetches score-section (一分一段表) data.
func FetchScoreSection(provinceID, year, typeName, level, score string) (*SectionResult, error) {
	// 1. Fetch config
	url := fmt.Sprintf("%s/section2021/dicScoreSection.json", baseURL)
	var configEnvelope struct {
		Data       json.RawMessage `json:"data"`
		ProvinceID []string        `json:"province_id"`
	}
	if err := FetchData(url, &configEnvelope); err != nil {
		return nil, fmt.Errorf("fetch section config: %w", err)
	}

	// 2. Parse top-level: { "data": {provinceId: ...}, "province_id": [...] }
	var allProvinces map[string]json.RawMessage
	if err := json.Unmarshal(configEnvelope.Data, &allProvinces); err != nil {
		return nil, fmt.Errorf("parse config data: %w", err)
	}

	provinceRaw, ok := allProvinces[provinceID]
	if !ok {
		return nil, fmt.Errorf("no section data for province %s", provinceID)
	}

	// 3. Parse province config: { [year]: { type/level config } }
	var yearMap map[string]json.RawMessage
	if err := json.Unmarshal(provinceRaw, &yearMap); err != nil {
		return nil, fmt.Errorf("parse province year map: %w", err)
	}

	// Pick latest year if empty
	if year == "" {
		year = latestYearFromKeys(yearMap)
		if year == "" {
			return nil, fmt.Errorf("no years available for province %s", provinceID)
		}
	}

	yearRaw, ok := yearMap[year]
	if !ok {
		available := make([]string, 0, len(yearMap))
		for y := range yearMap {
			available = append(available, y)
		}
		sort.Strings(available)
		return nil, fmt.Errorf("no data for year %s, available years: %v", year, available)
	}

	// 4. Parse year config to find type and level
	// Structure: { "0": {"id":3,"name":"综合"}, "1": {"id":2,"name":"文科"}, "level": { "0": {"id":1,"name":"本科"}, "1": {"id":2,"name":"专科"}, "kemu": [...] } }
	var yearConfig map[string]json.RawMessage
	if err := json.Unmarshal(yearRaw, &yearConfig); err != nil {
		return nil, fmt.Errorf("parse year config: %w", err)
	}

	// Extract types (numeric keys, skip "level")
	var types []typeInfo
	for k, v := range yearConfig {
		if k == "level" {
			continue
		}
		// Check if key is numeric
		if _, err := strconv.Atoi(k); err != nil {
			continue
		}
		var t typeInfo
		if err := json.Unmarshal(v, &t); err != nil {
			continue
		}
		types = append(types, t)
	}
	sort.Slice(types, func(i, j int) bool { return types[i].ID < types[j].ID })

	if len(types) == 0 {
		return nil, fmt.Errorf("no types found for province %s year %s", provinceID, year)
	}

	// Resolve type
	var selectedType typeInfo
	if typeName == "" {
		selectedType = types[0]
	} else {
		found := false
		for _, t := range types {
			if t.Name == typeName {
				selectedType = t
				found = true
				break
			}
		}
		if !found {
			names := make([]string, len(types))
			for i, t := range types {
				names[i] = t.Name
			}
			return nil, fmt.Errorf("type %q not found, available: %v", typeName, names)
		}
	}

	// Extract levels from "level" key
	levelRaw, ok := yearConfig["level"]
	if !ok {
		return nil, fmt.Errorf("no level info for province %s year %s", provinceID, year)
	}
	var levelMap map[string]json.RawMessage
	if err := json.Unmarshal(levelRaw, &levelMap); err != nil {
		return nil, fmt.Errorf("parse level config: %w", err)
	}

	var levels []typeInfo
	for k, v := range levelMap {
		if k == "kemu" {
			continue
		}
		if _, err := strconv.Atoi(k); err != nil {
			continue
		}
		var l typeInfo
		if err := json.Unmarshal(v, &l); err != nil {
			continue
		}
		levels = append(levels, l)
	}
	sort.Slice(levels, func(i, j int) bool { return levels[i].ID < levels[j].ID })

	if len(levels) == 0 {
		return nil, fmt.Errorf("no levels found for province %s year %s", provinceID, year)
	}

	// Resolve level
	var selectedLevel typeInfo
	if level == "" {
		selectedLevel = levels[0]
	} else {
		found := false
		for _, l := range levels {
			if l.Name == level {
				selectedLevel = l
				found = true
				break
			}
		}
		if !found {
			names := make([]string, len(levels))
			for i, l := range levels {
				names[i] = l.Name
			}
			return nil, fmt.Errorf("level %q not found, available: %v", level, names)
		}
	}

	// 5. Fetch the actual data
	dataURL := fmt.Sprintf("%s/section2021/%s/%s/%d/%d/lists.json",
		baseURL, year, provinceID, selectedType.ID, selectedLevel.ID)

	var listData struct {
		Search map[string]json.RawMessage `json:"search"`
		List   []SectionEntry             `json:"list"`
	}
	if err := FetchData(dataURL, &listData); err != nil {
		return nil, fmt.Errorf("fetch section data: %w", err)
	}

	result := &SectionResult{
		Year:     year,
		TypeName: selectedType.Name,
		Level:    selectedLevel.Name,
	}

	// 6. If score is specified, look up in search map
	if score != "" {
		raw, ok := listData.Search[score]
		if !ok {
			return nil, fmt.Errorf("score %s not found in section data", score)
		}
		var entry SectionEntry
		if err := json.Unmarshal(raw, &entry); err != nil {
			return nil, fmt.Errorf("parse score entry: %w", err)
		}
		result.Entries = []SectionEntry{entry}
	} else {
		result.Entries = listData.List
	}

	return result, nil
}

// latestYearFromKeys returns the largest year string from a map's keys.
func latestYearFromKeys(m map[string]json.RawMessage) string {
	best := ""
	for y := range m {
		if y > best {
			best = y
		}
	}
	return best
}
