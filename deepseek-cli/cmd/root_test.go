package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
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

func TestLoginStatusReportsDaemonStatusWhenBridgeReady(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"running":true,"extension_connected":true,"extension_version":"1.11.3","version":"v1.10.3"}`))
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"--webbridge-url", server.URL, "login-status"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var envelope struct {
		OK   bool `json:"ok"`
		Data struct {
			BridgeReady      bool   `json:"bridge_ready"`
			ExtensionVersion string `json:"extension_version"`
			LoginState       string `json:"login_state"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	if !envelope.OK || !envelope.Data.BridgeReady || envelope.Data.ExtensionVersion != "1.11.3" || envelope.Data.LoginState != "unchecked" {
		t.Fatalf("envelope = %#v", envelope)
	}
}

func TestLoginStatusReadsAuthenticatedDeepSeekPageFromCDP(t *testing.T) {
	serverURL, cleanup := fakeDeepSeekCDP(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"--cdp-url", serverURL, "login-status"}, &stdout, &stderr)
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
	if !envelope.OK || envelope.Data.Browser != "cdp" || !envelope.Data.Authenticated || !envelope.Data.PromptAvailable {
		t.Fatalf("envelope = %#v", envelope)
	}
	if !containsString(envelope.Data.Capabilities, "deepthink") || !containsString(envelope.Data.Capabilities, "web_search") {
		t.Fatalf("capabilities = %#v", envelope.Data.Capabilities)
	}
}

func TestCapabilitiesReadsDeepSeekPageFromCDP(t *testing.T) {
	serverURL, cleanup := fakeDeepSeekCDP(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"--cdp-url", serverURL, "capabilities"}, &stdout, &stderr)
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
}

func TestChatAskReadsStableAnswerFromCDP(t *testing.T) {
	serverURL, cleanup := fakeDeepSeekAskCDP(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"--cdp-url", serverURL, "chat", "ask", "--prompt", "hello"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var envelope struct {
		OK   bool `json:"ok"`
		Data struct {
			Prompt string `json:"prompt"`
			Answer string `json:"answer"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	if !envelope.OK || envelope.Data.Prompt != "hello" || envelope.Data.Answer != "stable answer" {
		t.Fatalf("envelope = %#v", envelope)
	}
}

func fakeDeepSeekCDP(t *testing.T) (string, func()) {
	t.Helper()
	var serverURL string
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/json/list":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `[{"type":"page","url":"https://chat.deepseek.com/","title":"DeepSeek","webSocketDebuggerUrl":%q}]`, strings.Replace(serverURL, "http://", "ws://", 1)+"/devtools/page/1")
		case "/devtools/page/1":
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Errorf("upgrade websocket: %v", err)
				return
			}
			defer conn.Close()
			for {
				var req struct {
					ID     int    `json:"id"`
					Method string `json:"method"`
				}
				if err := conn.ReadJSON(&req); err != nil {
					return
				}
				if req.Method == "Runtime.evaluate" {
					_ = conn.WriteJSON(map[string]any{
						"id": req.ID,
						"result": map[string]any{
							"result": map[string]any{
								"value": map[string]any{
									"href":           "https://chat.deepseek.com/",
									"title":          "DeepSeek",
									"bodyText":       "使用快捷模式开始对话 默认模式 专家模式 识图模式 深度思考 联网搜索 上传文件",
									"hasPromptInput": true,
									"inputs": []map[string]any{{
										"tag":         "TEXTAREA",
										"placeholder": "给 DeepSeek 发送消息",
									}},
									"controls": []map[string]any{
										{"text": "深度思考"},
										{"text": "联网搜索"},
										{"text": "上传文件"},
									},
								},
							},
						},
					})
					continue
				}
				_ = conn.WriteJSON(map[string]any{"id": req.ID, "result": map[string]any{}})
			}
		default:
			http.NotFound(w, r)
		}
	}))
	serverURL = server.URL
	return server.URL, server.Close
}

func fakeDeepSeekAskCDP(t *testing.T) (string, func()) {
	t.Helper()
	var serverURL string
	answerReads := 0
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/json/list":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `[{"type":"page","url":"https://chat.deepseek.com/","title":"DeepSeek","webSocketDebuggerUrl":%q}]`, strings.Replace(serverURL, "http://", "ws://", 1)+"/devtools/page/ask")
		case "/devtools/page/ask":
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Errorf("upgrade websocket: %v", err)
				return
			}
			defer conn.Close()
			for {
				var req struct {
					ID     int            `json:"id"`
					Method string         `json:"method"`
					Params map[string]any `json:"params"`
				}
				if err := conn.ReadJSON(&req); err != nil {
					return
				}
				if req.Method == "Runtime.evaluate" {
					expr, _ := req.Params["expression"].(string)
					if strings.Contains(expr, "ds-assistant-message-main-content") {
						answerReads++
						count := 0
						answer := ""
						if answerReads >= 2 {
							count = 1
							answer = "stable answer"
						}
						_ = conn.WriteJSON(map[string]any{
							"id": req.ID,
							"result": map[string]any{
								"result": map[string]any{
									"value": map[string]any{"count": count, "latest": answer},
								},
							},
						})
						continue
					}
					_ = conn.WriteJSON(map[string]any{
						"id": req.ID,
						"result": map[string]any{
							"result": map[string]any{
								"value": map[string]any{"ok": true, "x": 50, "y": 60},
							},
						},
					})
					continue
				}
				_ = conn.WriteJSON(map[string]any{"id": req.ID, "result": map[string]any{}})
			}
		default:
			http.NotFound(w, r)
		}
	}))
	serverURL = server.URL
	return server.URL, server.Close
}

func containsString(values []string, value string) bool {
	for _, got := range values {
		if got == value {
			return true
		}
	}
	return false
}
