# Site Exploration Protocol (Archaeology)

Run this full protocol for each planned feature **before** writing any implementation code.

All `curl` commands below use the kimi-webbridge daemon at `http://127.0.0.1:10086`.
Replace `<PLATFORM>` with your session name (e.g. `myblog`, `myshop`).

---

## The 6-Step Protocol

### Step 1: Navigate to the target page

```bash
curl -s -X POST http://127.0.0.1:10086/command \
  -H 'Content-Type: application/json' \
  -d '{"action":"navigate","args":{"url":"<TARGET_URL>","newTab":true},"session":"<PLATFORM>"}'
```

### Step 2: Snapshot — understand page structure

```bash
curl -s -X POST http://127.0.0.1:10086/command \
  -H 'Content-Type: application/json' \
  -d '{"action":"snapshot","session":"<PLATFORM>"}'
```

Review the `tree` field (accessibility tree). Note `@e` refs for interactive elements.
If the feature is purely DOM-based (clicking a button, reading visible text), you may be able to skip to Step 6.

### Step 3: Start network capture

```bash
curl -s -X POST http://127.0.0.1:10086/command \
  -H 'Content-Type: application/json' \
  -d '{"action":"network","args":{"cmd":"start"},"session":"<PLATFORM>"}'
```

Then **manually trigger the action in Chrome** (scroll, click the button, submit the form, etc.).

### Step 4: Stop capture and list requests

```bash
# Stop
curl -s -X POST http://127.0.0.1:10086/command \
  -d '{"action":"network","args":{"cmd":"stop"},"session":"<PLATFORM>"}'

# List
curl -s -X POST http://127.0.0.1:10086/command \
  -d '{"action":"network","args":{"cmd":"list"},"session":"<PLATFORM>"}'
```

Scan request URLs to identify the relevant XHR/GraphQL/REST call.
Ignore static assets (`.js`, `.css`, images). Focus on API calls with meaningful responses.

### Step 5: Inspect request detail

```bash
curl -s -X POST http://127.0.0.1:10086/command \
  -d '{"action":"network","args":{"cmd":"detail","requestId":"<ID>"},"session":"<PLATFORM>"}'
```

Extract:
- Full URL (including query params)
- Method (GET/POST)
- Request headers — especially `Authorization`, `Cookie`, `x-csrf-token`, `x-client-transaction-id`
- Request body (for POST/GraphQL)
- Response body shape (field names, nested structure, pagination cursor)

### Step 6: Verify with `evaluate` (the critical check)

Replicate the API call inside the browser using `evaluate`. The browser context provides cookies and auth headers automatically.

```bash
curl -s -X POST http://127.0.0.1:10086/command \
  -H 'Content-Type: application/json' \
  -d '{
    "action": "evaluate",
    "args": {
      "code": "const r = await fetch('"'"'/api/v2/feed'"'"', {headers: {'"'"'x-csrf-token'"'"': document.cookie.match(/ct0=([^;]+)/)?.[1]}}); return await r.json();"
    },
    "session": "<PLATFORM>"
  }'
```

**If Step 6 returns expected data → archaeology done. Proceed to Phase 4.**
**If Step 6 fails → check auth headers, adjust request, retry Step 6.**

---

## When to Skip Network Capture (Steps 3–5)

Skip network capture if the feature is purely DOM-based:
- Reading text already visible in the page
- Clicking a button whose effect is visible in the DOM
- In these cases: Step 2 snapshot gives you `@e` refs, Step 6 uses `evaluate` to click/read

## Common Patterns

| Pattern | What it means | How to handle |
|---------|---------------|---------------|
| GraphQL endpoint | Single URL, `operationName` in body | Replicate same URL + body in `evaluate` |
| REST with cookie auth | Browser cookies provide auth | `evaluate` fetch inherits cookies automatically |
| `x-csrf-token` header | Token extracted from cookie | `document.cookie.match(/ct0=([^;]+)/)?.[1]` |
| Cursor-based pagination | `cursor`/`after` field in response | Pass cursor back in next request args |
| Data only in XHR, not DOM | Real content loads via JS | Must use network capture; DOM snapshot is insufficient |

## Archaeology Summary Template

After completing the protocol, document findings before writing code:

```
Feature: [home feed / search / post / ...]
URL: GET/POST https://example.com/api/v1/...
Auth: [cookie only / x-csrf-token / bearer token]
Key request params: [query, cursor, limit]
Response shape: {data: [{id, text, author: {name}, created_at}], next_cursor}
Evaluate call: [paste the working JS]
```
