# Repo conventions for Claude Code

This repo is a collection of CLIs that each drive a real Chrome session via [kimi-webbridge](https://www.kimi.com/features/webbridge). One CLI per `<name>-cli/` directory; shared skills under `skills/`; PR-review tooling under `.claude/skills/pr-review/`.

## Output contract

Every CLI emits a single JSON envelope on stdout:

```json
{"ok": true,  "data": ...}                                          // success
{"ok": false, "error": {"code": "...", "message": "..."}}           // failure (exit 1)
```

Cobra subcommands should set `SilenceUsage = true` and `SilenceErrors = true` so cobra's own messages never leak onto stdout.

## CLI template structure

New CLIs follow the shape established by `baidu-cli/`, `boss-cli/`, `scholar-cli/`:

```
<name>-cli/
├── main.go                  # entrypoint, calls cmd.Execute
├── cmd/root.go              # cobra command tree + flag definitions
├── browser/client.go        # kimi-webbridge HTTP client (port Status() + use checkDaemon helper)
├── output/output.go         # the JSON envelope helpers above
├── go.mod                   # module path matches dir name; cobra is a DIRECT dep, not indirect
├── .gitignore               # /<name>-cli  *.exe  *.test  *.out
└── README.md
```

Two specific things to copy from existing CLIs:

1. `browser.Status()` + a `checkDaemon()` helper at the top of every WebBridge-using subcommand, so daemon-down surfaces as `daemon_unreachable` / `daemon_not_running` / `extension_not_connected` — not a raw connect error.
2. JS-side login detection in the extraction snippets, returning `{"error": "not_logged_in"}` that the Go side maps to a `not_logged_in` error code. Don't silently return empty results when the user is logged out.

## Hard rules

- **Don't commit build binaries.** Add `/<cli-name>` to that CLI's `.gitignore`. PRs that ship a Mach-O binary at `<cli>/<cli>` get blocked at review.
- **Don't ship legally-grey behaviors as default-on.** Sci-Hub, scraper-of-paywalled-content, etc. must be explicitly opt-in (`--scihub <domain>`), never on by default.
- **Don't hardcode contact details.** Polite-pool APIs (CrossRef, Unpaywall, NCBI) expect the caller's own email — make it env-configurable; default to an RFC-reserved `example.com` address, not a real Gmail.

## Testing / pr-review hygiene

The `.claude/skills/pr-review/` workflow checks out the PR, builds the CLI, and drives it against the real target site via kimi-webbridge to verify the user-facing claim. When doing this (or any manual WebBridge test):

- **Close every WebBridge tab / tab group / session you opened during testing or review before reporting done.** This includes the `pr-review` session itself and any per-CLI sessions (`boss`, `scholar-google`, etc.). Leaving them open litters the user's browser group across sessions.
- Screenshots posted to PR comments must be PII-scrubbed: scan for usernames, real names, account avatars in nav bars, request IPs / UUIDs on CAPTCHA pages. Redact before uploading via the helper script.

## After merge

When a CLI PR lands, run `update contributors for PR <N>` (see the pr-review skill). New contribution types only — returning contributors don't get re-listed.
