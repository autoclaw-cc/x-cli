package unsplash

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Downloads go straight to Unsplash's imgix CDN, which — unlike unsplash.com —
// is not behind the Anubis bot check. So the bytes never travel through the browser:
// no base64 round trip through the daemon, no multi-hundred-MB tab. The browser
// is only ever needed to learn *which* URL to fetch.
//
// unsplash.com/photos/<id>/download is the endpoint the site's own button uses,
// but it answers 401 to anything that hasn't cleared Anubis, so it is not an
// option here.

// httpUA identifies the CLI. The CDN serves anonymous requests, but sending a
// real UA keeps us out of default bot heuristics.
const httpUA = "unsplash-cli (+https://github.com/xpzouying/x-cli)"

var downloader = &http.Client{Timeout: 5 * time.Minute}

// SizeSpec describes the imgix transform to request. A zero Width asks for the
// original — Unsplash's imgix source returns the full-resolution file when no
// sizing params are present.
type SizeSpec struct {
	Width   int
	Quality int
	Format  string // jpg | png | webp | "" for the source format
}

// BuildImageURL applies the size spec to a CDN base URL.
func BuildImageURL(base string, s SizeSpec) string {
	q := url.Values{}
	if s.Format != "" {
		q.Set("fm", s.Format)
	}
	if s.Quality > 0 {
		q.Set("q", fmt.Sprint(s.Quality))
	}
	if s.Width > 0 {
		q.Set("w", fmt.Sprint(s.Width))
		// fit=max scales down without cropping and never upscales, so a --width
		// larger than the original is a no-op instead of a blurry enlargement.
		q.Set("fit", "max")
	}
	if len(q) == 0 {
		return base
	}
	return base + "?" + q.Encode()
}

// Result records one completed download.
type Result struct {
	ID        string `json:"id"`
	SourceURL string `json:"source_url"`
	Path      string `json:"path"`
	Bytes     int64  `json:"bytes"`
	Skipped   bool   `json:"skipped"` // already on disk, --force not set
}

// safeName strips anything that would break a filename on any of the three
// major platforms, and caps length so long slugs stay openable.
func safeName(s string) string {
	s = strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', 0:
			return '-'
		}
		return r
	}, s)
	s = strings.Trim(strings.TrimSpace(s), ".")
	if len(s) > 120 {
		s = s[:120]
	}
	return s
}

// Filename renders the on-disk name for a photo.
func Filename(id, slug, format string) string {
	if format == "" {
		format = "jpg"
	}
	stem := id
	if slug != "" {
		stem = safeName(slug) + "-" + id
	}
	return stem + "." + format
}

// Download fetches imageURL into dir under name. It writes to a .part file and
// renames on success, so an interrupted run never leaves a truncated image that
// a later --skip-existing would mistake for a finished download.
func Download(imageURL, dir, name string, force bool) (*Result, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create %s: %w", dir, err)
	}
	dest := filepath.Join(dir, name)

	if !force {
		if st, err := os.Stat(dest); err == nil && st.Size() > 0 {
			return &Result{Path: dest, Bytes: st.Size(), Skipped: true, SourceURL: imageURL}, nil
		}
	}

	req, err := http.NewRequest(http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", httpUA)

	resp, err := downloader.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", imageURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: HTTP %d", imageURL, resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "image/") {
		return nil, fmt.Errorf("fetch %s: expected an image, got %s", imageURL, ct)
	}

	tmp := dest + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", tmp, err)
	}
	n, err := io.Copy(f, resp.Body)
	closeErr := f.Close()
	if err != nil {
		os.Remove(tmp)
		return nil, fmt.Errorf("write %s: %w", tmp, err)
	}
	if closeErr != nil {
		os.Remove(tmp)
		return nil, fmt.Errorf("close %s: %w", tmp, closeErr)
	}
	if err := os.Rename(tmp, dest); err != nil {
		return nil, fmt.Errorf("rename %s: %w", tmp, err)
	}

	return &Result{Path: dest, Bytes: n, SourceURL: imageURL}, nil
}
