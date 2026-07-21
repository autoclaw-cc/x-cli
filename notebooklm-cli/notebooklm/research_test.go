package notebooklm

import (
	"context"
	"testing"
	"time"
)

func TestFillResearchQueryTargetsEnabledDiscoveryBox(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
	}}
	if err := fillResearchQuery(context.Background(), b, "Example Domain", time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if !hasCallContaining(b.calls, "基于输入的查询发现来源") || !hasCallContaining(b.calls, "!e.disabled") {
		t.Fatalf("research query must target the enabled discovery query box, calls = %#v", b.calls)
	}
	if !hasCall(b.calls, "mouse_click:#notebooklm-research-query") || !hasCallContaining(b.calls, "cdp:Input.insertText") {
		t.Fatalf("research query must be entered through browser text insertion, calls = %#v", b.calls)
	}
}

func TestClickResearchSubmitUsesEnabledSubmitButton(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true, "disabled": false},
	}}
	if err := clickResearchSubmit(context.Background(), b, time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if !hasCall(b.calls, "click:#notebooklm-research-submit") {
		t.Fatalf("submit click missing: %#v", b.calls)
	}
	if !hasCallContaining(b.calls, "submit-button") || !hasCallContaining(b.calls, "!e.disabled") {
		t.Fatalf("research submit must use an enabled submit control, calls = %#v", b.calls)
	}
}

func TestClearResearchResultDeletesExistingDiscoveryResult(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"present": true},
		map[string]any{"ready": true},
	}}
	if err := clearResearchResultIfPresent(context.Background(), b, time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if !hasCallContaining(b.calls, "source-discovery-completed-action-delete-button") || !hasCallContaining(b.calls, "scrollIntoView") || !hasCall(b.calls, "click:#notebooklm-research-delete") {
		t.Fatalf("existing research result must be cleared through its delete action: %#v", b.calls)
	}
}

func TestSelectResearchModeConfirmsSelectedMode(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ready": true, "selected": false},
		map[string]any{"ok": false},
		map[string]any{"ok": true},
		map[string]any{"ready": true, "selected": false, "menuOpen": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true, "selected": true},
	}}
	if err := selectResearchMode(context.Background(), b, "deep", time.Now().Add(5*time.Second)); err != nil {
		t.Fatal(err)
	}
	if !hasCall(b.calls, "click:#notebooklm-research-mode-choice") {
		t.Fatalf("research mode choice click missing: %#v", b.calls)
	}
	if !hasCall(b.calls, "click:#notebooklm-research-mode-menu") {
		t.Fatalf("research mode menu must be opened through WebBridge click: %#v", b.calls)
	}
	if !hasCallContaining(b.calls, "selected") || !hasCallContaining(b.calls, "Deep Research") {
		t.Fatalf("research mode must be confirmed after click: %#v", b.calls)
	}
}

func TestSelectResearchModeClosesAlreadySelectedOpenMenu(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ready": true, "selected": true, "menuOpen": true},
		map[string]any{"ready": true, "selected": true, "menuOpen": false},
	}}
	if err := selectResearchMode(context.Background(), b, "deep", time.Now().Add(5*time.Second)); err != nil {
		t.Fatal(err)
	}
	if !hasCall(b.calls, "send_keys:Escape") {
		t.Fatalf("open selected menu should be closed with Escape: %#v", b.calls)
	}
	if hasCall(b.calls, "click:#notebooklm-research-mode-choice") {
		t.Fatalf("already selected mode must not click menu choice again: %#v", b.calls)
	}
}
