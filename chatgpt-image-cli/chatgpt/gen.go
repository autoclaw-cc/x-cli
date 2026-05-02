package chatgpt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"chatgpt-image-cli/browser"
)

const imagesURL = "https://chatgpt.com/images/"

type Result struct {
	Prompt          string `json:"prompt"`
	Path            string `json:"path"`
	Bytes           int    `json:"bytes"`
	Caption         string `json:"caption,omitempty"`
	ConversationURL string `json:"conversation_url,omitempty"`
	ElapsedMS       int64  `json:"elapsed_ms"`
}

type Options struct {
	Prompt  string
	OutDir  string
	Timeout time.Duration
}

// Gen orchestrates:
//
//	navigate to /images/ → inject prompt into ProseMirror → click send →
//	wait for conversation URL → wait for generated <img> → fetch PNG bytes
//	from inside the page → write to disk.
//
// The page is a React SPA; execCommand('insertText') alone is not enough to
// enable the send button — we must also dispatch an InputEvent so React's
// state machine re-evaluates the composer.
func Gen(c *browser.Client, opts Options) (*Result, error) {
	start := time.Now()
	if opts.Prompt == "" {
		return nil, fmt.Errorf("prompt is empty")
	}
	if opts.OutDir == "" {
		opts.OutDir = "."
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 120 * time.Second
	}
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir output dir: %w", err)
	}

	if err := c.Navigate(imagesURL, true); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}
	if err := waitTextbox(c, 15*time.Second); err != nil {
		return nil, err
	}
	if err := injectPrompt(c, opts.Prompt); err != nil {
		return nil, err
	}
	if err := waitSendEnabled(c, 5*time.Second); err != nil {
		return nil, err
	}
	// Snapshot any images already in /images/ (the gallery shows recent
	// generations as thumbnails) so we can tell our new one apart from them.
	baseline, err := captureBaselineImages(c)
	if err != nil {
		return nil, err
	}
	if err := c.Click("#composer-submit-button"); err != nil {
		return nil, fmt.Errorf("click send: %w", err)
	}

	convURL, err := waitConversationURL(c, 30*time.Second)
	if err != nil {
		return nil, err
	}
	imgInfo, err := waitForGeneratedImage(c, baseline, opts.Timeout)
	if err != nil {
		return nil, err
	}
	pngBytes, err := downloadImage(c, imgInfo.FileID, imgInfo.Src)
	if err != nil {
		return nil, err
	}

	stem := time.Now().Format("20060102-150405")
	outPath := filepath.Join(opts.OutDir, "chatgpt-"+stem+".png")
	if err := os.WriteFile(outPath, pngBytes, 0o644); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}
	abs, _ := filepath.Abs(outPath)

	return &Result{
		Prompt:          opts.Prompt,
		Path:            abs,
		Bytes:           len(pngBytes),
		Caption:         imgInfo.Alt,
		ConversationURL: convURL,
		ElapsedMS:       time.Since(start).Milliseconds(),
	}, nil
}

// --- browser-side helpers ---

