package deepseek

import "testing"

func TestAnalyzePageReportsAuthenticatedPromptAndCapabilities(t *testing.T) {
	snapshot := PageSnapshot{
		Href:           "https://chat.deepseek.com/",
		Title:          "DeepSeek - 探索未至之境",
		BodyText:       "使用快捷模式开始对话 默认模式 专家模式 识图模式 深度思考 联网搜索",
		HasPromptInput: true,
		Inputs: []InputSnapshot{
			{Tag: "TEXTAREA", Placeholder: "给 DeepSeek 发送消息"},
		},
		Controls: []ControlSnapshot{
			{Text: "深度思考"},
			{Text: "联网搜索"},
			{Text: "上传文件"},
		},
	}

	got := AnalyzePage(snapshot)
	if !got.Authenticated {
		t.Fatalf("Authenticated = false, state = %#v", got)
	}
	if !got.PromptAvailable {
		t.Fatalf("PromptAvailable = false, state = %#v", got)
	}
	want := []string{"chat", "deepthink", "web_search", "file_upload", "vision"}
	for _, capability := range want {
		if !contains(got.Capabilities, capability) {
			t.Fatalf("Capabilities missing %q: %#v", capability, got.Capabilities)
		}
	}
}

func TestAnalyzePageDoesNotTreatLoginScreenAsAuthenticated(t *testing.T) {
	snapshot := PageSnapshot{
		Href:     "https://chat.deepseek.com/sign_in",
		Title:    "DeepSeek",
		BodyText: "登录 手机号 验证码",
	}

	got := AnalyzePage(snapshot)
	if got.Authenticated {
		t.Fatalf("Authenticated = true for login page: %#v", got)
	}
	if got.PromptAvailable {
		t.Fatalf("PromptAvailable = true for login page: %#v", got)
	}
}

func TestAnalyzePageTreatsSmartSearchAsWebSearch(t *testing.T) {
	snapshot := PageSnapshot{
		Href:           "https://chat.deepseek.com/",
		Title:          "DeepSeek",
		BodyText:       "快速模式 专家模式 识图模式 深度思考 智能搜索",
		HasPromptInput: true,
	}

	got := AnalyzePage(snapshot)
	if !contains(got.Capabilities, "web_search") {
		t.Fatalf("Capabilities missing web_search for smart-search label: %#v", got.Capabilities)
	}
}

func contains(values []string, value string) bool {
	for _, got := range values {
		if got == value {
			return true
		}
	}
	return false
}
