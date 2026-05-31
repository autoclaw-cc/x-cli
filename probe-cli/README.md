# probe-cli

Automates the [Site Archaeology protocol](https://github.com/better-world-ai/agent-cli-creator) from the x-cli ecosystem.

One command replaces 30-60 minutes of manual `curl` calls with 30 seconds of auto-detection.

## What it does

Given any URL, probe-cli connects to the kimi-webbridge daemon and automatically:

1. **Detects auth mechanism** — scans localStorage, sessionStorage, and cookies for token/session/csrf patterns
2. **Extracts DOM elements** — finds all forms, inputs, buttons with their selectors
3. **Captures network traffic** — filters XHR/Fetch requests, inspects auth headers, identifies API endpoints
4. **Recommends a CLI pattern** — scores dom-scrape / api-reverse / form-submit / async-poll based on findings
5. **Outputs a JSON site profile** — ready for the agent-cli-creator Phase 4 (implementation)

## Protocol Mapping

| Site Archaeology Step | probe-cli automation |
|-----------------------|---------------------|
| Step 1: Navigate | ✅ Auto-navigate to URL |
| Step 2: Snapshot DOM | ✅ Auto-extract forms, inputs, buttons |
| Step 3: Start network capture | ✅ Auto-start before navigate |
| Step 4: Stop + list requests | ✅ Auto-stop after page settles |
| Step 5: Inspect request detail | ✅ Auto-inspect auth headers |
| Step 6: Verify with evaluate | ✅ Auth pattern detection + recommendation |

## Install

```bash
# Build from source
git clone https://github.com/RachelXiaolan/probe-cli.git
cd probe-cli
go build -o probe-cli .
```

## Usage

```bash
# Full probe (auth + DOM + network)
probe-cli https://example.com

# Quick mode (auth + DOM only, skip network capture)
probe-cli https://example.com --quick

# Custom session name
probe-cli https://example.com --session mysite
```

**Prerequisites:** kimi-webbridge daemon running at `http://127.0.0.1:10086`

## Output Format

```json
{
  "ok": true,
  "data": {
    "url": "https://fish.audio/zh-CN/app/text-to-speech/",
    "timestamp": "2026-06-01T12:00:00+09:00",
    "auth": {
      "detected": true,
      "method": "bearer-localstorage",
      "storage_type": "localStorage",
      "token_keys": ["token"]
    },
    "dom": {
      "forms": [],
      "standalone_inputs": [
        {"tag": "div", "content_editable": true, "role": "textbox"}
      ],
      "buttons": [
        {"text": "生成语音", "type": "button"}
      ]
    },
    "network": {
      "total_requests": 42,
      "api_endpoints": [
        {
          "method": "POST",
          "url": "https://api.fish.audio/task",
          "has_auth_header": true,
          "auth_headers": ["Authorization: Bearer eyJhbG***"]
        }
      ],
      "static_filtered": 37
    },
    "pattern": {
      "type": "api-reverse",
      "confidence": 0.85,
      "reason": "Data loaded via XHR/Fetch API calls with bearer-localstorage auth. Replicate API calls in evaluate()."
    }
  }
}
```

## How it works

```
┌─────────────────────────────────────────┐
│ 1. Start network capture                │
│ 2. Navigate to URL                      │
│ 3. Wait 3s for page to settle           │
├─────────────────────────────────────────┤
│ 4. Inject auth detection JS             │
│    → localStorage / cookie / sessionStorage scan │
│ 5. Inject DOM extraction JS             │
│    → forms, inputs, buttons             │
├─────────────────────────────────────────┤
│ 6. Stop network capture                 │
│ 7. List + filter XHR/Fetch requests     │
│ 8. Inspect auth headers per endpoint    │
├─────────────────────────────────────────┤
│ 9. Score patterns + recommend           │
│ 10. Output site profile JSON            │
└─────────────────────────────────────────┘
```

## Integration with agent-cli-creator

After probe-cli outputs a site profile, the AI agent can skip directly to Phase 4 (Implementation) of the agent-cli-creator workflow:

```
Before: Agent reads site-exploration.md → manually runs 6 curl steps → 30-60 min
After:  Agent runs probe-cli → gets site-profile.json → skip to Phase 4
```

The site profile contains everything the agent needs:
- `auth.method` → tells the agent how to authenticate
- `network.api_endpoints` → tells the agent which APIs to call
- `dom.forms/inputs/buttons` → tells the agent which elements to interact with
- `pattern.type` → tells the agent which template to use

## Auth Detection Patterns

| Detected Method | What it means | How CLI should handle |
|----------------|---------------|----------------------|
| `bearer-localstorage` | JWT/Bearer token in localStorage | Extract token via evaluate, add to fetch headers |
| `csrf-cookie` | CSRF token in cookie (e.g. ct0) | Extract from cookie, add x-csrf-token header |
| `cookie-only` | Session cookie, no extra headers | Just use evaluate (cookies sent automatically) |
| `session-storage` | Auth token in sessionStorage | Extract via evaluate, add to headers |
| `localstorage-other` | Non-token auth data in localStorage | Inspect keys for API key or session ID |
| *(none detected)* | No login required, or complex auth | Manual archaeology needed |

## CLI Pattern Types

| Pattern | When | Implementation approach |
|---------|------|------------------------|
| `dom-scrape` | Data visible in DOM, no API | Snapshot + evaluate to extract text |
| `api-reverse` | Data from XHR/Fetch with auth | Replicate fetch calls in evaluate |
| `form-submit` | Forms to fill and submit | Fill fields + click submit via evaluate |
| `async-poll` | POST action → poll GET for result | Submit + loop poll until done |

## License

MIT
