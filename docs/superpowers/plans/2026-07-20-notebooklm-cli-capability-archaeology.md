# NotebookLM CLI Capability Archaeology Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Produce a sanitized, account-verified capability profile and reproducible WebBridge interaction record for a full NotebookLM CLI without touching pre-existing notebooks.

**Architecture:** Use one dedicated Kimi WebBridge session at `http://127.0.0.1:10400` and one CLI-owned notebook named `CLI TEST - notebooklm-cli - 20260720-01`. Every live operation is recorded as an allowlisted capability result, selector/RPC family, state transition, and output type; private account fields and existing notebook data are discarded.

**Tech Stack:** Kimi WebBridge 1.11.3 extension, local daemon 1.10.3, PowerShell 7, Chrome, NotebookLM web UI, Markdown and JSON evidence.

---

### Task 1: Prepare the sanitized case pack

**Files:**
- Create: `.codex/.tmp/notebooklm-cli/case-pack/source.txt`
- Create: `.codex/.tmp/notebooklm-cli/case-pack/source.md`
- Create: `.codex/.tmp/notebooklm-cli/capability-profile.json`

- [ ] **Step 1: Create deterministic non-secret source text**

Use this exact content in both text fixtures:

```text
CLI TEST CASE PACK
Alpha evidence: daily walking is the case pack's preferred low-intensity activity.
Beta evidence: the synthetic participant completed 42 minutes on 2026-07-20.
Gamma evidence: hydration was recorded before and after the activity.
The verification phrase is ORCHID-7421.
```

- [ ] **Step 2: Initialize the capability profile**

```json
{
  "schema_version": 1,
  "account_alias": "notebooklm-pro-01",
  "notebook_name": "CLI TEST - notebooklm-cli - 20260720-01",
  "existing_notebooks_touched": 0,
  "capabilities": []
}
```

- [ ] **Step 3: Verify fixture secrecy and encoding**

Run:

```powershell
rg -n -i "cookie|authorization|token|gmail|account_id" .codex/.tmp/notebooklm-cli/case-pack
Get-Content .codex/.tmp/notebooklm-cli/case-pack/source.txt -Encoding utf8
```

Expected: the secret scan returns no matches and the verification phrase is `ORCHID-7421`.

### Task 2: Verify account and capability surfaces

**Files:**
- Create: `.codex/.tmp/notebooklm-cli/evidence/account.md`
- Modify: `.codex/.tmp/notebooklm-cli/capability-profile.json`

- [ ] **Step 1: Verify WebBridge health**

Run:

```powershell
Invoke-RestMethod http://127.0.0.1:10400/status | Select-Object running,extension_connected,version,extension_version
```

Expected: `running=true`, `extension_connected=true`, extension `1.11.3`.

- [ ] **Step 2: Navigate with a dedicated session**

Use session `notebooklm-cli-archaeology-20260720` and open `https://notebooklm.google.com/` in a new tab group named `NotebookLM CLI TEST`.

- [ ] **Step 3: Record an allowlisted login result**

Record only `logged_in`, visible plan label, locale, and capability control names. Do not record account names, emails, avatars, account IDs, notebook titles, or raw snapshot trees.

- [ ] **Step 4: Verify the not-logged-in detector design**

Freeze DOM predicates for sign-in controls, redirect URLs, CAPTCHA text, and the authenticated `New notebook` control without logging out or changing the live account.

### Task 3: Exercise the CLI-owned notebook lifecycle

**Files:**
- Create: `.codex/.tmp/notebooklm-cli/evidence/notebook.md`
- Modify: `.codex/.tmp/notebooklm-cli/capability-profile.json`

- [ ] **Step 1: Create the owned notebook**

Create exactly `CLI TEST - notebooklm-cli - 20260720-01` through the visible new-notebook control. Record its URL-derived notebook ID locally and mark it `owned=true`.

- [ ] **Step 2: Verify direct reopen**

Close the WebBridge session, create a fresh session, navigate directly to the recorded notebook URL, and verify the live ID and title.

- [ ] **Step 3: Exercise rename and restore**

Rename the test notebook to `CLI TEST - notebooklm-cli - 20260720-01-renamed`, verify persistence after a fresh reopen, then restore the original title.

- [ ] **Step 4: Probe duplicate and delete controls safely**

If duplicate is visible, duplicate only the owned notebook and immediately delete only the duplicate. Record unavailable controls as `observed=false`; never fall back to another notebook.

### Task 4: Exercise source ingestion and research

**Files:**
- Create: `.codex/.tmp/notebooklm-cli/evidence/sources.md`
- Modify: `.codex/.tmp/notebooklm-cli/capability-profile.json`

- [ ] **Step 1: Add pasted text**

Add `source.txt` as pasted text, wait for ready state, and verify a source summary can recover `ORCHID-7421`.

- [ ] **Step 2: Add local files**

Probe and, when accepted, upload the owned `source.txt` and `source.md` fixtures individually. Record supported extensions, duplicate behavior, processing states, and removal semantics.

- [ ] **Step 3: Add a public webpage**

Submit `https://en.wikipedia.org/wiki/NotebookLM`, wait for processing, and verify the source title and ready/error state without copying article content into evidence.

