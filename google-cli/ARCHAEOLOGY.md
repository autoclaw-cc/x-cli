# Site Archaeology — Google Search

Findings from applying the `references/site-exploration.md` 6-step protocol against Google Search (en-US) on 2026-04-22.

## Feature A: `search <query>`

**URL:** `GET https://www.google.com/search?q=<urlencode(query)>&hl=en&num=<N>`

- `hl=en` stabilizes the result-card markup across locales.
- `num` is a soft hint; Google returns fewer than requested when ads / People-Also-Ask / feature cards take priority. Over-fetch (e.g. `num=limit*2`) and trim client-side.

**Auth:** none (the browser's normal cookies are enough — no consent interstitial was observed in this session).

**Delivery model:** pure SSR HTML. Initial results are in the DOM on first paint; no XHR is needed for the first page. Network capture (protocol Steps 3–5) was **not** required.

**Response shape (extracted via `evaluate`):**

```
[
  { "title": string, "url": string, "snippet": string },
  ...
]
```

**Working evaluate call:**

```js
const seen = new Set();
const items = Array.from(document.querySelectorAll('div#search div[data-hveid]'))
  .filter(el => el.querySelector('h3') && el.querySelector('a[href]'))
  .map(el => {
    const a = el.querySelector('a[href]');
    const h = el.querySelector('h3');
    return {
      title: h.innerText,
      url: a.href,
      snippet: (el.querySelector('[data-sncf]')?.innerText || '').replace(/\s*Read more\s*$/, '')
    };
  })
  .filter(r => { if (seen.has(r.url)) return false; seen.add(r.url); return true; });
JSON.stringify({ count: items.length, items });
```

**Notes / gotchas:**

- `div[data-hveid]` also matches nested cards inside a single result (site-links block), which is why the outer cards produce duplicate URLs. Dedup by `url` — see `const seen` above.
- `[data-sncf]` and `.VwiC3b` return the same snippet text (nested), but `[data-sncf]` is a semantic `data-` attribute and less likely to rename than a hashed class.
- Snippets often end with literal `"Read more"` from Google's UI — stripped with a regex.
- `data-hveid` cards without an `h3` are "People Also Ask" / feature boxes — filtered out.
- When logged into Google with personalization on, the result ordering differs from incognito. This is expected; don't treat it as a bug.

**Known failure mode (not observed in this session, but possible):**

- EU / anonymous browsers may hit a consent interstitial at `consent.google.com`. Detect by checking `location.host.startsWith('consent.')` before extraction; error with code `consent_required` and ask the user to accept once in Chrome.

## Feature B: `result <url>`

**Delivery model:** any page. DOM-based extraction via `evaluate`.

**Response shape:**

```
{
  "url":         string,   // location.href (post-redirect)
  "title":       string,
  "description": string,   // meta[name=description] or og:description, "" if none
  "text":        string    // document.body.innerText, capped at 5000 chars
}
```

**Working evaluate call:**

```js
JSON.stringify({
  url: location.href,
  title: document.title,
  description: document.querySelector('meta[name="description"]')?.content
            || document.querySelector('meta[property="og:description"]')?.content
            || '',
  text: document.body.innerText.slice(0, 5000)
});
```

**Notes / gotchas:**

- `document.body.innerText` starts with the page's header / nav text (e.g. "Skip to main content"). Acceptable for an MVP — a reader-mode pass is out of scope.
- For JS-heavy pages, the first `evaluate` after `navigate` may return before content mounts. If `text.length < 50`, retry once after a 1.5s wait.
- No network capture needed.

## Daemon `evaluate` protocol (important for the Go client)

From probing `http://127.0.0.1:10086` directly:

- `code` is executed as a **top-level expression**, not wrapped in an async function.
  - `return <expr>` at top level → `"SyntaxError: Illegal return statement"`.
  - The value of the last expression is the returned value.
  - For async work, wrap explicitly: `(async () => { ... return ...; })()`.
- The daemon always wraps the return as `{ "type": "<type>", "value": <v> }`.
  - When the code ends with `JSON.stringify(...)`, `type` is `"string"` and `value` is the stringified JSON.
  - The Go client must unwrap this envelope before unmarshalling into domain structs.
