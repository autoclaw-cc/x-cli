# google-cli

CLI wrapper around Google Search backed by the [kimi-webbridge](https://www.kimi.com/features/webbridge) browser daemon. Runs inside the user's real Chrome session — no API key, no login loop — and emits JSON on stdout.

## Commands

| Command | Usage | Returns |
|---|---|---|
| `search` | `google-cli search <query> [--limit 10] [--hl en]` | `[{title, url, snippet}]` |
| `result` | `google-cli result <url>` | `{url, title, description, text}` |

All output:

```json
{"ok": true, "data": ...}
```
or on failure (non-zero exit):

```json
{"ok": false, "error": {"code": "...", "message": "..."}}
```

## Prerequisites

1. **kimi-webbridge daemon** running on `127.0.0.1:10086` — install from https://www.kimi.com/features/webbridge.
2. **Go 1.25+** for building.

## Build

```bash
go build -o google-cli .
./google-cli --help
```

## Quick test

```bash
./google-cli search "claude code" --limit 3
./google-cli result "https://code.claude.com/docs/en/overview"
```

## Consent pages

EU / anonymous browsers may hit `consent.google.com` on first Google navigation. When detected, `search` errors with code `consent_required` — accept it once in Chrome, then retry.

## How selectors stay alive

Google frequently renames DOM classes. The working selectors, rationale, and failure modes are pinned in [`ARCHAEOLOGY.md`](./ARCHAEOLOGY.md). When parsing breaks, re-run the site-archaeology protocol from that file and update both `google/search.go` and `ARCHAEOLOGY.md` together.

## Layout

```
google-cli/
├── main.go                # entrypoint
├── browser/client.go      # kimi-webbridge HTTP client (+ EvaluateJSON helper)
├── output/output.go       # JSON contract: {ok, data} / {ok, error}
├── google/
│   ├── common.go          # evaluateWithRetry (handles CDP context races)
│   ├── search.go          # search command backend
│   └── result.go          # result command backend
├── cmd/
│   ├── root.go
│   ├── search.go
│   └── result.go
├── ARCHAEOLOGY.md         # DOM-selector field notes
└── README.md
```

## License

MIT (see `LICENSE` — to be added).
