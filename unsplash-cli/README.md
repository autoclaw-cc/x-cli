# unsplash-cli

CLI for searching and downloading [Unsplash](https://unsplash.com) photos, backed by the [kimi-webbridge](https://www.kimi.com/features/webbridge) browser daemon. Emits JSON on stdout.

No Unsplash account, API key, or login is required.

## Commands

| Command | Usage | Returns |
|---|---|---|
| `search` | `unsplash-cli search <query> [-n 20] [--orientation ...] [--color ...] [--order ...] [--include-plus]` | `{query, count, requested, truncated, search_url, results: [{rank, id, slug, description, page_url, image_url, author, author_url, plus}], note?}` |
| `download` | `unsplash-cli download <id\|url>... [-o DIR] [-w PX] [-q 1-100] [--format jpg\|png\|webp] [-f]` | `{dir, downloaded, failed, results: [{ref, id, path, bytes, skipped, source}], errors?}` |
| `close` | `unsplash-cli close` | `{closed, session}` — closes the Chrome tab group this CLI opened |

Every command also takes `--activate tab\|window\|off` (default `tab`) — see [Pagination, and the background-tab trap](#pagination-and-the-background-tab-trap). `WEBBRIDGE_ACTIVATE` sets the default for all CLIs in this repo; the flag overrides it.

All output:

```json
{"ok": true, "data": ...}
```

or, on failure (non-zero exit):

```json
{"ok": false, "error": {"code": "...", "message": "..."}}
```

A `download` batch where some refs succeed and some fail returns `ok: true` with an `errors` array **and** exits non-zero — partial results are reported rather than thrown away.

## Prerequisites

1. **kimi-webbridge daemon** running on `127.0.0.1:10086` — install from https://www.kimi.com/features/webbridge.
2. **Go 1.25+** to build.

## Build

```bash
go build -o unsplash-cli .
./unsplash-cli --help
```

## Quick test

```bash
./unsplash-cli search "misty forest" -n 5
./unsplash-cli search sunset -n 100 --color orange --order latest
./unsplash-cli download 6ArTTluciuA --out ~/Pictures --width 1920
./unsplash-cli search ocean -n 5 \
  | jq -r '.data.results[].image_url' \
  | xargs ./unsplash-cli download --out ./shots -w 1200
./unsplash-cli close
```

## How it works

unsplash.com is behind [Anubis](https://github.com/TecharoHQ/anubis), a proof-of-work bot check. That shapes the whole design:

- **Metadata must come from a real browser.** The site's internal JSON API (`unsplash.com/napi/*`) answers `401` with a challenge page to plain HTTP clients — and to `fetch()` from inside the page as well. Only Unsplash's own frontend knows how to retry through the challenge, so `search` navigates a Chrome tab to the **server-rendered** search page, reads `figure[data-testid^="asset-grid-"]` out of the DOM, and lets the site's own JS do the paging. A document navigation is the one request shape the bot check reliably lets through; `unsplash/page.go` polls the tab until it clears.
- **Image bytes do not.** `images.unsplash.com` / `plus.unsplash.com` (Unsplash's imgix CDN) are *not* behind the bot check. `download` fetches them with plain Go `net/http`, so the file never round-trips through the daemon as base64. Given a CDN URL — e.g. piped from `search` — a download touches Chrome zero times.
- `--width` / `--quality` / `--format` map onto imgix params (`w`, `q`, `fm`) with `fit=max`, so images are scaled down but never upscaled or cropped. Omit `--width` for the original file.

### Pagination, and the background-tab trap

`--limit` is not capped: results load 20 at a time — a *Load more* button for the first batch, infinite scroll after that — and `search` drives both. 200 results takes roughly 15 seconds. `?page=N` exists in Unsplash's URLs but the server ignores it, so paging only ever happens inside the live page.

Two things had to be right for that to work, and both are easy to get wrong:

1. **The tab must render.** A background Chrome tab reports `visibilityState: "hidden"`, skips layout entirely, and never fires the `IntersectionObserver` callbacks behind lazy loading and infinite scroll. On Unsplash the grid then measures 0px tall and the *Load more* button has no box to click — which is indistinguishable, from the outside, from Unsplash refusing to paginate. `browser.Client.Activate` fixes it, and `--activate` picks how:

   | mode | mechanism | frontmost app afterwards | pagination |
   |---|---|---|---|
   | `tab` *(default)* | `Page.setWebLifecycleState` + `Emulation.setFocusEmulationEnabled` | **unchanged** — you keep your screen | works |
   | `window` | `Page.bringToFront` | **becomes Google Chrome** — takes over your screen | works |
   | `off` | nothing | unchanged | first 20 only |

   Those columns are measured, not assumed. Under `tab` the page reports `visibilityState: "visible"` and lays out fully (`scrollHeight` 8593 vs 907) while the frontmost application never changes; `window` really does raise Chrome over whatever you were doing, which is why it is opt-in.

   `tab` leaves the tab un-*selected* in its Chrome window — Chrome offers no way to select it without raising the window, and `Target.activateTarget` is refused over `chrome.debugger`. That distinction doesn't matter to any extractor here: what they need is a page that renders, which is what they get.

   Set `WEBBRIDGE_ACTIVATE` to change the default without passing the flag every time.

2. **The click must be retried.** The button ships in the server-rendered HTML but is inert until React hydrates. A single click fired a second too early does nothing, and looks like the end of the results. `advanceJS` retries for a few seconds, then falls back to scrolling — in that order, because infinite scroll only arms itself once the button has been used.

If Unsplash genuinely runs out of photos, the response sets `truncated: true` with a `note` naming the count it stopped at.

### Unsplash+

Unsplash+ is Unsplash's paid subscription tier. Those assets are **excluded by default** (the CLI pins `license=free`); pass `--include-plus` to include them. Results are flagged with `"plus": true`. Respect the [Unsplash License](https://unsplash.com/license) for anything you download.

### Browser tabs

`search` and `download`-by-id reuse a single tab in a Chrome tab group labelled `unsplash-cli`, rather than opening a new one per run. Leaving it open also keeps the Anubis cookie warm for the next command. Run `unsplash-cli close` when you want it gone.

## Layout

```
unsplash-cli/
├── main.go                 # entrypoint
├── browser/client.go       # kimi-webbridge HTTP client (status, navigate, activate, evaluate, cdp)
├── output/output.go        # JSON contract: {ok, data} / {ok, error}
├── unsplash/
│   ├── page.go             # navigate, activate the tab, wait out the bot check
│   ├── search.go           # SSR DOM extractor + Load more / infinite-scroll paging
│   ├── photo.go            # ref parsing (id / page URL / CDN URL) + og:image lookup
│   └── download.go         # imgix URL building + direct CDN fetch
└── cmd/
    ├── root.go             # cobra root, daemon preflight, error envelope
    ├── search.go
    ├── download.go
    └── close.go
```
