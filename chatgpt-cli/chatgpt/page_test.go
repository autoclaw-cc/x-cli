package chatgpt

import "testing"

func TestAnalyzePageReturnsAllowlistedAuthenticatedCapabilities(t *testing.T) {
	got := AnalyzePage(PageSnapshot{
		Href:          "https://chatgpt.com/",
		Locale:        "zh-CN",
		HasPrompt:     true,
		LoginControls: false,
		ToolLabels: []string{
			"创建图片 可视化呈现任何内容",
			"网页搜索 查找实时新闻和信息",
			"深度研究 获取详细报告",
		},
	})

	if !got.Authenticated || !got.PromptAvailable {
		t.Fatalf("state = %#v", got)
	}
	if got.URL != "https://chatgpt.com/" || got.Locale != "zh-CN" {
		t.Fatalf("state = %#v", got)
	}
	for _, capability := range []string{"chat", "web_search", "deep_research", "image_generation"} {
		if !contains(got.Capabilities, capability) {
			t.Fatalf("capabilities missing %q: %#v", capability, got.Capabilities)
		}
	}
}

func TestAnalyzePageRejectsLoginScreenAndDropsPrivateFields(t *testing.T) {
	got := AnalyzePage(PageSnapshot{
		Href:          "https://chatgpt.com/auth/login",
		Locale:        "en-US",
		HasPrompt:     true,
		LoginControls: true,
		ToolLabels:    []string{"Web search"},
	})

	if got.Authenticated || got.PromptAvailable {
		t.Fatalf("state = %#v", got)
	}
	if len(got.Capabilities) != 0 {
		t.Fatalf("capabilities = %#v", got.Capabilities)
	}
}

func TestAnalyzePageRecognizesEnglishLabels(t *testing.T) {
	got := AnalyzePage(PageSnapshot{
		Href:          "https://chatgpt.com/",
		Locale:        "en-US",
		HasPrompt:     true,
		LoginControls: false,
		ToolLabels:    []string{"Create image", "Web search", "Deep research"},
	})
	for _, capability := range []string{"chat", "web_search", "deep_research", "image_generation"} {
		if !contains(got.Capabilities, capability) {
			t.Fatalf("capabilities missing %q: %#v", capability, got.Capabilities)
		}
	}
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
