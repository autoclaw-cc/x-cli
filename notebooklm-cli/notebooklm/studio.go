package notebooklm

import (
	"context"
	"fmt"
	"time"
)

type StudioCapabilities struct {
	Types []string `json:"types"`
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
