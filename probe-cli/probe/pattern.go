package probe

import (
	"strings"
)

// DetectPattern analyzes auth, DOM, and network results to recommend a CLI pattern.
func DetectPattern(auth *AuthResult, dom *DOMResult, network *NetworkResult) *PatternResult {
	score := map[string]float64{
		"dom-scrape":   0,
		"api-reverse":  0,
		"form-submit":  0,
		"async-poll":   0,
	}

	// ── Signal: API endpoints found ──
	if network != nil && len(network.APIEndpoints) > 0 {
		score["api-reverse"] += 0.4

		// If multiple endpoints with auth headers → strong API-reverse signal
		authCount := 0
		for _, ep := range network.APIEndpoints {
			if ep.HasAuthHeader {
				authCount++
			}
		}
		if authCount > 0 {
			score["api-reverse"] += 0.3
		}

		// If there's a POST followed by GET with same base → async-poll pattern
		hasPost := false
		hasGet := false
		for _, ep := range network.APIEndpoints {
			if ep.Method == "POST" {
				hasPost = true
			}
			if ep.Method == "GET" {
				hasGet = true
			}
		}
		if hasPost && hasGet {
			score["async-poll"] += 0.2
		}
	}

	// ── Signal: Auth detected ──
	if auth != nil && auth.Detected {
		switch {
		case strings.Contains(auth.Method, "bearer") || strings.Contains(auth.Method, "localstorage"):
			score["api-reverse"] += 0.2
		case strings.Contains(auth.Method, "csrf"):
			score["api-reverse"] += 0.15
		case strings.Contains(auth.Method, "cookie"):
			score["api-reverse"] += 0.1
		}
	}

	// ── Signal: DOM has forms ──
	if dom != nil {
		if len(dom.Forms) > 0 {
			score["form-submit"] += 0.3
			// Forms with few inputs (search box) → simpler pattern
			for _, f := range dom.Forms {
				if f.InputCount <= 2 && f.Method == "GET" {
					score["dom-scrape"] += 0.1
				}
			}
		}
		// Standalone contenteditable → rich text / input for generation
		for _, inp := range dom.StandaloneInputs {
			if inp.ContentEditable {
				score["form-submit"] += 0.15
				score["async-poll"] += 0.1
			}
		}
		// Buttons with submit/generate keywords → async-poll hint
		for _, btn := range dom.Buttons {
			lower := strings.ToLower(btn.Text)
			if strings.Contains(lower, "生成") || strings.Contains(lower, "generate") ||
				strings.Contains(lower, "create") || strings.Contains(lower, "submit") {
				score["async-poll"] += 0.15
				score["form-submit"] += 0.1
			}
		}
	}

	// ── Signal: No API endpoints found → DOM scrape is the only option ──
	if network == nil || len(network.APIEndpoints) == 0 {
		score["dom-scrape"] += 0.5
	}

	// ── Pick highest scoring pattern ──
	best := "dom-scrape"
	bestScore := 0.0
	for pattern, s := range score {
		if s > bestScore {
			bestScore = s
			best = pattern
		}
	}

	// Cap confidence at 0.95
	if bestScore > 0.95 {
		bestScore = 0.95
	}

	reason := patternReason(best, auth, dom, network)

	return &PatternResult{
		Type:       best,
		Confidence: bestScore,
		Reason:     reason,
	}
}

func patternReason(pattern string, auth *AuthResult, dom *DOMResult, network *NetworkResult) string {
	switch pattern {
	case "dom-scrape":
		return "Data is available in page DOM; no API calls detected. Use snapshot + evaluate to extract."
	case "api-reverse":
		reason := "Data loaded via XHR/Fetch API calls"
		if auth != nil && auth.Detected {
			reason += " with " + auth.Method + " auth"
		}
		reason += ". Replicate API calls in evaluate()."
		return reason
	case "form-submit":
		return "Page has forms that accept user input. Fill form fields + submit via evaluate()."
	case "async-poll":
		return "Action triggers async processing (POST → poll GET for result). Implement submit + poll loop."
	default:
		return "Pattern unclear. Manual site archaeology recommended."
	}
}
