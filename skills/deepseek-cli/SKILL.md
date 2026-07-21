---
name: deepseek-cli
description: Use when Codex needs to inspect or ask DeepSeek through the local x-cli DeepSeek browser CLI, including login checks, DeepThink, web search, file or image input, stable-answer retrieval, local answer persistence, and starting a fresh conversation through an isolated Kimi WebBridge browser session.
---

# DeepSeek CLI

Use the local `deepseek-cli` module at `D:\Tools\x-cli\deepseek-cli` or the active x-cli worktree equivalent.

## Prerequisites

- Use only the isolated DeepSeek Chrome profile connected through Kimi WebBridge. The default is `--webbridge-url http://127.0.0.1:10086`; use `http://127.0.0.1:10400` on the current Windows machine because 10086 is reserved.
- Login must be completed manually by the user.
- Do not read, export, save, or print cookies, local storage, authorization headers, or account identity.
- Do not output, export, save, or print existing private chat titles.
- Do not automate CAPTCHA, 2FA, payment, account settings, or credential entry.

## Commands

```powershell
deepseek-cli --webbridge-url http://127.0.0.1:10086 login-status
deepseek-cli --webbridge-url http://127.0.0.1:10086 capabilities
deepseek-cli --webbridge-url http://127.0.0.1:10086 --timeout 2m chat ask --prompt "Reply with exactly: lapis" --output ".codex\tmp\deepseek-answer.txt"
deepseek-cli --webbridge-url http://127.0.0.1:10400 chat new
deepseek-cli --webbridge-url http://127.0.0.1:10400 --timeout 5m chat ask --deepthink --search --file ".codex\tmp\case.md" --image ".codex\tmp\chart.png" --prompt "Analyze the attached material"
```

All commands return JSON:

- success: `{"ok":true,"data":...}`
- failure: `{"ok":false,"error":{"code":"...","message":"..."}}`

## Workflow

1. Run `login-status` first. Continue only if `authenticated:true` and `prompt_available:true`.
2. Run `capabilities` when you need to confirm visible modes such as `chat`, `deepthink`, `web_search`, and `vision`.
3. Use `chat new` when the task requires a clean conversation. This creates a new visible DeepSeek conversation.
4. Use `chat ask` for prompt/answer work. Add `--deepthink` and/or `--search` only when needed; the CLI waits until each corresponding visible toggle reports `aria-pressed=true`.
5. Add repeatable `--file PATH` and `--image PATH` flags for explicit local attachments. The CLI validates each path, selects and verifies vision mode for images, opens the visible upload control, and waits for the send control before submitting through Kimi WebBridge.
6. The command waits until the newest assistant answer is unchanged across stable samples before returning. Increase the root `--timeout` for DeepThink, web search, or large attachments.
7. Use `--output` for local persistence when the user asks to save an answer. Keep generated files inside the active project or `.codex\tmp` unless the user specifies another safe location.
