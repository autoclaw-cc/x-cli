# Repo conventions

This repo collects CLIs. The ones in it today are all Go and drive Chrome via [kimi-webbridge](https://www.kimi.com/features/webbridge), but that's how they happen to have been built — it's not a constraint on new CLIs in other languages or architectures. If you're adding one in the existing pattern, `baidu-cli/` is a fine reference.

## Hard rules

- **Don't commit build artifacts.** Whatever your language emits — the `<name>-cli` binary for Go, `dist/` / `target/` / `__pycache__/` / `node_modules/`, etc. — belongs in that CLI's `.gitignore`, not in git.
- **Don't ship legally-grey behaviors as default-on.** Sci-Hub, paywall scrapers, and the like must require an explicit opt-in flag.
- **Don't hardcode personal contact details or credentials.** API contact emails should be env-configurable with an RFC-reserved (`example.com`) default; no real Gmail addresses, tokens, cookies, or session keys.

## Before fixing an issue

- **Check for an in-flight PR before writing any code for an existing issue** — `gh pr list --state open --search "<issue-number>"`. GitHub only lets you assign someone who has write access or has already commented, so an outside contributor who picked up a `good first issue` **cannot** be set as the assignee — the issue looks unclaimed when it isn't. (`gh issue edit --add-assignee` fails *silently* on such a user: it returns success, but `assignees` stays empty. Verify with `gh issue view --json assignees`.) If a PR already exists, review it — don't write a competing one.
- **`good first issue` is a promise.** Once that label is on, the issue is no longer yours to quietly fix — you've invited a stranger to spend their time. If you need it done sooner, say so in the issue first and give whoever may be working on it a chance to reply.

## Testing / pr-review hygiene

- **Close every WebBridge tab / tab group / session opened during testing or review before reporting done** — including the `pr-review` session itself and any per-CLI sessions. Don't litter the user's browser group.
- **PII-scrub screenshots before uploading to PR comments.** Scan for usernames, real names, account avatars in nav bars, request IPs / UUIDs on CAPTCHA or error pages. Redact before posting.

## After merge

When a CLI PR lands, update the all-contributors table for it. Returning contributors only get NEW contribution types added — don't re-list types they already have.
