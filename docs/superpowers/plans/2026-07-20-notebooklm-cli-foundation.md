# NotebookLM CLI Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship a tested Go `notebooklm-cli` with account inspection, owned-notebook registration, pasted-text ingestion, grounded chat, and live Studio capability discovery.

**Architecture:** A configurable WebBridge client owns one unique browser session per invocation and always closes it. Provider workflows depend on a narrow browser interface for unit testing, while an atomic local registry prevents write commands from targeting notebooks that were not created or explicitly authorized by the caller.

**Tech Stack:** Go 1.23, Cobra, `net/http`, `httptest`, Kimi WebBridge, NotebookLM web UI.

---

### Task 1: JSON output and daemon configuration

**Files:**
- Create: `notebooklm-cli/go.mod`
- Create: `notebooklm-cli/main.go`
- Create: `notebooklm-cli/output/output.go`
- Create: `notebooklm-cli/output/output_test.go`
- Create: `notebooklm-cli/config/config.go`
- Create: `notebooklm-cli/config/config_test.go`
- Create: `notebooklm-cli/.gitignore`

- [ ] Write failing tests for success/error envelopes and URL precedence:
  flag value, `KIMI_WEBBRIDGE_URL`, then `http://127.0.0.1:10086`.
- [ ] Run `go test ./output ./config` and verify missing packages/functions fail.
- [ ] Implement the minimal JSON encoder and `ResolveWebBridgeURL`.
- [ ] Run `go test ./output ./config` and verify pass.

Expected public APIs:

```go
func Success(w io.Writer, data any) error
func Error(w io.Writer, code, message string) error
func ResolveWebBridgeURL(flagValue string, getenv func(string) string) string
```

### Task 2: WebBridge client and session cleanup

**Files:**
- Create: `notebooklm-cli/browser/client.go`
- Create: `notebooklm-cli/browser/client_test.go`

- [ ] Write failing `httptest` cases for `/status`, `/command`, evaluate
  wrappers, daemon errors, configurable URL, and `close_session`.
- [ ] Run `go test ./browser` and verify the client is missing.
- [ ] Implement `Client`, `Status`, `Call`, `Navigate`, `EvaluateValue`,
  `MouseClick`, `KeyType`, `SendKeys`, `CDP`, and `CloseSession`.
- [ ] Run `go test ./browser` and verify pass.

The provider-facing interface is:

```go
type Bridge interface {
    Navigate(url string, newTab bool, groupTitle string) error
    EvaluateValue(code string, dst any) error
    MouseClick(selector string) error
    KeyType(text string) error
    SendKeys(keys string) error
    CDP(method string, params map[string]any) error
    CloseSession() error
}
```

### Task 3: Atomic ownership registry

**Files:**
- Create: `notebooklm-cli/registry/registry.go`
- Create: `notebooklm-cli/registry/registry_test.go`

- [ ] Write failing tests for empty load, authorize, exact ID lookup,
  duplicate update, invalid NotebookLM URL rejection, and atomic JSON save.
- [ ] Run `go test ./registry` and verify failure for missing implementation.
- [ ] Implement `Registry`, `Notebook`, `Authorize`, `RequireOwned`, and
  `DefaultPath` under `%LOCALAPPDATA%/notebooklm-cli/owned-notebooks.json`.
- [ ] Run `go test ./registry` and verify pass.

No registry field may contain cookies, headers, account IDs, source text, chat
content, or artifact content.

### Task 4: Login and account capability workflows

**Files:**
- Create: `notebooklm-cli/notebooklm/account.go`
- Create: `notebooklm-cli/notebooklm/account_test.go`
- Create: `notebooklm-cli/notebooklm/testbridge_test.go`

- [ ] Write failing workflow tests that assert navigation to the home page,
  allowlisted login predicates, plan-label omission, CAPTCHA classification,
  and capability names without account identity fields.
- [ ] Run `go test ./notebooklm -run 'TestLogin|TestCapabilities'` and verify
  failure for missing workflows.
- [ ] Implement `LoginStatus` and `AccountCapabilities` with compact IIFE
  evaluate scripts that return only booleans, locale, plan label, and observed
  control names.
- [ ] Run the focused tests and verify pass.

### Task 5: Owned notebook, pasted source, chat, and Studio workflows

**Files:**
- Create: `notebooklm-cli/notebooklm/notebook.go`
- Create: `notebooklm-cli/notebooklm/notebook_test.go`
- Create: `notebooklm-cli/notebooklm/source.go`
- Create: `notebooklm-cli/notebooklm/source_test.go`
- Create: `notebooklm-cli/notebooklm/chat.go`
- Create: `notebooklm-cli/notebooklm/chat_test.go`
- Create: `notebooklm-cli/notebooklm/studio.go`
- Create: `notebooklm-cli/notebooklm/studio_test.go`

- [ ] Write failing notebook tests for create URL parsing, exact title update,
  source modal close, and direct reopen.
- [ ] Write failing source tests for `复制的文字`, CDP `key_type`, disabled
  insert protection, and ready-state polling.
- [ ] Write failing chat tests for visible Chat tab selection, non-empty input,
  nearest submit control, last-answer extraction, citation count, and timeout.
- [ ] Write failing Studio tests that return the exact observed output labels:
  audio, presentation, video, mind map, report, flashcards, quiz, infographic,
  and data table.
- [ ] Run `go test ./notebooklm` and verify failures for missing workflows.
- [ ] Implement the minimal workflows with `Page.bringToFront` before every
  interactive batch and post-action state verification.
- [ ] Run `go test ./notebooklm` and verify pass.

### Task 6: Cobra command surface

**Files:**
- Create: `notebooklm-cli/cmd/root.go`
- Create: `notebooklm-cli/cmd/root_test.go`
- Create: `notebooklm-cli/README.md`

- [ ] Write failing command tests for help, invalid arguments, ownership
  rejection, non-zero errors, and one JSON envelope per invocation.
- [ ] Run `go test ./cmd` and verify failure for missing commands.
- [ ] Implement:

```text
notebooklm-cli login-status
notebooklm-cli capabilities
notebooklm-cli notebook list
notebooklm-cli notebook authorize --url URL --title TITLE --confirm
notebooklm-cli notebook create --title TITLE
notebooklm-cli source add-text --notebook ID (--text TEXT | --file PATH)
notebooklm-cli chat ask --notebook ID --question QUESTION
notebooklm-cli studio capabilities --notebook ID
```

- [ ] Add persistent `--webbridge-url`, `--registry`, and `--timeout` flags.
- [ ] Defer `CloseSession` in every live command, including error paths.
- [ ] Run `go test ./cmd` and verify pass.

### Task 7: Static and live verification

**Files:**
- Create: `skills/notebooklm-cli/SKILL.md`
- Modify: `notebooklm-cli/README.md`

- [ ] Run `gofmt -w notebooklm-cli`.
- [ ] Run `go test ./...` and `go vet ./...` inside `notebooklm-cli`.
- [ ] Build only to `.codex/.tmp/notebooklm-cli/bin/notebooklm-cli.exe`.
- [ ] Verify every command's `--help`.
- [ ] Run live `login-status` and `capabilities` against port `10400`.
- [ ] Authorize only the CLI-created test notebook ID in the isolated registry.
- [ ] Run live `studio capabilities`, one pasted-text source into a separate
  CLI-created verification notebook, and one deterministic grounded chat.
- [ ] Close all WebBridge sessions and verify no session tabs remain.
- [ ] Scan tracked files for credentials, account identities, private notebook
  titles, and build artifacts.
- [ ] Commit only source, tests, README, skill, and sanitized documentation.
