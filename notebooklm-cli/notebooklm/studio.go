package notebooklm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

type StudioExportResult struct {
	Artifact       StudioArtifact `json:"artifact"`
	Body           string         `json:"body"`
	BodyCharacters int            `json:"body_characters"`
}

type StudioAttributionResult struct {
	Artifact    StudioArtifact `json:"artifact"`
	Prompt      string         `json:"prompt"`
	Sources     []string       `json:"sources"`
	SourceCount int            `json:"source_count"`
}

type StudioRenameResult struct {
	Type     string `json:"type"`
	OldTitle string `json:"old_title"`
	NewTitle string `json:"new_title"`
	Renamed  bool   `json:"renamed"`
}

type StudioDeleteResult struct {
	Type          string `json:"type"`
	Title         string `json:"title"`
	Deleted       bool   `json:"deleted"`
	ArtifactCount int    `json:"artifact_count"`
}

type StudioDownloadResult struct {
	Artifact        StudioArtifact `json:"artifact"`
	OutputPath      string         `json:"output_path"`
	BytesWritten    int64          `json:"bytes_written"`
	ContentType     string         `json:"content_type,omitempty"`
	DownloadStarted bool           `json:"download_started"`
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

var studioMediaHTTPClient = http.DefaultClient
var allowStudioMediaHost = isAllowedStudioMediaHost
var studioMediaRangeChunkSize int64 = 1 << 20
var studioMediaRangeConcurrency = 4
var studioMediaRangeRequestTimeout = 90 * time.Second

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
			inferStudioArtifactTypes(state.Artifacts)
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

func WaitStudioArtifactReady(ctx context.Context, bridge Bridge, notebookURL, kind string, timeout time.Duration) (*StudioArtifact, error) {
	kind = normalizeStudioType(kind)
	if _, ok := studioLabelsByType[kind]; !ok {
		return nil, fmt.Errorf("unsupported Studio type: %s", kind)
	}
	if err := openOwnedNotebook(ctx, bridge, notebookURL, timeout); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(timeout)
	if err := openStudioArtifacts(ctx, bridge, "notebooklm-wait-studio-tab", deadline); err != nil {
		return nil, err
	}
	baselineReady := map[string]bool{}
	baselineCaptured := false
	sawGenerating := false
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		artifacts, loading, err := readStudioArtifacts(bridge)
		if err == nil && !loading {
			readyMatches := make([]StudioArtifact, 0)
			generating := false
			for _, artifact := range artifacts {
				if artifact.Type != kind {
					continue
				}
				switch artifact.State {
				case "generating":
					generating = true
					sawGenerating = true
				case "ready":
					readyMatches = append(readyMatches, artifact)
				}
			}
			if !baselineCaptured {
				baselineCaptured = true
				for _, artifact := range readyMatches {
					baselineReady[studioArtifactReadyKey(artifact)] = true
				}
				if !generating && len(readyMatches) > 0 {
					return &readyMatches[0], nil
				}
			}
			for _, artifact := range readyMatches {
				if !baselineReady[studioArtifactReadyKey(artifact)] {
					return &artifact, nil
				}
			}
			if !sawGenerating && len(readyMatches) > 0 {
				return &readyMatches[0], nil
			}
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout: Studio artifact did not reach ready state")
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func studioArtifactReadyKey(artifact StudioArtifact) string {
	return strings.Join([]string{artifact.Type, artifact.Title, artifact.Details}, "\x1f")
}

func ExportStudioArtifact(ctx context.Context, bridge Bridge, notebookURL, kind, title string, timeout time.Duration) (*StudioExportResult, error) {
	kind = normalizeStudioType(kind)
	if _, ok := studioLabelsByType[kind]; !ok {
		return nil, fmt.Errorf("unsupported Studio type: %s", kind)
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("artifact title is empty")
	}
	if err := openOwnedNotebook(ctx, bridge, notebookURL, timeout); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(timeout)
	if err := openStudioArtifacts(ctx, bridge, "notebooklm-export-studio-tab", deadline); err != nil {
		return nil, err
	}
	encodedKind, _ := json.Marshal(kind)
	encodedTitle, _ := json.Marshal(title)
	openArtifact := fmt.Sprintf(`(() => {const kind=%s;const title=%s;const typeByIcon={sticky_note_2:'note',audio_magic_eraser:'audio',subscriptions:'video',tablet:'presentation',flowchart:'mind_map',auto_tab_group:'report',cards_star:'flashcards',quiz:'quiz',stacked_bar_chart:'infographic',table_view:'data_table'};const items=[...document.querySelectorAll('*')].filter(e=>e.tagName.startsWith('ARTIFACT-LIBRARY-'));const artifacts=items.map(e=>{const icon=(e.querySelector('.artifact-icon')?.textContent||'').trim();const cardTitle=(e.querySelector('.artifact-title')?.textContent||'').trim();const details=(e.querySelector('.artifact-details')?.textContent||'').trim().replace(/\s+/g,' ');const text=(e.innerText||'').trim();return {node:e,type:typeByIcon[icon]||'unknown',title:cardTitle,details,state:/正在生成|Generating|处理中|Processing/i.test(text)?'generating':cardTitle?'ready':'unknown',playable:[...e.querySelectorAll('button')].some(b=>/播放|Play/i.test(b.getAttribute('aria-label')||'')),hasMenu:!!e.querySelector('.artifact-actions')||[...e.querySelectorAll('button')].some(b=>/更多|More/i.test(b.getAttribute('aria-label')||''))};}).filter(e=>e.title);const matches=artifacts.filter(e=>e.type===kind&&e.title===title);if(matches.length!==1)return {ok:false,matches:matches.length,available:artifacts.map(e=>({type:e.type,title:e.title,state:e.state})).slice(0,20)};const selected=matches[0];if(selected.state!=='ready')return {ok:false,matches:1,state:selected.state};const b=selected.node.querySelector('button.artifact-stretched-button')||selected.node.querySelector('.artifact-item-button')||selected.node.querySelector('button');if(!b)return {ok:false,matches:1,state:selected.state,noButton:true};b.id='notebooklm-export-artifact';b.click();return {ok:true,artifact:{type:selected.type,title:selected.title,details:selected.details,state:selected.state,playable:selected.playable,has_menu:selected.hasMenu}};})()`, encodedKind, encodedTitle)
	type openArtifactState struct {
		OK        bool           `json:"ok"`
		Matches   int            `json:"matches"`
		State     string         `json:"state"`
		NoButton  bool           `json:"noButton"`
		Artifact  StudioArtifact `json:"artifact"`
		Available []struct {
			Type  string `json:"type"`
			Title string `json:"title"`
			State string `json:"state"`
		} `json:"available"`
	}
	var opened openArtifactState
	var lastErr error
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var state openArtifactState
		if err := bridge.EvaluateValue(openArtifact, &state); err == nil {
			if state.OK {
				opened = state
				break
			}
			switch {
			case state.Matches > 1:
				return nil, fmt.Errorf("artifact title/type is not unique")
			case state.NoButton:
				lastErr = fmt.Errorf("artifact open control not found")
			case state.Matches == 1:
				lastErr = fmt.Errorf("artifact is not ready: %s", state.State)
			default:
				lastErr = fmt.Errorf("artifact not found")
			}
		} else {
			lastErr = err
		}
		if time.Now().After(deadline) {
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, fmt.Errorf("timeout: Studio artifact did not become exportable")
		}
		time.Sleep(500 * time.Millisecond)
	}
	body, err := waitStudioArtifactBody(ctx, bridge, title, deadline)
	if err != nil {
		return nil, err
	}
	return &StudioExportResult{
		Artifact:       opened.Artifact,
		Body:           body,
		BodyCharacters: len([]rune(body)),
	}, nil
}

func InspectStudioAttribution(ctx context.Context, bridge Bridge, notebookURL, kind, title string, timeout time.Duration) (*StudioAttributionResult, error) {
	kind = normalizeStudioType(kind)
	if _, ok := studioLabelsByType[kind]; !ok {
		return nil, fmt.Errorf("unsupported Studio type: %s", kind)
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("artifact title is empty")
	}
	if err := openOwnedNotebook(ctx, bridge, notebookURL, timeout); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(timeout)
	if err := openStudioArtifacts(ctx, bridge, "notebooklm-inspect-studio-tab", deadline); err != nil {
		return nil, err
	}
	artifact, err := openStudioArtifactMenu(ctx, bridge, kind, title, "notebooklm-inspect-artifact-menu", deadline)
	if err != nil {
		return nil, err
	}
	if err := clickStudioMenuItem(ctx, bridge, "notebooklm-inspect-attribution", "查看提示和来源|View prompt|Prompt.*source|Sources", deadline); err != nil {
		return nil, err
	}
	attribution, err := waitStudioAttributionDialog(ctx, bridge, deadline)
	if err != nil {
		return nil, err
	}
	attribution.Artifact = *artifact
	return attribution, nil
}

func RenameStudioArtifact(ctx context.Context, bridge Bridge, notebookURL, kind, oldTitle, newTitle string, timeout time.Duration) (*StudioRenameResult, error) {
	kind = normalizeStudioType(kind)
	if _, ok := studioLabelsByType[kind]; !ok {
		return nil, fmt.Errorf("unsupported Studio type: %s", kind)
	}
	oldTitle = strings.TrimSpace(oldTitle)
	newTitle = strings.TrimSpace(newTitle)
	if oldTitle == "" || newTitle == "" {
		return nil, fmt.Errorf("artifact title is empty")
	}
	if err := openOwnedNotebook(ctx, bridge, notebookURL, timeout); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(timeout)
	if err := openStudioArtifacts(ctx, bridge, "notebooklm-rename-studio-tab", deadline); err != nil {
		return nil, err
	}
	if _, err := openStudioArtifactMenu(ctx, bridge, kind, oldTitle, "notebooklm-rename-artifact-menu", deadline); err != nil {
		return nil, err
	}
	if err := clickStudioMenuItem(ctx, bridge, "notebooklm-rename-action", "重命名|Rename", deadline); err != nil {
		return nil, err
	}
	if err := fillInlineArtifactTitle(ctx, bridge, oldTitle, newTitle, deadline); err != nil {
		return nil, err
	}
	return &StudioRenameResult{Type: kind, OldTitle: oldTitle, NewTitle: newTitle, Renamed: true}, nil
}

func DeleteStudioArtifact(ctx context.Context, bridge Bridge, notebookURL, kind, title string, timeout time.Duration) (*StudioDeleteResult, error) {
	kind = normalizeStudioType(kind)
	if _, ok := studioLabelsByType[kind]; !ok {
		return nil, fmt.Errorf("unsupported Studio type: %s", kind)
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("artifact title is empty")
	}
	if err := openOwnedNotebook(ctx, bridge, notebookURL, timeout); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(timeout)
	if err := openStudioArtifacts(ctx, bridge, "notebooklm-delete-studio-tab", deadline); err != nil {
		return nil, err
	}
	if _, err := openStudioArtifactMenu(ctx, bridge, kind, title, "notebooklm-delete-artifact-menu", deadline); err != nil {
		return nil, err
	}
	if err := clickStudioMenuItem(ctx, bridge, "notebooklm-delete-action", "删除|Delete", deadline); err != nil {
		return nil, err
	}
	if err := confirmStudioDelete(ctx, bridge, deadline); err != nil {
		return nil, err
	}
	count, err := waitStudioArtifactRemoved(ctx, bridge, kind, title, deadline)
	if err != nil {
		return nil, err
	}
	return &StudioDeleteResult{Type: kind, Title: title, Deleted: true, ArtifactCount: count}, nil
}

func DownloadStudioArtifact(ctx context.Context, bridge Bridge, notebookURL, kind, title, outputPath string, timeout time.Duration) (*StudioDownloadResult, error) {
	kind = normalizeStudioType(kind)
	if _, ok := studioLabelsByType[kind]; !ok {
		return nil, fmt.Errorf("unsupported Studio type: %s", kind)
	}
	if kind != "audio" && kind != "video" {
		return nil, fmt.Errorf("download_unavailable: raw media download is currently supported only for audio and video artifacts")
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("artifact title is empty")
	}
	outputPath = strings.TrimSpace(outputPath)
	if outputPath == "" {
		return nil, fmt.Errorf("download output path is empty")
	}
	if err := openOwnedNotebook(ctx, bridge, notebookURL, timeout); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(timeout)
	if err := openStudioArtifacts(ctx, bridge, "notebooklm-download-studio-tab", deadline); err != nil {
		return nil, err
	}
	var networkStart struct {
		Success bool `json:"success"`
	}
	if err := bridge.NetworkValue("start", "", "", &networkStart); err != nil {
		return nil, fmt.Errorf("download_unavailable: start network capture: %w", err)
	}
	defer func() {
		var ignored struct {
			Success bool `json:"success"`
		}
		_ = bridge.NetworkValue("stop", "", "", &ignored)
	}()
	artifact, playbackClicked, err := openStudioMediaArtifactPlayback(ctx, bridge, kind, title, "notebooklm-download-playback", deadline)
	if err != nil {
		return nil, err
	}
	if !playbackClicked {
		if err := startStudioMediaPlayback(ctx, bridge, kind, deadline); err != nil {
			return nil, err
		}
	}
	mediaURL, err := waitStudioNetworkMediaURL(ctx, bridge, kind, deadline)
	if err != nil {
		return nil, err
	}
	downloadCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	bytesWritten, contentType, err := downloadStudioSignedMedia(downloadCtx, mediaURL, kind, outputPath)
	if err != nil {
		return nil, err
	}
	return &StudioDownloadResult{
		Artifact:        *artifact,
		OutputPath:      outputPath,
		BytesWritten:    bytesWritten,
		ContentType:     contentType,
		DownloadStarted: true,
	}, nil
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

func waitStudioArtifactBody(ctx context.Context, bridge Bridge, title string, deadline time.Time) (string, error) {
	encodedTitle, _ := json.Marshal(title)
	inspect := fmt.Sprintf(`(() => {const title=%s;const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));const viewer=document.querySelector('artifact-viewer');if(!visible(viewer))return {ready:false,blocks:[],body:''};const content=viewer.querySelector('.artifact-content')||viewer;const clone=content.cloneNode(true);clone.querySelectorAll('button,mat-icon,.mat-mdc-button-touch-target,.mat-ripple,.cdk-visually-hidden').forEach(e=>e.remove());const structural=[...clone.querySelectorAll('labs-tailwind-structural-element-view-v2')];let blocks=structural.map(e=>(e.innerText||e.textContent||'').replace(/\s+/g,' ').trim()).filter(Boolean);if(blocks.length===0){const raw=(clone.innerText||clone.textContent||'').replace(/[ \t]+\n/g,'\n').replace(/\n{3,}/g,'\n\n').trim();blocks=raw?raw.split(/\n+/).map(e=>e.trim()).filter(Boolean):[];}const body=blocks.join('\n\n').trim();return {ready:body.length>0&&body.includes(title),blocks,body};})()`, encodedTitle)
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		var state struct {
			Ready  bool     `json:"ready"`
			Blocks []string `json:"blocks"`
			Body   string   `json:"body"`
		}
		if err := bridge.EvaluateValue(inspect, &state); err == nil && state.Ready {
			body := normalizeStudioBlocks(state.Blocks)
			if body == "" {
				body = strings.TrimSpace(state.Body)
			}
			if body != "" && strings.Contains(body, title) {
				return body, nil
			}
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("timeout: Studio artifact body did not become ready")
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func openStudioArtifactMenu(ctx context.Context, bridge Bridge, kind, title, elementID string, deadline time.Time) (*StudioArtifact, error) {
	encodedKind, _ := json.Marshal(kind)
	encodedTitle, _ := json.Marshal(title)
	openMenu := fmt.Sprintf(`(() => {const kind=%s;const title=%s;const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));const typeByIcon={sticky_note_2:'note',audio_magic_eraser:'audio',subscriptions:'video',tablet:'presentation',flowchart:'mind_map',auto_tab_group:'report',cards_star:'flashcards',quiz:'quiz',stacked_bar_chart:'infographic',table_view:'data_table'};const items=[...document.querySelectorAll('*')].filter(e=>e.tagName.startsWith('ARTIFACT-LIBRARY-'));const artifacts=items.map(e=>{const icon=(e.querySelector('.artifact-icon')?.textContent||'').trim();const cardTitle=(e.querySelector('.artifact-title')?.textContent||'').trim();const details=(e.querySelector('.artifact-details')?.textContent||'').trim().replace(/\s+/g,' ');const text=(e.innerText||'').trim();return {node:e,type:typeByIcon[icon]||'unknown',title:cardTitle,details,state:/正在生成|Generating|处理中|Processing/i.test(text)?'generating':cardTitle?'ready':'unknown',playable:[...e.querySelectorAll('button')].some(b=>/播放|Play/i.test(b.getAttribute('aria-label')||'')),hasMenu:!!e.querySelector('.artifact-actions')||[...e.querySelectorAll('button')].some(b=>/更多|More/i.test(b.getAttribute('aria-label')||''))};}).filter(e=>e.title);const matches=artifacts.filter(e=>e.type===kind&&e.title===title);if(matches.length!==1)return {ok:false,matches:matches.length,available:artifacts.map(e=>({type:e.type,title:e.title,state:e.state})).slice(0,20)};const selected=matches[0];const menu=selected.node.querySelector('button.artifact-more-button')||[...selected.node.querySelectorAll('button')].find(b=>/更多|More|more_vert/i.test((b.getAttribute('aria-label')||'')+' '+(b.textContent||'')));if(!menu)return {ok:false,matches:1,noMenu:true};menu.id=%q;menu.click();return {ok:true,artifact:{type:selected.type,title:selected.title,details:selected.details,state:selected.state,playable:selected.playable,has_menu:selected.hasMenu}};})()`, encodedKind, encodedTitle, elementID)
	type menuState struct {
		OK       bool           `json:"ok"`
		Matches  int            `json:"matches"`
		NoMenu   bool           `json:"noMenu"`
		Artifact StudioArtifact `json:"artifact"`
	}
	var lastErr error
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var state menuState
		if err := bridge.EvaluateValue(openMenu, &state); err == nil {
			if state.OK {
				return &state.Artifact, nil
			}
			switch {
			case state.Matches > 1:
				return nil, fmt.Errorf("artifact title/type is not unique")
			case state.NoMenu:
				lastErr = fmt.Errorf("artifact menu control not found")
			default:
				lastErr = fmt.Errorf("artifact not found")
			}
		} else {
			lastErr = err
		}
		if time.Now().After(deadline) {
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, fmt.Errorf("timeout: Studio artifact menu did not become ready")
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func openStudioArtifactViewer(ctx context.Context, bridge Bridge, kind, title, elementID string, deadline time.Time) (*StudioArtifact, error) {
	encodedKind, _ := json.Marshal(kind)
	encodedTitle, _ := json.Marshal(title)
	openArtifact := fmt.Sprintf(`(() => {const kind=%s;const title=%s;const typeByIcon={sticky_note_2:'note',audio_magic_eraser:'audio',subscriptions:'video',tablet:'presentation',flowchart:'mind_map',auto_tab_group:'report',cards_star:'flashcards',quiz:'quiz',stacked_bar_chart:'infographic',table_view:'data_table'};const items=[...document.querySelectorAll('*')].filter(e=>e.tagName.startsWith('ARTIFACT-LIBRARY-'));const artifacts=items.map(e=>{const icon=(e.querySelector('.artifact-icon')?.textContent||'').trim();const cardTitle=(e.querySelector('.artifact-title')?.textContent||'').trim();const details=(e.querySelector('.artifact-details')?.textContent||'').trim().replace(/\s+/g,' ');const text=(e.innerText||'').trim();return {node:e,type:typeByIcon[icon]||'unknown',title:cardTitle,details,state:/正在生成|Generating|处理中|Processing/i.test(text)?'generating':cardTitle?'ready':'unknown',playable:[...e.querySelectorAll('button')].some(b=>/播放|Play/i.test(b.getAttribute('aria-label')||'')),hasMenu:!!e.querySelector('.artifact-actions')||[...e.querySelectorAll('button')].some(b=>/更多|More/i.test(b.getAttribute('aria-label')||''))};}).filter(e=>e.title);const matches=artifacts.filter(e=>e.type===kind&&e.title===title);if(matches.length!==1)return {ok:false,matches:matches.length};const selected=matches[0];if(selected.state!=='ready')return {ok:false,matches:1,state:selected.state};const b=selected.node.querySelector('button.artifact-stretched-button')||selected.node.querySelector('.artifact-item-button')||selected.node.querySelector('button');if(!b)return {ok:false,matches:1,noButton:true};b.id=%q;b.click();return {ok:true,artifact:{type:selected.type,title:selected.title,details:selected.details,state:selected.state,playable:selected.playable,has_menu:selected.hasMenu}};})()`, encodedKind, encodedTitle, elementID)
	type openState struct {
		OK       bool           `json:"ok"`
		Matches  int            `json:"matches"`
		State    string         `json:"state"`
		NoButton bool           `json:"noButton"`
		Artifact StudioArtifact `json:"artifact"`
	}
	var lastErr error
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var state openState
		if err := bridge.EvaluateValue(openArtifact, &state); err == nil {
			if state.OK {
				return &state.Artifact, nil
			}
			switch {
			case state.Matches > 1:
				return nil, fmt.Errorf("artifact title/type is not unique")
			case state.NoButton:
				lastErr = fmt.Errorf("artifact open control not found")
			case state.Matches == 1:
				lastErr = fmt.Errorf("artifact is not ready: %s", state.State)
			default:
				lastErr = fmt.Errorf("artifact not found")
			}
		} else {
			lastErr = err
		}
		if time.Now().After(deadline) {
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, fmt.Errorf("timeout: Studio artifact did not become downloadable")
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func openStudioMediaArtifactPlayback(ctx context.Context, bridge Bridge, kind, title, elementID string, deadline time.Time) (*StudioArtifact, bool, error) {
	encodedKind, _ := json.Marshal(kind)
	encodedTitle, _ := json.Marshal(title)
	openArtifact := fmt.Sprintf(`(() => {const kind=%s;const title=%s;const typeByIcon={sticky_note_2:'note',audio_magic_eraser:'audio',subscriptions:'video',tablet:'presentation',flowchart:'mind_map',auto_tab_group:'report',cards_star:'flashcards',quiz:'quiz',stacked_bar_chart:'infographic',table_view:'data_table'};const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));const items=[...document.querySelectorAll('*')].filter(e=>e.tagName.startsWith('ARTIFACT-LIBRARY-'));const artifacts=items.map(e=>{const icon=(e.querySelector('.artifact-icon')?.textContent||'').trim();const cardTitle=(e.querySelector('.artifact-title')?.textContent||'').trim();const details=(e.querySelector('.artifact-details')?.textContent||'').trim().replace(/\s+/g,' ');const text=(e.innerText||'').trim();return {node:e,type:typeByIcon[icon]||'unknown',title:cardTitle,details,state:/正在生成|Generating|处理中|Processing/i.test(text)?'generating':cardTitle?'ready':'unknown',playable:[...e.querySelectorAll('button')].some(b=>/播放|Play/i.test((b.getAttribute('aria-label')||'')+' '+(b.textContent||''))),hasMenu:!!e.querySelector('.artifact-actions')||[...e.querySelectorAll('button')].some(b=>/更多|More/i.test(b.getAttribute('aria-label')||''))};}).filter(e=>e.title);const matches=artifacts.filter(e=>e.type===kind&&e.title===title);if(matches.length!==1)return {ok:false,matches:matches.length};const selected=matches[0];if(selected.state!=='ready')return {ok:false,matches:1,state:selected.state};const playButton=[...selected.node.querySelectorAll('button')].filter(visible).find(b=>/播放|Play/i.test((b.getAttribute('aria-label')||'')+' '+(b.textContent||''))||/(^|\\s)play_arrow(\\s|$)/i.test(b.textContent||''));const fallback=selected.node.querySelector('button.artifact-stretched-button')||selected.node.querySelector('.artifact-item-button')||selected.node.querySelector('button');const b=playButton||fallback;if(!b)return {ok:false,matches:1,noButton:true};b.id=%q;b.scrollIntoView({block:'center',inline:'center'});return {ok:true,playbackClicked:!!playButton,artifact:{type:selected.type,title:selected.title,details:selected.details,state:selected.state,playable:selected.playable,has_menu:selected.hasMenu}};})()`, encodedKind, encodedTitle, elementID)
	type openState struct {
		OK              bool           `json:"ok"`
		Matches         int            `json:"matches"`
		State           string         `json:"state"`
		NoButton        bool           `json:"noButton"`
		PlaybackClicked bool           `json:"playbackClicked"`
		Artifact        StudioArtifact `json:"artifact"`
	}
	var lastErr error
	for {
		if err := ctx.Err(); err != nil {
			return nil, false, err
		}
		var state openState
		if err := bridge.EvaluateValue(openArtifact, &state); err == nil {
			if state.OK {
				if err := bridge.Click("#" + elementID); err != nil {
					return nil, false, fmt.Errorf("open Studio media artifact: %w", err)
				}
				return &state.Artifact, state.PlaybackClicked, nil
			}
			switch {
			case state.Matches > 1:
				return nil, false, fmt.Errorf("artifact title/type is not unique")
			case state.NoButton:
				lastErr = fmt.Errorf("artifact open control not found")
			case state.Matches == 1:
				lastErr = fmt.Errorf("artifact is not ready: %s", state.State)
			default:
				lastErr = fmt.Errorf("artifact not found")
			}
		} else {
			lastErr = err
		}
		if time.Now().After(deadline) {
			if lastErr != nil {
				return nil, false, lastErr
			}
			return nil, false, fmt.Errorf("timeout: Studio media artifact did not become playable")
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func startStudioMediaPlayback(ctx context.Context, bridge Bridge, kind string, deadline time.Time) error {
	encodedKind, _ := json.Marshal(kind)
	inspect := fmt.Sprintf(`(() => {const kind=%s;const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));const viewer=document.querySelector('artifact-viewer')||document.body;const media=[...viewer.querySelectorAll('audio,video')].filter(visible)[0];if(media){media.muted=true;media.preload='auto';try{media.load();}catch(e){}try{const promise=media.play();if(promise&&promise.catch)promise.catch(()=>{});}catch(e){}return {ready:true,tag:media.tagName,readyState:media.readyState};}const inLibrary=e=>!!e.closest('ARTIFACT-LIBRARY-ITEM,ARTIFACT-LIBRARY-NOTE,.artifact-library-container');const label=kind==='audio'?'播放音频|Play audio':'播放视频|Play video';const re=new RegExp(label,'i');const button=[...document.querySelectorAll('button')].filter(e=>visible(e)&&!inLibrary(e)).find(e=>re.test((e.getAttribute('aria-label')||'')+' '+(e.textContent||''))||(/playback-control-button/i.test(String(e.className))&&/播放|Play/i.test((e.getAttribute('aria-label')||'')+' '+(e.textContent||''))));if(button){button.id='notebooklm-media-play-button';return {ready:true,tag:kind,clickButton:true};}return {ready:false};})()`, encodedKind)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var state struct {
			Ready       bool   `json:"ready"`
			Tag         string `json:"tag"`
			ClickButton bool   `json:"clickButton"`
		}
		if err := bridge.EvaluateValue(inspect, &state); err == nil && state.Ready {
			if state.ClickButton {
				if err := bridge.Click("#notebooklm-media-play-button"); err != nil {
					return fmt.Errorf("start Studio media playback: %w", err)
				}
				return nil
			}
			tag := strings.ToLower(state.Tag)
			if kind == "audio" && tag != "audio" {
				return fmt.Errorf("download_unavailable: ready artifact does not expose an audio player")
			}
			if kind == "video" && tag != "video" {
				return fmt.Errorf("download_unavailable: ready artifact does not expose a video player")
			}
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("download_unavailable: Studio media player did not become visible")
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func waitStudioNetworkMediaURL(ctx context.Context, bridge Bridge, kind string, deadline time.Time) (string, error) {
	type request struct {
		URL      string `json:"url"`
		Method   string `json:"method"`
		Status   int    `json:"status"`
		MimeType string `json:"mimeType"`
	}
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		var state struct {
			Requests []request `json:"requests"`
		}
		if err := bridge.NetworkValue("list", "", "", &state); err == nil {
			for i := len(state.Requests) - 1; i >= 0; i-- {
				req := state.Requests[i]
				if req.Method != "" && !strings.EqualFold(req.Method, http.MethodGet) {
					continue
				}
				if req.Status != http.StatusOK && req.Status != http.StatusPartialContent {
					continue
				}
				if !isExpectedStudioMediaContent(kind, req.MimeType) {
					continue
				}
				u, err := url.Parse(req.URL)
				if err == nil && u.Scheme == "https" && allowStudioMediaHost(u.Hostname()) {
					return u.String(), nil
				}
			}
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("download_unavailable: signed Studio media request was not observed")
		}
		time.Sleep(750 * time.Millisecond)
	}
}

func downloadStudioSignedMedia(ctx context.Context, mediaURL, kind, outputPath string) (int64, string, error) {
	u, err := url.Parse(mediaURL)
	if err != nil || u.Scheme != "https" || !allowStudioMediaHost(u.Hostname()) {
		return 0, "", fmt.Errorf("download_unavailable: Studio media URL is not a supported downloadable Google media URL")
	}
	dir := filepath.Dir(outputPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return 0, "", fmt.Errorf("download_unavailable: create output directory: %w", err)
		}
	}
	firstChunk, err := fetchStudioMediaRange(ctx, u.String(), kind, 0, studioMediaRangeChunkSize-1)
	if err != nil {
		return 0, "", err
	}
	contentType := firstChunk.contentType
	total := firstChunk.total

	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return 0, "", fmt.Errorf("download_unavailable: create output file: %w", err)
	}
	closeAndRemove := func() {
		_ = file.Close()
		_ = os.Remove(outputPath)
	}
	defer func() {
		_ = file.Close()
	}()
	if _, err := file.WriteAt(firstChunk.body, firstChunk.first); err != nil {
		closeAndRemove()
		return 0, "", fmt.Errorf("download_unavailable: save Studio media: %w", err)
	}
	if firstChunk.status == http.StatusOK || total <= 0 || firstChunk.last+1 >= total {
		return int64(len(firstChunk.body)), contentType, nil
	}
	if err := file.Truncate(total); err != nil {
		closeAndRemove()
		return 0, "", fmt.Errorf("download_unavailable: size output file: %w", err)
	}

	downloadCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	type mediaRange struct {
		start int64
		end   int64
	}
	jobs := make(chan mediaRange)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup
	workerCount := studioMediaRangeConcurrency
	if workerCount < 1 {
		workerCount = 1
	}
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				chunk, err := fetchStudioMediaRange(downloadCtx, u.String(), kind, job.start, job.end)
				if err == nil && chunk.first != job.start {
					err = fmt.Errorf("download_unavailable: signed Studio media returned an unexpected range")
				}
				if err == nil {
					_, err = file.WriteAt(chunk.body, chunk.first)
				}
				if err != nil {
					select {
					case errCh <- err:
						cancel()
					default:
					}
					return
				}
			}
		}()
	}
	for start := firstChunk.last + 1; start < total; start += studioMediaRangeChunkSize {
		end := start + studioMediaRangeChunkSize - 1
		if end >= total {
			end = total - 1
		}
		select {
		case err := <-errCh:
			close(jobs)
			wg.Wait()
			closeAndRemove()
			return 0, "", err
		case <-downloadCtx.Done():
			close(jobs)
			wg.Wait()
			closeAndRemove()
			return 0, "", downloadCtx.Err()
		case jobs <- mediaRange{start: start, end: end}:
		}
	}
	close(jobs)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case err := <-errCh:
		<-done
		closeAndRemove()
		return 0, "", err
	case <-downloadCtx.Done():
		<-done
		closeAndRemove()
		return 0, "", downloadCtx.Err()
	case <-done:
	}
	return total, contentType, nil
}

func newSignedStudioMediaRequest(ctx context.Context, rawURL string, start, end int64) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("download_unavailable: create media request")
	}
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	request.Header.Set("Origin", "https://notebooklm.google.com")
	request.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	request.Header.Set("Referer", "https://notebooklm.google.com/")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36")
	return request, nil
}

type studioMediaRangeResult struct {
	status      int
	body        []byte
	first       int64
	last        int64
	total       int64
	contentType string
}

func fetchStudioMediaRange(ctx context.Context, rawURL, kind string, start, end int64) (*studioMediaRangeResult, error) {
	request, err := newSignedStudioMediaRequest(ctx, rawURL, start, end)
	if err != nil {
		return nil, err
	}
	client := *studioMediaHTTPClient
	if client.Timeout == 0 {
		client.Timeout = studioMediaRangeRequestTimeout
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("download_unavailable: fetch signed Studio media failed")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("download_unavailable: signed Studio media returned HTTP %d", response.StatusCode)
	}
	contentType := response.Header.Get("Content-Type")
	if !isExpectedStudioMediaContent(kind, contentType) {
		return nil, fmt.Errorf("download_unavailable: signed Studio media returned %s instead of %s media", contentType, kind)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("download_unavailable: read signed Studio media: %w", err)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("download_unavailable: signed Studio media returned no bytes")
	}
	result := &studioMediaRangeResult{
		status:      response.StatusCode,
		body:        body,
		first:       0,
		last:        int64(len(body)) - 1,
		total:       int64(len(body)),
		contentType: contentType,
	}
	if response.StatusCode == http.StatusPartialContent {
		first, last, total, ok := parseContentRange(response.Header.Get("Content-Range"))
		if !ok {
			return nil, fmt.Errorf("download_unavailable: signed Studio media returned an unsupported range")
		}
		result.first = first
		result.last = last
		result.total = total
	}
	return result, nil
}

func parseContentRange(value string) (int64, int64, int64, bool) {
	var first, last, total int64
	if _, err := fmt.Sscanf(value, "bytes %d-%d/%d", &first, &last, &total); err != nil {
		return 0, 0, 0, false
	}
	return first, last, total, true
}

func isAllowedStudioMediaHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	return host == "googleusercontent.com" ||
		strings.HasSuffix(host, ".googleusercontent.com") ||
		host == "googlevideo.com" ||
		strings.HasSuffix(host, ".googlevideo.com")
}

func isExpectedStudioMediaContent(kind, contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if contentType == "application/octet-stream" {
		return true
	}
	switch kind {
	case "audio":
		return strings.HasPrefix(contentType, "audio/")
	case "video":
		return strings.HasPrefix(contentType, "video/")
	default:
		return false
	}
}

func clickStudioMenuItem(ctx context.Context, bridge Bridge, elementID, pattern string, deadline time.Time) error {
	tagAction := fmt.Sprintf(`(() => {const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));const re=new RegExp(%q,'i');const item=[...document.querySelectorAll('[role=menuitem],button')].filter(visible).find(e=>re.test(((e.getAttribute('aria-label')||'')+'\n'+(e.textContent||'')).replace(/\s+/g,' ')));if(!item)return {ok:false};item.id=%q;return {ok:true,text:(item.textContent||'').replace(/\s+/g,' ').trim()};})()`, pattern, elementID)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var state struct {
			OK bool `json:"ok"`
		}
		if err := bridge.EvaluateValue(tagAction, &state); err == nil && state.OK {
			return bridge.Click("#" + elementID)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout: Studio menu item did not become ready")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func waitStudioAttributionDialog(ctx context.Context, bridge Bridge, deadline time.Time) (*StudioAttributionResult, error) {
	const inspect = `(() => {const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));const dialog=document.querySelector('SOURCE-ATTRIBUTION-DIALOG,source-attribution-dialog,[aria-label*="来源"],[aria-label*="source"]');if(!visible(dialog))return {ready:false};const prompt=(dialog.querySelector('.prompt-text')?.textContent||'').replace(/\s+/g,' ').trim();const sources=[...dialog.querySelectorAll('.source-title')].map(e=>(e.textContent||'').replace(/\s+/g,' ').trim()).filter(Boolean);return {ready:prompt.length>0||sources.length>0,prompt,sources};})()`
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var state struct {
			Ready   bool     `json:"ready"`
			Prompt  string   `json:"prompt"`
			Sources []string `json:"sources"`
		}
		if err := bridge.EvaluateValue(inspect, &state); err == nil && state.Ready {
			if state.Sources == nil {
				state.Sources = []string{}
			}
			return &StudioAttributionResult{Prompt: state.Prompt, Sources: state.Sources, SourceCount: len(state.Sources)}, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout: Studio attribution dialog did not become ready")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func fillInlineArtifactTitle(ctx context.Context, bridge Bridge, oldTitle, newTitle string, deadline time.Time) error {
	encodedOld, _ := json.Marshal(oldTitle)
	encodedNew, _ := json.Marshal(newTitle)
	fillTitle := fmt.Sprintf(`(() => {const oldTitle=%s;const newTitle=%s;const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));const input=[...document.querySelectorAll('input.artifact-title.artifact-title-input,input.artifact-title-input,input')].filter(visible).find(e=>(e.value||'').trim()===oldTitle||e.className.includes('artifact-title-input'));if(!input)return {ok:false};input.focus();const setter=Object.getOwnPropertyDescriptor(HTMLInputElement.prototype,'value').set;setter.call(input,newTitle);input.dispatchEvent(new InputEvent('input',{bubbles:true,inputType:'insertText',data:newTitle}));input.dispatchEvent(new Event('change',{bubbles:true}));return {ok:true,value:input.value};})()`, encodedOld, encodedNew)
	confirm := fmt.Sprintf(`(() => {const newTitle=%s;const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));const titles=[...document.querySelectorAll('.artifact-title,input.artifact-title-input')].filter(visible).map(e=>((e.value||e.textContent||'').trim()));return {ready:titles.includes(newTitle)};})()`, encodedNew)
	filled := false
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !filled {
			var state struct {
				OK bool `json:"ok"`
			}
			if err := bridge.EvaluateValue(fillTitle, &state); err == nil && state.OK {
				if err := bridge.SendKeys("Enter"); err != nil {
					return fmt.Errorf("commit Studio artifact title: %w", err)
				}
				filled = true
			}
		} else {
			var state struct {
				Ready bool `json:"ready"`
			}
			if err := bridge.EvaluateValue(confirm, &state); err == nil && state.Ready {
				return nil
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout: Studio artifact title did not update")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func confirmStudioDelete(ctx context.Context, bridge Bridge, deadline time.Time) error {
	const tagConfirm = `(() => {const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));const dialog=[...document.querySelectorAll('[role=dialog],mat-dialog-container,.mat-mdc-dialog-container')].filter(visible).find(d=>/确认删除|Delete|删除/i.test((d.innerText||'')+' '+(d.getAttribute('aria-label')||'')));if(!dialog)return {ok:false};dialog.id='notebooklm-confirm-delete-dialog';const b=[...dialog.querySelectorAll('button')].filter(visible).find(e=>/确认删除|Delete|删除/i.test(((e.getAttribute('aria-label')||'')+'\n'+(e.textContent||'')).replace(/\s+/g,' '))&&!/取消|Cancel/i.test(((e.getAttribute('aria-label')||'')+'\n'+(e.textContent||'')).replace(/\s+/g,' ')));if(!b)return {ok:false};b.id='notebooklm-confirm-delete';return {ok:true};})()`
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var state struct {
			OK bool `json:"ok"`
		}
		if err := bridge.EvaluateValue(tagConfirm, &state); err == nil && state.OK {
			return bridge.Click("#notebooklm-confirm-delete")
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout: Studio delete confirmation did not become ready")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func waitStudioArtifactRemoved(ctx context.Context, bridge Bridge, kind, title string, deadline time.Time) (int, error) {
	encodedKind, _ := json.Marshal(kind)
	encodedTitle, _ := json.Marshal(title)
	inspect := fmt.Sprintf(`(() => {const kind=%s;const title=%s;const typeByIcon={sticky_note_2:'note',audio_magic_eraser:'audio',subscriptions:'video',tablet:'presentation',flowchart:'mind_map',auto_tab_group:'report',cards_star:'flashcards',quiz:'quiz',stacked_bar_chart:'infographic',table_view:'data_table'};const items=[...document.querySelectorAll('*')].filter(e=>e.tagName.startsWith('ARTIFACT-LIBRARY-'));const artifacts=items.map(e=>{const icon=(e.querySelector('.artifact-icon')?.textContent||'').trim();const cardTitle=(e.querySelector('.artifact-title')?.textContent||'').trim();return {type:typeByIcon[icon]||'unknown',title:cardTitle};}).filter(e=>e.title);return {removed:!artifacts.some(e=>e.type===kind&&e.title===title),artifactCount:artifacts.length};})()`, encodedKind, encodedTitle)
	for {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		var state struct {
			Removed       bool `json:"removed"`
			ArtifactCount int  `json:"artifactCount"`
		}
		if err := bridge.EvaluateValue(inspect, &state); err == nil && state.Removed {
			return state.ArtifactCount, nil
		}
		if time.Now().After(deadline) {
			return 0, fmt.Errorf("timeout: Studio artifact was not removed")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func normalizeStudioBlocks(blocks []string) string {
	cleaned := make([]string, 0, len(blocks))
	for _, block := range blocks {
		block = strings.Join(strings.Fields(block), " ")
		if block == "" {
			continue
		}
		if len(cleaned) > 0 && cleaned[len(cleaned)-1] == block {
			continue
		}
		cleaned = append(cleaned, block)
	}
	filtered := make([]string, 0, len(cleaned))
	for i, block := range cleaned {
		compactBlock := compactStudioText(block)
		contained := 0
		for j, other := range cleaned {
			if i == j {
				continue
			}
			compactOther := compactStudioText(other)
			if len([]rune(compactOther)) >= 2 && compactBlock != compactOther && strings.Contains(compactBlock, compactOther) {
				contained++
			}
		}
		if len([]rune(compactBlock)) > 40 && contained >= 3 {
			continue
		}
		filtered = append(filtered, block)
	}
	return strings.Join(filtered, "\n\n")
}

func compactStudioText(text string) string {
	return strings.Join(strings.Fields(text), "")
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
	inferStudioArtifactTypes(state.Artifacts)
	return state.Artifacts, state.Loading, nil
}

func inferStudioArtifactTypes(artifacts []StudioArtifact) {
	for i := range artifacts {
		if artifacts[i].Type != "" && artifacts[i].Type != "unknown" {
			continue
		}
		text := strings.ToLower(artifacts[i].Title + "\n" + artifacts[i].Details)
		switch {
		case strings.Contains(text, "演示文稿") || strings.Contains(text, "presentation"):
			artifacts[i].Type = "presentation"
		case strings.Contains(text, "信息图") || strings.Contains(text, "infographic"):
			artifacts[i].Type = "infographic"
		case strings.Contains(text, "音频概览") || strings.Contains(text, "audio overview"):
			artifacts[i].Type = "audio"
		case strings.Contains(text, "视频概览") || strings.Contains(text, "video overview"):
			artifacts[i].Type = "video"
		}
	}
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
