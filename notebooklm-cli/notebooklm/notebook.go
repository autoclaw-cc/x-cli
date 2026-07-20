package notebooklm

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var notebookIDPattern = regexp.MustCompile(`^/notebook/([0-9a-fA-F-]{36})$`)

type NotebookResult struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Title string `json:"title"`
}

func CreateNotebook(ctx context.Context, bridge Bridge, title string, timeout time.Duration) (*NotebookResult, error) {
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("title is required")
	}
	if err := bridge.Navigate(homeURL, true, "NotebookLM CLI"); err != nil {
		return nil, fmt.Errorf("navigate NotebookLM: %w", err)
	}
	if err := bridge.CDP("Page.bringToFront", map[string]any{}); err != nil {
		return nil, fmt.Errorf("activate page: %w", err)
	}
	const clickNew = `(() => {
  const b=[...document.querySelectorAll('button')].find(e=>/新建笔记本|create new notebook|new notebook/i.test((e.getAttribute('aria-label')||'')+' '+(e.textContent||'')));
  if(!b) return {ok:false}; b.click(); return {ok:true};
})()`
	var click struct {
		OK bool `json:"ok"`
	}
	if err := bridge.EvaluateValue(clickNew, &click); err != nil || !click.OK {
		if err == nil {
			err = fmt.Errorf("new notebook control not found")
		}
		return nil, err
	}

	deadline := time.Now().Add(timeout)
	var state struct {
		URL   string `json:"url"`
		Ready bool   `json:"ready"`
	}
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		const inspect = `(() => ({url:location.href,ready:/\/notebook\/[0-9a-f-]{36}/i.test(location.pathname)}))()`
		if err := bridge.EvaluateValue(inspect, &state); err == nil && state.Ready {
			break
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout: notebook creation did not complete")
		}
		time.Sleep(250 * time.Millisecond)
	}

	encodedTitle := fmt.Sprintf("%q", title)
	setTitle := `(() => {
  const close=[...document.querySelectorAll('button[aria-label="关闭"],button[aria-label="Close"]')].find(e=>!!(e.offsetWidth||e.offsetHeight||e.getClientRects().length));
  if(close) close.click();
  const input=document.querySelector('input.title-input');
  if(!input) return {title:''};
  const set=Object.getOwnPropertyDescriptor(HTMLInputElement.prototype,'value').set;
  set.call(input,` + encodedTitle + `);
  input.dispatchEvent(new InputEvent('input',{bubbles:true,inputType:'insertText',data:` + encodedTitle + `}));
  input.dispatchEvent(new Event('change',{bubbles:true})); input.blur();
  return {title:input.value};
})()`
	var renamed struct {
		Title string `json:"title"`
	}
	if err := bridge.EvaluateValue(setTitle, &renamed); err != nil {
		return nil, fmt.Errorf("set notebook title: %w", err)
	}
	if renamed.Title != title {
		return nil, fmt.Errorf("notebook title verification failed")
	}

	u, err := url.Parse(state.URL)
	if err != nil {
		return nil, fmt.Errorf("parse notebook URL: %w", err)
	}
	match := notebookIDPattern.FindStringSubmatch(u.Path)
	if match == nil {
		return nil, fmt.Errorf("created notebook URL has no stable ID")
	}
	u.RawQuery = ""
	u.Fragment = ""
	return &NotebookResult{ID: match[1], URL: u.String(), Title: title}, nil
}
