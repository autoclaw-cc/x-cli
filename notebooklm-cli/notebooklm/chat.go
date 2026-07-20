package notebooklm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type ChatResult struct {
	Answer    string `json:"answer"`
	Citations int    `json:"citations"`
}

func Ask(ctx context.Context, bridge Bridge, notebookURL, question string, timeout time.Duration) (*ChatResult, error) {
	if strings.TrimSpace(question) == "" {
		return nil, fmt.Errorf("question is empty")
	}
	if err := bridge.Navigate(notebookURL, true, "NotebookLM CLI"); err != nil {
		return nil, fmt.Errorf("navigate notebook: %w", err)
	}
	if err := bridge.CDP("Page.bringToFront", map[string]any{}); err != nil {
		return nil, fmt.Errorf("activate page: %w", err)
	}
	const openChat = `(() => {const e=[...document.querySelectorAll('[role=tab]')].find(x=>/^(对话|Chat)$/.test((x.textContent||'').trim()));if(!e)return {ok:false,answerCount:0};e.click();return {ok:true,answerCount:document.querySelectorAll('button[aria-label="将消息保存到笔记中"],button[aria-label="Save message to note"]').length};})()`
	var opened struct {
		OK          bool `json:"ok"`
		AnswerCount int  `json:"answerCount"`
	}
	if err := bridge.EvaluateValue(openChat, &opened); err != nil || !opened.OK {
		if err == nil {
			err = fmt.Errorf("chat tab not found")
		}
		return nil, err
	}
	if err := bridge.MouseClick(`textarea[aria-label="查询框"]`); err != nil {
		return nil, fmt.Errorf("focus chat: %w", err)
	}
	if err := bridge.KeyType(question); err != nil {
		return nil, fmt.Errorf("type question: %w", err)
	}
	const tagSubmit = `(() => {const t=document.querySelector('textarea[aria-label="查询框"],textarea[aria-label="Query box"]');if(!t)return {ok:false,disabled:true};const tr=t.getBoundingClientRect();const bs=[...document.querySelectorAll('button[aria-label="提交"],button[aria-label="Submit"]')];const b=bs.sort((a,b)=>Math.abs(a.getBoundingClientRect().top-tr.top)-Math.abs(b.getBoundingClientRect().top-tr.top))[0];if(!b)return {ok:false,disabled:true};b.id='notebooklm-chat-submit';return {ok:true,disabled:!!b.disabled};})()`
	var submit struct {
		OK       bool `json:"ok"`
		Disabled bool `json:"disabled"`
	}
	if err := bridge.EvaluateValue(tagSubmit, &submit); err != nil {
		return nil, err
	}
	if !submit.OK || submit.Disabled {
		return nil, fmt.Errorf("chat submit control is disabled")
	}
	if err := bridge.MouseClick("#notebooklm-chat-submit"); err != nil {
		return nil, fmt.Errorf("submit question: %w", err)
	}

	deadline := time.Now().Add(timeout)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		inspect := fmt.Sprintf(`(() => {const bs=[...document.querySelectorAll('button[aria-label="将消息保存到笔记中"],button[aria-label="Save message to note"]')];const b=bs[bs.length-1];const card=b?.closest('mat-card')||b?.parentElement?.parentElement;const answer=(card?.innerText||'').replace(/\s*(保存到笔记|Save to note|copy_all|thumb_up|thumb_down).*$/s,'').trim();const citations=card?[...card.querySelectorAll('button,a')].filter(e=>/引用|citation|^\d+$/.test((e.getAttribute('aria-label')||'')+' '+(e.textContent||'').trim())).length:0;return {done:bs.length>%d,answer,citations};})()`, opened.AnswerCount)
		var state struct {
			Done      bool   `json:"done"`
			Answer    string `json:"answer"`
			Citations int    `json:"citations"`
		}
		if err := bridge.EvaluateValue(inspect, &state); err == nil && state.Done && state.Answer != "" {
			return &ChatResult{Answer: state.Answer, Citations: state.Citations}, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout: answer did not complete")
		}
		time.Sleep(500 * time.Millisecond)
	}
}
