# gaokao-cli

高考信息查询 CLI，基于[掌上高考 (gaokao.cn)](https://www.gaokao.cn) 公开 CDN 数据。直接请求公开 JSON 接口，无需 kimi-webbridge 或浏览器。

## Commands

| Command | Usage | Returns |
|---|---|---|
| `provinces` | `gaokao-cli provinces` | `[{id, name}]` |
| `score-line` | `gaokao-cli score-line --province 北京 [--year 2025] [--type 综合]` | `{province, year, lines: [{year, type_name, batch_name, score, ...}]}` |
| `score-section` | `gaokao-cli score-section --province 北京 [--year 2025] [--score 650]` | `{province, year, type, entries: [{score, rank, count, ...}]}` |
| `school` | `gaokao-cli school --name 清华 [--985] [--211] [--dual-class]` | `{count, schools: [{id, name, province, f985, f211, ...}]}` |
| `batch-history` | `gaokao-cli batch-history [--province 北京]` | `[{year, province, content}]` |

All output:

```json
{"ok": true, "data": ...}
```
or on failure (non-zero exit):

```json
{"ok": false, "error": {"code": "...", "message": "..."}}
```

## Prerequisites

1. **Go 1.25+** for building.
2. No browser or kimi-webbridge required — data is fetched from public CDN.

## Build

```bash
go build -o gaokao-cli .
./gaokao-cli --help
```

## Quick test

```bash
./gaokao-cli provinces
./gaokao-cli score-line --province 北京 --year 2025
./gaokao-cli score-section --province 北京 --year 2025 --score 650
./gaokao-cli school --name 清华
./gaokao-cli batch-history --province 北京
```

## Layout

```
gaokao-cli/
├── main.go                # entrypoint
├── output/output.go       # JSON contract: {ok, data} / {ok, error}
├── gaokao/
│   ├── client.go          # HTTP client, FetchJSON/FetchData
│   ├── provinces.go       # 31-province map, ResolveProvince
│   ├── scoreline.go       # score-line (省控线) backend
│   ├── section.go         # score-section (一分一段表) backend
│   ├── school.go          # school search backend
│   └── batchhistory.go    # batch-history (批次改革) backend
├── cmd/
│   └── root.go            # Cobra CLI wiring
└── README.md
```

## License

MIT (see repo root `LICENSE`).
