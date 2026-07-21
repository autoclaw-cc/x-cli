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
			Browser         string   `json:"browser"`
			Authenticated   bool     `json:"authenticated"`
			PromptAvailable bool     `json:"prompt_available"`
			Locale          string   `json:"locale"`
			URL             string   `json:"url"`
			Capabilities    []string `json:"capabilities"`
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

func TestChatNewStartsFreshOwnedConversation(t *testing.T) {
	server, state := fakeAskServer(t)
	defer server.Close()

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"--webbridge-url", server.URL, "chat", "new"}, &stdout, &stderr)
	if code != 0 || !strings.Contains(stdout.String(), `"started":true`) {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if state.navigateCalls != 1 || containsAction(state.actions, "close_session") {
		t.Fatalf("state = %#v", state)
	}
}

func TestChatAskSubmitsExactlyOnceAndReturnsStableAnswer(t *testing.T) {
	server, state := fakeAskServer(t)
	defer server.Close()

	var stdout, stderr bytes.Buffer
	code := Execute([]string{
		"--webbridge-url", server.URL,
		"--timeout", "3s",
		"chat", "ask",
		"--prompt", "Return OAK-420",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if state.submitCalls != 1 || state.fillValue != "Return OAK-420" {
		t.Fatalf("state = %#v", state)
	}
	if !strings.Contains(stdout.String(), `"answer":"OAK-420"`) || !strings.Contains(stdout.String(), `"stable_samples":3`) {
		t.Fatalf("stdout = %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "https://example.com/source") {
		t.Fatalf("citation missing: %s", stdout.String())
	}
}

func TestChatAskSelectsEachResearchModeOnce(t *testing.T) {
	for _, test := range []struct {
		name string
		flag string
		mode string
	}{
		{name: "web search", flag: "--search", mode: "web_search"},
		{name: "deep research", flag: "--deep-research", mode: "deep_research"},
	} {
		t.Run(test.name, func(t *testing.T) {
			server, state := fakeAskServer(t)
			defer server.Close()
			var stdout, stderr bytes.Buffer
			code := Execute([]string{
				"--webbridge-url", server.URL,
				"--timeout", "3s",
				"chat", "ask", "--prompt", "research", test.flag,
			}, &stdout, &stderr)
			if code != 0 {
				t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
			}
			if len(state.selectedModes) != 1 || state.selectedModes[0] != test.mode {
				t.Fatalf("selected modes = %#v", state.selectedModes)
			}
			if !strings.Contains(stdout.String(), `"mode":"`+test.mode+`"`) {
				t.Fatalf("stdout = %s", stdout.String())
			}
		})
	}
}

func TestChatAskRejectsConflictingModesBeforeBrowserWork(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Execute([]string{"chat", "ask", "--prompt", "conflict", "--search", "--deep-research"}, &stdout, &stderr)
	if code == 0 || !strings.Contains(stdout.String(), "modes_conflict") {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestImageGenerateRequiresExplicitConfirmation(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Execute([]string{"image", "generate", "--prompt", "a test image"}, &stdout, &stderr)
	if code == 0 || !strings.Contains(stdout.String(), "confirmation_required") {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestImageGenerateWritesVerifiedImageBytes(t *testing.T) {
	server, state := fakeImageServer(t)
	defer server.Close()
	outDir := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := Execute([]string{
		"--webbridge-url", server.URL,
		"--timeout", "3s",
		"image", "generate",
		"--prompt", "a blue square",
		"--out", outDir,
		"--confirm",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if state.submitCalls != 1 || state.fillValue != "a blue square" {
		t.Fatalf("state=%#v", state)
	}
	var envelope struct {
		Data struct {
			Path  string `json:"path"`
			Bytes int    `json:"bytes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(envelope.Data.Path) != outDir || envelope.Data.Bytes != 4 {
		t.Fatalf("data=%#v", envelope.Data)
	}
	got, err := os.ReadFile(envelope.Data.Path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string([]byte{0x89, 'P', 'N', 'G'}) {
		t.Fatalf("bytes=%v", got)
	}
}

type imageServerState struct {
	submitCalls int
	fillValue   string
}

func fakeImageServer(t *testing.T) (*httptest.Server, *imageServerState) {
	t.Helper()
	state := &imageServerState{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/status" {
			_, _ = w.Write([]byte(`{"running":true,"extension_connected":true}`))
			return
		}
		var call struct {
			Action string `json:"action"`
			Args   struct {
				Code  string `json:"code"`
				Value string `json:"value"`
			} `json:"args"`
		}
		if err := json.NewDecoder(r.Body).Decode(&call); err != nil {
			t.Fatal(err)
		}
		switch call.Action {
		case "fill":
			state.fillValue = call.Args.Value
			_, _ = w.Write([]byte(`{"ok":true,"data":{"success":true}}`))
		case "evaluate":
			value := `{"ok":true}`
			switch {
			case strings.Contains(call.Args.Code, "chatgptImageReady"):
				value = `{"ready":true}`
			case strings.Contains(call.Args.Code, "chatgptImageBaseline"):
				value = `{"fileIds":["file_old"]}`
			case strings.Contains(call.Args.Code, "chatgptSubmitPrompt"):
				state.submitCalls++
				value = `{"ok":true}`
			case strings.Contains(call.Args.Code, "chatgptImageSnapshot"):
				value = `{"images":[{"src":"https://chatgpt.com/backend-api/estuary/content?id=file_new","alt":"Generated image","fileId":"file_new","width":1024,"height":1024,"complete":true}]}`
			case strings.Contains(call.Args.Code, "chatgptDownloadImage"):
				value = `{"ok":true,"contentType":"image/png","base64":"iVBORw=="}`
			}
			payload, _ := json.Marshal(map[string]any{"type": "string", "value": value})
			response, _ := json.Marshal(map[string]any{"ok": true, "data": json.RawMessage(payload)})
			_, _ = w.Write(response)
		default:
			_, _ = w.Write([]byte(`{"ok":true,"data":{"success":true}}`))
		}
	}))
	return server, state
}

type askServerState struct {
	actions       []string
	navigateCalls int
	submitCalls   int
	answerPolls   int
	fillValue     string
	selectedModes []string
}

func fakeAskServer(t *testing.T) (*httptest.Server, *askServerState) {
	t.Helper()
	state := &askServerState{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/status" {
			_, _ = w.Write([]byte(`{"running":true,"extension_connected":true}`))
			return
		}
		var call struct {
			Action string `json:"action"`
			Args   struct {
				Code  string `json:"code"`
				Value string `json:"value"`
			} `json:"args"`
		}
		if err := json.NewDecoder(r.Body).Decode(&call); err != nil {
			t.Fatal(err)
		}
		state.actions = append(state.actions, call.Action)
		switch call.Action {
		case "navigate":
			state.navigateCalls++
			_, _ = w.Write([]byte(`{"ok":true,"data":{"success":true}}`))
		case "fill":
			state.fillValue = call.Args.Value
			_, _ = w.Write([]byte(`{"ok":true,"data":{"success":true}}`))
		case "evaluate":
			value := `{"ok":true}`
			switch {
			case strings.Contains(call.Args.Code, "chatgptPageReady"):
				value = `{"ready":true}`
			case strings.Contains(call.Args.Code, "chatgptSelectMode"):
				mode := "web_search"
				if strings.Contains(call.Args.Code, `const mode="deep_research"`) {
					mode = "deep_research"
				}
				state.selectedModes = append(state.selectedModes, mode)
				value = `{"ok":true,"mode":"` + mode + `"}`
			case strings.Contains(call.Args.Code, "chatgptSubmitPrompt"):
				state.submitCalls++
				value = `{"ok":true}`
			case strings.Contains(call.Args.Code, "chatgptAnswerSnapshot"):
				state.answerPolls++
				if state.answerPolls == 1 {
					value = `{"count":0,"latest":"","streaming":false,"citations":[]}`
				} else {
					value = `{"count":1,"latest":"OAK-420","streaming":false,"citations":[{"url":"https://example.com/source","label":"source"}]}`
				}
			}
			payload, _ := json.Marshal(map[string]any{"type": "string", "value": value})
			response, _ := json.Marshal(map[string]any{"ok": true, "data": json.RawMessage(payload)})
			_, _ = w.Write(response)
		default:
			_, _ = w.Write([]byte(`{"ok":true,"data":{"success":true}}`))
		}
	}))
	return server, state
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
