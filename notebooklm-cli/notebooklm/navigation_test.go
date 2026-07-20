package notebooklm

import (
	"context"
	"testing"
	"time"
)

func TestOpenOwnedNotebookUsesStaticSameOriginBootstrap(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
	}}
	url := "https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532"
	if err := openOwnedNotebook(context.Background(), b, url, 2*time.Second); err != nil {
		t.Fatal(err)
	}
	if !hasCall(b.calls, "navigate:https://notebooklm.google.com/robots.txt:true:NotebookLM CLI") {
		t.Fatalf("calls = %#v", b.calls)
	}
	if hasCall(b.calls, "navigate:"+url+":true:NotebookLM CLI") {
		t.Fatalf("direct heavy navigation is forbidden: %#v", b.calls)
	}
}
