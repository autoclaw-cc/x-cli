package deepseek

import "strings"

type PageSnapshot struct {
	Href           string            `json:"href"`
	Title          string            `json:"title"`
	BodyText       string            `json:"bodyText"`
	HasPromptInput bool              `json:"hasPromptInput"`
	Inputs         []InputSnapshot   `json:"inputs"`
	Controls       []ControlSnapshot `json:"controls"`
}

type InputSnapshot struct {
	Tag         string `json:"tag"`
	Role        string `json:"role"`
	Type        string `json:"type"`
	Aria        string `json:"aria"`
	Placeholder string `json:"placeholder"`
	Text        string `json:"text"`
	Disabled    bool   `json:"disabled"`
	Class       string `json:"cls"`
}

type ControlSnapshot struct {
	Tag      string `json:"tag"`
	Role     string `json:"role"`
	Aria     string `json:"aria"`
	Text     string `json:"text"`
	Disabled bool   `json:"disabled"`
	Class    string `json:"cls"`
}

type PageState struct {
	Authenticated   bool     `json:"authenticated"`
	PromptAvailable bool     `json:"prompt_available"`
	Capabilities    []string `json:"capabilities"`
	URL             string   `json:"url"`
	Title           string   `json:"title"`
}

func AnalyzePage(snapshot PageSnapshot) PageState {
	text := strings.ToLower(strings.Join([]string{
		snapshot.Href,
		snapshot.Title,
		snapshot.BodyText,
		controlText(snapshot.Controls),
		inputText(snapshot.Inputs),
	}, " "))
	loginScreen := containsAny(text, "登录", "log in", "sign in", "手机号", "验证码")
	promptAvailable := snapshot.HasPromptInput || containsAny(text, "给 deepseek 发送消息", "message deepseek")
	capabilities := []string{}
	if promptAvailable {
		capabilities = append(capabilities, "chat")
	}
	if containsAny(text, "深度思考", "deepthink", "deep think", "r1") {
		capabilities = append(capabilities, "deepthink")
	}
	if containsAny(text, "联网搜索", "智能搜索", "联网", "web search", "search") {
		capabilities = append(capabilities, "web_search")
	}
	if containsAny(text, "上传", "文件", "attach", "upload") {
		capabilities = append(capabilities, "file_upload")
	}
	if containsAny(text, "识图", "vision", "image") {
		capabilities = append(capabilities, "vision")
	}
	return PageState{
		Authenticated:   strings.Contains(snapshot.Href, "chat.deepseek.com") && promptAvailable && !loginScreen,
		PromptAvailable: promptAvailable && !loginScreen,
		Capabilities:    unique(capabilities),
		URL:             snapshot.Href,
		Title:           snapshot.Title,
	}
}

func controlText(controls []ControlSnapshot) string {
	parts := make([]string, 0, len(controls)*2)
	for _, control := range controls {
		parts = append(parts, control.Text, control.Aria)
	}
	return strings.Join(parts, " ")
}

func inputText(inputs []InputSnapshot) string {
	parts := make([]string, 0, len(inputs)*3)
	for _, input := range inputs {
		parts = append(parts, input.Placeholder, input.Aria, input.Text)
	}
	return strings.Join(parts, " ")
}

func containsAny(text string, values ...string) bool {
	for _, value := range values {
		if strings.Contains(text, strings.ToLower(value)) {
			return true
		}
	}
	return false
}

func unique(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}
