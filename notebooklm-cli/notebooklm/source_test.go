package notebooklm

import (
	"context"
	"testing"
	"time"
)

func TestAddTextSourceUsesPastedTextPath(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true, "sourceCount": 1},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true, "disabled": true},
		map[string]any{"ok": true, "disabled": false},
		map[string]any{"ready": true, "sourceCount": 1},
		map[string]any{"ready": true, "sourceCount": 2},
	}}
	got, err := AddTextSource(context.Background(), b, "https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532", "ORCHID-7421", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got.SourceCount != 2 {
		t.Fatalf("result = %#v", got)
	}
	if !hasCall(b.calls, "mouse_click:textarea[aria-label=\"粘贴的文字\"]") ||
		!hasCall(b.calls, "mouse_click:#notebooklm-sources-tab") ||
		!hasCall(b.calls, "mouse_click:#notebooklm-add-source") ||
		!hasCall(b.calls, "mouse_click:#notebooklm-pasted-text") ||
		!hasCall(b.calls, "key_type:ORCHID-7421") ||
		!hasCall(b.calls, "mouse_click:#notebooklm-insert-source") {
		t.Fatalf("calls = %#v", b.calls)
	}
}

func TestAddURLSourceWaitsForStableSourceIncrement(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true, "sourceCount": 0},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true, "disabled": true},
		map[string]any{"ok": true, "disabled": false},
		map[string]any{"ready": true, "sourceCount": 0},
		map[string]any{"ready": true, "sourceCount": 1},
	}}

	got, err := AddURLSource(context.Background(), b,
		"https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532",
		"https://example.com/", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got.SourceCount != 1 {
		t.Fatalf("result = %#v", got)
	}
	if !hasCall(b.calls, "mouse_click:#notebooklm-url-input") ||
		!hasCall(b.calls, "key_type:https://example.com/") ||
		!hasCall(b.calls, "mouse_click:#notebooklm-website-source") ||
		!hasCall(b.calls, "mouse_click:#notebooklm-insert-url-source") {
		t.Fatalf("calls = %#v", b.calls)
	}
}
