---
name: agent-cli-creator
description: Use when the user wants to build a CLI tool to automate browser interactions on a specific website using kimi-webbridge. Invoke when user says "create a CLI for X site", "build a tool to automate X", or wants to control a website programmatically via an AI agent.
---

# Agent CLI Creator

Guide an AI Coding Agent to build a website-automating CLI tool backed by the kimi-webbridge browser daemon.

## Phase 1: Prerequisites

```bash
~/.kimi-webbridge/bin/kimi-webbridge status
```

| Result | Action |
|--------|--------|
| `running: true` + `extension_connected: true` | Proceed |
| Command not found | Tell user to visit https://www.kimi.team/features/webbridge (中文: https://www.kimi.team/zh-cn/features/webbridge) to install |
| `running: false` or `extension_connected: false` | Invoke the `kimi-webbridge` skill → references/operations.md |

## Phase 2: Requirements Interview

Ask the user in one message, wait for reply:

1. **Target website URL** (required)
2. **Programming language** — Go (recommended if in this repo) / Python / Node.js / Other
3. **Login required?** — Yes / No / Unknown (can skip for now)
4. **First 1–3 features** — pick from common categories:
   - Read: home feed, search, profile page, post/item detail
   - Write: create post, like/unlike, comment, bookmark/save
   - Account: login-status, user info

**Explain iterative development to the user:**
> "We'll start with 1–3 features to validate the approach end-to-end before adding more. Site archaeology for features you don't need yet wastes time. You can always add features later by re-running from Phase 3."

## Phase 3: Site Archaeology (mandatory before writing any code)

**This phase is non-negotiable. Do not write business logic until archaeology is complete.**

For each planned feature, run the full protocol in `references/site-exploration.md`.

The protocol yields: API endpoint, required headers, response shape, and a verified `evaluate` call that proves the API works inside the browser session.

**Only proceed to Phase 4 when you have a working `evaluate` call for every planned feature.**

## Phase 4: Implement

Order matters — do not skip ahead:

1. **Project scaffold** — see `references/go-layout.md` for Go; adapt conventions for other languages
2. **`login-status` command** — if the site requires login; see `references/login-handling.md`
3. **Read commands** — no side effects; implement and test first
4. **Write commands** — side effects (post, like, etc.); implement after reads work

**After each command is implemented, immediately verify before moving on:**
```bash
{platform}-cli {command} --help          # --help must work
{platform}-cli {command} [args]          # must return {"ok": true, "data": ...}
```
For write commands, also verify the error path (e.g., wrong ID, missing flag) returns `{"ok": false, ...}` with non-zero exit.

If verification fails, fix before implementing the next command.

**Universal CLI contract (all languages):**
- `--help` / `-h` must work on every command
- All output: `{"ok": true, "data": ...}` or `{"ok": false, "error": {"code": "...", "message": "..."}}`
- Non-zero exit code on error

**Go:** 用 `go mod init {platform}-cli` 初始化独立 module，按 `references/go-layout.md` 的结构搭建，包含自己的轻量 browser client 和 output helper。

## Phase 5: Write Companion Skill

After the CLI works, create `~/.claude/skills/{platform}-cli/SKILL.md` using the template in `references/companion-skill-template.md`.

**Purpose:** tells a future AI agent *how to use* the CLI, not how to build it. The two skills serve different audiences at different times.
