package probe

import (
	"encoding/json"
	"fmt"
	"strings"

	"probe-cli/browser"
)

// Static file extensions to filter out.
var staticExts = []string{
	".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico",
	".woff", ".woff2", ".ttf", ".eot", ".otf",
	".mp4", ".mp3", ".webm", ".webp", ".avif",
}

// rawNetworkList matches the expected daemon response for "list" command.
type rawNetworkItem struct {
	RequestID string `json:"requestId"`
	URL       string `json:"url"`
	Method    string `json:"method"`
	Status    int    `json:"status"`
	Type      string `json:"type"` // XHR, Fetch, Document, Script, Stylesheet, Image, etc.
}

// rawNetworkDetail matches the expected daemon response for "detail" command.
type rawNetworkDetail struct {
	RequestID   string            `json:"requestId"`
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Status      int               `json:"status"`
	RequestHeaders map[string]string `json:"requestHeaders"`
	RequestBody string            `json:"requestBody"`
	ResponseHeaders map[string]string `json:"responseHeaders"`
	ContentType string            `json:"contentType"`
}

// AnalyzeNetwork stops capture, lists requests, filters to API calls, and inspects details.
func AnalyzeNetwork(client *browser.Client) (*NetworkResult, []string) {
	var warnings []string

	// Get list of captured requests
	rawList, err := client.NetworkList()
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("network list failed: %v", err))
		return &NetworkResult{}, warnings
	}

	var items []rawNetworkItem
	if err := json.Unmarshal(rawList, &items); err != nil {
		// Try as map with "requests" key
		var wrapper struct {
			Requests []rawNetworkItem `json:"requests"`
		}
		if err2 := json.Unmarshal(rawList, &wrapper); err2 != nil {
			warnings = append(warnings, fmt.Sprintf("network list parse failed: %v", err))
			return &NetworkResult{}, warnings
		}
		items = wrapper.Requests
	}

	result := &NetworkResult{
		TotalRequests: len(items),
	}

	// Filter: keep only XHR and Fetch requests
	var apiItems []rawNetworkItem
	for _, item := range items {
		if isStaticURL(item.URL) {
			result.StaticFiltered++
			continue
		}
		typ := strings.ToUpper(item.Type)
		if typ == "XHR" || typ == "FETCH" || typ == "" {
			// Empty type might be from older daemon versions; include and check later
			apiItems = append(apiItems, item)
		}
	}

	// Deduplicate by URL (keep first occurrence per unique URL)
	seen := make(map[string]bool)
	var uniqueAPI []rawNetworkItem
	for _, item := range apiItems {
		if !seen[item.URL] {
			seen[item.URL] = true
			uniqueAPI = append(uniqueAPI, item)
		}
	}

	// Get details for up to 15 unique API endpoints
	limit := 15
	if len(uniqueAPI) < limit {
		limit = len(uniqueAPI)
	}

	for i := 0; i < limit; i++ {
		item := uniqueAPI[i]
		endpoint := APIEndpoint{
			Method: item.Method,
			URL:    item.URL,
			Status: item.Status,
		}

		// Try to get details for auth header inspection
		detail, err := client.NetworkDetail(item.RequestID)
		if err != nil {
			// Detail failed — still include the endpoint with basic info
			result.APIEndpoints = append(result.APIEndpoints, endpoint)
			continue
		}

		var d rawNetworkDetail
		if err := json.Unmarshal(detail, &d); err != nil {
			result.APIEndpoints = append(result.APIEndpoints, endpoint)
			continue
		}

		// Check for auth-related headers
		for key, val := range d.RequestHeaders {
			lower := strings.ToLower(key)
			if strings.Contains(lower, "auth") || strings.Contains(lower, "token") ||
				strings.Contains(lower, "csrf") || strings.Contains(lower, "x-csrf") {
				endpoint.HasAuthHeader = true
				// Mask token value for safety (show first 8 chars + ...)
				masked := maskSecret(val)
				endpoint.AuthHeaders = append(endpoint.AuthHeaders, fmt.Sprintf("%s: %s", key, masked))
			}
		}

		if ct, ok := d.RequestHeaders["Content-Type"]; ok {
			endpoint.ContentType = ct
		}

		result.APIEndpoints = append(result.APIEndpoints, endpoint)
	}

	return result, warnings
}

// isStaticURL returns true for static asset URLs.
func isStaticURL(url string) bool {
	// Strip query string for extension check
	path := url
	if idx := strings.Index(url, "?"); idx >= 0 {
		path = url[:idx]
	}
	lower := strings.ToLower(path)
	for _, ext := range staticExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// maskSecret hides most of a secret value, keeping just a prefix.
func maskSecret(val string) string {
	if len(val) <= 12 {
		return "***"
	}
	return val[:8] + "***"
}
