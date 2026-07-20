package notebooklm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type StudioCapabilities struct {
	Types []string `json:"types"`
}

type StudioArtifact struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Details  string `json:"details"`
	State    string `json:"state"`
	Playable bool   `json:"playable"`
	HasMenu  bool   `json:"has_menu"`
}

type StudioArtifactList struct {
	Artifacts []StudioArtifact `json:"artifacts"`
}

var studioTypes = map[string]string{
	"音频概览": "audio", "Audio Overview": "audio",
	"演示文稿": "presentation", "Presentation": "presentation",
	"视频概览": "video", "Video Overview": "video",
	"思维导图": "mind_map", "Mind Map": "mind_map",
	"报告": "report", "Report": "report",
	"闪卡": "flashcards", "Flashcards": "flashcards",
	"测验": "quiz", "Quiz": "quiz",
	"信息图": "infographic", "Infographic": "infographic",
	"数据表格": "data_table", "Data Table": "data_table",
}

func InspectStudio(ctx context.Context, bridge Bridge, notebookURL string) (*StudioCapabilities, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := openOwnedNotebook(ctx, bridge, notebookURL, 2*time.Minute); err != nil {
		return nil, err
	}
	const openStudio = `(() => {const e=[...document.querySelectorAll('[role=tab]')].find(x=>(x.textContent||'').trim()==='Studio');if(!e)return {ok:false};e.id='notebooklm-studio-tab';return {ok:true};})()`
	var opened struct {
		OK bool `json:"ok"`
	}
	if err := bridge.EvaluateValue(openStudio, &opened); err != nil || !opened.OK {
		if err == nil {
			err = fmt.Errorf("Studio tab not found")
		}
		return nil, err
	}
	if err := bridge.MouseClick("#notebooklm-studio-tab"); err != nil {
		return nil, fmt.Errorf("open Studio: %w", err)
	}
	const inspect = `(() => ({labels:[...document.querySelectorAll('.create-artifact-button-container[aria-label]')].map(e=>e.getAttribute('aria-label')).filter(Boolean)}))()`
	deadline := time.Now().Add(15 * time.Second)
	var result struct {
		Labels []string `json:"labels"`
	}
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := bridge.EvaluateValue(inspect, &result); err == nil && len(result.Labels) > 0 {
			break
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout: Studio controls did not become ready")
		}
		time.Sleep(250 * time.Millisecond)
	}
	seen := map[string]bool{}
	types := make([]string, 0, len(result.Labels))
	for _, label := range result.Labels {
		kind, ok := studioTypes[label]
		if ok && !seen[kind] {
			seen[kind] = true
			types = append(types, kind)
		}
	}
	return &StudioCapabilities{Types: types}, nil
}

func ListStudioArtifacts(ctx context.Context, bridge Bridge, notebookURL string, timeout time.Duration) (*StudioArtifactList, error) {
	if err := openOwnedNotebook(ctx, bridge, notebookURL, timeout); err != nil {
		return nil, err
	}
	const tagStudio = `(() => {const e=[...document.querySelectorAll('[role=tab]')].find(x=>(x.textContent||'').trim()==='Studio');if(!e)return {ok:false};e.id='notebooklm-artifacts-studio-tab';return {ok:true};})()`
	var tagged struct {
		OK bool `json:"ok"`
	}
	if err := bridge.EvaluateValue(tagStudio, &tagged); err != nil || !tagged.OK {
		if err == nil {
			err = fmt.Errorf("Studio tab not found")
		}
		return nil, err
	}
	if err := bridge.MouseClick("#notebooklm-artifacts-studio-tab"); err != nil {
		return nil, fmt.Errorf("open Studio: %w", err)
	}
	deadline := time.Now().Add(timeout)
	const readyScript = `(() => {const tab=document.querySelector('#notebooklm-artifacts-studio-tab');const add=[...document.querySelectorAll('button')].find(e=>/添加笔记|Add note/i.test((e.textContent||'').trim()));return {ready:tab?.getAttribute('aria-selected')==='true'&&!!add};})()`
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var state struct {
			Ready bool `json:"ready"`
		}
		if err := bridge.EvaluateValue(readyScript, &state); err == nil && state.Ready {
			break
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout: Studio artifact library did not become ready")
		}
		time.Sleep(250 * time.Millisecond)
	}

	const inspect = `(() => {const visible=e=>!!(e.offsetWidth||e.offsetHeight||e.getClientRects().length);const loading=[...document.querySelectorAll('mat-spinner,mat-progress-spinner,[role=progressbar]')].some(visible);const typeByIcon={sticky_note_2:'note',audio_magic_eraser:'audio',subscriptions:'video',tablet:'presentation',flowchart:'mind_map',auto_tab_group:'report',cards_star:'flashcards',quiz:'quiz',stacked_bar_chart:'infographic',table_view:'data_table'};const items=[...document.querySelectorAll('*')].filter(e=>e.tagName.startsWith('ARTIFACT-LIBRARY-'));return {loading,artifacts:items.map(e=>{const icon=(e.querySelector('.artifact-icon')?.textContent||'').trim();const title=(e.querySelector('.artifact-title')?.textContent||'').trim();const details=(e.querySelector('.artifact-details')?.textContent||'').trim().replace(/\s+/g,' ');const text=(e.innerText||'').trim();const buttons=[...e.querySelectorAll('button')];return {type:typeByIcon[icon]||'unknown',title,details,state:/正在生成|Generating|处理中|Processing/i.test(text)?'generating':title?'ready':'unknown',playable:buttons.some(b=>/播放|Play/i.test(b.getAttribute('aria-label')||'')),hasMenu:!!e.querySelector('.artifact-actions')||buttons.some(b=>/更多|More/i.test(b.getAttribute('aria-label')||''))};}).filter(e=>e.title)}})()`
	lastKey := ""
	stable := 0
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var state struct {
			Loading   bool             `json:"loading"`
			Artifacts []StudioArtifact `json:"artifacts"`
		}
		if err := bridge.EvaluateValue(inspect, &state); err == nil && !state.Loading {
			parts := make([]string, 0, len(state.Artifacts))
			for _, artifact := range state.Artifacts {
				parts = append(parts, fmt.Sprintf("%s\x1f%s\x1f%s\x1f%s\x1f%t\x1f%t", artifact.Type, artifact.Title, artifact.Details, artifact.State, artifact.Playable, artifact.HasMenu))
			}
			key := strings.Join(parts, "\x00")
			if key == lastKey {
				stable++
			} else {
				lastKey = key
				stable = 1
			}
			if stable >= 4 {
				if state.Artifacts == nil {
					state.Artifacts = []StudioArtifact{}
				}
				return &StudioArtifactList{Artifacts: state.Artifacts}, nil
			}
		} else {
			stable = 0
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout: Studio artifact list did not stabilize")
		}
		time.Sleep(250 * time.Millisecond)
	}
}
