# bing-cli Architecture

## Design Principles

- **CLI stays simple** — one action per command, no orchestration. Agent composes calls.
- **Raw JSON by default** — primary consumer is AI agents, not humans.
- **No API key** — uses kimi-webbridge for real browser sessions → cookies → no bot detection.

## Dedup Strategy

| Flag | Behavior |
|------|----------|
| *(none)* | Raw Bing top 10, no filtering |
| `-u` | Title dedup (>80% char overlap) + max 3 per domain |

Only `-u` triggers dedup. Default is pass-through — agent decides if results are too noisy and retries with `-u`.

## Market Strategy

Default is `-m us` (international). Bing's `cc` parameter controls the market. Only use `-m cn` when Chinese-local sources (Baidu/Zhihu/etc) are specifically needed.

## Platform Patterns

### General Search Engines (Bing, Google, Baidu, Sogou)
- Same strategy: default pass-through, `-u` for dedup
- Domain diversity matters — results from multiple sources

### Platform-Internal Search (Xiaohongshu, Douyin, WeChat)
- All results from same domain → `-u` meaningless
- Need different dedup (content-based, not domain-based)
- Separate CLI per platform

## Project Structure

```
bing-cli/
├── main.go           # Entry point
├── browser/          # WebBridge HTTP client (reusable across all CLI tools)
│   └── client.go
├── output/           # JSON output helpers
│   └── output.go
├── bing/             # Bing-specific search logic
│   └── search.go
└── cmd/              # CLI commands (cobra)
    └── root.go
```
