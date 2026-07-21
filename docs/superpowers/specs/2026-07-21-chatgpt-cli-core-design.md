# ChatGPT CLI Core Design

## Goal

Ship a focused `chatgpt-cli` that drives an already authenticated ChatGPT web
account exclusively through Kimi WebBridge. The first public version covers
the highest-value browser workflows without copying credentials, enumerating
private history, or growing into a general ChatGPT account manager.

## Public Commands

- `chatgpt-cli login-status`: report daemon health, authentication state,
  locale, and allowlisted capability names.
- `chatgpt-cli capabilities`: report only capabilities observed in the active
  account UI.
- `chatgpt-cli chat new`: open a clean ChatGPT composer.
- `chatgpt-cli chat ask --prompt <text> [--search|--deep-research]`: submit one
  prompt, wait for the final answer, and return answer text plus visible source
  citations when present.
- `chatgpt-cli image generate --prompt <text> --out <dir> --confirm`: generate
  one image with the account's included web capability and save it locally.

`--search` and `--deep-research` are mutually exclusive. Creation commands use
fresh conversations and never inspect or modify pre-existing conversation
titles.

## Architecture

The independent Go module follows the established x-cli layout:

- `browser`: daemon status, WebBridge commands, configurable URL, unique
  sessions, and best-effort session cleanup.
- `chatgpt`: allowlisted page inspection plus DOM workflows for chat, research,
  and image creation.
- `cmd`: Cobra command tree, validation, timeouts, and stable error codes.
- `output`: exactly one JSON envelope per invocation.

The implementation prefers semantic DOM signals and provider-owned test IDs.
It does not call private ChatGPT APIs directly, export browser state, or persist
headers, cookies, tokens, identity, plan billing details, or conversation lists.

## Completion And Download Detection

Chat commands record the baseline assistant-message count, submit once, and
wait until a new assistant response is present and the stop-streaming control
has disappeared for two consecutive polls. Deep Research uses the same stable
completion rule with a longer timeout and preserves visible citation URLs.

Image generation records baseline estuary file IDs, waits for a new completed
image, fetches that same-origin signed asset inside the browser, and writes the
decoded bytes beneath `--out`. The command never purchases credits or retries
by submitting a second prompt.

## Safety And Output Contract

- Every invocation uses a unique `chatgpt-cli-*` WebBridge session. Result-producing
  commands close their tabs; `chat new` intentionally leaves its fresh blank tab open.
- `--webbridge-url`, then `KIMI_WEBBRIDGE_URL`, then
  `http://127.0.0.1:10086` selects the daemon.
- Success: `{"ok":true,"data":...}`.
- Failure: `{"ok":false,"error":{"code":"...","message":"..."}}` and a
  non-zero process exit.
- Image generation requires `--confirm`; read-only status commands do not.
- Existing chats are never selected, renamed, deleted, or summarized.

## Release Boundary

Version `v0.1.0` is complete when unit tests, vet, cross-build checks, and one
live smoke test per workflow pass. The same publication change adds
`chatgpt-cli`, `notebooklm-cli`, and `deepseek-cli` to the repository release
workflow and publishes platform archives plus SHA-256 checksums.
