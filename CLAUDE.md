# Repo conventions

This repo collects CLIs. The ones in it today are all Go and drive Chrome via [kimi-webbridge](https://www.kimi.com/features/webbridge), but that's how they happen to have been built — it's not a constraint on new CLIs in other languages or architectures. If you're adding one in the existing pattern, `baidu-cli/` is a fine reference.

## Hard rules

- **Don't commit build artifacts.** Whatever your language emits — the `<name>-cli` binary for Go, `dist/` / `target/` / `__pycache__/` / `node_modules/`, etc. — belongs in that CLI's `.gitignore`, not in git.
- **Don't ship legally-grey behaviors as default-on.** Sci-Hub, paywall scrapers, and the like must require an explicit opt-in flag.
- **Don't hardcode personal contact details or credentials.** API contact emails should be env-configurable with an RFC-reserved (`example.com`) default; no real Gmail addresses, tokens, cookies, or session keys.

## Browser-driven CLIs

- **Make the tab render before reading it.** Tabs the daemon opens sit in the background: Chrome reports `visibilityState: "hidden"`, skips layout, and never fires `IntersectionObserver`. Clicks hit 0×0 elements, infinite scroll and lazy images never fire, and the failure looks exactly like the site blocking you. Every `browser/client.go` here now carries `browser/activate.go`; `Navigate` calls `Activate()` for you. Keep that when adding a CLI, and expose the `--activate front|auto|off` flag (`cmd/activate.go`) plus the shared `WEBBRIDGE_ACTIVATE` env var. **Activating is the default** — `browser.DefaultActivateMode = ActivateFront` in every CLI here, meaning the tab is really raised so it really renders. A new CLI that only ever reads the server-rendered first screen may set that constant to `ActivateOff` to stay out of the user's way; say why in a comment, because getting it wrong fails invisibly.
- **Before writing down a site-side limit, check a human wouldn't hit it.** If a person clicking through the same flow gets further than the CLI does, the difference is in our setup — tab visibility, event timing, hydration — not in the site.

## Testing / pr-review hygiene

- **Close every WebBridge tab / tab group / session opened during testing or review before reporting done** — including the `pr-review` session itself and any per-CLI sessions. Don't litter the user's browser group.
- **PII-scrub screenshots before uploading to PR comments.** Scan for usernames, real names, account avatars in nav bars, request IPs / UUIDs on CAPTCHA or error pages. Redact before posting.

## After merge

When a CLI PR lands, update the all-contributors table for it. Returning contributors only get NEW contribution types added — don't re-list types they already have.
