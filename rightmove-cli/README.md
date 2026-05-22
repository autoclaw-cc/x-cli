# rightmove-cli

CLI wrapper around Rightmove (UK) rental listings backed by the [kimi-webbridge](https://www.kimi.com/features/webbridge) browser daemon. Runs inside the user's real Chrome session — no API key, no login required — and emits JSON on stdout.

## Commands

| Command | Usage | Returns |
|---|---|---|
| `search` | `rightmove-cli search --location London [--min-beds 1] [--max-beds 2] [--min-price 800] [--max-price 2000] [--limit 20] [--page 1]` | `{properties: [{id, name, address, price, bedrooms, ...}]}` |
| `detail` | `rightmove-cli detail --url "https://www.rightmove.co.uk/properties/xxx"` | `{name, address, price, bedrooms, bathrooms, features, images, ...}` |

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
go build -o rightmove-cli .
./rightmove-cli --help
```

## Quick test

```bash
./rightmove-cli search --location London --min-beds 1 --max-price 2000 --limit 3
./rightmove-cli detail --url "https://www.rightmove.co.uk/properties/157584631"
```

## Locations

London, Manchester, Birmingham, Edinburgh, Bristol, Leeds, Liverpool, Glasgow, Cambridge, Oxford

## Layout

```
rightmove-cli/
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

MIT (see `LICENSE` — to be added).
