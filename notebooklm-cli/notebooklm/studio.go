package notebooklm

import (
	"context"
	"encoding/json"
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

type StudioGenerateResult struct {
	Type          string `json:"type"`
	Title         string `json:"title"`
	Details       string `json:"details"`
	State         string `json:"state"`
	ArtifactCount int    `json:"artifact_count"`
	Submitted     bool   `json:"submitted"`
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

var studioLabelsByType = map[string][]string{
	"audio":        {"音频概览", "Audio Overview"},
	"presentation": {"演示文稿", "Presentation"},
	"video":        {"视频概览", "Video Overview"},
	"mind_map":     {"思维导图", "Mind Map"},
	"report":       {"报告", "Report"},
	"flashcards":   {"闪卡", "Flashcards"},
	"quiz":         {"测验", "Quiz"},
	"infographic":  {"信息图", "Infographic"},
	"data_table":   {"数据表格", "Data Table"},
}

const studioArtifactInspectScript = `(() => {const visible=e=>!!(e.offsetWidth||e.offsetHeight||e.getClientRects().length);const loading=[...document.querySelectorAll('mat-spinner,mat-progress-spinner,[role=progressbar]')].some(visible);const typeByIcon={sticky_note_2:'note',audio_magic_eraser:'audio',subscriptions:'video',tablet:'presentation',flowchart:'mind_map',auto_tab_group:'report',cards_star:'flashcards',quiz:'quiz',stacked_bar_chart:'infographic',table_view:'data_table'};const items=[...document.querySelectorAll('*')].filter(e=>e.tagName.startsWith('ARTIFACT-LIBRARY-'));return {loading,artifacts:items.map(e=>{const icon=(e.querySelector('.artifact-icon')?.textContent||'').trim();const title=(e.querySelector('.artifact-title')?.textContent||'').trim();const details=(e.querySelector('.artifact-details')?.textContent||'').trim().replace(/\s+/g,' ');const text=(e.innerText||'').trim();const buttons=[...e.querySelectorAll('button')];return {type:typeByIcon[icon]||'unknown',title,details,state:/正在生成|Generating|处理中|Processing/i.test(text)?'generating':title?'ready':'unknown',playable:buttons.some(b=>/播放|Play/i.test(b.getAttribute('aria-label')||'')),hasMenu:!!e.querySelector('.artifact-actions')||buttons.some(b=>/更多|More/i.test(b.getAttribute('aria-label')||''))};}).filter(e=>e.title)}})()`

func InspectStudio(ctx context.Context, bridge Bridge, notebookURL string) (*StudioCapabilities, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := openOwnedNotebook(ctx, bridge, notebookURL, 2*time.Minute); err != nil {
		return nil, err
	}
	const openStudio = `(() => {const e=[...document.querySelectorAll('[role=tab]')].find(x=>(x.textContent||'').trim()==='Studio');if(!e)return {ok:false};e.id='notebooklm-studio-tab';e.click();return {ok:true};})()`
	var opened struct {
		OK bool `json:"ok"`
	}
	if err := bridge.EvaluateValue(openStudio, &opened); err != nil || !opened.OK {
		if err == nil {
			err = fmt.Errorf("Studio tab not found")
		}
		return nil, err
	}
	const inspect = `(() => {const tab=document.querySelector('#notebooklm-studio-tab');if(tab&&tab.getAttribute('aria-selected')!=='true'){tab.click();return {labels:[]};}return {labels:[...document.querySelectorAll('.create-artifact-button-container[aria-label]')].map(e=>e.getAttribute('aria-label')).filter(Boolean)}})()`
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
	deadline := time.Now().Add(timeout)
	if err := openStudioArtifacts(ctx, bridge, "notebooklm-artifacts-studio-tab", deadline); err != nil {
		return nil, err
	}
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
		if err := bridge.EvaluateValue(studioArtifactInspectScript, &state); err == nil && !state.Loading {
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

func GenerateStudioArtifact(ctx context.Context, bridge Bridge, notebookURL, kind, prompt string, waitReady bool, timeout time.Duration) (*StudioGenerateResult, error) {
	kind = normalizeStudioType(kind)
	labels, ok := studioLabelsByType[kind]
	if !ok {
		return nil, fmt.Errorf("unsupported Studio type: %s", kind)
	}
	if err := openOwnedNotebook(ctx, bridge, notebookURL, timeout); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(timeout)
	if err := openStudioArtifacts(ctx, bridge, "notebooklm-generate-studio-tab", deadline); err != nil {
		return nil, err
	}
	baseline, err := readStudioArtifactsStable(ctx, bridge, deadline)
	if err != nil {
		return nil, err
	}
	baselineCount := countStudioType(baseline, kind)
	encodedLabels, _ := json.Marshal(labels)
	tagCreate := fmt.Sprintf(`(() => {const labels=%s;const controls=[...document.querySelectorAll('.create-artifact-button-container[aria-label]')];const e=controls.find(x=>labels.includes((x.getAttribute('aria-label')||'').trim()));if(!e)return {ok:false,available:controls.map(x=>x.getAttribute('aria-label')).filter(Boolean)};e.id='notebooklm-create-studio-artifact';return {ok:true,label:e.getAttribute('aria-label')};})()`, encodedLabels)
	var tagged struct {
		OK    bool   `json:"ok"`
		Label string `json:"label"`
	}
	if err := bridge.EvaluateValue(tagCreate, &tagged); err != nil || !tagged.OK {
		if err == nil {
			err = fmt.Errorf("Studio type control not found")
		}
		return nil, err
	}
	if err := bridge.Click("#notebooklm-create-studio-artifact"); err != nil {
		return nil, fmt.Errorf("start Studio generation: %w", err)
	}
	submitted, err := submitStudioGenerationDialog(ctx, bridge, prompt, deadline)
	if err != nil {
		return nil, err
	}
	return waitGeneratedStudioArtifact(ctx, bridge, kind, baselineCount, waitReady, submitted, deadline)
}

func normalizeStudioType(kind string) string {
	kind = strings.ToLower(strings.TrimSpace(kind))
	kind = strings.ReplaceAll(kind, "-", "_")
	kind = strings.ReplaceAll(kind, " ", "_")
	return kind
}

func openStudioArtifacts(ctx context.Context, bridge Bridge, elementID string, deadline time.Time) error {
	tagStudio := fmt.Sprintf(`(() => {const e=[...document.querySelectorAll('[role=tab]')].find(x=>(x.textContent||'').trim()==='Studio');if(!e)return {ok:false};e.id=%q;e.click();return {ok:true};})()`, elementID)
	var tagged struct {
		OK bool `json:"ok"`
	}
	if err := bridge.EvaluateValue(tagStudio, &tagged); err != nil || !tagged.OK {
		if err == nil {
			err = fmt.Errorf("Studio tab not found")
		}
		return err
	}
	readyScript := fmt.Sprintf(`(() => {const tab=document.querySelector(%q);if(tab&&tab.getAttribute('aria-selected')!=='true'){tab.click();return {ready:false};}const add=[...document.querySelectorAll('button')].find(e=>/添加笔记|Add note/i.test((e.textContent||'').trim()));return {ready:tab?.getAttribute('aria-selected')==='true'&&!!add};})()`, "#"+elementID)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var state struct {
			Ready bool `json:"ready"`
		}
		if err := bridge.EvaluateValue(readyScript, &state); err == nil && state.Ready {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout: Studio artifact library did not become ready")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func readStudioArtifactsStable(ctx context.Context, bridge Bridge, deadline time.Time) ([]StudioArtifact, error) {
	lastKey := ""
	stable := 0
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		artifacts, loading, err := readStudioArtifacts(bridge)
		if err == nil && !loading {
			parts := make([]string, 0, len(artifacts))
			for _, artifact := range artifacts {
				parts = append(parts, fmt.Sprintf("%s\x1f%s\x1f%s\x1f%s", artifact.Type, artifact.Title, artifact.Details, artifact.State))
			}
			key := strings.Join(parts, "\x00")
			if key == lastKey {
				stable++
			} else {
				lastKey = key
				stable = 1
			}
			if stable >= 4 {
				return artifacts, nil
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

func readStudioArtifacts(bridge Bridge) ([]StudioArtifact, bool, error) {
	var state struct {
		Loading   bool             `json:"loading"`
		Artifacts []StudioArtifact `json:"artifacts"`
	}
	if err := bridge.EvaluateValue(studioArtifactInspectScript, &state); err != nil {
		return nil, false, err
	}
	if state.Artifacts == nil {
		state.Artifacts = []StudioArtifact{}
	}
	return state.Artifacts, state.Loading, nil
}

func countStudioType(artifacts []StudioArtifact, kind string) int {
	count := 0
	for _, artifact := range artifacts {
		if artifact.Type == kind {
			count++
		}
	}
	return count
}

func submitStudioGenerationDialog(ctx context.Context, bridge Bridge, prompt string, deadline time.Time) (bool, error) {
	softDeadline := time.Now().Add(15 * time.Second)
	if deadline.Before(softDeadline) {
		softDeadline = deadline
	}
	promptFilled := false
	optionSelected := false
	const inspect = `(() => {const visible=e=>!!(e.offsetWidth||e.offsetHeight||e.getClientRects().length);const fields=[...document.querySelectorAll('textarea,input[type=text],.ProseMirror,[contenteditable=true]')].filter(e=>visible(e)&&!e.disabled&&!e.readOnly&&!/查询框|query box/i.test(e.getAttribute('aria-label')||''));const field=fields[0];if(field)field.id='notebooklm-studio-prompt';const buttons=[...document.querySelectorAll('button')].filter(visible);const submit=buttons.find(b=>!b.disabled&&/^(生成|创建|Generate|Create|生成.*|Create.*|Generate.*)$/i.test((b.textContent||'').trim())&&!/取消|Cancel/i.test((b.textContent||'').trim()));if(submit)submit.id='notebooklm-studio-submit';const defaults=['简报文档','Briefing document','Briefing Doc','Briefing doc','学习指南','Study Guide','Study guide'];const option=buttons.find(b=>!b.disabled&&defaults.some(label=>((b.getAttribute('aria-label')||'')+'\n'+(b.textContent||'')).includes(label)));if(option)option.id='notebooklm-studio-default-option';const active=!!field||!!submit||!!option||[...document.querySelectorAll('[role=dialog],mat-dialog-container,.mat-mdc-dialog-container')].some(visible);return {active,hasPrompt:!!field,canSubmit:!!submit,hasDefaultOption:!!option};})()`
	for {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		var state struct {
			Active           bool `json:"active"`
			HasPrompt        bool `json:"hasPrompt"`
			CanSubmit        bool `json:"canSubmit"`
			HasDefaultOption bool `json:"hasDefaultOption"`
		}
		if err := bridge.EvaluateValue(inspect, &state); err == nil {
			if strings.TrimSpace(prompt) != "" && state.HasPrompt && !promptFilled {
				if err := bridge.Fill("#notebooklm-studio-prompt", prompt); err != nil {
					if focusErr := bridge.MouseClick("#notebooklm-studio-prompt"); focusErr != nil {
						return false, fmt.Errorf("focus Studio prompt: %w", focusErr)
					}
					if typeErr := bridge.KeyType(prompt); typeErr != nil {
						return false, fmt.Errorf("type Studio prompt: %w", typeErr)
					}
				}
				promptFilled = true
			}
			if !state.CanSubmit && state.HasDefaultOption {
				if err := bridge.Click("#notebooklm-studio-default-option"); err != nil {
					return false, fmt.Errorf("select Studio default option: %w", err)
				}
				optionSelected = true
				time.Sleep(250 * time.Millisecond)
				continue
			}
			if state.CanSubmit {
				if err := bridge.Click("#notebooklm-studio-submit"); err != nil {
					return false, fmt.Errorf("submit Studio generation: %w", err)
				}
				return true, nil
			}
			if !state.Active && optionSelected {
				return true, nil
			}
			if !state.Active && time.Now().After(softDeadline) {
				return optionSelected, nil
			}
		}
		if time.Now().After(softDeadline) {
			return optionSelected, nil
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func waitGeneratedStudioArtifact(ctx context.Context, bridge Bridge, kind string, baselineCount int, waitReady, submitted bool, deadline time.Time) (*StudioGenerateResult, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		artifacts, loading, err := readStudioArtifacts(bridge)
		if err == nil && !loading {
			matches := make([]StudioArtifact, 0)
			generating := false
			for _, artifact := range artifacts {
				if artifact.Type != kind {
					continue
				}
				matches = append(matches, artifact)
				if artifact.State == "generating" {
					generating = true
				}
			}
			if len(matches) > baselineCount && (!waitReady || !generating) {
				selected := matches[0]
				for _, artifact := range matches {
					if artifact.State == "ready" {
						selected = artifact
						break
					}
				}
				return &StudioGenerateResult{
					Type:          kind,
					Title:         selected.Title,
					Details:       selected.Details,
					State:         selected.State,
					ArtifactCount: len(matches),
					Submitted:     submitted,
				}, nil
			}
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout: Studio artifact did not reach requested state")
		}
		time.Sleep(500 * time.Millisecond)
	}
}
