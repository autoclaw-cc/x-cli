package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoginStatusReportsExtensionDisconnectedAsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/status" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"running":true,"extension_connected":false,"version":"v1.10.3"}`))
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"--webbridge-url", server.URL, "login-status"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("exit code = 0, stdout=%s", stdout.String())
	}
	var envelope struct {
		OK    bool `json:"ok"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	if envelope.OK || envelope.Error.Code != "extension_not_connected" {
		t.Fatalf("envelope = %#v", envelope)
	}
	if !strings.Contains(envelope.Error.Message, "Kimi WebBridge extension") {
		t.Fatalf("message should guide extension recovery: %q", envelope.Error.Message)
	}
}

func TestLoginStatusReadsAuthenticatedDeepSeekPageThroughWebBridge(t *testing.T) {
	server, actions := fakeWebBridge(t)
	defer server.Close()

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"--webbridge-url", server.URL, "login-status"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var envelope struct {
		OK   bool `json:"ok"`
		Data struct {
			Browser         string   `json:"browser"`
			Authenticated   bool     `json:"authenticated"`
			PromptAvailable bool     `json:"prompt_available"`
			Capabilities    []string `json:"capabilities"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	if !envelope.OK || envelope.Data.Browser != "webbridge" || !envelope.Data.Authenticated || !envelope.Data.PromptAvailable {
		t.Fatalf("envelope = %#v", envelope)
	}
	if !containsString(envelope.Data.Capabilities, "deepthink") || !containsString(envelope.Data.Capabilities, "web_search") {
		t.Fatalf("capabilities = %#v", envelope.Data.Capabilities)
	}
	assertActions(t, *actions, []string{"find_tab", "evaluate"})
}

