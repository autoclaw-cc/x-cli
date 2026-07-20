package notebooklm

import (
	"context"
	"fmt"
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
	if err := bridge.Navigate(notebookURL, true, "NotebookLM CLI"); err != nil {
		return nil, fmt.Errorf("navigate notebook: %w", err)
	}
	if err := bridge.CDP("Page.bringToFront", map[string]any{}); err != nil {
		return nil, fmt.Errorf("activate page: %w", err)
	}
	const openStudio = `(() => {const e=[...document.querySelectorAll('[role=tab]')].find(x=>(x.textContent||'').trim()==='Studio');if(!e)return {ok:false};e.click();return {ok:true};})()`
	var opened struct {
		OK bool `json:"ok"`
	}
	if err := bridge.EvaluateValue(openStudio, &opened); err != nil || !opened.OK {
		if err == nil {
			err = fmt.Errorf("Studio tab not found")
		}
		return nil, err
	}
	const inspect = `(() => ({labels:[...document.querySelectorAll('.create-artifact-button-container[aria-label]')].map(e=>e.getAttribute('aria-label')).filter(Boolean)}))()`
	var result struct {
		Labels []string `json:"labels"`
	}
	if err := bridge.EvaluateValue(inspect, &result); err != nil {
		return nil, fmt.Errorf("inspect Studio: %w", err)
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
