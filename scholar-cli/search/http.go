package search

import (
	"context"
	"net/http"
	"os"
	"regexp"
)

const userAgent = "scholar-cli/1.0 (https://github.com/scholar-cli; mailto:scholar-cli@example.com)"

// ContactEmail returns the email to attach to polite-pool API requests
// (CrossRef, OpenAlex, NCBI E-utilities). Set SCHOLAR_CLI_EMAIL to your own
// address; falls back to the RFC-reserved example.com so we never bake in a
// real one.
func ContactEmail() string {
	if e := os.Getenv("SCHOLAR_CLI_EMAIL"); e != "" {
		return e
	}
	return "scholar-cli@example.com"
}

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

func stripHTMLTags(s string) string {
	return htmlTagRe.ReplaceAllString(s, "")
}

func newRequest(ctx context.Context, method, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	return req, nil
}
