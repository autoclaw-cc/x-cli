---
name: notebooklm-cli
description: Use when the user wants to create or operate explicitly authorized NotebookLM notebooks, add pasted-text sources, ask grounded questions, or inspect NotebookLM Studio capabilities through the signed-in browser session.
---

# NotebookLM CLI

Operate NotebookLM through Kimi WebBridge and the user's already signed-in
Chrome session. Never automate login or extract cookies, credentials, headers,
account identity, or unrequested notebook metadata.

## Prerequisites

1. Verify the WebBridge daemon and extension are connected.
2. Verify the CLI is available with `notebooklm-cli --help`.
3. Run `notebooklm-cli login-status`; ask the user to log in manually when it
   returns `logged_in: false`.
4. Set `KIMI_WEBBRIDGE_URL` or pass `--webbridge-url` when the daemon does not
   use its default port.

## Safety boundary

- Treat every pre-existing notebook as read-only unless the user explicitly
  authorizes its exact URL and title.
- Prefer `notebook create` for automation work. It registers the new notebook
  automatically.
- Write commands accept only notebook IDs present in the local ownership
  registry.
- Never enumerate or report unrelated notebook titles.
- Never bypass CAPTCHA, automate login, or export browser authentication data.

## Commands

| Command | Purpose |
|---|---|
| `login-status` | Report login state, locale, and visible CAPTCHA state |
| `capabilities` | Report visible account-level NotebookLM controls |
| `notebook list` | List locally authorized notebooks only |
| `notebook create --title TITLE` | Create, title, and register a notebook |
| `notebook authorize --url URL --title TITLE --confirm` | Deliberately register one exact notebook |
| `source add-text --notebook ID --text TEXT` | Add a pasted-text source |
| `source add-text --notebook ID --file PATH` | Add a local UTF-8 file as pasted text |
| `source add-url --notebook ID --url URL` | Add a public website or YouTube URL source |
| `chat ask --notebook ID --question QUESTION` | Return a grounded answer and citation count |
| `note create --notebook ID --title TITLE --text TEXT` | Create an editable note and verify it appears in Studio |
| `note create --notebook ID --title TITLE --file PATH` | Create an editable note from UTF-8 text |
| `note list --notebook ID` | List note titles after the Studio library stabilizes |
| `studio capabilities --notebook ID` | Discover visible Studio output types |
| `studio list --notebook ID` | List typed artifacts, state, details, playback, and menu availability |
| `studio generate --notebook ID --type TYPE [--prompt TEXT] [--wait started|ready]` | Generate a Studio artifact in an owned notebook |

Run `notebooklm-cli <command> --help` for flags. All commands return one JSON
envelope on stdout and use a non-zero exit status on error.

## Common workflow

```text
notebooklm-cli login-status
notebooklm-cli notebook create --title "CLI TEST - research"
notebooklm-cli source add-text --notebook ID --file case-pack.md
notebooklm-cli source add-url --notebook ID --url https://example.com/
notebooklm-cli chat ask --notebook ID --question "Summarize the evidence with citations."
notebooklm-cli note create --notebook ID --title "Evidence summary" --file summary.md
notebooklm-cli note list --notebook ID
notebooklm-cli studio capabilities --notebook ID
notebooklm-cli studio generate --notebook ID --type data_table --prompt "Build a compact comparison table."
notebooklm-cli studio generate --notebook ID --type mind_map --prompt "Map the key evidence."
notebooklm-cli studio generate --notebook ID --type report --wait ready
notebooklm-cli studio generate --notebook ID --type quiz --prompt "Create a short quiz." --wait ready
notebooklm-cli studio generate --notebook ID --type flashcards --prompt "Create a small flashcard set." --wait ready
notebooklm-cli studio generate --notebook ID --type presentation --prompt "Create slides." --wait ready
notebooklm-cli studio generate --notebook ID --type infographic --prompt "Create an infographic." --wait ready
notebooklm-cli studio generate --notebook ID --type audio --wait started
notebooklm-cli studio generate --notebook ID --type video --wait started
notebooklm-cli studio list --notebook ID
```

Use the exact `id` returned by `notebook create`. Do not substitute an ID found
by browsing the user's existing notebooks.

## Known limitations

- Login is manual.
- Direct local file upload is blocked by the current CDP boundary; `--file`
  pastes UTF-8 text instead.
- The Google Drive picker is cross-origin and is not automated.
- Browser-level download routing and direct media downloads require additional
  browser authorization and are not exposed by this CLI.
- `studio generate` supports all nine visible Studio type labels. Long media
  outputs should be run with a larger `--timeout`, or with `--wait started`
  followed by `studio list` polling.
