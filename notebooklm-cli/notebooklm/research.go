package notebooklm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type ResearchResult struct {
	Mode              string `json:"mode"`
	Query             string `json:"query"`
	Status            string `json:"status"`
	Summary           string `json:"summary"`
	ResultCount       int    `json:"result_count"`
	Imported          bool   `json:"imported"`
	SourceCountBefore int    `json:"source_count_before"`
	SourceCountAfter  int    `json:"source_count_after"`
}

func RunResearch(ctx context.Context, bridge Bridge, notebookURL, mode, query string, importResults bool, timeout time.Duration) (*ResearchResult, error) {
	mode = normalizeResearchMode(mode)
	if mode != "fast" && mode != "deep" {
		return nil, fmt.Errorf("unsupported research mode: %s", mode)
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("research query is empty")
	}
	if err := openOwnedNotebook(ctx, bridge, notebookURL, timeout); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(timeout)
	baseline, err := openSourcesAndCount(ctx, bridge, deadline)
	if err != nil {
		return nil, err
	}
	if err := clearResearchResultIfPresent(ctx, bridge, deadline); err != nil {
		return nil, err
	}
	if err := fillResearchQuery(ctx, bridge, query, deadline); err != nil {
		return nil, err
	}
	if mode == "deep" {
		if err := selectResearchMode(ctx, bridge, mode, deadline); err != nil {
			return nil, err
		}
		if err := fillResearchQuery(ctx, bridge, query, deadline); err != nil {
			return nil, err
		}
	}
	if err := clickResearchSubmit(ctx, bridge, deadline); err != nil {
		return nil, err
	}
	result, err := waitResearchComplete(ctx, bridge, mode, query, deadline)
	if err != nil {
		return nil, err
	}
	result.SourceCountBefore = baseline
	if importResults {
		if err := importResearchResults(ctx, bridge, deadline); err != nil {
			return nil, err
		}
		after, err := waitResearchSourceIncrement(ctx, bridge, baseline, deadline)
		if err != nil {
			return nil, err
		}
		result.Imported = true
		result.SourceCountAfter = after
	} else {
		result.SourceCountAfter = baseline
	}
	return result, nil
}

func normalizeResearchMode(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	switch mode {
	case "fast", "fast_research", "fast-research":
		return "fast"
	case "deep", "deep_research", "deep-research":
		return "deep"
	default:
		return mode
	}
}

