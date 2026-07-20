package notebooklm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type SourceResult struct {
	SourceCount int `json:"source_count"`
}

func AddTextSource(ctx context.Context, bridge Bridge, notebookURL, text string, timeout time.Duration) (*SourceResult, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("source text is empty")
	}
	if err := bridge.Navigate(notebookURL, true, "NotebookLM CLI"); err != nil {
		return nil, fmt.Errorf("navigate notebook: %w", err)
	}
	if err := bridge.CDP("Page.bringToFront", map[string]any{}); err != nil {
		return nil, fmt.Errorf("activate page: %w", err)
	}
	steps := []string{
		`(() => {const e=[...document.querySelectorAll('[role=tab]')].find(x=>/^(来源|Sources?)$/.test((x.textContent||'').trim()));if(!e)return {ok:false};e.click();return {ok:true};})()`,
		`(() => {const b=document.querySelector('button[aria-label="添加来源"],button[aria-label="Add source"]');if(!b)return {ok:false};b.click();return {ok:true};})()`,
		`(() => {const b=[...document.querySelectorAll('button')].find(e=>/复制的文字|copied text|paste text/i.test(e.textContent||''));if(!b)return {ok:false};b.click();return {ok:true};})()`,
	}
	for _, script := range steps {
		var result struct {
			OK bool `json:"ok"`
		}
		if err := bridge.EvaluateValue(script, &result); err != nil || !result.OK {
			if err == nil {
				err = fmt.Errorf("source control not found")
			}
			return nil, err
		}
	}
	if err := bridge.MouseClick(`textarea[aria-label="粘贴的文字"]`); err != nil {
		return nil, fmt.Errorf("focus pasted text: %w", err)
	}
	if err := bridge.KeyType(text); err != nil {
		return nil, fmt.Errorf("type source text: %w", err)
	}
	const tagInsert = `(() => {const b=[...document.querySelectorAll('button')].find(e=>/^(插入|Insert)$/.test((e.textContent||'').trim()));if(!b)return {ok:false,disabled:true};b.id='notebooklm-insert-source';return {ok:true,disabled:!!b.disabled};})()`
	var insert struct {
		OK       bool `json:"ok"`
		Disabled bool `json:"disabled"`
	}
	if err := bridge.EvaluateValue(tagInsert, &insert); err != nil {
		return nil, err
	}
	if !insert.OK || insert.Disabled {
		return nil, fmt.Errorf("source insert control is disabled")
	}
	if err := bridge.MouseClick("#notebooklm-insert-source"); err != nil {
		return nil, fmt.Errorf("insert source: %w", err)
	}

	deadline := time.Now().Add(timeout)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		const inspect = `(() => {const xs=[...document.querySelectorAll('input[type=checkbox][aria-label]')].filter(e=>!/(选择所有来源|Select all sources)/i.test(e.getAttribute('aria-label')||''));return {ready:!document.querySelector('textarea[aria-label="粘贴的文字"]'),sourceCount:xs.length};})()`
		var state struct {
			Ready       bool `json:"ready"`
			SourceCount int  `json:"sourceCount"`
		}
		if err := bridge.EvaluateValue(inspect, &state); err == nil && state.Ready {
			return &SourceResult{SourceCount: state.SourceCount}, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout: source did not become ready")
		}
		time.Sleep(250 * time.Millisecond)
	}
}
