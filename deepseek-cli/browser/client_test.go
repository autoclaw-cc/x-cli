package browser

import (
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
		_, _ = w.Write([]byte(`{"running":true,"extension_connected":true,"extension_version":"1.11.3","version":"v1.10.3"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "deepseek-test")
	got, err := client.Status()
	if err != nil {
		t.Fatal(err)
	}
	if !got.Running || !got.ExtensionConnected || got.ExtensionVersion != "1.11.3" {
		t.Fatalf("status = %#v", got)
	}
}
