package search

import (
	"context"
	"net/http"
	"os"
	"regexp"
	"time"
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

// s2GET issues a GET to a Semantic Scholar URL, retrying once after a 1s
// backoff on 429. S2's no-key tier is a shared global pool that 429s during
// peak hours regardless of the caller's pace; one retry rides out most of the
// short transients. If it still 429s after the retry, callers treat S2 as a
// best-effort enrichment source (search-en's other 6 sources keep working).
func s2GET(ctx context.Context, url string) (*http.Response, error) {
	req, err := newRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}
	resp, err := defaultHTTPClient.Do(req)
	if err != nil || resp.StatusCode != 429 {
		return resp, err
	}
	resp.Body.Close()
	time.Sleep(1 * time.Second)
	req2, err := newRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}
	return defaultHTTPClient.Do(req2)
}