func TestCapabilitiesReadsDeepSeekPageThroughWebBridge(t *testing.T) {
	server, actions := fakeWebBridge(t)
	defer server.Close()

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"--webbridge-url", server.URL, "capabilities"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var envelope struct {
		OK   bool `json:"ok"`
		Data struct {
			Capabilities []string `json:"capabilities"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	for _, capability := range []string{"chat", "deepthink", "web_search", "file_upload", "vision"} {
		if !containsString(envelope.Data.Capabilities, capability) {
			t.Fatalf("capabilities missing %q: %#v", capability, envelope.Data.Capabilities)
		}
	}
	assertActions(t, *actions, []string{"find_tab", "evaluate"})
}

func TestChatAskUsesWebBridgeCommands(t *testing.T) {
	server, actions := fakeWebBridge(t)
	defer server.Close()

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"--webbridge-url", server.URL, "chat", "ask", "--prompt", "hello"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var envelope struct {
		OK   bool `json:"ok"`
		Data struct {
			Browser string `json:"browser"`
			Prompt  string `json:"prompt"`
			Answer  string `json:"answer"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	if !envelope.OK || envelope.Data.Browser != "webbridge" || envelope.Data.Prompt != "hello" || envelope.Data.Answer != "stable answer" {
		t.Fatalf("envelope = %#v", envelope)
	}
	assertActions(t, *actions, []string{"find_tab", "cdp", "evaluate", "fill", "evaluate", "evaluate", "evaluate"})
}

func TestChatAskCanEnableModesAndUploadFiles(t *testing.T) {
	server, actions := fakeWebBridge(t)
	defer server.Close()

	attachment := filepath.Join(t.TempDir(), "case.txt")
	image := filepath.Join(t.TempDir(), "diagram.png")
	if err := os.WriteFile(attachment, []byte("case"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(image, []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Execute([]string{
		"--webbridge-url", server.URL,
		"chat", "ask",
		"--prompt", "hello",
		"--deepthink",
		"--search",
		"--file", attachment,
		"--image", image,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var envelope struct {
		OK   bool `json:"ok"`
		Data struct {
			Modes []string `json:"modes"`
			Files []string `json:"files"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	if !envelope.OK || !containsString(envelope.Data.Modes, "deepthink") || !containsString(envelope.Data.Modes, "web_search") || !containsString(envelope.Data.Modes, "vision") {
		t.Fatalf("envelope = %#v", envelope)
	}
	if len(envelope.Data.Files) != 2 {
		t.Fatalf("files = %#v", envelope.Data.Files)
	}
	assertActions(t, *actions, []string{"find_tab", "cdp", "evaluate", "evaluate", "evaluate", "evaluate", "evaluate", "evaluate", "evaluate", "upload", "evaluate", "fill", "evaluate", "evaluate", "evaluate", "evaluate"})
}

func TestChatNewNavigatesTheActiveDeepSeekTabToANewConversation(t *testing.T) {
	server, actions := fakeWebBridge(t)
	defer server.Close()

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"--webbridge-url", server.URL, "chat", "new"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var envelope struct {
		OK   bool `json:"ok"`
		Data struct {
			Started bool `json:"started"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	if !envelope.OK || !envelope.Data.Started {
		t.Fatalf("envelope = %#v", envelope)
	}
	assertActions(t, *actions, []string{"find_tab", "cdp", "navigate"})
}

func fakeWebBridge(t *testing.T) (*httptest.Server, *[]string) {
	t.Helper()
	actions := []string{}
	answerReads := 0
	uploadSeen := false
	submitAttempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/status":
			_, _ = w.Write([]byte(`{"running":true,"extension_connected":true,"extension_version":"1.11.3","version":"v1.10.3"}`))
		case "/command":
			var req struct {
				Action string         `json:"action"`
				Args   map[string]any `json:"args"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode command request: %v", err)
			}
			actions = append(actions, req.Action)
			switch req.Action {
			case "find_tab":
				if active, _ := req.Args["active"].(bool); !active {
					t.Fatalf("find_tab must prefer the active DeepSeek tab: %#v", req.Args)
				}
				writeBridgeData(w, map[string]any{"success": true})
			case "navigate", "fill", "cdp":
				writeBridgeData(w, map[string]any{"success": true})
			case "upload":
				files, _ := req.Args["files"].([]any)
				if len(files) == 0 {
					t.Fatalf("upload missing files: %#v", req.Args)
				}
				uploadSeen = true
				writeBridgeData(w, map[string]any{"success": true, "fileCount": len(files)})
			case "evaluate":
				code, _ := req.Args["code"].(string)
				switch {
				case strings.Contains(code, "deepseekSetModes"):
					if strings.Contains(code, "|on") {
						t.Fatalf("mode active detection must not match the 'on' inside button: %s", code)
					}
					enabled := []string{}
					if strings.Contains(code, "deepthink:true") {
						enabled = append(enabled, "deepthink")
					}
					if strings.Contains(code, "web_search:true") {
						enabled = append(enabled, "web_search")
					}
					if strings.Contains(code, "vision:true") {
						enabled = append(enabled, "vision")
					}
					if containsString(enabled, "vision") && (!strings.Contains(code, "data-model-type=\"vision\"") || !strings.Contains(code, "[aria-pressed]")) {
						t.Fatalf("mode script must use semantic DeepSeek controls: %s", code)
					}
					writeBridgeEvaluate(w, map[string]any{"ok": true, "enabled": enabled})
				case strings.Contains(code, "deepseekModeReady") || strings.Contains(code, "deepseekVisionModeReady"):
					writeBridgeEvaluate(w, map[string]any{"ready": true})
				case strings.Contains(code, "deepseekPrepareUpload"):
					writeBridgeEvaluate(w, map[string]any{"ok": true, "selector": "input[type=file]"})
				case strings.Contains(code, "ds-assistant-message-main-content"):
					answerReads++
					count := 0
					answer := ""
					if answerReads >= 2 {
						count = 1
						answer = "stable answer"
					}
					writeBridgeEvaluate(w, map[string]any{"count": count, "latest": answer})
				case strings.Contains(code, "deepseekSubmitPrompt"):
					if uploadSeen && submitAttempts == 0 {
						submitAttempts++
						writeBridgeEvaluate(w, map[string]any{"ok": false, "error": "send_button_not_ready"})
						return
					}
					writeBridgeEvaluate(w, map[string]any{"ok": true})
				default:
					writeBridgeEvaluate(w, map[string]any{
						"href":           "https://chat.deepseek.com/",
						"title":          "DeepSeek",
						"bodyText":       "使用快捷模式开始对话 默认模式 专家模式 识图模式 深度思考 智能搜索 上传文件",
						"hasPromptInput": true,
						"inputs": []map[string]any{{
							"tag":         "TEXTAREA",
							"placeholder": "给 DeepSeek 发送消息",
						}},
						"controls": []map[string]any{
							{"text": "深度思考"},
							{"text": "智能搜索"},
							{"text": "上传文件"},
						},
					})
				}
			default:
				t.Fatalf("unexpected action %s", req.Action)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	return server, &actions
}

func writeBridgeData(w http.ResponseWriter, data any) {
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": data})
}

func writeBridgeEvaluate(w http.ResponseWriter, value any) {
	writeBridgeData(w, map[string]any{"type": "object", "value": value})
}

func assertActions(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("actions = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("actions = %#v, want %#v", got, want)
		}
	}
}

func containsString(values []string, value string) bool {
	for _, got := range values {
		if got == value {
			return true
		}
	}
	return false
}
