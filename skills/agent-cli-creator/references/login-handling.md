# Login Handling

## Recommended Pattern: Manual Login + Status Check

**Do not automate login forms.**

Login flows involve CAPTCHA, 2FA, security fingerprinting, and anti-bot detection. They break constantly and are highly site-specific. The kimi-webbridge daemon keeps browser tabs alive between CLI calls — a user logs in once manually in Chrome; every subsequent CLI call reuses the authenticated session.

## Implementation

### 1. Add `login-status` as the first command

Always the first command to implement. It must:
- Navigate to a lightweight authenticated endpoint (or check DOM for user identity)
- Return `{"ok": true, "data": {"logged_in": true, "user": "username"}}` if authenticated
- Return `{"ok": false, "error": {"code": "not_logged_in", "message": "..."}}` + exit 1 if not

### 2. Find the auth check endpoint

Use the site archaeology protocol (site-exploration.md). Navigate to the site, start network capture, reload the page or click on user-profile link. Look for:
- A `/me`, `/account/verify_credentials`, `/user/profile` API call
- A `/session` or `/auth/status` endpoint
- Any call that returns user identity (`user_id`, `username`, `is_guest`)

### 3. Go implementation pattern

```go
func CheckLogin(client *browser.Client) (*LoginStatus, error) {
    // Navigate to a page that triggers the auth API
    if err := client.Navigate("https://example.com/explore"); err != nil {
        return nil, err
    }
    // Intercept the auth API response (discovered via site archaeology)
    data, err := interceptAPI(client, "https://api.example.com", "/user/me")
    if err != nil {
        return nil, err
    }
    // Parse the response — field names vary by site, adapt from archaeology findings
    var resp struct {
        Data struct {
            Guest  bool   `json:"guest"`   // false = logged in
            UserID string `json:"user_id"`
        } `json:"data"`
    }
    if err := json.Unmarshal(data, &resp); err != nil {
        return nil, err
    }
    return &LoginStatus{
        LoggedIn: !resp.Data.Guest,
        UserID:   resp.Data.UserID,
    }, nil
}
```

### 4. Standard not-logged-in error output

```json
{
  "ok": false,
  "error": {
    "code": "not_logged_in",
    "message": "Not logged in to [Site]. Please open Chrome, navigate to [URL], and log in manually. Then run this command again."
  }
}
```

### 5. Standard CAPTCHA error output

```json
{
  "ok": false,
  "error": {
    "code": "captcha_required",
    "message": "CAPTCHA detected on [Site]. Please complete the CAPTCHA in Chrome, then run this command again."
  }
}
```

## Decision Table

| Scenario | Action |
|----------|--------|
| Site requires login | Implement `login-status` as first command |
| Site doesn't require login | Skip login handling; no `login-status` needed |
| Not logged in at runtime | Print instructions + exit 1 |
| CAPTCHA detected | Print CAPTCHA instructions + exit 1 |
| Login form automation requested | Document as out-of-scope; direct user to log in manually |
| Session expires mid-use | Same as "not logged in" — user re-logs in Chrome, session resumes |
