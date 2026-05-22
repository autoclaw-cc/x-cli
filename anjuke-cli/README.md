# anjuke-cli

CLI wrapper around Anjuke (安居客) rental listings backed by the [kimi-webbridge](https://www.kimi.com/features/webbridge) browser daemon. Runs inside the user's real Chrome session — no API key, no login required — and emits JSON on stdout.

## Commands

| Command | Usage | Returns |
|---|---|---|
| `search` | `anjuke-cli search --city sh --keyword 张江 [--min-price 2000] [--max-price 5000] [--limit 20]` | `{listings: [{id, title, price, layout, area, ...}]}` |
| `detail` | `anjuke-cli detail --city sh --id 123456789` | `{title, price, layout, area, address, images, ...}` |

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
go build -o anjuke-cli .
./anjuke-cli --help
```

## Quick test

```bash
./anjuke-cli search --city sh --keyword 张江 --max-price 5000 --limit 3
./anjuke-cli detail --city sh --id 3717986222631942
```

## City names

Use short city codes: sz (深圳), bj (北京), sh (上海), gz (广州), hz (杭州), cd (成都), tj (天津), nj (南京), wh (武汉), cs (长沙), cq (重庆), xa (西安)

## Layout

```
anjuke-cli/
├── main.go
├── browser/client.go      # kimi-webbridge HTTP client
├── output/output.go       # JSON contract: {ok, data} / {ok, error}
├── house/
│   ├── search.go          # search command backend
│   └── detail.go          # detail command backend
├── cmd/root.go
└── README.md
```

## License

MIT (see `LICENSE` — to be added).
