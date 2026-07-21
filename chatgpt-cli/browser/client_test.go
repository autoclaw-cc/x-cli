package browser

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusUsesConfiguredDaemonURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/status" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"running":true,"extension_connected":true,"extension_version":"1.9.13","version":"v1.10.3"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL+"/", "chatgpt-test")
	got, err := client.Status()
	if err != nil {
		t.Fatal(err)
	}
	if !got.Running || !got.ExtensionConnected || got.ExtensionVersion != "1.9.13" {
		t.Fatalf("status = %#v", got)
	}
}

func TestCommandsCarrySessionAndArguments(t *testing.T) {
	var calls []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/command" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var call map[string]any
		if err := json.NewDecoder(r.Body).Decode(&call); err != nil {
			t.Fatal(err)
		}
		calls = append(calls, call)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"data":{"type":"string","value":"{}"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "chatgpt-session")
	if err := client.FindTab("https://chatgpt.com/", true); err != nil {
		t.Fatal(err)
	}
	if err := client.Navigate("https://chatgpt.com/", true); err != nil {
		t.Fatal(err)
	}
	if err := client.Fill("#prompt-textarea", "hello"); err != nil {
		t.Fatal(err)
	}
	if err := client.Click("[data-testid=send-button]"); err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := client.EvaluateValue("JSON.stringify({ok:true})", &value); err != nil {
		t.Fatal(err)
	}
	if err := client.CloseSession(); err != nil {
		t.Fatal(err)
	}

	if len(calls) != 6 {
		t.Fatalf("calls = %d", len(calls))
	}
	for _, call := range calls {
		if call["session"] != "chatgpt-session" {
			t.Fatalf("session = %#v", call["session"])
		}
	}
	if calls[0]["action"] != "find_tab" || calls[5]["action"] != "close_session" {
		t.Fatalf("actions = %#v ... %#v", calls[0]["action"], calls[5]["action"])
	}
}

func TestEvaluateValueParsesJSONStringResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"data":{"type":"string","value":"{\"answer\":\"ok\"}"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "chatgpt-test")
	var got struct {
		Answer string `json:"answer"`
	}
	if err := client.EvaluateValue("ignored", &got); err != nil {
		t.Fatal(err)
	}
	if got.Answer != "ok" {
		t.Fatalf("answer = %q", got.Answer)
	}
}
