# bing-cli

CLI wrapper around Bing Search backed by the [kimi-webbridge](https://www.kimi.com/features/webbridge) browser daemon. Runs inside your real Chrome session — no API key needed — and emits JSON on stdout.

## Commands

| Command | Usage | Returns |
|---|---|---|
| `search` | `bing-cli search <query> [-n N] [-m cn\|us] [-o N] [-r]` | `{query, count, results: [{title, url, snippet, source}]}` |
| `result` | `bing-cli result <url> [-r]` | `{url, title, description, text}` |

All output:

```json
{"ok": true, "data": ...}
```
or on failure (non-zero exit):

```json
{"ok": false, "error": {"code": "...", "message": "..."}}
```

### Agent consumption (recommended)

Use `--raw` (`-r`) to get clean JSON without the status wrapper — saves tokens:

```bash
bing-cli search "query" -r -n 10
# → [{"title":"...","url":"...","snippet":"...","source":"..."}]
```

## Prerequisites

**kimi-webbridge daemon** running on `127.0.0.1:10086` — install from https://www.kimi.com/features/webbridge.

## Install

**Binary** (Windows): copy `bing-cli.exe` into a directory on your `PATH`.

**Build from source** (requires Go 1.26+):

```bash
go build -o bing-cli .
./bing-cli search "claude code"
```

## Quick test

```bash
# English search
bing-cli search "claude code" --count 3

# Chinese search
bing-cli search "牛仔裤面料" --market cn -r -n 5

# International with pagination
bing-cli search "rust async patterns" -m us -r -n 10 -o 10

# Fetch page content
bing-cli result "https://example.com/article"
```

## Flags (search)

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--market` | `-m` | `""` (auto) | `cn` (China), `us` (International) |
| `--count` | `-n` | `10` | Results per page (max 50) |
| `--offset` | `-o` | `0` | Pagination offset |
| `--raw` | `-r` | `false` | Bare JSON array — no wrapper |

## Layout

```
bing-cli/
├── main.go                # entrypoint
├── browser/client.go      # kimi-webbridge HTTP client (+ EvaluateJSON, Status)
├── output/output.go       # JSON contract: {ok, data} / {ok, error}
├── bing/
│   ├── search.go          # search command backend (DOM extractor + /ck/a resolver)
│   ├── result.go          # result command backend (page content extraction)
│   └── common.go          # evaluateWithRetry + isTransientContextError
├── cmd/
│   └── root.go            # Cobra CLI definition (search + result)
├── ARCHAEOLOGY.md         # DOM-selector field notes
└── README.md              # this file
```

## Robustness

- **CDP retry**: 3-level backoff (300ms → 700ms → 1500ms) for transient context errors.
- **Consent detection**: Detects Bing privacy interstitials and reports a clear error.
- **Async DOM polling**: Waits up to 8 s for results to render (defense against lazy loading).
- **URL dedup**: Filters duplicate display URLs from the result set.
- **/ck/a redirect resolution**: Follows Bing's encrypted redirect pages to extract real target URLs.
- **JS-heavy page retry**: For `result`, retries with a 2 s settle delay if body text is < 50 chars.

## Notes

- Bing encrypts the real link target in the `href` of each result's `<h2 a>` (a proprietary redirect through `/ck/a`). The CLI automatically follows each redirect and extracts the real URL from the response HTML.
- Selectors pinned to `#b_results > li.b_algo` + `h2 a` + `.b_attribution cite` + `.b_caption p`. If Bing reflows, re-run the archaeology protocol and update `bing/search.go`.
- No login needed for basic search.
- `cn.bing.com` auto-redirect may happen based on IP when no `--market` flag.
