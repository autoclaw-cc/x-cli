package probe

import (
	"encoding/json"
	"fmt"
	"time"

	"probe-cli/browser"
)

// SiteProfile is the final output of a probe run.
type SiteProfile struct {
	URL           string          `json:"url"`
	Timestamp     string          `json:"timestamp"`
	Auth          *AuthResult     `json:"auth"`
	DOM           *DOMResult      `json:"dom"`
	Network       *NetworkResult  `json:"network"`
	Pattern       *PatternResult  `json:"pattern"`
	Warnings      []string        `json:"warnings,omitempty"`
}

// AuthResult holds auth detection findings.
type AuthResult struct {
	Detected    bool            `json:"detected"`
	Method      string          `json:"method,omitempty"`      // bearer-localstorage, csrf-cookie, cookie-only, session-storage, none
	StorageType string          `json:"storage_type,omitempty"` // localStorage, sessionStorage, cookie
	TokenKeys   []string        `json:"token_keys,omitempty"`
	AllKeys     json.RawMessage `json:"all_keys,omitempty"`
	LoginURL    string          `json:"login_url,omitempty"`
}

// DOMResult holds extracted page element info.
type DOMResult struct {
	Forms            []FormInfo `json:"forms"`
	StandaloneInputs []InputInfo `json:"standalone_inputs"`
	Buttons          []ButtonInfo `json:"buttons"`
}

type FormInfo struct {
	Index      int         `json:"index"`
	Action     string      `json:"action,omitempty"`
	Method     string      `json:"method"`
	ID         string      `json:"id,omitempty"`
	InputCount int         `json:"input_count"`
	Inputs     []InputInfo `json:"inputs"`
}

type InputInfo struct {
	Tag           string `json:"tag"`
	Type          string `json:"type,omitempty"`
	Name          string `json:"name,omitempty"`
	Placeholder   string `json:"placeholder,omitempty"`
	ID            string `json:"id,omitempty"`
	ContentEditable bool  `json:"content_editable"`
	Role          string `json:"role,omitempty"`
}

type ButtonInfo struct {
	Text       string `json:"text,omitempty"`
	Type       string `json:"type,omitempty"`
	ID         string `json:"id,omitempty"`
	AriaLabel  string `json:"aria_label,omitempty"`
	Disabled   bool   `json:"disabled"`
}

// NetworkResult holds captured network analysis.
type NetworkResult struct {
	TotalRequests  int           `json:"total_requests"`
	APIEndpoints   []APIEndpoint `json:"api_endpoints"`
	StaticFiltered int           `json:"static_filtered"`
}

type APIEndpoint struct {
	Method       string `json:"method"`
	URL          string `json:"url"`
	HasAuthHeader bool  `json:"has_auth_header"`
	AuthHeaders  []string `json:"auth_headers,omitempty"`
	ContentType  string `json:"content_type,omitempty"`
	Status       int    `json:"status"`
}

// PatternResult holds the detected CLI pattern recommendation.
type PatternResult struct {
	Type       string  `json:"type"`        // dom-scrape, api-reverse, form-submit, async-poll
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

// Run executes the full probe pipeline.
func Run(client *browser.Client, targetURL string) (*SiteProfile, error) {
	profile := &SiteProfile{
		URL:       targetURL,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// ── Phase 1: Start network capture BEFORE navigate ──
	if err := client.NetworkStart(); err != nil {
		profile.Warnings = append(profile.Warnings, fmt.Sprintf("network capture start failed: %v", err))
	}

	// ── Phase 2: Navigate (triggers page-load requests) ──
	if err := client.Navigate(targetURL); err != nil {
		return nil, fmt.Errorf("navigate failed: %w", err)
	}
	// Wait for page + async requests to settle
	time.Sleep(3 * time.Second)

	// ── Phase 3: Auth detection ──
	auth, warnings := DetectAuth(client)
	profile.Auth = auth
	profile.Warnings = append(profile.Warnings, warnings...)

	// ── Phase 4: DOM extraction ──
	dom, domWarnings := ExtractDOM(client)
	profile.DOM = dom
	profile.Warnings = append(profile.Warnings, domWarnings...)

	// ── Phase 5: Stop network capture + analyze ──
	if err := client.NetworkStop(); err != nil {
		profile.Warnings = append(profile.Warnings, fmt.Sprintf("network stop failed: %v", err))
	}
	network, netWarnings := AnalyzeNetwork(client)
	profile.Network = network
	profile.Warnings = append(profile.Warnings, netWarnings...)

	// ── Phase 6: Pattern detection ──
	profile.Pattern = DetectPattern(profile.Auth, profile.DOM, profile.Network)

	return profile, nil
}

// RunQuick runs a fast probe without network capture (DOM + auth only).
func RunQuick(client *browser.Client, targetURL string) (*SiteProfile, error) {
	profile := &SiteProfile{
		URL:       targetURL,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Navigate
	if err := client.Navigate(targetURL); err != nil {
		return nil, fmt.Errorf("navigate failed: %w", err)
	}
	time.Sleep(2 * time.Second)

	// Auth
	auth, warnings := DetectAuth(client)
	profile.Auth = auth
	profile.Warnings = append(profile.Warnings, warnings...)

	// DOM
	dom, domWarnings := ExtractDOM(client)
	profile.DOM = dom
	profile.Warnings = append(profile.Warnings, domWarnings...)

	// Pattern (limited without network data)
	profile.Pattern = DetectPattern(profile.Auth, profile.DOM, &NetworkResult{})

	return profile, nil
}
