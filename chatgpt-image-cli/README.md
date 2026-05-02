# chatgpt-image-cli

Generate images on [chatgpt.com/images](https://chatgpt.com/images/) from the command line and save the PNG to a local directory — using your real, logged-in Chrome session (no API key).

Backed by the [kimi-webbridge](https://kimi.com/features/webbridge) browser daemon: the CLI drives your actual Chrome tab with proper cookies + Sentinel anti-abuse handling, so you stay on your existing ChatGPT plan.

```bash
$ chatgpt-image-cli generate "a cute panda riding a bicycle through a bamboo forest" -o ./images
{
  "data": {
    "prompt": "a cute panda riding a bicycle through a bamboo forest",
    "path": "/Users/you/images/chatgpt-20260422-174111.png",
    "bytes": 2228437,
    "caption": "已生成图片：森林中的快乐熊猫骑行",
    "conversation_url": "https://chatgpt.com/c/69e8977f-...",
    "elapsed_ms": 59970
  },
  "ok": true
}
```

## Prerequisites

1. **kimi-webbridge daemon running.**
   ```bash
   ~/.kimi-webbridge/bin/kimi-webbridge status
   # running: true, extension_connected: true
   ```
   If not installed, see <https://kimi.com/features/webbridge>.

2. **Signed in to chatgpt.com in the same Chrome.** Open <https://chatgpt.com/images/> manually once, log in, and leave the profile as-is. The CLI reuses that session.

## Install

### Pre-built binary (recommended)

This CLI ships from the [autoclaw-cc/x-cli](https://github.com/autoclaw-cc/x-cli) monorepo. Find the latest release tag (formatted `chatgpt-image-cli/v<version>`) at <https://github.com/autoclaw-cc/x-cli/releases?q=chatgpt-image-cli>, then:

```bash
# macOS arm64 (Apple Silicon) — swap the suffix for your platform:
TAG=chatgpt-image-cli/v0.1.0   # replace with the latest tag
curl -L -o chatgpt-image-cli \
  "https://github.com/autoclaw-cc/x-cli/releases/download/${TAG}/chatgpt-image-cli-darwin-arm64"
chmod +x chatgpt-image-cli
./chatgpt-image-cli --help
```

> **macOS users**: browser-downloaded binaries get blocked by Gatekeeper. Clear the quarantine flag once:
> ```bash
> xattr -d com.apple.quarantine ./chatgpt-image-cli
> ```

Release assets (per tag):

| Platform | Asset |
|---|---|
| macOS Intel | `chatgpt-image-cli-darwin-amd64` |
| macOS Apple Silicon | `chatgpt-image-cli-darwin-arm64` |
| Linux x86_64 | `chatgpt-image-cli-linux-amd64` |
| Windows x86_64 | `chatgpt-image-cli-windows-amd64.exe` |
| Checksums | `checksums.txt` (SHA-256) |

### From source

```bash
git clone https://github.com/autoclaw-cc/x-cli.git
cd x-cli/chatgpt-image-cli
go build -o chatgpt-image-cli .
./chatgpt-image-cli --help
```

## Usage

```bash
chatgpt-image-cli generate <prompt> [-o <dir>] [--timeout <sec>]
chatgpt-image-cli gen <prompt> [-o <dir>] [--timeout <sec>]    # short alias
```

| Flag | Default | Description |
|---|---|---|
| `-o, --out` | `.` | Output directory for the saved PNG |
| `--timeout` | `180` | Max seconds to wait for the image to appear |

Examples:

```bash
# Save to current directory
chatgpt-image-cli generate "a red apple on a wooden table"

# Save to a specific folder
chatgpt-image-cli generate "夕阳下的富士山" -o ~/Pictures/chatgpt

# Longer timeout for complex prompts
chatgpt-image-cli gen "detailed watercolor painting of a cyberpunk tokyo street at night" --timeout 300
```

Filename format: `chatgpt-YYYYMMDD-HHMMSS.png`.

## Output format

All output goes to stdout as JSON.

**Success:**
```json
{
  "ok": true,
  "data": {
    "prompt": "...",
    "path": "/abs/path/to/chatgpt-YYYYMMDD-HHMMSS.png",
    "bytes": 2228437,
    "caption": "...",
    "conversation_url": "https://chatgpt.com/c/<id>",
    "elapsed_ms": 59970
  }
}
```

**Error** (exit code 1):
```json
{
  "ok": false,
  "error": { "code": "<code>", "message": "..." }
}
```

Error codes:

| Code | Meaning |
|---|---|
| `invalid_args` | Missing or extra positional args |
| `daemon_unreachable` | kimi-webbridge daemon not reachable at `127.0.0.1:10086` |
| `daemon_not_running` | Daemon process is down |
| `extension_not_connected` | Chrome WebBridge extension is not connected |
| `generate_failed` | Anything during the generation flow (see message) |

## How it works

1. Navigate a tab to `https://chatgpt.com/images/`.
2. Inject the prompt into the ProseMirror textbox (`#prompt-textarea`) via `execCommand('insertText', ...)` + a manual `InputEvent` dispatch (needed to wake React's state so the send button enables).
3. Click `#composer-submit-button`.
4. Poll for the URL to become `/c/<conversation-id>` — the signal that the submission passed ChatGPT's Sentinel anti-abuse check.
5. Poll for an `<img>` in `<main>` whose `src` contains `/backend-api/estuary/content`.
6. `fetch()` the image inside the browser page context (credentials included, signed URL honored), base64-encode, return to Go.
7. Decode and write PNG bytes to disk.

The only state on disk is the PNG itself; the conversation stays in your ChatGPT history.

## Known limitations

- **Plan limits apply.** If your ChatGPT plan has hit its image-generation cap, the send button will stay disabled and the CLI will error out (`generate_failed: send button never became enabled`).
- **Content policy.** If ChatGPT refuses a prompt, the CLI detects common refusal phrases and fails with `chatgpt refused generation: ...` rather than timing out.
- **Chinese/English UI.** The CLI works regardless of ChatGPT's UI language — all selectors are stable IDs (`#prompt-textarea`, `#composer-submit-button`), not text content.
- **Temporary chats.** Each generation creates a normal chat in your ChatGPT history. If you want temporary chats, that feature isn't wired up here.

## License

MIT — see [LICENSE](./LICENSE).
