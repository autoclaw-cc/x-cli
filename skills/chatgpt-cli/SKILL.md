---
name: chatgpt-cli
description: Use when an agent needs to inspect or operate ChatGPT through the local x-cli browser CLI, including login checks, ordinary chat, web search, Deep Research, citation retrieval, stable answer persistence, image generation, and local image download through an isolated Kimi WebBridge session.
---

# ChatGPT CLI

Use `chatgpt-cli` only with an isolated Chrome profile already signed in by the
user. It returns one JSON envelope per command and never exposes browser
credentials or existing chat history.

## Prerequisites

- Verify Kimi WebBridge is running and the extension is connected.
- Run `login-status`; continue only when `authenticated:true`.
- Pass `--webbridge-url http://127.0.0.1:10400` on the current Windows machine.
- Keep login, CAPTCHA, 2FA, payment, and account settings manual.

## Commands

```powershell
chatgpt-cli --webbridge-url http://127.0.0.1:10400 login-status
chatgpt-cli --webbridge-url http://127.0.0.1:10400 capabilities
chatgpt-cli chat new
chatgpt-cli --timeout 2m chat ask --prompt "Return exactly: cobalt"
chatgpt-cli --timeout 3m chat ask --search --prompt "Find the official source."
chatgpt-cli --timeout 10m chat ask --deep-research --prompt "Compare the evidence."
chatgpt-cli chat ask --prompt "Summarize this" --output artifacts/chatgpt-answer.txt
chatgpt-cli --timeout 10m image generate --prompt "A green circle on white" --out artifacts/images --confirm
```

Use only one of `--search` and `--deep-research`. Increase `--timeout` for slow
research and image generation; do not retry by submitting another prompt while
a job may still be running. `image generate` requires explicit `--confirm`.

## Result Handling

- Chat results include `answer`, `mode`, `stable_samples`, and visible
  `citations`.
- `--output` writes only the returned answer text.
- Image results include the absolute local `path`, byte count, media type, and
  elapsed time. The CLI validates `image/*` before writing.
- Save outputs inside the active project or its `.codex` evidence directory,
  never the Desktop unless the user explicitly requests it.

## Safety Boundary

- Never automate credentials, CAPTCHA, payment, purchases, or account changes.
- Never print or persist cookies, tokens, headers, identity, private chat
  titles, or raw account endpoints.
- Commands create fresh conversations and do not select or modify existing
  chats.
- `chat new` leaves only its new blank tab open. Other workflows close their
  own WebBridge tabs after returning.
