package unsplash

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// perPage is the batch size Unsplash renders server-side and adds per
// pagination step. It is not configurable through the URL, so --limit above
// this drives extra rounds of advanceJS.
const perPage = 20

// gridSelector is the figure wrapper Unsplash tags each asset with. Using the
// data-testid (rather than a hashed CSS-module class) keeps the extractor
// stable across their frontend deploys, and excludes the illustration/stock
// ad rails, which render as plain <figure> without it.
//
// The prefix match is load-bearing: Unsplash serves the same search results
// under at least two grid layouts — asset-grid-list-figure and
// asset-grid-masonry-figure — and which one you get varies between page loads.
// Their inner markup is identical, so one extractor handles both; matching only
// one testid silently returns zero results on the other.
const gridSelector = `figure[data-testid^="asset-grid-"]`

// Photo is one search hit.
type Photo struct {
	Rank        int    `json:"rank"`
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	PageURL     string `json:"page_url"`
	ImageURL    string `json:"image_url"` // CDN base, no sizing params
	Author      string `json:"author"`
	AuthorURL   string `json:"author_url"`
	Plus        bool   `json:"plus"` // Unsplash+ subscription asset
}

// SearchOptions mirrors the filters unsplash.com accepts as query params. Every
// one was verified against the live site during archaeology.
type SearchOptions struct {
	Query       string
	Limit       int
	Orientation string // landscape | portrait | squarish
	Color       string // black_and_white, black, white, yellow, orange, red, purple, magenta, green, teal, blue
	OrderBy     string // relevant | latest
	IncludePlus bool   // false pins license=free, which drops Unsplash+ results
}

// extractorJS is the verified evaluate call from archaeology. It reads the
// SSR'd DOM — Unsplash renders the first page of results server-side, so no
// XHR replay is needed (and none is possible: the bot check blocks /napi/*).
const extractorJS = `(() => {
  const out = [];
  const seen = new Set();
  document.querySelectorAll('figure[data-testid^="asset-grid-"]').forEach(f => {
    const a = f.querySelector('a[itemprop="contentUrl"]');
    if (!a) return;
    const href = a.getAttribute("href") || "";
    // Photo hrefs are /photos/<slug>-<id> or /photos/<id>; the id is always the
    // trailing 11-char nanoid.
    const m = href.match(/^\/photos\/(?:(.*)-)?([A-Za-z0-9_-]{11})$/);
    if (!m) return;
    const id = m[2];
    if (seen.has(id)) return;
    seen.add(id);
    const img = a.querySelector("img");
    const src = img ? (img.getAttribute("src") || "") : "";
    // The first /@handle link in a figure is the avatar (no text); the second
    // carries the display name.
    let author = "", authorUrl = "";
    f.querySelectorAll('a[href^="/@"]').forEach(x => {
      const t = (x.textContent || "").trim();
      if (!authorUrl) authorUrl = x.getAttribute("href");
      if (!author && t) { author = t; authorUrl = x.getAttribute("href"); }
    });
    out.push({
      id: id,
      slug: m[1] || "",
      description: a.getAttribute("title") || (img && img.getAttribute("alt")) || "",
      page_url: "https://unsplash.com" + href,
      image_url: src.split("?")[0],
      author: author,
      author_url: authorUrl ? "https://unsplash.com" + authorUrl : "",
      plus: src.indexOf("plus.unsplash.com") >= 0
    });
  });
  return out;
})()`

// BuildSearchURL renders the SSR search URL for a set of filters.
//
// Note the absence of a page param: unsplash.com accepts ?page=N but ignores it
// on the server-rendered response — pages 1, 2 and 3 come back byte-identical.
// Paging is a "Load more" button that calls the /napi endpoint the bot check blocks.
func BuildSearchURL(o SearchOptions) string {
	q := url.Values{}
	if o.Orientation != "" {
		q.Set("orientation", o.Orientation)
	}
	if o.Color != "" {
		q.Set("color", o.Color)
	}
	if o.OrderBy != "" && o.OrderBy != "relevant" {
		q.Set("order_by", o.OrderBy)
	}
	if !o.IncludePlus {
		// Verified: license=free returns zero plus.unsplash.com assets.
		q.Set("license", "free")
	}

	// Unsplash slugifies the query into the path; spaces become hyphens.
	slug := url.PathEscape(strings.Join(strings.Fields(o.Query), "-"))
	u := "https://unsplash.com/s/photos/" + slug
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	return u
}