func waitTextbox(c *browser.Client, timeout time.Duration) error {
	const code = `(function(){
		const tb = document.getElementById('prompt-textarea');
		return { ok: !!tb };
	})()`
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var out struct {
			OK bool `json:"ok"`
		}
		if err := c.EvaluateValue(code, &out); err == nil && out.OK {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for ChatGPT prompt textbox (not logged in? open %s in Chrome and sign in)", imagesURL)
}

func injectPrompt(c *browser.Client, prompt string) error {
	encoded, _ := json.Marshal(prompt)
	// execCommand inserts the text into ProseMirror; the explicit InputEvent
	// is what wakes React up so the send button stops being disabled.
	code := fmt.Sprintf(`(function(){
		const tb = document.getElementById('prompt-textarea');
		if (!tb) return { ok: false, err: 'textbox_not_found' };
		tb.focus();
		document.execCommand('selectAll', false, null);
		document.execCommand('insertText', false, %s);
		tb.dispatchEvent(new InputEvent('input', { bubbles: true, cancelable: true, inputType: 'insertText' }));
		return { ok: true };
	})()`, string(encoded))
	var out struct {
		OK  bool   `json:"ok"`
		Err string `json:"err"`
	}
	if err := c.EvaluateValue(code, &out); err != nil {
		return fmt.Errorf("inject prompt: %w", err)
	}
	if !out.OK {
		return fmt.Errorf("inject prompt failed: %s", out.Err)
	}
	return nil
}

func waitSendEnabled(c *browser.Client, timeout time.Duration) error {
	const code = `(function(){
		const b = document.getElementById('composer-submit-button');
		return { ok: !!b && !b.disabled };
	})()`
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var out struct {
			OK bool `json:"ok"`
		}
		if err := c.EvaluateValue(code, &out); err == nil && out.OK {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("send button never became enabled (may be rate-limited or plan-limited)")
}

func waitConversationURL(c *browser.Client, timeout time.Duration) (string, error) {
	const code = `(function(){
		const p = window.location.pathname;
		const m = p.match(/^\/c\/([0-9a-f-]+)/);
		return m ? { url: window.location.href } : { url: '' };
	})()`
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var out struct {
			URL string `json:"url"`
		}
		if err := c.EvaluateValue(code, &out); err != nil {
			return "", fmt.Errorf("poll url: %w", err)
		}
		if out.URL != "" {
			return out.URL, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return "", fmt.Errorf("timeout waiting for conversation URL (send may have been blocked by anti-abuse check)")
}

type imgInfo struct {
	Src    string `json:"src"`
	Alt    string `json:"alt"`
	FileID string `json:"fileId"`
	Width  int    `json:"w"`
	Height int    `json:"h"`
}

// captureBaselineImages records the file_id of every estuary image currently
// in <main>. We use this set as a "don't match these" filter when polling for
// the newly-generated image, because /images/ pre-renders recent generations
// as gallery thumbnails and those would otherwise race the real one.
func captureBaselineImages(c *browser.Client) (map[string]bool, error) {
	const code = `(function(){
		const out = [];
		for (const img of document.querySelectorAll('main img')) {
			const s = img.src || '';
			const m = s.match(/[?&]id=(file_[A-Za-z0-9]+)/);
			if (m) out.push(m[1]);
		}
		return out;
	})()`
	var ids []string
	if err := c.EvaluateValue(code, &ids); err != nil {
		return nil, fmt.Errorf("capture baseline: %w", err)
	}
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set, nil
}

func waitForGeneratedImage(c *browser.Client, baseline map[string]bool, timeout time.Duration) (*imgInfo, error) {
	const code = `(function(){
		const out = [];
		for (const img of document.querySelectorAll('main img')) {
			const s = img.src || '';
			if (!s.includes('/backend-api/estuary/content')) continue;
			if (!img.complete || img.naturalWidth === 0) continue;
			const m = s.match(/[?&]id=(file_[A-Za-z0-9]+)/);
			out.push({ src: s, alt: img.alt || '', fileId: m ? m[1] : '', w: img.naturalWidth, h: img.naturalHeight });
		}
		return out;
	})()`
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var candidates []imgInfo
		if err := c.EvaluateValue(code, &candidates); err != nil {
			return nil, fmt.Errorf("poll image: %w", err)
		}
		// Prefer the "main" generated image: large + conversation-style alt
		// prefix. If no candidate matches that shape, fall back to any new
		// image (covers English UI / future alt changes).
		var fallback *imgInfo
		for i := range candidates {
			img := candidates[i]
			if img.FileID == "" || baseline[img.FileID] {
				continue
			}
			if img.Width < 400 || img.Height < 400 {
				continue // small inline thumbnails in the sidebar / history strip
			}
			if isGeneratedAlt(img.Alt) {
				return &img, nil
			}
			if fallback == nil {
				fallback = &img
			}
		}
		if fallback != nil {
			return fallback, nil
		}
		if err := checkForError(c); err != nil {
			return nil, err
		}
		time.Sleep(1 * time.Second)
	}
	return nil, fmt.Errorf("timeout waiting for generated image (may have been blocked by content policy or quota)")
}

func isGeneratedAlt(alt string) bool {
	return strings.HasPrefix(alt, "已生成图片") ||
		strings.HasPrefix(alt, "Generated image") ||
		strings.Contains(alt, "generated image")
}

// checkForError looks for common error banners that ChatGPT surfaces when a
// generation is refused (content policy, quota, network). If we find one we
// fail fast instead of running out the timeout.
func checkForError(c *browser.Client) error {
	const code = `(function(){
		const text = document.body.innerText || '';
		const patterns = [
			/I (?:can|cannot|couldn't) (?:help|create|generate)/i,
			/content (?:policy|guidelines)/i,
			/rate[- ]limit|too many requests|try again later/i,
			/你已达到|超出限制|无法生成|违反了使用政策/i
		];
		for (const re of patterns) {
			const m = text.match(re);
			if (m) return { err: m[0].slice(0, 200) };
		}
		return { err: '' };
	})()`
	var out struct {
		Err string `json:"err"`
	}
	if err := c.EvaluateValue(code, &out); err != nil {
		return nil
	}
	if out.Err != "" {
		return fmt.Errorf("chatgpt refused generation: %s", strings.TrimSpace(out.Err))
	}
	return nil
}

func downloadImage(c *browser.Client, fileID, src string) ([]byte, error) {
	encodedSrc, _ := json.Marshal(src)
	encodedFileID, _ := json.Marshal(fileID)
	// Re-read the live img.src each call — the signed URL is time-sensitive
	// and the version in the DOM stays fresh while the signature rotates.
	// Scope the re-read by file_id so we don't accidentally grab a sibling
	// thumbnail that happens to sit in the same <main>.
	code := fmt.Sprintf(`(async function(){
		try {
			let url = %s;
			const fid = %s;
			if (fid) {
				const img = document.querySelector('main img[src*="' + fid + '"]');
				if (img && img.src) url = img.src;
			}
			const r = await fetch(url, { credentials: 'include' });
			if (!r.ok) return { ok: false, err: 'fetch_failed', status: r.status };
			const buf = await r.arrayBuffer();
			const u8 = new Uint8Array(buf);
			let s = '';
			const chunk = 32768;
			for (let i = 0; i < u8.length; i += chunk) {
				s += String.fromCharCode.apply(null, u8.subarray(i, i + chunk));
			}
			return { ok: true, contentType: r.headers.get('content-type') || '', size: u8.length, base64: btoa(s) };
		} catch (e) { return { ok: false, err: String(e).slice(0, 300) }; }
	})()`, string(encodedSrc), string(encodedFileID))
	var out struct {
		OK          bool   `json:"ok"`
		Err         string `json:"err"`
		Status      int    `json:"status"`
		ContentType string `json:"contentType"`
		Size        int    `json:"size"`
		Base64      string `json:"base64"`
	}
	if err := c.EvaluateValue(code, &out); err != nil {
		return nil, fmt.Errorf("fetch image: %w", err)
	}
	if !out.OK {
		return nil, fmt.Errorf("fetch image failed: %s (status=%d)", out.Err, out.Status)
	}
	if !strings.HasPrefix(out.ContentType, "image/") {
		return nil, fmt.Errorf("unexpected content-type: %s (size=%d)", out.ContentType, out.Size)
	}
	b, err := base64.StdEncoding.DecodeString(out.Base64)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}
	return b, nil
}