func selectResearchMode(ctx context.Context, bridge Bridge, mode string, deadline time.Time) error {
	className := "research-option-fast-research"
	label := "Fast Research"
	if mode == "deep" {
		className = "research-option-deep-research"
		label = "Deep Research"
	}
	confirm := fmt.Sprintf(`(() => {const target=%q;const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));const inMenu=e=>!!e.closest('.cdk-overlay-pane,[role=menu],[role=listbox]');const buttons=[...document.querySelectorAll('button')].filter(visible);const b=buttons.find(e=>/researcher-menu-trigger/i.test(String(e.className)))||buttons.find(e=>!inMenu(e)&&/Fast Research|Deep Research/i.test((e.textContent||'')+' '+(e.getAttribute('aria-label')||''))&&(/menu-trigger/i.test(String(e.className))||e.getAttribute('aria-haspopup')));const menuOpen=[...document.querySelectorAll('.cdk-overlay-pane,[role=menu],[role=listbox]')].some(visible);if(!b)return {ready:false,selected:false,menuOpen};const text=(b.textContent||'').replace(/\s+/g,' ').trim();return {ready:true,selected:text.includes(target),disabled:!!b.disabled,text,menuOpen};})()`, label)
	openMenu := `(() => {const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));const inMenu=e=>!!e.closest('.cdk-overlay-pane,[role=menu],[role=listbox]');const buttons=[...document.querySelectorAll('button')].filter(visible);const b=buttons.find(e=>/researcher-menu-trigger/i.test(String(e.className)))||buttons.find(e=>!inMenu(e)&&/Fast Research|Deep Research/i.test((e.textContent||'')+' '+(e.getAttribute('aria-label')||''))&&(/menu-trigger/i.test(String(e.className))||e.getAttribute('aria-haspopup')));if(!b||b.disabled)return {ok:false};b.id='notebooklm-research-mode-menu';return {ok:true,text:(b.textContent||'').replace(/\s+/g,' ').trim()};})()`
	choose := fmt.Sprintf(`(() => {const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));const items=[...document.querySelectorAll('.cdk-overlay-pane button,.cdk-overlay-pane [role=menuitem],[role=menu] button,[role=menuitem],[role=option]')].filter(visible);const item=items.find(e=>(e.className||'').includes(%q)||((e.textContent||'').includes(%q)));if(!item)return {ok:false};item.id='notebooklm-research-mode-choice';return {ok:true};})()`, className, label)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var current struct {
			Ready    bool `json:"ready"`
			Selected bool `json:"selected"`
			Disabled bool `json:"disabled"`
			MenuOpen bool `json:"menuOpen"`
		}
		_ = bridge.EvaluateValue(confirm, &current)
		if current.Selected {
			if current.MenuOpen {
				if err := bridge.SendKeys("Escape"); err != nil {
					return fmt.Errorf("close research mode menu: %w", err)
				}
				time.Sleep(250 * time.Millisecond)
				continue
			}
			return nil
		}
		var selected struct {
			OK bool `json:"ok"`
		}
		if err := bridge.EvaluateValue(choose, &selected); err == nil && selected.OK {
			if err := bridge.Click("#notebooklm-research-mode-choice"); err != nil {
				return fmt.Errorf("select research mode: %w", err)
			}
			time.Sleep(time.Second)
			continue
		}
		if current.Ready && !current.Disabled {
			var opened struct {
				OK bool `json:"ok"`
			}
			_ = bridge.EvaluateValue(openMenu, &opened)
			if opened.OK {
				if err := bridge.Click("#notebooklm-research-mode-menu"); err != nil {
					return fmt.Errorf("open research mode menu: %w", err)
				}
				time.Sleep(time.Second)
				continue
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout: research mode did not become selectable")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func clearResearchResultIfPresent(ctx context.Context, bridge Bridge, deadline time.Time) error {
	tagDelete := `(() => {const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));const importButton=[...document.querySelectorAll('button')].filter(visible).find(e=>(e.className||'').includes('source-discovery-completed-action-import-button')||/^(add\\s*)?导入$|Import/i.test((e.textContent||'').replace(/\s+/g,' ').trim()));if(!importButton)return {present:false};const buttons=[...document.querySelectorAll('button')].filter(visible);const b=buttons.find(e=>(e.className||'').includes('source-discovery-completed-action-delete-button'))||buttons.find(e=>/^(删除|Delete)$/.test((e.textContent||'').replace(/\s+/g,' ').trim())||/删除|Delete/i.test(e.getAttribute('aria-label')||''));if(!b)return {present:true,missingDelete:true};b.id='notebooklm-research-delete';b.scrollIntoView({block:'center',inline:'center'});return {present:true};})()`
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var state struct {
			Present       bool `json:"present"`
			MissingDelete bool `json:"missingDelete"`
		}
		if err := bridge.EvaluateValue(tagDelete, &state); err == nil {
			if !state.Present {
				return nil
			}
			if state.MissingDelete {
				return fmt.Errorf("research result delete control not found")
			}
			if err := bridge.Click("#notebooklm-research-delete"); err != nil {
				return fmt.Errorf("clear existing research result: %w", err)
			}
			resetDeadline := time.Now().Add(15 * time.Second)
			if deadline.Before(resetDeadline) {
				resetDeadline = deadline
			}
			if err := waitResearchControlsReset(ctx, bridge, resetDeadline); err != nil {
				return fmt.Errorf("research_result_present: existing research result could not be cleared")
			}
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout: existing research result did not become clearable")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func waitResearchControlsReset(ctx context.Context, bridge Bridge, deadline time.Time) error {
	inspect := `(() => {const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));const importButton=[...document.querySelectorAll('button')].filter(visible).find(e=>(e.className||'').includes('source-discovery-completed-action-import-button')||/^(add\\s*)?导入$|Import/i.test((e.textContent||'').replace(/\s+/g,' ').trim()));const q=[...document.querySelectorAll('textarea')].filter(visible).find(e=>/基于输入的查询发现来源|Discover sources/i.test((e.getAttribute('aria-label')||'')+' '+(e.getAttribute('placeholder')||'')));return {ready:!importButton&&!!q&&!q.disabled&&!q.readOnly};})()`
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
			return fmt.Errorf("timeout: research controls did not reset")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func fillResearchQuery(ctx context.Context, bridge Bridge, query string, deadline time.Time) error {
	tagQuery := `(() => {const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));const qs=[...document.querySelectorAll('textarea')].filter(e=>visible(e)&&!e.disabled&&!e.readOnly);const q=qs.find(e=>/基于输入的查询发现来源|Discover sources/i.test((e.getAttribute('aria-label')||'')+' '+(e.getAttribute('placeholder')||'')))||qs.find(e=>/查询框|query box/i.test(e.getAttribute('aria-label')||''));if(!q)return {ok:false,reason:'query'};q.id='notebooklm-research-query';q.focus();const setter=Object.getOwnPropertyDescriptor(HTMLTextAreaElement.prototype,'value').set;setter.call(q,'');q.dispatchEvent(new InputEvent('input',{bubbles:true,inputType:'deleteContentBackward',data:null}));q.dispatchEvent(new Event('change',{bubbles:true}));return {ok:true,value:q.value};})()`
	confirmValue := fmt.Sprintf(`(() => {const q=document.querySelector('#notebooklm-research-query');return {ready:!!q&&q.value.includes(%q),value:q?.value||''};})()`, query)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var state struct {
			OK bool `json:"ok"`
		}
		if err := bridge.EvaluateValue(tagQuery, &state); err == nil && state.OK {
			if err := bridge.MouseClick("#notebooklm-research-query"); err != nil {
				return fmt.Errorf("focus research query: %w", err)
			}
			_ = bridge.CDP("Page.bringToFront", map[string]any{"dummy": false})
			if err := bridge.CDP("Input.insertText", map[string]any{"text": query}); err != nil {
				return fmt.Errorf("insert research query: %w", err)
			}
			for i := 0; i < 10; i++ {
				var confirmed struct {
					Ready bool `json:"ready"`
				}
				if err := bridge.EvaluateValue(confirmValue, &confirmed); err == nil && confirmed.Ready {
					return nil
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout: research query box did not become ready")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func clickResearchSubmit(ctx context.Context, bridge Bridge, deadline time.Time) error {
	tagSubmit := `(() => {const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));const buttons=[...document.querySelectorAll('button')].filter(e=>visible(e)&&!e.disabled);const b=buttons.find(e=>/提交|Submit/i.test(e.getAttribute('aria-label')||'')&&(/actions-enter-button|submit-button/i.test(String(e.className))||/search|arrow_forward/.test(e.textContent||'')));if(!b)return {ok:false,disabled:true};b.id='notebooklm-research-submit';return {ok:true,disabled:false,text:(b.textContent||'').replace(/\s+/g,' ').trim()};})()`
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var state struct {
			OK       bool `json:"ok"`
			Disabled bool `json:"disabled"`
		}
		if err := bridge.EvaluateValue(tagSubmit, &state); err == nil && state.OK && !state.Disabled {
			return bridge.Click("#notebooklm-research-submit")
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout: research submit control did not enable")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func waitResearchComplete(ctx context.Context, bridge Bridge, mode, query string, deadline time.Time) (*ResearchResult, error) {
	inspect := `(() => {const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));const loading=[...document.querySelectorAll('mat-spinner,mat-progress-spinner,[role=progressbar]')].some(visible);const importButton=[...document.querySelectorAll('button')].filter(visible).find(e=>(e.className||'').includes('source-discovery-completed-action-import-button')||/导入|Import/i.test(e.textContent||''));const body=(document.body.innerText||'').replace(/\s+/g,' ').trim();const done=/Fast Research\s*已完成|Deep Research\s*已完成|Fast Research complete|Deep Research complete/i.test(body)||!!importButton;const summary=done?body.replace(/^.*?(Fast Research|Deep Research)/,'$1').slice(0,1200):'';const more=(body.match(/另外\s*(\d+)\s*个来源|(\d+)\s*more sources?/i)||[]);let extra=0;if(more[1])extra=parseInt(more[1],10);if(more[2])extra=parseInt(more[2],10);const visibleResults=[...document.querySelectorAll('a,button')].filter(visible).map(e=>(e.textContent||'').replace(/\s+/g,' ').trim()).filter(Boolean).filter(t=>!/导入|删除|thumb|sort|settings|share|Studio|来源|对话/.test(t));return {loading,done,summary,resultCount:done?Math.max(visibleResults.length, extra>0?extra+3:0):0};})()`
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var state struct {
			Loading     bool   `json:"loading"`
			Done        bool   `json:"done"`
			Summary     string `json:"summary"`
			ResultCount int    `json:"resultCount"`
		}
		if err := bridge.EvaluateValue(inspect, &state); err == nil && !state.Loading && state.Done {
			return &ResearchResult{
				Mode:        mode,
				Query:       query,
				Status:      "complete",
				Summary:     strings.TrimSpace(state.Summary),
				ResultCount: state.ResultCount,
			}, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout: research did not complete")
		}
		time.Sleep(1 * time.Second)
	}
}

func importResearchResults(ctx context.Context, bridge Bridge, deadline time.Time) error {
	tagImport := `(() => {const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));const b=[...document.querySelectorAll('button')].filter(visible).find(e=>(e.className||'').includes('source-discovery-completed-action-import-button')||/导入|Import/i.test(e.textContent||''));if(!b)return {ok:false};b.id='notebooklm-research-import';return {ok:true,disabled:!!b.disabled};})()`
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var state struct {
			OK       bool `json:"ok"`
			Disabled bool `json:"disabled"`
		}
		if err := bridge.EvaluateValue(tagImport, &state); err == nil && state.OK && !state.Disabled {
			return bridge.Click("#notebooklm-research-import")
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout: research import control did not become ready")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func waitResearchSourceIncrement(ctx context.Context, bridge Bridge, baseline int, deadline time.Time) (int, error) {
	for {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		count, err := openSourcesAndCount(ctx, bridge, deadline)
		if err == nil && count > baseline {
			return count, nil
		}
		if time.Now().After(deadline) {
			return 0, fmt.Errorf("timeout: imported research sources did not become ready")
		}
		time.Sleep(500 * time.Millisecond)
	}
}
