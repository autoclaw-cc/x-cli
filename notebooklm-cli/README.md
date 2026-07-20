# notebooklm-cli

`notebooklm-cli` drives an already signed-in NotebookLM browser session through
Kimi WebBridge. It never stores browser credentials and refuses write commands
for notebooks outside its local ownership registry.

## Build

```powershell
cd notebooklm-cli
go build -o notebooklm-cli.exe .
```

The browser must already be signed in. The CLI does not automate login, export
cookies, or persist account identity.

## Commands

```text
notebooklm-cli login-status
notebooklm-cli capabilities
notebooklm-cli notebook list
notebooklm-cli notebook create --title "CLI TEST"
notebooklm-cli notebook authorize --url URL --title "CLI TEST" --confirm
notebooklm-cli source add-text --notebook ID --text "source text"
notebooklm-cli source add-text --notebook ID --file source.md
notebooklm-cli source add-url --notebook ID --url https://example.com/
notebooklm-cli chat ask --notebook ID --question "What does the source say?"
notebooklm-cli note create --notebook ID --title "CLI NOTE" --text "note body"
notebooklm-cli note create --notebook ID --title "CLI NOTE" --file note.md
notebooklm-cli note list --notebook ID
notebooklm-cli studio capabilities --notebook ID
notebooklm-cli studio list --notebook ID
```

Every command prints one JSON envelope. Successful commands use
`{"ok":true,"data":...}`; failures use
`{"ok":false,"error":{"code":"...","message":"..."}}` and a non-zero exit
status.

Set `KIMI_WEBBRIDGE_URL=http://127.0.0.1:10400` or pass
`--webbridge-url http://127.0.0.1:10400` when the daemon does not use the
default `10086` port.

`source add-text --file` reads a local UTF-8 text file and pastes its contents;
it does not use the currently blocked browser file-upload path.

## Ownership boundary

Write commands resolve `--notebook` only through the local ownership registry.
`notebook create` registers the notebook it just created. To allow an existing
notebook deliberately, authorize its exact URL and title:

```text
notebooklm-cli notebook authorize --url URL --title TITLE --confirm
```

Authorization records only the notebook ID, URL, title, and authorization time.
It never records sources, answers, cookies, request headers, or account IDs.

## Verified workflow

```text
notebooklm-cli login-status
notebooklm-cli notebook create --title "CLI TEST - isolated verification"
notebooklm-cli source add-text --notebook ID --text "The checkpoint is 7319."
notebooklm-cli source add-url --notebook ID --url https://example.com/
notebooklm-cli chat ask --notebook ID --question "What is the checkpoint?"
notebooklm-cli note create --notebook ID --title "CLI NOTE" --text "Verified notes persist."
notebooklm-cli note list --notebook ID
notebooklm-cli studio capabilities --notebook ID
notebooklm-cli studio list --notebook ID
```

Live verification confirmed exact notebook creation and reopening, stable
text and URL source-count growth, grounded answers with citation counts,
editable note persistence after reopening, typed Studio artifact listing, and
discovery of `audio`, `presentation`, `video`, `mind_map`, `report`,
`flashcards`, `quiz`, `infographic`, and `data_table` Studio controls.

## Known limits

Current Chrome/WebBridge restrictions block browser-level download routing,
direct local file upload, and automated control inside the cross-origin Google
Drive picker. The CLI reports unsupported paths instead of silently succeeding.
