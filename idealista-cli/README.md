# idealista-cli

CLI wrapper around Idealista (Spain/Italy/Portugal) rental listings backed by the [kimi-webbridge](https://www.kimi.com/features/webbridge) browser daemon. Runs inside the user's real Chrome session — no API key, no login required — and emits JSON on stdout.

## Commands

| Command | Usage | Returns |
|---|---|---|
| `search` | `idealista-cli search --country spain --city madrid-madrid [--min-rooms 1] [--max-rooms 3] [--min-price 500] [--max-price 1200] [--limit 20]` | `{properties: [{id, title, price, rooms, area, ...}]}` |
| `detail` | `idealista-cli detail --country spain --url "https://www.idealista.com/inmueble/xxx/"` | `{title, price, rooms, bathrooms, area, floor, features, images, ...}` |

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
go build -o idealista-cli .
./idealista-cli --help
```

## Quick test

```bash
./idealista-cli search --country spain --city madrid-madrid --max-price 1200 --limit 3
./idealista-cli detail --country spain --url "https://www.idealista.com/inmueble/106838498/"
```

## Countries and cities

| Country | Domain | City slugs |
|---------|--------|------------|
| spain | idealista.com | madrid-madrid, barcelona-barcelona, valencia, sevilla |
| italy | idealista.it | roma, milano, firenze |
| portugal | idealista.pt | lisboa, porto |

## Layout

```
idealista-cli/
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
