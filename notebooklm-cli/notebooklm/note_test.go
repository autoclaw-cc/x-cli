package notebooklm

import (
	"context"
	"testing"
	"time"
)

func TestCreateNoteVerifiesReopenedTitleAndBody(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true, "noteCount": 0},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true, "title": "CLI NOTE"},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true, "noteCount": 1},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
	}}

	got, err := CreateNote(context.Background(), b,
		"https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532",
		"CLI NOTE", "LAPIS-5502", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "CLI NOTE" || got.BodyCharacters != 10 {
		t.Fatalf("note = %#v", got)
	}
	if !hasCall(b.calls, "fill:.ProseMirror:LAPIS-5502") ||
		!hasCall(b.calls, "mouse_click:#notebooklm-add-note") ||
		!hasCall(b.calls, "mouse_click:#notebooklm-reopen-note") {
		t.Fatalf("calls = %#v", b.calls)
	}
	if !hasCallContaining(b.calls, "HTMLInputElement.prototype") || !hasCallContaining(b.calls, "artifact-library-note") {
		t.Fatalf("note persistence guards missing: %#v", b.calls)
	}
}

func TestListNotesReturnsExactVisibleTitles(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"loading": true, "notes": []map[string]any{}},
		map[string]any{"notes": []map[string]any{{"title": "CLI NOTE A"}, {"title": "CLI NOTE B"}}},
		map[string]any{"notes": []map[string]any{{"title": "CLI NOTE A"}, {"title": "CLI NOTE B"}}},
		map[string]any{"notes": []map[string]any{{"title": "CLI NOTE A"}, {"title": "CLI NOTE B"}}},
		map[string]any{"notes": []map[string]any{{"title": "CLI NOTE A"}, {"title": "CLI NOTE B"}}},
	}}

	got, err := ListNotes(context.Background(), b,
		"https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Notes) != 2 || got.Notes[0].Title != "CLI NOTE A" || got.Notes[1].Title != "CLI NOTE B" {
		t.Fatalf("notes = %#v", got)
	}
}
