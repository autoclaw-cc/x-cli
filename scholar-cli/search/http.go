package search

import (
	"context"
	"errors"
	"fmt"
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

// ErrS2RateLimited reports that Semantic Scholar kept returning 429 after every
// retry. It is deliberately distinguishable from "no such paper": for an arXiv
// DOI, S2 is the only source that has the record (CrossRef legitimately 404s,
// since arXiv DOIs are registered with DataCite), so a rate limit there
// otherwise looks exactly like a missing paper.
var ErrS2RateLimited = errors.New("semantic scholar rate limited (HTTP 429)")

// s2Backoff is the wait before each retry. S2's no-key tier is a shared global
// pool that 429s during peak hours regardless of the caller's own pace, so a
// single fixed 1s retry rides out only the shortest transients.
var s2Backoff = []time.Duration{time.Second, 2 * time.Second, 4 * time.Second}

// s2GET issues a GET to a Semantic Scholar URL, backing off on 429. On success
// the caller owns the response body. Persistent 429s return ErrS2RateLimited
// rather than the 429 response, so callers cannot mistake a throttled request
// for an empty result.
func s2GET(ctx context.Context, url string) (*http.Response, error) {
	for attempt := 0; ; attempt++ {
		req, err := newRequest(ctx, "GET", url)
		if err != nil {
			return nil, err
		}
		resp, err := defaultHTTPClient.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}
		resp.Body.Close()
		if attempt >= len(s2Backoff) {
			return nil, fmt.Errorf("%w after %d attempts", ErrS2RateLimited, attempt+1)
		}
		select {
		case <-time.After(s2Backoff[attempt]):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
