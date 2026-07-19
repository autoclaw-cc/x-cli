# ChatGPT CLI First Slice Design

## Goal

Create a new Go CLI, `chatgpt-cli`, that drives the user's already signed-in
ChatGPT session through Kimi WebBridge. The first slice proves one account
read path and one useful creation path without storing credentials or changing
the existing `chatgpt-image-cli`.

## Commands

1. `chatgpt-cli login-status`
   - Verifies daemon and extension health.
   - Opens ChatGPT in an isolated WebBridge session.
   - Reports only non-secret login state, plan label when visible, and provider
     model slugs. It must not emit account IDs, email, names, cookies, headers,
     conversation titles, or tokens.
2. `chatgpt-cli capabilities`
   - Reads the current account UI and established read-only endpoint metadata.
   - Reports only capabilities actually observed for this account, with an
     `observed` or `official_included` evidence label. It must not describe
     homepage web search as Deep Research.
3. `chatgpt-cli image generate <prompt>`
   - Reuses the proven ChatGPT Images browser flow and downloads the resulting
     image into the requested project output directory.
   - Requires `--confirm` at action time because it creates a conversation and
     consumes the account's included image allowance.
   - Makes no upload, API call, credit purchase, or auto top-up.

Every command supports `--help` and returns exactly one JSON envelope:
`{"ok":true,"data":...}` or
`{"ok":false,"error":{"code":"...","message":"..."}}`.

## Architecture

The new independent Go module follows the repository's existing layout:

- `browser`: configurable WebBridge client, status checks, calls, evaluation,
  and best-effort `close_session` cleanup.
- `chatgpt`: read-only account inspection and image-generation orchestration.
- `cmd`: Cobra commands, shared flags, validation, confirmation gate, and exit
  behavior.
- `output`: stable JSON envelopes.

The daemon URL resolves in this order: `--webbridge-url`,
`KIMI_WEBBRIDGE_URL`, then `http://127.0.0.1:10086`. The local verification
uses `http://127.0.0.1:10400`; no Windows networking configuration changes are
needed.

## Session And Data Safety

Each invocation uses a unique `chatgpt-cli-*` session and defers
`close_session`, including error paths. Login credentials stay exclusively in
Chrome. Read commands return an allowlisted result schema, never raw endpoint
responses. Image output is written only beneath the caller-selected directory.

## Verification

Verification proceeds command by command:

- unit tests for configuration precedence, envelopes, redaction, and
  confirmation behavior;
- `go test ./...` and `go vet ./...`;
- live `login-status` and `capabilities` against port `10400`;
- `image generate --help` and an unconfirmed failure-path test;
- one real image generation only after the exact prompt, output path, account
  impact, and no-upload/no-extra-billing facts are shown to and approved by the
  user.

Future chat, search, Deep Research, file analysis, image editing, Projects,
Work, and Plugins commands are explicitly outside this first slice.
