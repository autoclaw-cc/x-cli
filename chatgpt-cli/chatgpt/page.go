package chatgpt

import "strings"

type PageSnapshot struct {
	Href          string   `json:"href"`
	Locale        string   `json:"locale"`
	HasPrompt     bool     `json:"hasPrompt"`
	LoginControls bool     `json:"loginControls"`
	ToolLabels    []string `json:"toolLabels"`
}

type PageState struct {
	Authenticated   bool     `json:"authenticated"`
	PromptAvailable bool     `json:"prompt_available"`
	Locale          string   `json:"locale"`
	Capabilities    []string `json:"capabilities"`
	URL             string   `json:"url"`
}

func AnalyzePage(snapshot PageSnapshot) PageState {
	authenticated := strings.HasPrefix(snapshot.Href, "https://chatgpt.com/") && snapshot.HasPrompt && !snapshot.LoginControls
	capabilities := []string{}
	if authenticated {
		capabilities = append(capabilities, "chat")
		labels := strings.ToLower(strings.Join(snapshot.ToolLabels, " "))
		if containsAny(labels, "网页搜索", "web search") {
			capabilities = append(capabilities, "web_search")
		}
		if containsAny(labels, "深度研究", "deep research") {
			capabilities = append(capabilities, "deep_research")
		}
		if containsAny(labels, "创建图片", "create image", "image creation") {
			capabilities = append(capabilities, "image_generation")
		}
	}
	return PageState{
		Authenticated:   authenticated,
		PromptAvailable: authenticated,
		Locale:          snapshot.Locale,
		Capabilities:    capabilities,
		URL:             "https://chatgpt.com/",
	}
}

func containsAny(text string, values ...string) bool {
	for _, value := range values {
		if strings.Contains(text, strings.ToLower(value)) {
			return true
		}
	}
	return false
}
