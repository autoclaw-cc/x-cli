package notebooklm

import (
	"context"
	"fmt"
	"net/url"
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
	baselineSourceCount, err := openSourcePicker(ctx, bridge, deadline)
	if err != nil {
		return nil, err
	}
	const tagPastedText = `(() => {const b=[...document.querySelectorAll('button')].find(e=>/复制的文字|copied text|paste text/i.test(e.textContent||''));if(!b)return {ok:false};b.id='notebooklm-pasted-text';return {ok:true};})()`
	if err := tagAndClickSourceControl(bridge, tagPastedText, "#notebooklm-pasted-text"); err != nil {
		return nil, err
	}
	const pastedTextReady = `(() => ({ready:!!document.querySelector('textarea[aria-label="粘贴的文字"],textarea[aria-label="Pasted text"]')}))()`
	if err := waitSourceReady(ctx, bridge, pastedTextReady, deadline); err != nil {
		return nil, err
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

	return waitSourceIncrement(ctx, bridge, baselineSourceCount, deadline)
}

func AddURLSource(ctx context.Context, bridge Bridge, notebookURL, rawURL string, timeout time.Duration) (*SourceResult, error) {
	rawURL = strings.TrimSpace(rawURL)
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil {
		return nil, fmt.Errorf("source URL must be an absolute public http/https URL without credentials")
	}
	if err := openOwnedNotebook(ctx, bridge, notebookURL, timeout); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(timeout)
	baselineSourceCount, err := openSourcePicker(ctx, bridge, deadline)
	if err != nil {
		return nil, err
	}
	const tagWebsite = `(() => {const visible=e=>!!(e.offsetWidth||e.offsetHeight||e.getClientRects().length);const b=[...document.querySelectorAll('button')].filter(visible).find(e=>/网站|Website/i.test((e.textContent||'').trim()));if(!b)return {ok:false};b.id='notebooklm-website-source';return {ok:true};})()`
	if err := tagAndClickSourceControl(bridge, tagWebsite, "#notebooklm-website-source"); err != nil {
		return nil, err
	}
	const urlReady = `(() => {const t=[...document.querySelectorAll('textarea')].find(e=>/输入网址|Enter URL/i.test(e.getAttribute('aria-label')||''));if(!t)return {ready:false};t.id='notebooklm-url-input';return {ready:true};})()`
	if err := waitSourceReady(ctx, bridge, urlReady, deadline); err != nil {
		return nil, err
	}
	if err := bridge.MouseClick("#notebooklm-url-input"); err != nil {
		return nil, fmt.Errorf("focus source URL: %w", err)
	}
	if err := bridge.KeyType(rawURL); err != nil {
		return nil, fmt.Errorf("type source URL: %w", err)
	}
	const tagInsert = `(() => {const visible=e=>!!(e.offsetWidth||e.offsetHeight||e.getClientRects().length);const dialogs=[...document.querySelectorAll('[role=dialog]')].filter(visible);const b=dialogs.flatMap(d=>[...d.querySelectorAll('button')]).find(e=>/^(插入|Insert)$/.test((e.textContent||'').trim()));if(!b)return {ok:false,disabled:true};b.id='notebooklm-insert-url-source';return {ok:true,disabled:!!b.disabled};})()`
	if err := waitSourceInsertEnabled(bridge, tagInsert, deadline); err != nil {
		return nil, err
	}
	if err := bridge.MouseClick("#notebooklm-insert-url-source"); err != nil {
		return nil, fmt.Errorf("insert URL source: %w", err)
	}
	return waitSourceIncrement(ctx, bridge, baselineSourceCount, deadline)
}

func openSourcePicker(ctx context.Context, bridge Bridge, deadline time.Time) (int, error) {
	const tagSources = `(() => {const e=[...document.querySelectorAll('[role=tab]')].find(x=>/^(来源|Sources?)$/.test((x.textContent||'').trim()));if(!e)return {ok:false,sourceCount:0};e.id='notebooklm-sources-tab';const xs=[...document.querySelectorAll('input[type=checkbox][aria-label]')].filter(x=>!/(选择所有来源|Select all sources)/i.test(x.getAttribute('aria-label')||''));return {ok:true,sourceCount:xs.length};})()`
	var tagged struct {
		OK          bool `json:"ok"`
		SourceCount int  `json:"sourceCount"`
	}
	if err := bridge.EvaluateValue(tagSources, &tagged); err != nil || !tagged.OK {
		if err == nil {
			err = fmt.Errorf("Sources tab not found")
		}
		return 0, err
	}
	baseline := tagged.SourceCount
	if err := bridge.MouseClick("#notebooklm-sources-tab"); err != nil {
		return 0, fmt.Errorf("open Sources: %w", err)
	}
	const sourcesReady = `(() => {const e=document.querySelector('#notebooklm-sources-tab');return {ready:e?.getAttribute('aria-selected')==='true'&&!!document.querySelector('button[aria-label="添加来源"],button[aria-label="Add source"]')};})()`
	if err := waitSourceReady(ctx, bridge, sourcesReady, deadline); err != nil {
		return 0, err
	}
	const tagAdd = `(() => {const b=document.querySelector('button[aria-label="添加来源"],button[aria-label="Add source"]');if(!b)return {ok:false};b.id='notebooklm-add-source';return {ok:true};})()`
	if err := tagAndClickSourceControl(bridge, tagAdd, "#notebooklm-add-source"); err != nil {
		return 0, err
	}
	const pickerReady = `(() => ({ready:[...document.querySelectorAll('button')].some(e=>/复制的文字|copied text|paste text/i.test(e.textContent||''))&&[...document.querySelectorAll('button')].some(e=>/网站|Website/i.test(e.textContent||''))}))()`
	if err := waitSourceReady(ctx, bridge, pickerReady, deadline); err != nil {
		return 0, err
	}
	return baseline, nil
}

func tagAndClickSourceControl(bridge Bridge, script, selector string) error {
	var tagged struct {
		OK bool `json:"ok"`
	}
	if err := bridge.EvaluateValue(script, &tagged); err != nil || !tagged.OK {
		if err == nil {
			err = fmt.Errorf("source control not found")
		}
		return err
	}
	if err := bridge.MouseClick(selector); err != nil {
		return fmt.Errorf("activate source control: %w", err)
	}
	return nil
}

func waitSourceInsertEnabled(bridge Bridge, script string, deadline time.Time) error {
	var insert struct {
		OK       bool `json:"ok"`
		Disabled bool `json:"disabled"`
	}
	for {
		if err := bridge.EvaluateValue(script, &insert); err == nil && insert.OK && !insert.Disabled {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout: source insert control did not enable")
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func waitSourceIncrement(ctx context.Context, bridge Bridge, baseline int, deadline time.Time) (*SourceResult, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		const inspect = `(() => {const xs=[...document.querySelectorAll('input[type=checkbox][aria-label]')].filter(e=>!/(选择所有来源|Select all sources)/i.test(e.getAttribute('aria-label')||''));const input=document.querySelector('textarea[aria-label="粘贴的文字"],textarea[aria-label="Pasted text"],textarea[aria-label="输入网址"],textarea[aria-label="Enter URL"]');return {ready:!input,sourceCount:xs.length};})()`
		var state struct {
			Ready       bool `json:"ready"`
			SourceCount int  `json:"sourceCount"`
		}
		if err := bridge.EvaluateValue(inspect, &state); err == nil && state.Ready && state.SourceCount > baseline {
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
