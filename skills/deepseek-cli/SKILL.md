---
name: deepseek-cli
description: Use when Codex needs to inspect or ask DeepSeek through the local x-cli DeepSeek browser CLI, including checking login state, listing visible DeepSeek web capabilities, sending a prompt, waiting for a stable answer, or saving the answer locally through an isolated Chrome CDP session.
---

# DeepSeek CLI

Use the local `deepseek-cli` module at `D:\Tools\x-cli\deepseek-cli` or the active x-cli worktree equivalent.

## Prerequisites

- Use only the isolated DeepSeek Chrome profile and its CDP endpoint, normally `--cdp-url http://127.0.0.1:9223`.
- Login must be completed manually by the user.
- Do not read, export, save, or print cookies, local storage, authorization headers, or account identity.
- Do not output, export, save, or print existing private chat titles.
- Do not automate CAPTCHA, 2FA, payment, account settings, or credential entry.

## Commands

```powershell
deepseek-cli --cdp-url http://127.0.0.1:9223 login-status
deepseek-cli --cdp-url http://127.0.0.1:9223 capabilities
deepseek-cli --cdp-url http://127.0.0.1:9223 --timeout 2m chat ask --prompt "Reply with exactly: lapis" --output ".codex\tmp\deepseek-answer.txt"
```

All commands return JSON:

- success: `{"ok":true,"data":...}`
- failure: `{"ok":false,"error":{"code":"...","message":"..."}}`

## Workflow

1. Run `login-status` first. Continue only if `authenticated:true` and `prompt_available:true`.
2. Run `capabilities` when you need to confirm visible modes such as `chat`, `deepthink`, `web_search`, and `vision`.
3. Use `chat ask` for normal prompt/answer work. It sends the prompt through Chrome DevTools input events, then waits for the newest assistant answer to remain stable before returning.
4. Use `--output` for local persistence when the user asks to save an answer. Keep generated files inside the active project or `.codex\tmp` unless the user specifies another safe location.
