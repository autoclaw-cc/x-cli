# Site Archaeology — Bing Search

Findings from applying the archaeology protocol against Bing (English + Chinese UI) on 2026-07-15. Merged from bing-cli-acc and updated for v2.

## Feature A: `search <query>`

**URL:** `GET https://www.bing.com/search?q=<urlencode(query)>&count=<N>&offset=<O>`

- No auth required for basic search.
- Results are SSR'd in the DOM — no XHR needed for page one.
- Chinese market (`cc=cn&setlang=zh-Hans`) vs international (`cc=us&setlang=en`).
- `count` max is ~50 per page. `offset` enables pagination.

### DOM Structure

Each organic result lives in:

```html
<li class="b_algo" iid="SERP.NNNN">
  <style>...</style>
  <h2><a href="https://www.bing.com/ck/a?...">Title text</a></h2>
  <div class="b_attribution"><cite>https://example.com</cite></div>
  <div class="b_caption"><p class="b_paractl">Snippet text...</p></div>
</li>
```

**Selectors:**

| Field   | Selector       | Notes |
|---------|----------------|-------|
| title   | `#b_results > li.b_algo h2 a` textContent | |
| url     | `#b_results > li.b_algo .b_attribution cite` textContent | Display URL (domain only) |
| snippet | `#b_results > li.b_algo .b_caption p` textContent | May include breadcrumb prefix |
| source  | same as `cite` | |
| bingHref| `h2 a` getAttribute("href") | Encrypted /ck/a redirect — resolved on Go side |

### Extract Strategy

Async IIFE with 8 s polling deadline (google-cli pattern):

1. Check for consent interstitial (`location.host.startsWith('consent.')`).
2. Poll `#b_results > li.b_algo` every 500 ms until results appear.
3. Deduplicate by display URL (Set).
4. Clean snippet: `replace(/\s+/g, " ")`, strip trailing "Read more", slice to 300.

### Gotchas

- **Encrypted href**: The `<h2 a>` `href` is a Bing redirect (`/ck/a?...&u=<encrypted>`). The CLI follows each redirect via HTTP GET and extracts the real URL from the response HTML body using a regex.
- **Inline `<style>` in results**: Each `li.b_algo` contains inline `<style>`. Extract by selector, never `innerHTML`.
- **Snippet prefix**: Chinese UI snippets sometimes start with `› ` (breadcrumb). Whitespace normalization handles it.
- **Result count**: `#b_results > li.b_algo` typically returns 10–12 per page.
- **Consent interstitial**: EU / anonymous browsers may hit `consent.bing.com`. Detected and reported as `consent_required`.

### Known Failure Modes

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| `consent_required` | EU consent interstitial | Accept once in Chrome, retry |
| `no_results` / empty items | Bing reflowed selectors | Re-run archaeology against `li.b_algo` |
| `daemon_unreachable` | kimi-webbridge not running | Start Kimi Desktop App |
| `extension_not_connected` | WebBridge extension not installed | Install from kimi.com/features/webbridge |
| CDP context errors | Tab still loading | Handled by evaluateWithRetry (3 retries) |

---

## Feature B: `result <url>`

**Delivery model:** Arbitrary page; DOM extraction via `evaluate`.

### Extraction

```json
{
  "url":         "location.href after redirects",
  "title":       "document.title",
  "description": "meta[name=description] → meta[property='og:description'] → ''",
  "text":        "document.body.innerText capped at 5000 chars"
}
```

### Gotchas

- `document.body.innerText` begins with header/nav text — acceptable for MVP.
- JS-heavy pages: if `text.length < 50`, retry once with a 2 s settle delay.
- No network capture required.

---

## Daemon `evaluate` Protocol (Critical for Go Client)

- Code runs as a top-level expression — `return <expr>` triggers `SyntaxError`.
- The value of the last expression is the return value.
- For async, wrap explicitly: `(async () => { ... })()`.
- Daemon wraps every return as `{"type": "<type>", "value": <v>}`.
- When code ends with `JSON.stringify(...)`, `type` is `"string"`.
- Go client **must** unwrap this envelope before unmarshalling.
- `EvaluateJSON` handles the unwrap automatically.
