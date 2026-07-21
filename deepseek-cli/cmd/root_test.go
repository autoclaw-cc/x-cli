package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
