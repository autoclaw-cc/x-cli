package browser

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusUsesConfiguredBaseURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/status" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"running": true, "extension_connected": true, "version": "v1.10.3",
		})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-session")
	status, err := c.Status()
	if err != nil {
		t.Fatal(err)
	}
	if !status.Running || !status.ExtensionConnected {
		t.Fatalf("status = %#v", status)
	}
}

func TestCallSendsSessionAndReturnsData(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Action  string         `json:"action"`
			Session string         `json:"session"`
			Args    map[string]any `json:"args"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Action != "close_session" || body.Session != "s-1" {
			t.Fatalf("request = %#v", body)
		}
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": map[string]any{"closed": 1}})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "s-1")
	if err := c.CloseSession(); err != nil {
		t.Fatal(err)
	}
}

func TestEvaluateValueUnwrapsValue(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"ok":   true,
			"data": map[string]any{"type": "object", "value": map[string]any{"loggedIn": true}},
		})
	}))
	defer ts.Close()
	c := NewClient(ts.URL, "s")
	var got struct {
		LoggedIn bool `json:"loggedIn"`
	}
	if err := c.EvaluateValue("({loggedIn:true})", &got); err != nil {
		t.Fatal(err)
	}
	if !got.LoggedIn {
		t.Fatal("loggedIn = false")
	}
}

func TestCallReturnsDaemonError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": map[string]any{"code": "extension_error", "message": "not connected"},
		})
	}))
	defer ts.Close()
	c := NewClient(ts.URL, "s")
	if _, err := c.Call("snapshot", map[string]any{}); err == nil {
		t.Fatal("expected error")
	}
}
