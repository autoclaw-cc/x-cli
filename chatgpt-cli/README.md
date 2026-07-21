# chatgpt-cli

`chatgpt-cli` drives an already signed-in ChatGPT web account through Kimi
WebBridge. It supports focused chat, web search, Deep Research, and image
generation without exporting browser credentials or reading existing chat
history.

## Prerequisites

1. Run Kimi WebBridge and connect its Chrome extension.
2. Use a dedicated Chrome profile and sign in to ChatGPT manually.
3. Pass the daemon URL when it is not the default
   `http://127.0.0.1:10086`. The current Windows setup uses port `10400`.

```powershell
chatgpt-cli --webbridge-url http://127.0.0.1:10400 login-status
chatgpt-cli --webbridge-url http://127.0.0.1:10400 capabilities
```

Every command prints exactly one JSON envelope. Errors use a non-zero exit
status.

## Chat And Research

```powershell
chatgpt-cli chat new
chatgpt-cli --timeout 2m chat ask --prompt "Return exactly: cobalt"
chatgpt-cli --timeout 3m chat ask --search --prompt "Find the official IANA reserved domains page."
chatgpt-cli --timeout 10m chat ask --deep-research --prompt "Compare the authoritative evidence."
chatgpt-cli chat ask --prompt "Summarize this" --output artifacts/chatgpt-answer.txt
```

`--search` and `--deep-research` are mutually exclusive. Each `chat ask`
opens its own fresh conversation, selects a requested mode once, submits once,
waits for a non-streaming answer to remain stable across three samples, and
returns visible citation URLs. `chat new` intentionally leaves its new blank
tab open; result-producing commands close their own WebBridge tabs.

## Image Generation

```powershell
chatgpt-cli --timeout 10m image generate \
  --prompt "A green circle centered on white, no text" \
  --out artifacts/images \
  --confirm
```

Image generation requires `--confirm` because it creates a ChatGPT
conversation and consumes the signed-in account's included allowance. The CLI
records the existing image IDs, submits once, waits for a new completed image,
downloads that exact signed asset inside the browser, validates its `image/*`
media type, and writes it beneath `--out`. It never buys credits or retries by
submitting another prompt.

## Configuration

The WebBridge URL resolves in this order:

1. `--webbridge-url`
2. `KIMI_WEBBRIDGE_URL`
3. `http://127.0.0.1:10086`

Use the root `--timeout` flag for slow research or image jobs. Ten minutes is
the default.

## Safety Boundary

- Login, CAPTCHA, 2FA, payment, and account settings remain manual.
- Cookies, local storage, headers, tokens, account identity, plan billing
  details, and existing conversation titles are never returned or persisted.
- Commands do not select, rename, delete, or summarize existing chats.
- Prompts are transmitted only when explicitly passed to `chat ask` or
  confirmed `image generate`.
- Local output stays in the caller-selected project directory.

See [ARCHAEOLOGY.md](./ARCHAEOLOGY.md) for the verified provider signals and
completion rules.
