package notebooklm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

func openOwnedNotebook(ctx context.Context, bridge Bridge, rawURL string, timeout time.Duration) error {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" || u.Host != "notebooklm.google.com" || !notebookIDPattern.MatchString(u.Path) {
		return fmt.Errorf("invalid owned NotebookLM URL")
	}
	if err := bridge.Navigate("https://notebooklm.google.com/robots.txt", true, "NotebookLM CLI"); err != nil {
		return fmt.Errorf("bootstrap NotebookLM session: %w", err)
	}
	if err := bridge.CDP("Page.bringToFront", map[string]any{}); err != nil {
		return fmt.Errorf("activate page: %w", err)
	}
	encodedPath, _ := json.Marshal(u.Path)
	jump := fmt.Sprintf(`(() => {const path=%s;setTimeout(()=>location.assign(path),0);return {ok:true};})()`, encodedPath)
	var scheduled struct {
		OK bool `json:"ok"`
	}
	if err := bridge.EvaluateValue(jump, &scheduled); err != nil || !scheduled.OK {
		if err == nil {
			err = fmt.Errorf("owned notebook navigation was not scheduled")
		}
		return err
	}
	deadline := time.Now().Add(timeout)
	inspect := fmt.Sprintf(`(() => {const pageReady=location.pathname===%s&&!!document.querySelector('input.title-input')&&!!document.querySelector('[role=tab]');if(!pageReady)return {ready:false};const visible=e=>!!(e.offsetWidth||e.offsetHeight||e.getClientRects().length);const dialog=[...document.querySelectorAll('[role=dialog]')].filter(visible).find(d=>/复制的文字|Copied text|Paste text|网站|Website|上传文件|Upload files?/i.test(d.innerText||''));if(!dialog)return {ready:true};const close=[...dialog.querySelectorAll('button')].find(e=>/^(关闭|Close)$/i.test(e.getAttribute('aria-label')||''));if(close){close.click();return {ready:false,dialogClosed:true}}return {ready:false};})()`, encodedPath)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var state struct {
			Ready bool `json:"ready"`
		}
		if err := bridge.EvaluateValue(inspect, &state); err == nil && state.Ready {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout: owned notebook did not become ready")
		}
		time.Sleep(500 * time.Millisecond)
	}
}
