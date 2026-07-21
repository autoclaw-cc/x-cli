# ChatGPT Web Archaeology

Verified on 2026-07-21 against `https://chatgpt.com/` through Kimi
WebBridge at `http://127.0.0.1:10400`. The browser used the isolated
`chatgpt-pro-01` profile. No cookies, headers, account identity, history
titles, or private conversation content were captured.

## Shared Composer

- Authenticated readiness: `#prompt-textarea` exists and neither
  `[data-testid="login-button"]` nor `[data-testid="signup-button"]` exists.
- Prompt input: `#prompt-textarea`, a contenteditable ProseMirror textbox.
- Send: `[data-testid="send-button"]`; on `/images/` it also has
  `#composer-submit-button`.
- Streaming: `[data-testid="stop-button"]` or an equivalent localized stop
  aria-label is present while the answer is running.
- Assistant content: the newest
  `[data-message-author-role="assistant"]` inside the latest conversation turn.
- Fresh conversation: navigate to `https://chatgpt.com/` and wait for the
  composer. This avoids selecting any existing history item.

## Tool Selection

Open `[data-testid="composer-plus-btn"]` once. The visible tool menu contains
localized entries for image creation, web search, and Deep Research. Each
entry's clickable container is the closest `[tabindex="0"]` ancestor of its
exact label.

- Web search label: `网页搜索` or `Web search`.
- Deep Research label: `深度研究` or `Deep research`.
- Image label: `创建图片` or `Create image`.

Selection is verified from the composer text before a prompt is submitted.
The workflow never retries by clicking the menu entry again. A missing
verification is an error.

## Ordinary Chat

1. Record the current assistant-message count.
2. Fill `#prompt-textarea` through WebBridge.
3. Confirm `[data-testid="send-button"]` is enabled.
4. Click it exactly once.
5. Wait for a new assistant message, no stop control, and identical non-empty
   text on three consecutive polls.

Live proof returned the requested exact sentinel in one new conversation.

## Web Search

The ordinary-chat flow is preceded by one verified web-search selection.
Visible citation anchors are collected only from the new assistant message.
Live proof returned an official IANA citation and the requested sentinel.

## Deep Research

The ordinary-chat flow is preceded by one verified Deep Research selection.
The page may initially have no assistant message while showing a stop control
and research progress. Completion therefore waits up to the command timeout
for the first assistant result, then applies the same three-poll stability
rule. Live proof completed with three authoritative citations and the requested
sentinel without a repeated click or repeated submit.

## Image Generation

- Page: `https://chatgpt.com/images/`.
- Prompt and send controls use the shared composer selectors.
- Before submit, record file IDs from completed
  `main img[src*="/backend-api/estuary/content"]` images.
- After one submit, wait for a new completed image with a new `file_` ID and
  useful natural dimensions.
- Re-read the live signed image URL, fetch it inside the authenticated page,
  validate `content-type` starts with `image/`, transfer base64 through
  WebBridge, and write bytes beneath the requested output directory.

This flow reuses the already proven `chatgpt-image-cli` behavior. It does not
use an API key, upload a file, buy credits, or auto-top-up.

## Provider Boundaries

- UI labels are localized and may change; English and Chinese labels are
  accepted, with semantic/test-ID selectors preferred.
- Deep Research can take minutes; timeout is user-configurable and submission
  is never retried automatically.
- WebBridge commands act only in the isolated account tab and close their
  session group after the invocation.
