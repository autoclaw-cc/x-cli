package notebooklm

import (
	"context"
	"testing"
	"time"
)

func TestAskReturnsLastGroundedAnswerAndCitations(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ok": true, "disabled": false},
		map[string]any{"done": true, "answer": "ORCHID-7421, 42 minutes.", "citations": 1},
	}}
	got, err := Ask(context.Background(), b, "https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532", "What is the phrase?", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got.Answer != "ORCHID-7421, 42 minutes." || got.Citations != 1 {
		t.Fatalf("answer = %#v", got)
	}
	if !hasCall(b.calls, "key_type:What is the phrase?") || !hasCall(b.calls, "mouse_click:#notebooklm-chat-submit") {
		t.Fatalf("calls = %#v", b.calls)
	}
}
