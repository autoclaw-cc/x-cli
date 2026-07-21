# NotebookLM CLI Design

## Goal

Build a provider-specific Go CLI, `notebooklm-cli`, that drives the user's
already signed-in NotebookLM account through Kimi WebBridge. It must expose the
account's real browser capabilities without storing credentials, fabricating
unsupported features, or modifying notebooks that the CLI does not own.

## Ownership Boundary

The CLI maintains a local registry of notebooks it creates. Write commands are
allowed only when the target notebook is in that registry or the caller passes
an explicit authorization flag for a specific notebook ID. Existing notebooks
are never opened, read, renamed, sourced, shared, or deleted during development
or live verification.

Test notebooks use `CLI TEST - notebooklm-cli - <timestamp>` names. Test files,
downloads, and redacted evidence live under the repository's `.codex/.tmp`
directory. Cookies, account identifiers, email addresses, request headers,
private source contents, and existing notebook titles are never persisted or
returned.

## Command Model

The CLI has six command families:

1. `account`
   - `login-status` reports authenticated state and non-secret plan labels.
   - `capabilities` reports only controls observed in the live account, with
     evidence and availability status.
2. `notebook`
   - list only CLI-owned notebooks, create, inspect, rename, duplicate, and
     delete owned notebooks.
3. `source`
   - add text, URL, supported local files, YouTube, Drive selections, and
     account-supported discovery/research results.
   - list, inspect processing status, summarize, select, and remove sources in
     an owned notebook.
4. `chat`
   - grounded questions, multi-turn follow-up, source selection, citations,
     history, and save-to-note.
5. `note`
   - create, list, inspect, edit, remove, and convert an owned note to a source
     when the account exposes that action.
6. `studio`
   - list the account's real output types, create outputs, inspect asynchronous
     status, list generated outputs, and download/export supported artifacts.
   - Candidate output types include audio overview, video overview, mind map,
     report, FAQ, timeline, study guide, flashcards, quiz, infographic, and
     slide deck. Each is enabled only after live verification.

All commands support `--help` and emit exactly one JSON envelope:
`{"ok":true,"data":...}` or
`{"ok":false,"error":{"code":"...","message":"..."}}`.

## Architecture

The independent Go module follows the repository's established layout:

- `browser`: configurable WebBridge client, health checks, calls, evaluation,
  upload support, downloads, and best-effort `close_session` cleanup.
- `notebooklm`: provider-specific page models, selectors, workflows, ownership
  checks, polling, and allowlisted response structures.
- `registry`: local non-secret ownership registry with atomic writes.
- `cmd`: Cobra commands, shared flags, validation, authorization gates, and
  exit behavior.
- `output`: stable JSON envelopes and redaction.

The daemon URL resolves in this order: `--webbridge-url`,
`KIMI_WEBBRIDGE_URL`, then `http://127.0.0.1:10086`. Local development uses
`http://127.0.0.1:10400` without changing Windows networking.

## Data Flow

Read commands navigate to a lightweight page and return only allowlisted DOM or
in-page API fields. Write commands first resolve the target through the local
ownership registry, navigate directly to its URL, verify the live notebook ID,
perform one bounded action, poll the visible result, and update the registry or
download record only after success. Every invocation uses a unique WebBridge
session and closes it on success and failure.

Source uploads are explicit external actions. The CLI reports the file path,
type, and size before upload, never uploads hidden files, and rejects paths
outside the caller's requested scope. Downloaded outputs are written only to a
caller-selected project directory with collision-safe names.

## Capability Archaeology

Development uses one dedicated CLI-owned notebook and a deterministic,
non-secret case pack. The live matrix covers:

- authentication, plan labels, visible quotas, and language controls;
- notebook lifecycle;
- text, URL, file, YouTube, Drive, and discovery/research sources exposed by
  the account;
- grounded single-turn and multi-turn chat, source selection, citations, and
  save-to-note;
- note lifecycle and note-to-source conversion;
- every visible Studio output, its customization controls, asynchronous state,
  failure behavior, and supported download/export path;
- invalid URL, unsupported file, empty prompt, duplicate source, unauthenticated
  state, timeout, quota, and stale-selector failures.

Archaeology records endpoint families, selectors, request semantics, and
allowlisted response shapes, never secret headers or raw private payloads.

## Safety And Errors

The CLI fails closed for unknown notebook IDs, ownership mismatches, login
expiry, CAPTCHA, unsupported controls, and ambiguous selectors. Public sharing,
credential export, automated login, CAPTCHA bypass, purchases, added credits,
and auto top-up are out of scope. Existing notebooks are never fallback
targets.

Errors use stable codes including `not_logged_in`, `notebook_not_owned`,
`capability_unavailable`, `source_failed`, `generation_failed`, `timeout`, and
`quota_exhausted`.

## Verification

Each command is verified immediately with help output, focused unit tests, a
successful live path in the CLI-owned notebook, and at least one safe failure
path. Repository gates are `go test ./...`, `go vet ./...`, build-to-temporary
output only, JSON-envelope checks, credential/PII scans, ownership-boundary
tests, and a final WebBridge session leak check.

Implementation is incremental: account and ownership foundations first,
sources and chat second, notes and Studio third. Capability discovery is broad,
but commands ship only when their real account path is reproducible.
