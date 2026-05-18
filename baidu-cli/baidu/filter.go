package baidu

import "strings"

// filterOrganic removes non-organic results from a Baidu SERP and keeps only
// genuine web results.  Baidu mixes organic results (tpl="www_index") with
// dozens of aladdin-card types: AI answer cards, Baike summaries, video
// recommendations, related-search stubs, etc.
//
// The --all flag bypasses this filter entirely (callers use includeAll=true).
func filterOrganic(results []Result) []Result {
	var out []Result
	for _, r := range results {
		// Drop empty entries (no title = no real result)
		if strings.TrimSpace(r.Title) == "" {
			continue
		}
		// Always drop "people also searched" recommendation stubs.
		if r.Tpl == "recommend_list" {
			continue
		}
		// Drop AI-generated answer cards that can hallucinate.
		if r.Tpl == "ai_agent_distribute" {
			continue
		}
		// Keep: organic web results, Baike entity cards, and generic fallbacks.
		switch r.Tpl {
		case "www_index", "se_com_default", "":
			out = append(out, r)
		default:
			if strings.HasPrefix(r.Tpl, "sg_kg_") {
				// Baidu Baike entity card — usually worth keeping.
				out = append(out, r)
			}
		}
	}
	return out
}