// advanceJS pulls in the next batch of results and waits for the grid to grow.
//
// Unsplash paginates in two stages: a "Load more" button for the first
// extension, then infinite scroll for everything after it. Both are handled
// here — the button when it is present and laid out, a scroll to the bottom
// otherwise — so the caller just calls this until it stops growing.
//
// Counting unique photo ids rather than <figure> elements is deliberate: the
// page keeps a second, hidden copy of the grid in an alternate layout, so the
// element count roughly triples the real number of photos.
const advanceJS = `(async () => {
  const sel = 'figure[data-testid^="asset-grid-"] a[itemprop="contentUrl"]';
  const count = () => new Set(
    Array.from(document.querySelectorAll(sel))
      .map(a => ((a.getAttribute("href") || "").match(/([A-Za-z0-9_-]{11})$/) || [])[1])
      .filter(Boolean)
  ).size;

  const before = count();
  const findBtn = () => Array.from(document.querySelectorAll("button"))
    .find(b => (b.textContent || "").trim() === "Load more");

  // Retry rather than click once and wait. The button ships in the
  // server-rendered HTML but does nothing until React hydrates and attaches its
  // handler, and we can easily arrive before that — a single early click looks
  // exactly like "Unsplash refused to paginate". Clicking again is harmless:
  // the button removes itself once it has fired.
  //
  // Order matters too. Infinite scroll only arms itself after the button has
  // been used once, so scrolling is the fallback, never the opener.
  let method = "none";
  for (let attempt = 0; attempt < 6; attempt++) {
    const btn = findBtn();
    if (btn) {
      btn.scrollIntoView({ block: "center" });
      btn.click();
      method = "load-more";
    } else {
      window.scrollTo(0, document.body.scrollHeight);
      method = "scroll";
    }
    for (let i = 0; i < 6; i++) {
      await new Promise(r => setTimeout(r, 500));
      if (count() > before) {
        return { before: before, after: count(), method: method, grew: true };
      }
    }
  }
  return { before: before, after: count(), method: method, grew: false };
})()`

type advanceState struct {
	Before int    `json:"before"`
	After  int    `json:"after"`
	Method string `json:"method"`
	Grew   bool   `json:"grew"`
}

// SearchResult carries the hits plus enough context for a caller to know
// whether it saw everything it asked for.
type SearchResult struct {
	Photos    []Photo
	SearchURL string
	Truncated bool   // fewer results than Limit, because Unsplash ran out
	Note      string // why, when Truncated
}

// Search loads the search page and pages through results until it has Limit
// photos or Unsplash stops producing new ones.
func Search(client Browser, o SearchOptions) (*SearchResult, error) {
	if strings.TrimSpace(o.Query) == "" {
		return nil, fmt.Errorf("query is empty")
	}
	if o.Limit <= 0 {
		o.Limit = perPage
	}

	searchURL := BuildSearchURL(o)
	state, err := Load(client, searchURL, gridSelector, defaultWait)
	if err != nil {
		return nil, err
	}

	out := &SearchResult{SearchURL: searchURL}
	if state.Count == 0 {
		out.Note = "no results for these filters"
		return out, nil
	}

	photos, err := extract(client)
	if err != nil {
		return nil, err
	}

	// Each advance yields about perPage more photos. The slack covers a batch
	// that lands short, and caps a pathological run rather than the normal one.
	maxAdvances := o.Limit/perPage + 3
	for advances := 0; len(photos) < o.Limit && advances < maxAdvances; advances++ {
		raw, err := client.Evaluate(advanceJS)
		if err != nil {
			break
		}
		var adv advanceState
		if err := json.Unmarshal(raw, &adv); err != nil {
			break
		}
		if !adv.Grew {
			out.Note = fmt.Sprintf("unsplash.com stopped returning new photos after %d (last tried: %s)", adv.After, adv.Method)
			break
		}
		photos, err = extract(client)
		if err != nil {
			return nil, err
		}
	}

	if len(photos) > o.Limit {
		photos = photos[:o.Limit]
	}
	for i := range photos {
		photos[i].Rank = i + 1
	}
	out.Photos = photos
	out.Truncated = len(photos) < o.Limit
	if !out.Truncated {
		out.Note = ""
	}
	return out, nil
}

// extract runs the DOM extractor and de-duplicates by photo id — Unsplash
// occasionally renders the same asset in two grid slots.
func extract(client Browser) ([]Photo, error) {
	raw, err := client.Evaluate(extractorJS)
	if err != nil {
		return nil, fmt.Errorf("extract results: %w", err)
	}
	var batch []Photo
	if err := json.Unmarshal(raw, &batch); err != nil {
		return nil, fmt.Errorf("parse extractor output: %w", err)
	}
	seen := map[string]bool{}
	out := batch[:0]
	for _, p := range batch {
		if seen[p.ID] {
			continue
		}
		seen[p.ID] = true
		out = append(out, p)
	}
	return out, nil
}
