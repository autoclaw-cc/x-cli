package download

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"scholar-cli/paper"
)

var httpClient = &http.Client{
	Timeout: 60 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("too many redirects")
		}
		return nil
	},
}

const userAgent = "scholar-cli/1.0"

// contactEmail returns the email to send to polite-pool APIs (Unpaywall,
// CrossRef, NCBI). Set SCHOLAR_CLI_EMAIL to your own address. Defaults to
// the RFC-reserved example.com so we never accidentally use a real one.
func contactEmail() string {
	if e := os.Getenv("SCHOLAR_CLI_EMAIL"); e != "" {
		return e
	}
	return "scholar-cli@example.com"
}

type DownloadResult struct {
	DOI      string `json:"doi"`
	Title    string `json:"title"`
	FilePath string `json:"file_path"`
	Source   string `json:"source"`
	Size     int64  `json:"size_bytes"`
}

// Download tries multiple channels in order:
// 1. Direct pdf_url from search results (open access)
// 2. Unpaywall API (legal OA copies)
// 3. Sci-Hub (only when scihubDomain is explicitly set — never by default,
//    since Sci-Hub is under court injunctions in several jurisdictions and
//    users must opt in consciously)
func Download(ctx context.Context, p *paper.Paper, outputDir string, scihubDomain string) (*DownloadResult, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	filename := sanitizeFilename(p.Title)
	if filename == "" {
		filename = "paper"
	}
	if len(filename) > 80 {
		filename = filename[:80]
	}
	filename += ".pdf"
	destPath := filepath.Join(outputDir, filename)

	// Build a list of candidate URLs to try
	var candidates []struct{ url, source string }

	// arXiv direct PDF (most reliable for arXiv papers)
	if arxivID, ok := p.IDs["arxiv_id"]; ok && arxivID != "" {
		candidates = append(candidates, struct{ url, source string }{
			fmt.Sprintf("https://arxiv.org/pdf/%s", arxivID), "arxiv",
		})
	}

	// Direct pdf_url from search results
	if p.PDFURL != "" {
		candidates = append(candidates, struct{ url, source string }{p.PDFURL, "open_access"})
	}

	// Try each candidate
	for _, c := range candidates {
		size, err := downloadFile(ctx, c.url, destPath)
		if err == nil && size > 1024 {
			return &DownloadResult{
				DOI:      p.DOI,
				Title:    p.Title,
				FilePath: destPath,
				Source:   c.source,
				Size:     size,
			}, nil
		}
	}

	if p.DOI == "" {
		return nil, fmt.Errorf("no PDF URL and no DOI available for download")
	}

	// Channel 2: Unpaywall
	unpURL := fmt.Sprintf("https://api.unpaywall.org/v2/%s?email=%s",
		url.PathEscape(p.DOI), url.QueryEscape(contactEmail()))
	if pdfURL, err := getUnpaywallURL(ctx, unpURL); err == nil && pdfURL != "" {
		size, err := downloadFile(ctx, pdfURL, destPath)
		if err == nil && size > 1024 {
			return &DownloadResult{
				DOI:      p.DOI,
				Title:    p.Title,
				FilePath: destPath,
				Source:   "unpaywall",
				Size:     size,
			}, nil
		}
	}

	// Channel 3: Sci-Hub — opt-in only. Skip unless the caller passed a
	// domain via --scihub; never default to sci-hub.se on the user's behalf.
	if scihubDomain != "" {
		scihubURL := fmt.Sprintf("https://%s/%s", scihubDomain, p.DOI)
		pdfURL, err := resolveSciHubPDF(ctx, scihubURL)
		if err == nil && pdfURL != "" {
			size, err := downloadFile(ctx, pdfURL, destPath)
			if err == nil && size > 1024 {
				return &DownloadResult{
					DOI:      p.DOI,
					Title:    p.Title,
					FilePath: destPath,
					Source:   "scihub",
					Size:     size,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("all download channels failed for DOI %s", p.DOI)
}

func getUnpaywallURL(ctx context.Context, apiURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unpaywall status %d", resp.StatusCode)
	}

	var data struct {
		BestOALocation *struct {
			URLForPDF string `json:"url_for_pdf"`
			URL       string `json:"url"`
		} `json:"best_oa_location"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if data.BestOALocation != nil {
		if data.BestOALocation.URLForPDF != "" {
			return data.BestOALocation.URLForPDF, nil
		}
		return data.BestOALocation.URL, nil
	}
	return "", fmt.Errorf("no OA location found")
}

func resolveSciHubPDF(ctx context.Context, pageURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return "", err
	}

	html := string(body)

	// Look for embedded PDF iframe or direct link
	patterns := []string{
		`src="//`,
		`src="https://`,
		`src="http://`,
	}
	for _, pat := range patterns {
		idx := strings.Index(html, pat)
		if idx < 0 {
			continue
		}
		start := idx + len(`src="`)
		end := strings.Index(html[start:], `"`)
		if end < 0 {
			continue
		}
		candidate := html[start : start+end]
		if strings.HasPrefix(candidate, "//") {
			candidate = "https:" + candidate
		}
		if strings.Contains(candidate, ".pdf") || strings.Contains(candidate, "/pdf/") || strings.Contains(candidate, "download") {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("could not find PDF link on Sci-Hub page")
}

func downloadFile(ctx context.Context, rawURL, destPath string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "text/html") {
		return 0, fmt.Errorf("server returned HTML instead of PDF (likely paywall)")
	}

	f, err := os.Create(destPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	n, err := io.Copy(f, resp.Body)
	if err != nil {
		os.Remove(destPath)
		return 0, err
	}

	return n, nil
}

func sanitizeFilename(s string) string {
	s = strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', '\n', '\r', '\t':
			return '_'
		}
		return r
	}, s)
	return strings.TrimSpace(s)
}
