# baidu-cli

CLI wrapper around Baidu search backed by the [kimi-webbridge](https://www.kimi.com/features/webbridge) browser daemon. Runs inside the user's real Chrome session — no cookie/header dance, full user context — and emits JSON on stdout.

## Commands

| Command | Usage | Returns |
|---|---|---|
| `search` | `baidu-cli search <query> [--limit 10] [--all]` | `{query, count, results: [{rank, id, tpl, title, url, abstract, source}]}` |

All output:

```json
{"ok": true, "data": ...}
```
or on failure (non-zero exit):

```json
{"ok": false, "error": {"code": "...", "message": "..."}}
```

## Flags

- `--limit, -n` (default `10`) — max results to return.
- `--all` — include "aladdin cards" (百科卡片、AI 回答卡、相关搜索 stub 等). By default the CLI calls `filterOrganic` to drop these and keep organic web results only. The filter logic in `baidu/filter.go` is a placeholder — see the TODO there to implement filtering for your use case.

## Prerequisites

1. **kimi-webbridge daemon** running on `127.0.0.1:10086` — install from https://www.kimi.com/features/webbridge.
2. **Go 1.25+** for building.

## Build

```bash
go build -o baidu-cli .
./baidu-cli --help
```

## Quick test

```bash
./baidu-cli search "claude code" --limit 3
./baidu-cli search "天气 北京"
./baidu-cli search "大模型" --all
```

## How it works

`search` navigates Chrome to `https://www.baidu.com/s?wd=<query>&rn=<limit>` and reads results from the **SSR'd DOM** (`.result.c-container, .result-op.c-container`) — no XHR replay needed. The real target URL is read from each container's `mu` attribute, which bypasses Baidu's `www.baidu.com/link?url=...` redirect wrapper.

Each result carries Baidu's internal `tpl` template name (e.g. `www_index` for organic, `sg_kg_entity_san` for baike entity cards, `ai_agent_distribute` for AI answers, `recommend_list` for "people also searched") — useful for downstream filtering or routing.

## Layout

```
baidu-cli/
├── main.go               # entrypoint
├── browser/client.go     # kimi-webbridge HTTP client
├── output/output.go      # JSON contract: {ok, data} / {ok, error}
├── baidu/
│   ├── search.go         # SERP navigation + DOM extraction
│   └── filter.go         # filterOrganic (TODO: implement per your needs)
├── cmd/
│   └── root.go
└── README.md
```

## License

MIT (see repo root `LICENSE`).
