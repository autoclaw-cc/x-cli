# deepseek-cli

`deepseek-cli` drives DeepSeek Web through the user's already signed-in,
isolated Chrome session via Kimi WebBridge. It does not automate login and
must not store cookies, authorization headers, account identity, or private
session data.

Current slice:

- `login-status`: verifies the visible DeepSeek page is authenticated and has a
  prompt box.
- `capabilities`: lists visible UI capabilities such as chat, DeepThink,
  search, and vision mode.
- `chat ask`: sends a prompt, waits until the newest assistant answer is stable,
  and can save that answer to a local text file. It can enable DeepThink and
  web search, and attach one or more local files or images.
- `chat new`: starts a fresh conversation from the visible DeepSeek page.

Use `--webbridge-url http://127.0.0.1:10086` with Kimi WebBridge running and
the browser extension connected to the isolated DeepSeek Chrome profile. On a
machine where port 10086 is reserved, start WebBridge on another port and pass
that address explicitly; the current Windows test machine uses port 10400.

```text
deepseek-cli --webbridge-url http://127.0.0.1:10086 login-status
deepseek-cli --webbridge-url http://127.0.0.1:10086 capabilities
deepseek-cli --webbridge-url http://127.0.0.1:10086 chat ask --prompt "Reply with exactly: lapis" --output .codex/tmp/deepseek-answer.txt
deepseek-cli --webbridge-url http://127.0.0.1:10400 chat new
deepseek-cli --webbridge-url http://127.0.0.1:10400 --timeout 5m chat ask --deepthink --search --file .codex/tmp/case.md --image .codex/tmp/chart.png --prompt "Analyze the attached material"
```

Successful commands print `{"ok":true,"data":...}`. Failures print
`{"ok":false,"error":{"code":"...","message":"..."}}` and exit non-zero.

## Safety boundary

- Login remains manual.
- CAPTCHA, 2FA, payment, account settings, and credentials are never automated.
- The WebBridge path inspects and operates on the visible page only. It never
  reads cookies, local storage, headers, or account identity, and it does not
  output or persist existing private chat titles.
- `chat ask` creates a normal DeepSeek chat interaction in the isolated account.
- `chat new` starts a new conversation. `--deepthink`, `--search`, `--file`, and
  `--image` operate only through visible controls and Kimi WebBridge upload.
- Attached files are read only from paths explicitly supplied on the command
  line. The CLI does not scan directories or retain file contents.
