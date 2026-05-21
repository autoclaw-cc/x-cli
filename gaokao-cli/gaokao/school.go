package gaokao

import (
	"sort"
	"strconv"
	"strings"
)

type School struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Province  string `json:"province"`
	City      string `json:"city"`
	Is985     bool   `json:"is_985"`
	Is211     bool   `json:"is_211"`
	DualClass bool   `json:"dual_class"` // 双一流
	Nature    string `json:"nature"`     // 公办/民办
	Level     string `json:"level"`      // 普通本科/专科（高职）
}

type SchoolFilter struct {
	Name      string
	Province  string
	Is985     bool
	Is211     bool
	DualClass bool
	Level     string // 本科 or 专科
}

type rawSchool struct {
	Name      string `json:"name"`
	F985      string `json:"f985"`
	F211      string `json:"f211"`
	DualClass string `json:"dual_class"`
	Province  string `json:"p"`
	City      string `json:"c"`
	Nature    string `json:"nature"`
	Level     string `json:"level"`
	AnswerURL string `json:"answerurl"`
}

// schoolRank returns a relevance score for sorting (lower = better match).
// When a name filter is active, exact matches rank highest, then shorter names
// (more relevant), then school tier (985 > 211 > 双一流 > others).
func schoolRank(s School, nameFilter string) int {
	rank := 0

	if nameFilter != "" {
		if s.Name == nameFilter {
			// Exact match — highest priority
			rank += 0
		} else {
			// Partial match — penalize by extra name length
			// e.g. "北京工业大学" (len 5) beats "北京工业大学耿丹学院" (len 9)
			extra := len([]rune(s.Name)) - len([]rune(nameFilter))
			rank += 1000 + extra*100
		}
	}

	// Tier bonus: 985 > 211 > 双一流 > others
	switch {
	case s.Is985:
		rank += 0
	case s.Is211:
		rank += 10
	case s.DualClass:
		rank += 20
	default:
		rank += 30
	}

	return rank
}

func FetchSchools(filter SchoolFilter) ([]School, error) {
	url := baseURL + "/school/list_v2.json"
	var raw map[string]rawSchool
	if err := FetchData(url, &raw); err != nil {
		return nil, err
	}

	var schools []School
	for id, r := range raw {
		s := School{
			ID:        id,
			Name:      r.Name,
			Province:  r.Province,
			City:      r.City,
			Is985:     r.F985 == "1",
			Is211:     r.F211 == "1",
			DualClass: r.DualClass == "1",
			Nature:    r.Nature,
			Level:     r.Level,
		}

		if filter.Name != "" && !strings.Contains(s.Name, filter.Name) {
			continue
		}
		if filter.Province != "" && !strings.Contains(s.Province, filter.Province) {
			continue
		}
		if filter.Is985 && !s.Is985 {
			continue
		}
		if filter.Is211 && !s.Is211 {
			continue
		}
		if filter.DualClass && !s.DualClass {
			continue
		}
		if filter.Level != "" && !strings.Contains(s.Level, filter.Level) {
			continue
		}

		schools = append(schools, s)
	}

	// Sort by relevance: exact name match first, then shorter name, then
	// school tier (985 > 211 > 双一流 > others), then numeric ID as tiebreaker.
	sort.Slice(schools, func(i, j int) bool {
		ri := schoolRank(schools[i], filter.Name)
		rj := schoolRank(schools[j], filter.Name)
		if ri != rj {
			return ri < rj
		}
		// Numeric ID tiebreaker (smaller ID = older/more established school)
		ni, _ := strconv.Atoi(schools[i].ID)
		nj, _ := strconv.Atoi(schools[j].ID)
		return ni < nj
	})

	return schools, nil
}
