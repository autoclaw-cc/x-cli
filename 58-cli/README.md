# 58-cli

CLI wrapper around 58.com (58同城) rental listings backed by the [kimi-webbridge](https://www.kimi.com/features/webbridge) browser daemon. Runs inside the user's real Chrome session — no API key, no login required — and emits JSON on stdout.

## Commands

| Command | Usage | Returns |
|---|---|---|
| `search` | `58-cli search --city sh --keyword 张江 [--min-price 2000] [--max-price 5000] [--limit 20]` | `{listings: [{id, title, rent_monthly, layout, area_sqm, ...}]}` |
| `detail` | `58-cli detail --url "https://sh.58.com/zufang/xxx.shtml"` | `{title, rent_monthly, layout, area_sqm, address, images, ...}` |

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
go build -o 58-cli .
./58-cli --help
```

## Quick test

```bash
./58-cli search --city sh --keyword 张江 --max-price 5000 --limit 3
./58-cli detail --url "https://sh.58.com/zufang/87395837006556x.shtml"
```

## City codes

sz=深圳, sh=上海, bj=北京, gz=广州, cd=成都, hz=杭州, nj=南京, wh=武汉, cq=重庆, xa=西安, tj=天津, su=苏州

## Layout

```
58-cli/
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
