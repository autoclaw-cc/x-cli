package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoginStatusUsesWebBridgeAndRedactsIdentity(t *testing.T) {
	server, calls := fakeChatGPTServer(t, pageSnapshotJSON(true))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"--webbridge-url", server.URL, "login-status"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code = %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var envelope struct {
		OK   bool `json:"ok"`
		Data struct {
			Browser          string   `json:"browser"`
			Authenticated    bool     `json:"authenticated"`
			PromptAvailable  bool     `json:"prompt_available"`
			Locale           string   `json:"locale"`
			URL              string   `json:"url"`
			Capabilities     []string `json:"capabilities"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if !envelope.OK || !envelope.Data.Authenticated || envelope.Data.Browser != "webbridge" {
		t.Fatalf("envelope = %#v", envelope)
	}
	if envelope.Data.URL != "https://chatgpt.com/" || envelope.Data.Locale != "zh-CN" {
		t.Fatalf("data = %#v", envelope.Data)
	}
	for _, forbidden := range []string{"email", "account", "cookie", "token", "DWG"} {
		if strings.Contains(strings.ToLower(stdout.String()), strings.ToLower(forbidden)) {
			t.Fatalf("output contains forbidden field %q: %s", forbidden, stdout.String())
		}
	}
	if !containsAction(*calls, "close_session") {
		t.Fatalf("calls did not close session: %#v", *calls)
	}
}

func TestCapabilitiesReturnsObservedTools(t *testing.T) {
	server, _ := fakeChatGPTServer(t, pageSnapshotJSON(true))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"--webbridge-url", server.URL, "capabilities"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code = %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	for _, capability := range []string{"chat", "web_search", "deep_research", "image_generation"} {
		if !strings.Contains(stdout.String(), capability) {
			t.Fatalf("missing %q: %s", capability, stdout.String())
		}
	}
}

func TestLoginStatusReturnsNotLoggedInWithoutIdentity(t *testing.T) {
	server, _ := fakeChatGPTServer(t, pageSnapshotJSON(false))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"--webbridge-url", server.URL, "login-status"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code = %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"authenticated":false`) {
		t.Fatalf("stdout = %s", stdout.String())
	}
}

func TestLoginStatusWaitsForFreshHomeBeforeInspecting(t *testing.T) {
	evaluateCalls := 0
	var actions []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/status" {
			_, _ = w.Write([]byte(`{"running":true,"extension_connected":true}`))
			return
		}
		var call struct {
			Action string `json:"action"`
			Args   struct {
				Code string `json:"code"`
			} `json:"args"`
		}
		if err := json.NewDecoder(r.Body).Decode(&call); err != nil {
			t.Fatal(err)
		}
		actions = append(actions, call.Action)
		if call.Action == "evaluate" {
			evaluateCalls++
			value := `{"ready":false}`
			if evaluateCalls == 2 {
				value = `{"ready":true}`
			}
			if evaluateCalls >= 3 {
				value = pageSnapshotJSON(true)
			}
			payload, _ := json.Marshal(map[string]any{"type": "string", "value": value})
			response, _ := json.Marshal(map[string]any{"ok": true, "data": json.RawMessage(payload)})
			_, _ = w.Write(response)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true,"data":{"success":true}}`))
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"--webbridge-url", server.URL, "--timeout", "2s", "login-status"}, &stdout, &stderr)
	if code != 0 || !strings.Contains(stdout.String(), `"authenticated":true`) {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if len(actions) < 5 || actions[0] != "navigate" || actions[len(actions)-1] != "close_session" {
		t.Fatalf("actions = %#v", actions)
	}
}

func fakeChatGPTServer(t *testing.T, snapshot string) (*httptest.Server, *[]string) {
	t.Helper()
	calls := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/status" {
			_, _ = w.Write([]byte(`{"running":true,"extension_connected":true,"extension_version":"1.9.13","version":"v1.10.3"}`))
			return
		}
		var call struct {
			Action string `json:"action"`
			Args   struct {
				Code string `json:"code"`
			} `json:"args"`
		}
		if err := json.NewDecoder(r.Body).Decode(&call); err != nil {
			t.Fatal(err)
		}
		calls = append(calls, call.Action)
		if call.Action == "evaluate" {
			if strings.Contains(call.Args.Code, "chatgptPageReady") {
				payload, _ := json.Marshal(map[string]any{"type": "string", "value": `{"ready":true}`})
				response, _ := json.Marshal(map[string]any{"ok": true, "data": json.RawMessage(payload)})
				_, _ = w.Write(response)
				return
			}
			payload, _ := json.Marshal(map[string]any{"type": "string", "value": snapshot})
			response, _ := json.Marshal(map[string]any{"ok": true, "data": json.RawMessage(payload)})
			_, _ = w.Write(response)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true,"data":{"success":true}}`))
	}))
	return server, &calls
}

func pageSnapshotJSON(authenticated bool) string {
	value := map[string]any{
		"href":          "https://chatgpt.com/",
		"locale":        "zh-CN",
		"hasPrompt":     true,
		"loginControls": !authenticated,
		"toolLabels":    []string{"创建图片", "网页搜索", "深度研究"},
	}
	raw, _ := json.Marshal(value)
	return string(raw)
}

func containsAction(actions []string, expected string) bool {
	for _, action := range actions {
		if action == expected {
			return true
		}
	}
	return false
}
