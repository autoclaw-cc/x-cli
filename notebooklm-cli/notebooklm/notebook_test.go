package notebooklm

import (
	"context"
	"testing"
	"time"
)

func TestCreateNotebookReturnsStableIDAndExactTitle(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"url": "https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532?addSource=true", "ready": true},
		map[string]any{"title": "CLI TEST"},
	}}
	got, err := CreateNotebook(context.Background(), b, "CLI TEST", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "7471c40e-b33c-4518-b952-3cd786a4e532" || got.Title != "CLI TEST" {
		t.Fatalf("notebook = %#v", got)
	}
	if !hasCall(b.calls, "cdp:Page.bringToFront") {
		t.Fatalf("calls = %#v", b.calls)
	}
}
