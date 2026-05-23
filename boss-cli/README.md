# boss-cli

CLI wrapper around Boss直聘 (BOSS Zhipin) job search backed by the [kimi-webbridge](https://www.kimi.com/features/webbridge) browser daemon. Runs inside the user's real Chrome session and emits JSON on stdout.

## Commands

| Command | Usage | Returns |
|---|---|---|
| `login-status` | `boss-cli login-status` | `{logged_in, username}` |
| `search-jobs` | `boss-cli search-jobs --query "前端开发" [--city 101020100] [--salary 404] [--experience 105] [--limit 10]` | `{jobs: [{title, salary, company, ...}]}` |
| `job-detail` | `boss-cli job-detail --id <encrypted_id>` | `{title, salary, company, description, ...}` |

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
3. **Login required** — must be logged in to Boss直聘 in Chrome. Run `boss-cli login-status` to check.

## Build

```bash
go build -o boss-cli .
./boss-cli --help
```

## Quick test

```bash
./boss-cli login-status
./boss-cli search-jobs --query "Java" --city 101020100 --limit 3
./boss-cli job-detail --id <id_from_search>
```

## City codes

| Code | City |
|------|------|
| 101010100 | 北京 |
| 101020100 | 上海 |
| 101280100 | 广州 |
| 101280600 | 深圳 |
| 101210100 | 杭州 |
| 101030100 | 天津 |
| 101190400 | 苏州 |
| 101110100 | 西安 |

## Filter codes

**Salary**: 401=3K以下, 402=3-5K, 403=5-10K, 404=10-20K, 405=20-50K, 406=50K以上

**Experience**: 101=在校生, 102=应届生, 103=经验不限, 104=1年以内, 105=1-3年, 106=3-5年, 107=5-10年, 108=10年以上

**Degree**: 209=初中及以下, 208=中专/中技, 206=高中, 202=大专, 203=本科, 204=硕士, 205=博士

## Layout

```
boss-cli/
├── main.go
├── browser/client.go      # kimi-webbridge HTTP client
├── output/output.go       # JSON contract: {ok, data} / {ok, error}
├── boss/
│   ├── search.go          # search-jobs command backend
│   ├── detail.go          # job-detail command backend
│   └── login.go           # login-status check
├── cmd/root.go
└── README.md
```

## License

MIT (see `LICENSE` — to be added).
