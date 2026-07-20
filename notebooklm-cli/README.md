# notebooklm-cli

`notebooklm-cli` drives an already signed-in NotebookLM browser session through
Kimi WebBridge. It never stores browser credentials and refuses write commands
for notebooks outside its local ownership registry.

## Commands

```text
notebooklm-cli login-status
notebooklm-cli capabilities
notebooklm-cli notebook list
notebooklm-cli notebook create --title "CLI TEST"
notebooklm-cli notebook authorize --url URL --title "CLI TEST" --confirm
notebooklm-cli source add-text --notebook ID --text "source text"
notebooklm-cli source add-text --notebook ID --file source.md
notebooklm-cli chat ask --notebook ID --question "What does the source say?"
notebooklm-cli studio capabilities --notebook ID
```

Set `KIMI_WEBBRIDGE_URL=http://127.0.0.1:10400` or pass
`--webbridge-url http://127.0.0.1:10400` when the daemon does not use the
default `10086` port.

`source add-text --file` reads a local UTF-8 text file and pastes its contents;
it does not use the currently blocked browser file-upload path.

## Known limits

Current Chrome/WebBridge restrictions block browser-level download routing,
direct local file upload, and automated control inside the cross-origin Google
Drive picker. The CLI reports unsupported paths instead of silently succeeding.
