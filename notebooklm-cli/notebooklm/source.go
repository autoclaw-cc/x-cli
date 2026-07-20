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
	if err := openOwnedNotebook(ctx, bridge, notebookURL, timeout); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(timeout)
	controls := []struct {
		tag      string
		selector string
		ready    string
	}{
		{
			tag:      `(() => {const e=[...document.querySelectorAll('[role=tab]')].find(x=>/^(来源|Sources?)$/.test((x.textContent||'').trim()));if(!e)return {ok:false,sourceCount:0};e.id='notebooklm-sources-tab';const xs=[...document.querySelectorAll('input[type=checkbox][aria-label]')].filter(x=>!/(选择所有来源|Select all sources)/i.test(x.getAttribute('aria-label')||''));return {ok:true,sourceCount:xs.length};})()`,
			selector: "#notebooklm-sources-tab",
			ready:    `(() => {const e=document.querySelector('#notebooklm-sources-tab');return {ready:e?.getAttribute('aria-selected')==='true'&&!!document.querySelector('button[aria-label="添加来源"],button[aria-label="Add source"]')};})()`,
		},
		{
			tag:      `(() => {const b=document.querySelector('button[aria-label="添加来源"],button[aria-label="Add source"]');if(!b)return {ok:false};b.id='notebooklm-add-source';return {ok:true};})()`,
			selector: "#notebooklm-add-source",
			ready:    `(() => ({ready:[...document.querySelectorAll('button')].some(e=>/复制的文字|copied text|paste text/i.test(e.textContent||''))}))()`,
		},
		{
			tag:      `(() => {const b=[...document.querySelectorAll('button')].find(e=>/复制的文字|copied text|paste text/i.test(e.textContent||''));if(!b)return {ok:false};b.id='notebooklm-pasted-text';return {ok:true};})()`,
			selector: "#notebooklm-pasted-text",
			ready:    `(() => ({ready:!!document.querySelector('textarea[aria-label="粘贴的文字"],textarea[aria-label="Pasted text"]')}))()`,
		},
	}
	baselineSourceCount := -1
	for i, control := range controls {
		var tagged struct {
			OK          bool `json:"ok"`
			SourceCount int  `json:"sourceCount"`
		}
		if err := bridge.EvaluateValue(control.tag, &tagged); err != nil || !tagged.OK {
			if err == nil {
				err = fmt.Errorf("source control not found")
			}
			return nil, err
		}
		if i == 0 {
			baselineSourceCount = tagged.SourceCount
		}
		if err := bridge.MouseClick(control.selector); err != nil {
			return nil, fmt.Errorf("activate source control: %w", err)
		}
		if err := waitSourceReady(ctx, bridge, control.ready, deadline); err != nil {
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
	for {
		if err := bridge.EvaluateValue(tagInsert, &insert); err == nil && insert.OK && !insert.Disabled {
			break
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout: source insert control did not enable")
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err := bridge.MouseClick("#notebooklm-insert-source"); err != nil {
		return nil, fmt.Errorf("insert source: %w", err)
	}

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		const inspect = `(() => {const xs=[...document.querySelectorAll('input[type=checkbox][aria-label]')].filter(e=>!/(选择所有来源|Select all sources)/i.test(e.getAttribute('aria-label')||''));return {ready:!document.querySelector('textarea[aria-label="粘贴的文字"]'),sourceCount:xs.length};})()`
		var state struct {
			Ready       bool `json:"ready"`
			SourceCount int  `json:"sourceCount"`
		}
		if err := bridge.EvaluateValue(inspect, &state); err == nil && state.Ready && state.SourceCount > baselineSourceCount {
			return &SourceResult{SourceCount: state.SourceCount}, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout: source did not become ready")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func waitSourceReady(ctx context.Context, bridge Bridge, script string, deadline time.Time) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var state struct {
			Ready bool `json:"ready"`
		}
		if err := bridge.EvaluateValue(script, &state); err == nil && state.Ready {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout: source control did not become ready")
		}
		time.Sleep(250 * time.Millisecond)
	}
}
