# ctrip-cli

CLI wrapper around Ctrip (携程) — hotels, flights, attractions, destination guides — backed by the [kimi-webbridge](https://www.kimi.com/features/webbridge) browser daemon. Runs inside the user's real Chrome session — no API key, no login required — and emits JSON on stdout.

## Commands

| Command | Usage | Returns |
|---|---|---|
| `destination` | `ctrip-cli destination --name kyoto430` | `{must_do: [...], travel_notes: [...]}` |
| `search-hotels` | `ctrip-cli search-hotels --keyword Kyoto [--checkin 2026/06/01] [--checkout 2026/06/02] [--city-id N] [--country-id N] [--limit 10]` | `{hotels: [{name, star, address, score, price, currency, ...}]}` |
| `search-flights` | `ctrip-cli search-flights --from BJS --to SHA [--date 2026-06-15] [--limit 10]` | `{flights: [{airline, flight_no, dep_time, arr_time, price, ...}]}` |
| `search-attractions` | `ctrip-cli search-attractions --destination kyoto430 [--limit 20]` | `{attractions: [{name, score, review_count, price, is_free, ...}]}` |

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

**Pre-built binary (recommended)**: download `ctrip-cli-{platform}.tar.gz` from [Releases](https://github.com/better-world-ai/x-cli/releases?q=ctrip-cli) and extract. macOS users: clear the Gatekeeper quarantine once with `xattr -d com.apple.quarantine ./ctrip-cli`.

Available platforms: `darwin-{arm64,amd64}`, `linux-{amd64,arm64}`, `windows-{amd64,arm64}`.

**Build from source** (requires Go 1.21+):

```bash
go build -o ctrip-cli .
./ctrip-cli --help
```

## Quick test

```bash
./ctrip-cli destination --name kyoto430
./ctrip-cli search-attractions --destination kyoto430 --limit 5
```

## Notes

- Hotel and attraction queries use Ctrip's destination slug (e.g. `kyoto430`, `tokyo300`); browse https://you.ctrip.com/ to find slugs.
- Flight queries use IATA-like 3-letter codes (e.g. `BJS`, `SHA`, `CAN`).
- Login is optional but improves price accuracy on hotels — non-member prices are typically much higher.

## Layout

```
ctrip-cli/
├── main.go
├── browser/client.go      # kimi-webbridge HTTP client
├── output/output.go       # JSON contract: {ok, data} / {ok, error}
├── ctrip/
│   ├── hotels.go          # search-hotels backend
│   ├── flights.go         # search-flights backend
│   ├── attractions.go     # search-attractions backend
│   └── destination.go     # destination backend
├── cmd/root.go
└── README.md
```

## License

MIT — see [LICENSE](../LICENSE).
