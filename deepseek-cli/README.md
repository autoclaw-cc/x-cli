# deepseek-cli

`deepseek-cli` drives DeepSeek Web through the user's already signed-in,
isolated Chrome session. It does not automate login and must not store cookies,
authorization headers, account identity, or private session data.

Current slice:

- `login-status`: verifies the visible DeepSeek page is authenticated and has a
  prompt box.
- `capabilities`: lists visible UI capabilities such as chat, DeepThink,
  search, and vision mode.
- `chat ask`: sends a prompt, waits until the newest assistant answer is stable,
  and can save that answer to a local text file.

Use `--cdp-url http://127.0.0.1:9223` with the isolated DeepSeek Chrome profile
launched with remote debugging enabled.

```text
deepseek-cli --cdp-url http://127.0.0.1:9223 login-status
deepseek-cli --cdp-url http://127.0.0.1:9223 capabilities
deepseek-cli --cdp-url http://127.0.0.1:9223 chat ask --prompt "Reply with exactly: lapis" --output .codex/tmp/deepseek-answer.txt
```

Successful commands print `{"ok":true,"data":...}`. Failures print
`{"ok":false,"error":{"code":"...","message":"..."}}` and exit non-zero.

## Safety boundary

- Login remains manual.
- CAPTCHA, 2FA, payment, account settings, and credentials are never automated.
- The CDP path inspects visible DOM only. It never reads cookies, local storage,
  headers, or account identity, and it does not output or persist existing
  private chat titles.
- `chat ask` creates a normal DeepSeek chat interaction in the isolated account.
- Future file, image, mode-selection, and conversation-management commands must
  be implemented only after live site archaeology identifies the exact visible
  UI boundary and a test covers the behavior.
