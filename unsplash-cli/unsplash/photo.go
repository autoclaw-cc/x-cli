package unsplash

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// idPattern matches Unsplash's 11-char nanoid photo id.
var idPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

// ParseRef turns anything a user might paste — a bare id, a photo page URL, or
// a CDN image URL — into either a resolved CDN base URL or a photo id that
// still needs a page lookup.
//
// Recognising the CDN form matters for cost: piping `search` output straight
// into `download` then costs zero browser round trips.
func ParseRef(ref string) (imageURL, id string, err error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", "", fmt.Errorf("empty photo reference")
	}

	if strings.HasPrefix(ref, "https://images.unsplash.com/") ||
		strings.HasPrefix(ref, "https://plus.unsplash.com/") {
		return strings.SplitN(ref, "?", 2)[0], "", nil
	}

	if idPattern.MatchString(ref) {
		return "", ref, nil
	}

	if strings.Contains(ref, "unsplash.com/photos/") {
		path := strings.SplitN(ref, "unsplash.com/photos/", 2)[1]
		path = strings.SplitN(path, "?", 2)[0]
		path = strings.TrimSuffix(path, "/")
		if i := strings.LastIndex(path, "-"); i >= 0 {
			path = path[i+1:]
		}
		if idPattern.MatchString(path) {
			return "", path, nil
		}
	}

	return "", "", fmt.Errorf("cannot read a photo id out of %q (want an 11-char id, an unsplash.com/photos/... URL, or an images.unsplash.com URL)", ref)
}

// Meta is what a photo page tells us without an API.
type Meta struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	PageURL     string `json:"page_url"`
	ImageURL    string `json:"image_url"`
	Author      string `json:"author"`
	AuthorURL   string `json:"author_url"`
	Plus        bool   `json:"plus"`
}

// metaJS reads the photo page's OpenGraph tags. og:image is the only reliable
// handle on the CDN base URL — the rendered <img> is lazily swapped and its src
// carries display-sized params.
const metaJS = `(() => {
  const m = n => { const e = document.querySelector('meta[property="' + n + '"]'); return e ? e.getAttribute("content") : ""; };
  const og = m("og:image");
  const href = location.pathname;
  const idm = href.match(/([A-Za-z0-9_-]{11})$/);
  const slugm = href.match(/^\/photos\/(.*)-[A-Za-z0-9_-]{11}$/);
  const author = document.querySelector('a[href^="/@"][title], main a[href^="/@"]');
  return {
    id: idm ? idm[1] : "",
    slug: slugm ? slugm[1] : "",
    description: m("og:title").replace(/ - Free .*$/, "").trim(),
    page_url: location.href.split("?")[0],
    image_url: og ? og.split("?")[0] : "",
    author: author ? (author.textContent || "").trim() : "",
    author_url: author ? "https://unsplash.com" + author.getAttribute("href") : "",
    plus: (og || "").indexOf("plus.unsplash.com") >= 0
  };
})()`

// FetchMeta loads a photo page and reads its metadata. The bare-id URL works —
// Unsplash redirects /photos/<id> to the slugged canonical URL.
func FetchMeta(client Browser, id string) (*Meta, error) {
	pageURL := "https://unsplash.com/photos/" + id
	if _, err := Load(client, pageURL, `meta[property="og:image"]`, defaultWait); err != nil {
		return nil, err
	}
	raw, err := client.Evaluate(metaJS)
	if err != nil {
		return nil, fmt.Errorf("read photo metadata: %w", err)
	}
	var m Meta
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("parse photo metadata: %w", err)
	}
	if m.ImageURL == "" {
		return nil, fmt.Errorf("no image found on %s (photo removed, or id is wrong)", pageURL)
	}
	if m.ID == "" {
		m.ID = id
	}
	return &m, nil
}

// AssetName is the imgix asset segment of a CDN URL, e.g.
// "photo-1518837695005-2083093ee35b". It stands in for the photo id when the
// caller handed us a bare CDN URL, which carries no id of its own.
func AssetName(imageURL string) string {
	u := strings.SplitN(imageURL, "?", 2)[0]
	if i := strings.LastIndex(u, "/"); i >= 0 {
		u = u[i+1:]
	}
	if u == "" {
		return "unsplash-photo"
	}
	return u
}
