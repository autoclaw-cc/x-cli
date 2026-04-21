# Companion Skill Template

After the CLI is working, create `~/.claude/skills/{platform}-cli/SKILL.md` using this template.

**Purpose of a companion skill:** tells a future AI agent *how to use* the CLI — not how to build it.
The agent-cli-creator skill is for building; the companion skill is for operating.

---

Copy everything below the `---` line into the new file, then fill in the blanks.

---

```markdown
---
name: {platform}-cli
description: Use when the user wants to [read/post/search/interact with] [Site Name]. Invoke when user mentions "[site name]", asks to automate [site] tasks, or needs to [key action verbs].
---

# {Platform} CLI

Automates [Site Name] via the kimi-webbridge daemon using the user's real logged-in browser session.

## Prerequisites

1. kimi-webbridge daemon running:
   ```bash
   ~/.kimi-webbridge/bin/kimi-webbridge status
   ```
   If not running: invoke the `kimi-webbridge` skill.

2. CLI binary available:
   ```bash
   {platform}-cli --help
   ```

3. Logged in to [Site] in Chrome:
   ```bash
   {platform}-cli login-status
   ```
   If not logged in: open Chrome, navigate to [Site URL], log in manually, then retry.

## Commands

| Command | Args / Flags | Returns |
|---------|-------------|---------|
| `login-status` | — | `{logged_in, user_id, username}` |
| `home` | `[--limit N]` | `[{id, text, author, created_at}]` |
| `search` | `<query> [--limit N]` | `[{id, text, author}]` |
| `post` | `--content "text"` | `{id, url}` |
| _(add rows for each command)_ | | |

Run `{platform}-cli <command> --help` for full flag documentation.

## Output Format

All commands return JSON on stdout:

```json
{"ok": true, "data": ...}
```

On error (non-zero exit):

```json
{"ok": false, "error": {"code": "error_code", "message": "human-readable message"}}
```

## Common Workflows

**Check login then read home feed:**
```bash
{platform}-cli login-status && {platform}-cli home --limit 20
```

**Search and act on results:**
```bash
# Get IDs from search
{platform}-cli search "keyword" --limit 10

# Act on a specific item
{platform}-cli like <id>
```

_(Add more real workflows specific to this site)_

## Known Limitations

- Login must be performed manually in Chrome (no automated login)
- Rate limits: [describe if known, e.g. "search is limited to ~30 requests/minute"]
- [Any site-specific quirks discovered during development]
```

---

## Fill-In Checklist

- [ ] Replace `{platform}` with lowercase hyphenated name (e.g. `myblog`, `myshop`)
- [ ] Replace `[Site Name]` and `[Site URL]` with actual values
- [ ] Update `description` with real trigger phrases users would say
- [ ] Fill Commands table from `{platform}-cli --help` output
- [ ] Add 2–3 real common workflows
- [ ] Document known limitations and rate limits discovered during development
- [ ] Remove `login-status` row if site doesn't require login
