# nanobanana-cli

Generate images via Google Gemini's **Nano Banana** (Gemini 2.5 Flash Image) model from the command line and save a full-size PNG plus a locally-scaled thumbnail.

The CLI drives your **real Chrome session** through [`kimi-webbridge`](https://www.kimi.com/features/webbridge), so it reuses your existing Gemini login — no API key, no OAuth setup, no separate Gemini for Developers quota.

## What it does

```bash
nanobanana-cli gen "画一朵粉色月季花，微距特写" -o ./out
```

```json
{
  "ok": true,
  "data": {
    "prompt": "画一朵粉色月季花，微距特写",
    "full":  "/abs/path/out/20260422-143045-full.png",
    "thumb": "/abs/path/out/20260422-143045-thumb.png",
    "width":  2816,
    "height": 1536,
    "thumb_width": 256,
    "elapsed_ms": 184230
  }
}
```

Each run saves two files into `-o <dir>`:

- `<timestamp>-full.png` — the **real** high-resolution original (currently 2816×1536 for Nano Banana; ~5–7 MB), intercepted from Gemini's "Download full-size" response chain
- `<timestamp>-thumb.png` — PNG scaled to `--thumb-width` px (aspect preserved, default 256)

## Requirements

- **macOS** (Linux/Windows likely work but untested)
- **Kimi Desktop App** running — bundles the `kimi-webbridge` daemon on `http://127.0.0.1:10086`. Install: <https://www.kimi.com/features/webbridge>
- **Chrome** with the WebBridge extension installed and connected (status check: `curl http://127.0.0.1:10086/status` should report `extension_connected: true`)
- **Gemini logged in** in that Chrome — the CLI reuses your cookies via the real browser
- **Go 1.22+** to build

## Install

### Pre-built binary (recommended)

This CLI ships from the [autoclaw-cc/x-cli](https://github.com/autoclaw-cc/x-cli) monorepo. Find the latest release tag (formatted `nanobanana-cli/v<version>`) at <https://github.com/autoclaw-cc/x-cli/releases?q=nanobanana-cli>, then:

```bash
# macOS arm64 (Apple Silicon) — swap the suffix for your platform:
TAG=nanobanana-cli/v0.1.0   # replace with the latest tag
curl -LO "https://github.com/autoclaw-cc/x-cli/releases/download/${TAG}/nanobanana-cli-darwin-arm64.tar.gz"
tar -xzf nanobanana-cli-darwin-arm64.tar.gz
./nanobanana-cli --help
```

> **macOS users**: after extracting, clear the Gatekeeper quarantine flag once:
> ```bash
> xattr -d com.apple.quarantine ./nanobanana-cli
> ```

Available archives per release: `nanobanana-cli-{darwin-arm64,darwin-amd64,linux-amd64,linux-arm64}.tar.gz`, `nanobanana-cli-{windows-amd64,windows-arm64}.zip`, plus `checksums.txt`.

### Build from source

```bash
git clone https://github.com/autoclaw-cc/x-cli.git
cd x-cli/nanobanana-cli
go build -o nanobanana-cli .
```

## Usage

```
nanobanana-cli gen <prompt> [flags]

Flags:
  -o, --out string        output directory (default ".")
      --thumb-width int   thumbnail width in px (default 256)
      --timeout int       max seconds to wait for image generation (default 300)
```

Output is **always JSON** on stdout. Non-zero exit code on error. Error shape:

```json
{ "ok": false, "error": { "code": "...", "message": "..." } }
```

Common error codes: `daemon_unreachable`, `daemon_not_running`, `extension_not_connected`, `invalid_args`, `gen_failed`.

## How it works

```
user prompt
    │
    ▼
POST :10086/command  ─────▶  Chrome extension  ─────▶  gemini.google.com
    navigate                                           (your real session)
    evaluate(inject prompt via execCommand)
    evaluate(click button.send-button)
    evaluate(poll .generated-image img.loaded)
    evaluate(install step-3 fetch hook)
    evaluate(click [data-test-id="download-generated-image-button"])
    evaluate(poll window.__nbFinalURL)
    evaluate(fetch final URL → base64 encode)
    │                                                         │
    │ ◀─── base64 PNG (2816×1536, ~5–7 MB)  ◀──────────────────┘
    ▼
Go: decode PNG → write *-full.png → resize (Catmull-Rom) → write *-thumb.png
```

**The `<img>` in the chat is NOT the original.** Gemini renders a 1024×559 display-sized copy inline — fine for viewing, useless for saving. The real original only becomes accessible when you click "Download full-size", which kicks off a 4-hop URL chain:

```
POST c8o8Fe                                 → JSON; body has a signed gg-dl URL
GET  lh3.../gg-dl/...?alr=yes               → text/plain; body = fife URL
GET  work.fife.usercontent.google.com/...   → text/plain; body = final lh3 URL  ← step 3
GET  lh3.../rd-gg-dl/...                    → image/png  (Chrome downloads)     ← step 4
```

Letting step 4 complete normally pops up Chrome's "Save As" dialog — nasty for a CLI. Instead, the CLI installs a `window.fetch` hook before clicking download: when step 3 fires, the hook captures the final URL out of the response body into a window variable, then returns an empty `Response` so Gemini's own code has no URL to navigate to. Chrome never sees a Content-Disposition load, no dialog. The CLI then runs its own `fetch(window.__nbFinalURL)` from evaluate — `fetch()` is JS-initiated, stays in the renderer, no download manager. We base64-encode the bytes and ship them back to Go.

**Why generate the thumbnail locally instead of asking Gemini?** There is no separate thumbnail resource — the chat UI just CSS-scales the 1024×559 display copy. Scaling from the original via `golang.org/x/image/draw.CatmullRom` is deterministic, offline, and `--thumb-width` lets the caller pick any size.

## Project layout

```
nanobanana-cli/
├── main.go                   entry point
├── browser/client.go         HTTP client for kimi-webbridge daemon
├── output/output.go          structured JSON output helper
├── nanobanana/gen.go         the one feature: prompt → full + thumb
└── cmd/root.go               cobra command registration
```

## Troubleshooting

| Symptom | Likely cause |
|---------|--------------|
| `daemon_unreachable` | Kimi Desktop App not running. Open it. |
| `extension_not_connected` | Chrome WebBridge extension not installed/enabled. See <https://www.kimi.com/features/webbridge>. |
| `timeout waiting for generated image` | Gemini routed your prompt to text response, not image. Rephrase to be clearly an image-gen request (e.g., start with `画` / `generate an image of`). |
| `prompt is empty` | `gen ""` — pass a non-empty prompt. |

## License

MIT (see `LICENSE`).
