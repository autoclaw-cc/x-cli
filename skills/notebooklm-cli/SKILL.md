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
| `chat ask --notebook ID --question QUESTION` | Return a grounded answer and citation count |
| `studio capabilities --notebook ID` | Discover visible Studio output types |

Run `notebooklm-cli <command> --help` for flags. All commands return one JSON
envelope on stdout and use a non-zero exit status on error.

## Common workflow

```text
notebooklm-cli login-status
notebooklm-cli notebook create --title "CLI TEST - research"
notebooklm-cli source add-text --notebook ID --file case-pack.md
notebooklm-cli chat ask --notebook ID --question "Summarize the evidence with citations."
notebooklm-cli studio capabilities --notebook ID
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
