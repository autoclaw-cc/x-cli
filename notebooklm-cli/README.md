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
notebooklm-cli note to-source --notebook ID --title "CLI NOTE"
notebooklm-cli research run --notebook ID --mode fast --query "Research query" --out artifacts/research.json
notebooklm-cli studio capabilities --notebook ID
notebooklm-cli studio list --notebook ID
notebooklm-cli studio generate --notebook ID --type data_table --prompt "Table request"
notebooklm-cli studio generate --notebook ID --type mind_map --wait started
notebooklm-cli studio wait --notebook ID --type video --timeout 20m --out artifacts/video-ready.json
notebooklm-cli studio export --notebook ID --type report --title "CLI Report" --out artifacts/report.json
notebooklm-cli studio inspect --notebook ID --type video --title "CLI Video" --out artifacts/video-attribution.json
notebooklm-cli studio rename --notebook ID --type data_table --title "Old" --new-title "New"
notebooklm-cli studio delete --notebook ID --type data_table --title "Disposable" --confirm
notebooklm-cli studio download --notebook ID --type video --title "CLI Video" --out artifacts/video.mp4
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
notebooklm-cli note to-source --notebook ID --title "CLI NOTE"
notebooklm-cli research run --notebook ID --mode fast --query "Reserved example domains" --out artifacts/research-fast.json
notebooklm-cli research run --notebook ID --mode deep --query "Deep comparison question" --out artifacts/research-deep.json
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
notebooklm-cli studio wait --notebook ID --type video --timeout 20m --out artifacts/notebooklm-video.json
notebooklm-cli studio export --notebook ID --type report --title "CLI Report" --out artifacts/notebooklm-report.json
notebooklm-cli studio inspect --notebook ID --type video --title "CLI Video" --out artifacts/video-attribution.json
notebooklm-cli studio download --notebook ID --type video --title "CLI Video" --out artifacts/video.mp4
notebooklm-cli studio download --notebook ID --type audio --title "CLI Audio" --out artifacts/audio.m4a
notebooklm-cli studio rename --notebook ID --type data_table --title "Disposable Table" --new-title "CLI DISPOSABLE"
notebooklm-cli studio delete --notebook ID --type data_table --title "CLI DISPOSABLE" --confirm
notebooklm-cli studio list --notebook ID
```

Live verification confirmed exact notebook creation and reopening, stable
text and URL source-count growth, grounded answers with citation counts,
editable note persistence in Studio, note-to-source conversion, typed Studio
artifact listing, Studio generation for all nine visible Studio types: `audio`,
`presentation`, `video`, `mind_map`, `report`, `flashcards`, `quiz`,
`infographic`, and `data_table`.
`studio generate --wait ready` waits for completion; the default `--wait started`
returns once a new artifact of the requested type is visible. Long media outputs
such as audio and video may require a larger `--timeout`. `studio wait --out`
polls until the requested type is ready and writes a local JSON evidence file
with the artifact title, details, state, playable flag, and observation time.
`studio export --out` opens one exact ready Studio artifact by type and title,
then writes visible artifact text plus metadata to a local JSON file. It is
intended for text artifacts such as reports, quizzes, flashcards, data tables,
mind maps, presentations, and infographics.
`research run --mode fast` submits a new source-discovery query and can write
local JSON evidence; live verification passed on a clean CLI-created notebook.
`research run --mode deep` selects Deep Research without re-clicking an already
open selected mode menu; live verification completed with 57 discovered sources.
`studio inspect --out` reads the artifact prompt and source attribution dialog.
`studio rename` and `studio delete --confirm` were live-verified against a
disposable CLI-created artifact.
`studio download --out` captures the signed media request made by NotebookLM's
player and writes audio/video bytes through bounded Range requests; live
verification downloaded a 46,745,559-byte `video/mp4` file and a
14,911,788-byte `audio/mp4` file.

## Known limits

Current Chrome/WebBridge restrictions block direct local file upload and
automated control inside the cross-origin Google Drive picker. Browser-level
download routing is also blocked, so `studio download` uses observed player
media requests rather than `Page.setDownloadBehavior`. `studio wait --out`
writes metadata evidence, not raw media bytes.
`studio export --out` exports visible text content and metadata; it does not
claim raw audio/video byte downloads. NotebookLM can keep one visible Research
result in Sources; if that result cannot be cleared through the current browser
surface, `research run` returns `research_result_present`.
The CLI reports unsupported paths instead of silently succeeding.
