# deepseek-cli

`deepseek-cli` drives DeepSeek Web through Kimi WebBridge and the user's
already signed-in Chrome session. It does not automate login and must not store
cookies, authorization headers, account identity, or private session data.

Current slice:

- `login-status`: checks that the Kimi WebBridge daemon and Chrome extension
  are connected before DeepSeek account-specific archaeology starts.

Use `--webbridge-url http://127.0.0.1:10400` when the daemon is running on the
local desktop port used by this project.

```text
deepseek-cli --webbridge-url http://127.0.0.1:10400 login-status
```

Successful commands print `{"ok":true,"data":...}`. Failures print
`{"ok":false,"error":{"code":"...","message":"..."}}` and exit non-zero.

## Safety boundary

- Login remains manual.
- CAPTCHA, 2FA, payment, account settings, and credentials are never automated.
- Future write commands must be implemented only after live site archaeology
  identifies the exact visible UI/API boundary and a test covers the behavior.
