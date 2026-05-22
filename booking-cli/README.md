# booking-cli

CLI wrapper around Booking.com hotel search backed by the [kimi-webbridge](https://www.kimi.com/features/webbridge) browser daemon. Runs inside the user's real Chrome session — no API key, no login required — and emits JSON on stdout.

## Commands

| Command | Usage | Returns |
|---|---|---|
| `search-hotels` | `booking-cli search-hotels --destination Kyoto [--checkin 2026-06-01] [--checkout 2026-06-03] [--adults 2] [--rooms 1] [--limit 10]` | `{hotels: [{name, location, score, reviews, price, ...}]}` |

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
go build -o booking-cli .
./booking-cli --help
```

## Quick test

```bash
./booking-cli search-hotels --destination Kyoto --limit 3
```

## Notes

- Booking.com renders in Chinese or English depending on the Chrome locale; the parser handles both.
- Login is optional but improves price accuracy — member prices can be significantly lower than the public listing. If a "Verify email" popup appears, complete it once in Chrome and retry.

## Layout

```
booking-cli/
├── main.go
├── browser/client.go      # kimi-webbridge HTTP client
├── output/output.go       # JSON contract: {ok, data} / {ok, error}
├── booking/search.go      # search-hotels command backend
├── cmd/root.go
└── README.md
```

## License

MIT (see `LICENSE` — to be added).