- [ ] **Step 4: Probe YouTube and Drive paths**

Record the available YouTube and Google Drive controls, required fields, pickers, and cancellation behavior. Import only an explicitly non-private public video or a CLI-created Drive fixture; otherwise record the path as UI-observed but not imported.

- [ ] **Step 5: Exercise source discovery/research**

Use the query `official NotebookLM source grounding documentation`. Record whether the account exposes source discovery, Fast Research, or Deep Research, their parameters, result review step, and import behavior. Import at most two public results into the owned notebook.

- [ ] **Step 6: Verify source selection and removal**

Toggle individual owned sources for chat grounding, remove one disposable duplicate, and verify the remaining source count and IDs.

### Task 5: Exercise grounded chat and notes

**Files:**
- Create: `.codex/.tmp/notebooklm-cli/evidence/chat-notes.md`
- Modify: `.codex/.tmp/notebooklm-cli/capability-profile.json`

- [ ] **Step 1: Ask a deterministic grounded question**

Ask `What is the verification phrase, how many minutes were recorded, and which activity was preferred? Cite the supporting source.` Expected facts: `ORCHID-7421`, `42`, and daily walking.

- [ ] **Step 2: Verify citation navigation**

Record citation count, source linkage, quoted-span availability, and whether clicking a citation opens the correct owned source. Do not persist generated prose beyond short field-level assertions.

- [ ] **Step 3: Verify multi-turn context**

Ask `Was hydration recorded before the activity, after it, or both?` and verify the answer `both` without repeating the first prompt.

- [ ] **Step 4: Save and manage an owned note**

Save the first answer to a note, rename it `CLI TEST grounded answer`, edit it with the suffix `Verified by notebooklm-cli archaeology.`, list it, and probe note-to-source conversion if exposed.

- [ ] **Step 5: Verify safe chat failures**

Exercise an empty prompt, a prompt with all sources deselected, and a stale notebook ID. Record stable error classifications for the future CLI.

### Task 6: Exercise every visible Studio output

**Files:**
- Create: `.codex/.tmp/notebooklm-cli/evidence/studio.md`
- Create: `.codex/.tmp/notebooklm-cli/downloads/.gitkeep`
- Modify: `.codex/.tmp/notebooklm-cli/capability-profile.json`

- [ ] **Step 1: Enumerate output types and controls**

Record every visible Studio output type and its actual customization controls. Candidate types are audio overview, video overview, mind map, report, FAQ, timeline, study guide, flashcards, quiz, infographic, and slide deck; do not mark a candidate supported unless it is visible.

- [ ] **Step 2: Generate fast structured outputs**

Create one mind map, one report/study guide, one flashcard set, and one quiz when available. Record create parameters, progress states, completion markers, inspect/open behavior, and regeneration controls.

- [ ] **Step 3: Generate media outputs**

Create one shortest/default audio overview, video overview, infographic, and slide deck when available. Use Chinese only when the control exposes it; otherwise keep the account default. Poll without duplicate submissions and record elapsed state transitions.

- [ ] **Step 4: Download supported outputs**

Download one completed artifact of each downloadable type to `.codex/.tmp/notebooklm-cli/downloads`, verify non-zero size and MIME/extension agreement, and record only filenames, sizes, hashes, and output types.

- [ ] **Step 5: Verify generation failure handling**

Record behavior for an unavailable output type, duplicate generation request, navigation away during generation, timeout, and visible quota exhaustion without purchasing credits or changing plan settings.

### Task 7: Freeze the CLI profile and close all sessions

**Files:**
- Create: `docs/NOTEBOOKLM_CLI_CAPABILITY_PROFILE.md`
- Modify: `.codex/.tmp/notebooklm-cli/capability-profile.json`

- [ ] **Step 1: Normalize capability records**

Each capability entry must contain `name`, `observed`, `verified`, `side_effect`, `selector_or_rpc_family`, `inputs`, `result_shape`, `failure_codes`, and `notes` with no secrets.

- [ ] **Step 2: Run privacy and ownership scans**

Run:

```powershell
rg -n -i "cookie|authorization|bearer|gmail|account_id|xsrf|csrf" .codex/.tmp/notebooklm-cli docs/NOTEBOOKLM_CLI_CAPABILITY_PROFILE.md
```

Expected: no credential or account-identity material. Verify `existing_notebooks_touched` remains `0`.

- [ ] **Step 3: Close the archaeology session**

Call `close_session` for `notebooklm-cli-archaeology-20260720`, then call `list_tabs` and verify zero tabs remain for that session.

- [ ] **Step 4: Freeze implementation slices**

Write separate implementation plans for account/ownership, source/chat/note, and Studio commands. Each plan must use TDD, ship stable JSON envelopes, and include immediate live verification per command.

- [ ] **Step 5: Commit sanitized durable documentation only**

```powershell
git add -- docs/NOTEBOOKLM_CLI_CAPABILITY_PROFILE.md docs/superpowers/plans
git diff --cached --check
git -c user.name=Codex -c user.email=codex@example.com commit -m "docs: map notebooklm cli capabilities"
```

Expected: test artifacts under `.codex/.tmp` and downloaded media are not staged.
