# NotebookLM CLI Capability Profile

Verified on 2026-07-20 against the dedicated account alias
`notebooklm-pro-01` through Kimi WebBridge at `http://127.0.0.1:10400`.
This document contains no account identity, credentials, private headers,
existing notebook titles, or private source content.

## Safety Boundary

- Existing notebooks touched: **0**.
- All write tests used one CLI-created notebook named
  `CLI TEST - notebooklm-cli - 20260720-01`.
- Public sharing, deletion of existing data, purchases, added credits, login
  automation, and CAPTCHA automation were not performed.

## Account

| Capability | Result |
|---|---|
| Authenticated home | Verified |
| Locale | `zh-CN` |
| Visible plan label on home | Not exposed |
| New notebook control | Verified |
| CAPTCHA | Not present |

The login check can be implemented with allowlisted DOM predicates; no account
profile endpoint or identity field is required.

## Notebook Lifecycle

- Create through the `新建笔记本` control: verified.
- Direct URL contains a stable `/notebook/<id>` identifier.
- Rename through `input.title-input`: verified with native value setter plus
  `input`/`change` events.
- Close session and reopen directly by URL: title, sources, and chat persisted.
- Duplicate/delete controls were not exercised because the test notebook holds
  the generated capability evidence.

## Sources And Research

| Path | Result | Notes |
|---|---|---|
| Pasted text | Verified | `key_type` enables the insert control; 289-byte deterministic fixture imported. |
| Public webpage | Verified | URL textarea plus insert; processing completed. |
| Local file UI | Observed, automation blocked | Accepts PDF, TXT, MD, DOCX, CSV, PPTX, EPUB, many audio/video/image extensions. WebBridge `upload` returned CDP `Not allowed`. |
| Google Drive | UI verified | Opens `docs.google.com/picker/v2/home` in a cross-origin iframe; picker contents are outside top-frame WebBridge control. |
| YouTube | UI observed | Shares the website source path; no public video was imported in this pass. |
| Fast Research | CLI implemented and live-verified | Clean CLI-created notebook returned completed JSON evidence; import also verified earlier with source count growth. |
| Deep Research | CLI implemented and live-verified | Clean CLI-created notebook completed with 57 discovered sources after fixing selected-menu handling. |
| Source selection | Verified | Native checkbox state changes asynchronously after DOM click; state must be re-read before chat. |
| Note to source | Verified | Saved-answer note converted to a selectable source. |

The write RPC family is `POST /_/LabsTailwindUi/data/batchexecute` with status
200. The CLI should use stable visible controls first and treat the RPC family
as diagnostic evidence, not persist opaque request payloads.

## Chat And Notes

- Grounded single-turn answer verified exact facts `ORCHID-7421`, `42`, daily
  walking, and a citation control.
- Multi-turn follow-up verified `both` for hydration before and after activity.
- Saving an answer creates a read-only note and switches the visible panel to
  Studio.
- Saved-answer notes can be renamed and converted to a source.
- A blank note created from `添加笔记` uses ProseMirror. WebBridge `fill` works
  for its content; the note auto-saves and reopens with exact title/body.
- Direct keystroke replacement can leave composition suffixes in note titles;
  use the native `HTMLInputElement.value` setter and dispatch `input` and
  `change`, then verify persistence.

## Studio

Nine output types were visible and account-verified:

1. Audio Overview
2. Presentation (Beta)
3. Video Overview
4. Mind Map
5. Report
6. Flashcards
7. Quiz
8. Infographic (Beta)
9. Data Table

Representative live outputs all completed from one selected test source:

| Type | Verified controls | Approximate completion |
|---|---|---:|
| Mind Map | source, topic | under 30 s |
| Report | custom, briefing, study guide, blog | under 30 s |
| Flashcards | fewer/standard/more, easy/medium/hard, source, topic | 50 s |
| Quiz | fewer/standard/more, easy/medium/hard, source, topic | 30 s |
| Data Table | language, source, requested columns/shape | under 15 s |
| Presentation | detailed/presentation slides, language, short/default, source, prompt | 196 s |
| Infographic | language, orientation, 11 visual styles, brief/standard/detailed, source, prompt | 191 s |
| Audio | deep dive/summary/critique/debate, language, optional length, source, host focus | 221-276 s |
| Video | explainer/summary, language, source, 10 visual styles, host focus | 546 s |

For Angular Material tile radios, clicking the hidden `<input>` is not enough.
The CLI must click/focus the visible radio container, re-read checked state, and
fail before generation when the requested value is not selected.

Every inspected artifact menu exposed share, rename, download, view prompt and
sources, and delete. Playback/open flows were verified for audio and video.
The CLI now exports ready artifact visible text plus metadata with
`studio export --out`; live verification covered a generated report and wrote a
local JSON file without using browser-level download routing. Prompt/source
inspection is implemented with `studio inspect --out` and was live-verified on a
ready video artifact. Artifact rename and delete are implemented and were
live-verified on a disposable CLI-created data-table artifact. Raw media
download is implemented with `studio download --out`: it captures the signed
player media request through WebBridge network observation and writes bounded
Range chunks locally. Live video verification produced a 46,745,559-byte
`video/mp4` file with an MP4 `ftyp` header. Live audio verification produced a
14,911,788-byte `audio/mp4` file with an MP4 `ftyp` header.

## Current Automation Limits

- Browser-level CDP download control is blocked with
  `Cannot access browser-level commands`; the CLI therefore does not rely on
  `Page.setDownloadBehavior`.
- Media player entry URLs are served from
  `lh3.googleusercontent.com/notebooklm/...`. In-page fetch is blocked by CORS
  and unauthenticated external GET returns HTML; the working path is the
  observed signed `googlevideo.com/videoplayback` request emitted by the player.
- File upload is visible but `DOM.setFileInputFiles` returns `Not allowed` in
  the current Chrome/WebBridge pairing.
- Drive picker automation is blocked by the top-frame-only cross-origin iframe
  boundary.

The CLI reports these as unavailable rather than pretending the operation
succeeded. Pasted text, public URL, grounded chat, editable note creation/list,
note-to-source conversion, Fast Research, Studio capability discovery, typed
artifact listing, Studio generation, slow-output ready polling, visible text
export, prompt/source inspection, raw audio/video media download, artifact rename,
and artifact delete are implemented. Live generation verification covers all
nine visible Studio types:
`audio`, `presentation`, `video`, `mind_map`, `report`, `flashcards`, `quiz`,
`infographic`, and `data_table`; long media outputs require larger timeouts or
`--wait started` plus `studio wait --out` / `studio list` polling.
`studio wait --out` persists local JSON evidence for the ready artifact, and
`studio export --out` persists visible text content for ready text artifacts.
`studio download --out` persists ready audio/video bytes when the player emits a
signed media request; both video and audio are live-verified under the active
provider/browser state.
