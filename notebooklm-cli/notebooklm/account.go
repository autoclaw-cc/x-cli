package notebooklm

import (
	"context"
	"fmt"
)

const homeURL = "https://notebooklm.google.com/"

type LoginStatus struct {
	LoggedIn       bool   `json:"logged_in"`
	Locale         string `json:"locale"`
	PlanLabel      string `json:"plan_label,omitempty"`
	CaptchaVisible bool   `json:"captcha_visible"`
}

type AccountCapabilities struct {
	LoggedIn bool     `json:"logged_in"`
	Controls []string `json:"controls"`
}

func CheckLogin(ctx context.Context, bridge Bridge) (*LoginStatus, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := bridge.Navigate(homeURL, true, "NotebookLM CLI"); err != nil {
		return nil, fmt.Errorf("navigate NotebookLM: %w", err)
	}
	const script = `(() => {
  const texts=[...document.querySelectorAll('button,a')].map(e=>((e.getAttribute('aria-label')||'')+' '+(e.textContent||'')).trim());
  const has=re=>texts.some(t=>re.test(t));
  const body=document.body.innerText||'';
  return {
    loggedIn:has(/新建笔记本|create new notebook|new notebook/i)&&!has(/sign in|登录/i),
    locale:document.documentElement.lang||navigator.language||'',
    planLabel:(body.match(/NotebookLM\s*(?:Plus|Pro)|Google AI (?:Pro|Ultra)/i)||[])[0]||'',
    captchaVisible:/captcha|验证您是真人/i.test(body)
  };
})()`
	var result struct {
		LoggedIn       bool   `json:"loggedIn"`
		Locale         string `json:"locale"`
		PlanLabel      string `json:"planLabel"`
		CaptchaVisible bool   `json:"captchaVisible"`
	}
	if err := bridge.EvaluateValue(script, &result); err != nil {
		return nil, fmt.Errorf("inspect login: %w", err)
	}
	status := &LoginStatus{
		LoggedIn: result.LoggedIn, Locale: result.Locale,
		PlanLabel: result.PlanLabel, CaptchaVisible: result.CaptchaVisible,
	}
	if status.CaptchaVisible {
		return nil, fmt.Errorf("captcha_required: complete the CAPTCHA in Chrome")
	}
	if !status.LoggedIn {
		return nil, fmt.Errorf("not_logged_in: sign in to NotebookLM manually in Chrome")
	}
	return status, nil
}

func InspectAccountCapabilities(ctx context.Context, bridge Bridge) (*AccountCapabilities, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := bridge.Navigate(homeURL, true, "NotebookLM CLI"); err != nil {
		return nil, fmt.Errorf("navigate NotebookLM: %w", err)
	}
	const script = `(() => {
  const text=(document.body.innerText||'');
  const controls=[];
  if(/新建笔记本|create new notebook|new notebook/i.test(text)) controls.push('new_notebook');
  if(/Fast Research/i.test(text)) controls.push('fast_research');
  if(/Deep Research/i.test(text)) controls.push('deep_research');
  return {loggedIn:controls.includes('new_notebook'),controls};
})()`
	var result struct {
		LoggedIn bool     `json:"loggedIn"`
		Controls []string `json:"controls"`
	}
	if err := bridge.EvaluateValue(script, &result); err != nil {
		return nil, fmt.Errorf("inspect capabilities: %w", err)
	}
	if !result.LoggedIn {
		return nil, fmt.Errorf("not_logged_in: sign in to NotebookLM manually in Chrome")
	}
	return &AccountCapabilities{LoggedIn: result.LoggedIn, Controls: result.Controls}, nil
}
