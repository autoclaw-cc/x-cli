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
	if err := openOwnedNotebook(ctx, bridge, notebookURL, timeout); err != nil {
		return nil, err
	}
	const openChat = `(() => {const e=[...document.querySelectorAll('[role=tab]')].find(x=>/^(对话|Chat)$/.test((x.textContent||'').trim()));if(!e)return {ok:false};e.id='notebooklm-chat-tab';return {ok:true};})()`
	var opened struct {
		OK bool `json:"ok"`
	}
	if err := bridge.EvaluateValue(openChat, &opened); err != nil || !opened.OK {
		if err == nil {
			err = fmt.Errorf("chat tab not found")
		}
		return nil, err
	}
	if err := bridge.MouseClick("#notebooklm-chat-tab"); err != nil {
		return nil, fmt.Errorf("open chat: %w", err)
	}
	deadline := time.Now().Add(timeout)
	var ready struct {
		Ready       bool `json:"ready"`
		AnswerCount int  `json:"answerCount"`
	}
	for {
		const inspectChat = `(() => {const visible=e=>!!(e.offsetWidth||e.offsetHeight||e.getClientRects().length);const tab=document.querySelector('#notebooklm-chat-tab');const input=document.querySelector('textarea[aria-label="查询框"],textarea[aria-label="Query box"]');const r=input?.getBoundingClientRect();const loading=[...document.querySelectorAll('mat-spinner,mat-progress-spinner,[role="progressbar"]')].some(visible);return {ready:tab?.getAttribute('aria-selected')==='true'&&!!r&&r.width>0&&r.height>0&&r.right>0&&!loading,answerCount:document.querySelectorAll('button[aria-label="将消息保存到笔记中"],button[aria-label="Save message to note"]').length};})()`
		if err := bridge.EvaluateValue(inspectChat, &ready); err == nil && ready.Ready {
			break
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout: chat input did not become ready")
		}
		time.Sleep(250 * time.Millisecond)
	}
	if err := bridge.MouseClick(`textarea[aria-label="查询框"]`); err != nil {
		return nil, fmt.Errorf("focus chat: %w", err)
	}
	if err := bridge.KeyType(question); err != nil {
		return nil, fmt.Errorf("type question: %w", err)
	}
	const tagSubmit = `(() => {const t=document.querySelector('textarea[aria-label="查询框"],textarea[aria-label="Query box"]');if(!t)return {ok:false,disabled:true,valueLength:0};const tr=t.getBoundingClientRect();const bs=[...document.querySelectorAll('button[aria-label="提交"],button[aria-label="Submit"]')];const b=bs.sort((a,b)=>Math.abs(a.getBoundingClientRect().top-tr.top)-Math.abs(b.getBoundingClientRect().top-tr.top))[0];if(!b)return {ok:false,disabled:true,valueLength:t.value.length};b.id='notebooklm-chat-submit';return {ok:true,disabled:!!b.disabled,valueLength:t.value.length};})()`
	var submit struct {
		OK          bool `json:"ok"`
		Disabled    bool `json:"disabled"`
		ValueLength int  `json:"valueLength"`
	}
	for {
		if err := bridge.EvaluateValue(tagSubmit, &submit); err == nil && submit.OK && !submit.Disabled && submit.ValueLength > 0 {
			break
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout: chat input did not enable submit")
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err := bridge.MouseClick("#notebooklm-chat-submit"); err != nil {
		return nil, fmt.Errorf("submit question: %w", err)
	}

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		inspect := fmt.Sprintf(`(() => {const bs=[...document.querySelectorAll('button[aria-label="将消息保存到笔记中"],button[aria-label="Save message to note"]')];const b=bs[bs.length-1];const card=b?.closest('mat-card')||b?.parentElement?.parentElement;const citations=card?card.querySelectorAll('button.citation-marker').length:0;const clone=card?.cloneNode(true);clone?.querySelectorAll('mat-card-actions,button.citation-marker').forEach(e=>e.remove());const answer=(clone?.textContent||'').replace(/\s+/g,' ').trim();return {done:bs.length>%d,answer,citations};})()`, ready.AnswerCount)
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
