# apartments-cli

CLI wrapper around Apartments.com (US) rental listings backed by the [kimi-webbridge](https://www.kimi.com/features/webbridge) browser daemon. Runs inside the user's real Chrome session — no API key, no login required — and emits JSON on stdout.

## Commands

| Command | Usage | Returns |
|---|---|---|
| `search` | `apartments-cli search --location new-york-ny [--min-beds 1] [--max-beds 2] [--min-price 1000] [--max-price 3000] [--limit 20] [--page 1]` | `{properties: [{id, name, address, pricing, ...}]}` |
| `detail` | `apartments-cli detail --url "https://www.apartments.com/xxx/yyy/"` | `{name, address, rent_range, units, amenities, images, ...}` |

All output:

```json
{"ok": true, "data": ...}
```
or on failure (non-zero exit):

```json
{"ok": false, "error": {"code": "...", "message": "..."}}
```

## Prerequisites

**kimi-webbridge daemon** running on `127.0.0.1:10086` — install from https://www.kimi.com/features/webbridge.

## Install

**Pre-built binary (recommended)**: download `apartments-cli-{platform}.tar.gz` from [Releases](https://github.com/better-world-ai/x-cli/releases?q=apartments-cli) and extract. macOS users: clear the Gatekeeper quarantine once with `xattr -d com.apple.quarantine ./apartments-cli`.

Available platforms: `darwin-{arm64,amd64}`, `linux-{amd64,arm64}`, `windows-{amd64,arm64}`.

**Build from source** (requires Go 1.25+):

```bash
go build -o apartments-cli .
./apartments-cli --help
```

## Quick test

```bash
./apartments-cli search --location new-york-ny --min-beds 1 --max-price 3000 --limit 3
./apartments-cli detail --url "https://www.apartments.com/10-halletts-point-astoria-ny/1j2c5h6/"
```

## Location slugs

Format: `city-name-state` — e.g. new-york-ny, los-angeles-ca, san-francisco-ca, chicago-il, seattle-wa, boston-ma, washington-dc, austin-tx, denver-co, miami-fl, san-diego-ca, portland-or

## Layout

```
apartments-cli/
├── main.go
├── browser/client.go      # kimi-webbridge HTTP client
├── output/output.go       # JSON contract: {ok, data} / {ok, error}
├── property/
│   ├── search.go          # search command backend
│   └── detail.go          # detail command backend
├── cmd/root.go
└── README.md
```

## License

MIT — see [LICENSE](../LICENSE).
