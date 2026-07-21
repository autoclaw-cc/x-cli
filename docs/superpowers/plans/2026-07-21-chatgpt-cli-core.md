# ChatGPT CLI Core Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build, verify, integrate, and publish the three WebBridge CLIs, including the missing full `chatgpt-cli`.

**Architecture:** Add a self-contained Go module whose browser boundary is Kimi WebBridge and whose provider layer exposes allowlisted page state, chat/research orchestration, and image download. Integrate the three completed CLI branches only after each module is clean, then extend the existing per-CLI release workflow.

**Tech Stack:** Go 1.23, Cobra, `net/http`, Kimi WebBridge DOM actions, GitHub Actions, GitHub Releases.

---

### Task 1: Freeze ChatGPT Archaeology

**Files:**
- Create: `chatgpt-cli/ARCHAEOLOGY.md`

- [ ] Record the authenticated composer, send, mode-selection, assistant-result,
  citation, image-result, and completion signals without recording history or
  account identity.
- [ ] Verify each selector with an allowlisted WebBridge `evaluate` call.
- [ ] Record provider limitations and exact live smoke prompts.

### Task 2: Browser And Output Foundations

**Files:**
- Create: `chatgpt-cli/go.mod`
- Create: `chatgpt-cli/main.go`
- Create: `chatgpt-cli/browser/client.go`
- Create: `chatgpt-cli/browser/client_test.go`
- Create: `chatgpt-cli/output/output.go`
- Create: `chatgpt-cli/output/output_test.go`

- [ ] Write failing tests for daemon URL precedence, command envelopes,
  `find_tab`, `navigate`, `fill`, `click`, `evaluate`, and `close_session`.
- [ ] Run `go test ./browser ./output` and confirm RED for missing behavior.
- [ ] Implement the smallest WebBridge client and JSON writer that pass.
- [ ] Run `go test ./browser ./output` and confirm GREEN.

### Task 3: Login And Capability Inspection

**Files:**
- Create: `chatgpt-cli/chatgpt/page.go`
- Create: `chatgpt-cli/chatgpt/page_test.go`
- Create: `chatgpt-cli/chatgpt/webbridge.go`
- Create: `chatgpt-cli/cmd/root.go`
- Create: `chatgpt-cli/cmd/root_test.go`

- [ ] Write failing tests for logged-in, logged-out, redacted, and capability
  discovery states.
- [ ] Confirm RED with `go test ./chatgpt ./cmd -run 'Login|Capabil'`.
- [ ] Implement `login-status` and `capabilities` using only allowlisted DOM
  fields.
- [ ] Confirm GREEN and verify both command help pages.

### Task 4: New Chat And Ask Workflows

**Files:**
- Modify: `chatgpt-cli/chatgpt/webbridge.go`
- Modify: `chatgpt-cli/cmd/root.go`
- Modify: `chatgpt-cli/cmd/root_test.go`

- [ ] Write failing fake-WebBridge tests for `chat new`, ordinary ask, web
  search, Deep Research, mutually exclusive modes, one-submit behavior,
  streaming completion, timeout, and citations.
- [ ] Confirm RED with `go test ./cmd -run 'Chat|Research|Search'`.
- [ ] Implement the semantic DOM scripts and bounded polling loop.
- [ ] Confirm GREEN and run one live smoke prompt in each supported mode.

### Task 5: Image Generation And Local Persistence

**Files:**
- Create: `chatgpt-cli/chatgpt/image.go`
- Create: `chatgpt-cli/chatgpt/image_test.go`
- Modify: `chatgpt-cli/cmd/root.go`
- Modify: `chatgpt-cli/cmd/root_test.go`

- [ ] Write failing tests for confirmation, output containment, baseline image
  filtering, content-type validation, timeout, and a single submit.
- [ ] Confirm RED with `go test ./chatgpt ./cmd -run 'Image|Confirm'`.
- [ ] Adapt the proven estuary-image workflow from `chatgpt-image-cli` behind
  the new browser interface.
- [ ] Confirm GREEN and save one live test image under the worktree `.codex`
  evidence directory.

### Task 6: User Documentation And Companion Skill

**Files:**
- Create: `chatgpt-cli/README.md`
- Create: `skills/chatgpt-cli/SKILL.md`
- Modify: `README.md`
- Modify: `README_EN.md`

- [ ] Document installation, commands, JSON schemas, profile isolation, daemon
  selection, timeouts, and billing guardrails.
- [ ] Verify examples with the built binary and scan docs for secrets or private
  titles.

### Task 7: Module Verification

**Files:**
- Modify only files required by defects reproduced with a failing test.

- [ ] Run `gofmt -w chatgpt-cli`.
- [ ] Run `go test ./...` and `go vet ./...` in `chatgpt-cli`.
- [ ] Run the same gates in `notebooklm-cli` and `deepseek-cli`.
- [ ] Cross-build all three modules for Windows, Linux, and macOS amd64/arm64.
- [ ] Scan tracked diffs and build outputs for cookies, tokens, account identity,
  private titles, and generated binaries.

### Task 8: Integrate And Publish

**Files:**
- Modify: `.github/workflows/release.yml`
- Modify: `README.md`
- Modify: `README_EN.md`

- [ ] Merge the three isolated feature branches into one `codex/` publication
  branch and audit the exact staged manifest.
- [ ] Add `chatgpt-cli`, `notebooklm-cli`, and `deepseek-cli` tag triggers and
  workflow-dispatch choices.
- [ ] Re-run all module tests, vet, cross-build, YAML parse, and `git diff --check`.
- [ ] Push the publication branch, create the GitHub pull request, merge after
  checks, then create and push `*/v0.1.0` tags.
- [ ] Watch all release workflows to completion and verify six archives plus
  `checksums.txt` on each public release.
