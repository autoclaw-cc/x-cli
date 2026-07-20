package notebooklm

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestInspectStudioReturnsExactObservedTypes(t *testing.T) {
	want := []string{"audio", "presentation", "video", "mind_map", "report", "flashcards", "quiz", "infographic", "data_table"}
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"labels": []string{}},
		map[string]any{"labels": []string{"音频概览", "演示文稿", "视频概览", "思维导图", "报告", "闪卡", "测验", "信息图", "数据表格"}},
	}}
	got, err := InspectStudio(context.Background(), b, "https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Types, want) {
		t.Fatalf("types = %#v, want %#v", got.Types, want)
	}
	if !hasCallContaining(b.calls, "notebooklm-studio-tab") || !hasCallContaining(b.calls, "e.click();") {
		t.Fatalf("calls = %#v", b.calls)
	}
}

func TestListStudioArtifactsReturnsTypedReadyItems(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"loading": true, "artifacts": []map[string]any{}},
		map[string]any{"artifacts": []map[string]any{
			{"type": "audio", "title": "CLI Audio", "details": "1:10 · 1 source", "state": "ready", "playable": true, "hasMenu": true},
			{"type": "data_table", "title": "CLI Table", "details": "1 source", "state": "ready", "playable": false, "hasMenu": true},
		}},
		map[string]any{"artifacts": []map[string]any{
			{"type": "audio", "title": "CLI Audio", "details": "1:10 · 1 source", "state": "ready", "playable": true, "hasMenu": true},
			{"type": "data_table", "title": "CLI Table", "details": "1 source", "state": "ready", "playable": false, "hasMenu": true},
		}},
		map[string]any{"artifacts": []map[string]any{
			{"type": "audio", "title": "CLI Audio", "details": "1:10 · 1 source", "state": "ready", "playable": true, "hasMenu": true},
			{"type": "data_table", "title": "CLI Table", "details": "1 source", "state": "ready", "playable": false, "hasMenu": true},
		}},
		map[string]any{"artifacts": []map[string]any{
			{"type": "audio", "title": "CLI Audio", "details": "1:10 · 1 source", "state": "ready", "playable": true, "hasMenu": true},
			{"type": "data_table", "title": "CLI Table", "details": "1 source", "state": "ready", "playable": false, "hasMenu": true},
		}},
	}}

	got, err := ListStudioArtifacts(context.Background(), b,
		"https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Artifacts) != 2 || got.Artifacts[0].Type != "audio" || !got.Artifacts[0].Playable || got.Artifacts[1].Type != "data_table" {
		t.Fatalf("artifacts = %#v", got)
	}
	if !hasCallContaining(b.calls, "audio_magic_eraser") || !hasCallContaining(b.calls, "artifact-details") || !hasCallContaining(b.calls, "artifact-actions") {
		t.Fatalf("artifact semantics missing: %#v", b.calls)
	}
	if !hasCallContaining(b.calls, "trim()==='Studio'") || !hasCallContaining(b.calls, "e.click();") {
		t.Fatalf("Studio tab activation must use DOM click verification: %#v", b.calls)
	}
	if !hasCallContaining(b.calls, "tab.click();return {ready:false") {
		t.Fatalf("Studio tab readiness must retry tab click until selected: %#v", b.calls)
	}
}

func TestGenerateStudioArtifactClicksTypeAndWaitsForReadyIncrement(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"artifacts": []map[string]any{}},
		map[string]any{"artifacts": []map[string]any{}},
		map[string]any{"artifacts": []map[string]any{}},
		map[string]any{"artifacts": []map[string]any{}},
		map[string]any{"ok": true, "label": "思维导图"},
		map[string]any{"active": true, "hasPrompt": true, "canSubmit": true},
		map[string]any{"artifacts": []map[string]any{{"type": "mind_map", "title": "CLI Mind Map", "details": "1 source", "state": "generating"}}},
		map[string]any{"artifacts": []map[string]any{{"type": "mind_map", "title": "CLI Mind Map", "details": "1 source", "state": "ready"}}},
	}}

	got, err := GenerateStudioArtifact(context.Background(), b,
		"https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532",
		"mind_map", "CLI topic", true, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != "mind_map" || got.Title != "CLI Mind Map" || got.State != "ready" || got.ArtifactCount != 1 {
		t.Fatalf("generated = %#v", got)
	}
	if !hasCall(b.calls, "click:#notebooklm-create-studio-artifact") ||
		!hasCall(b.calls, "fill:#notebooklm-studio-prompt:CLI topic") ||
		!hasCall(b.calls, "click:#notebooklm-studio-submit") {
		t.Fatalf("calls = %#v", b.calls)
	}
}

func TestGenerateStudioArtifactSelectsDialogDefaultOptionBeforeSubmit(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"artifacts": []map[string]any{}},
		map[string]any{"artifacts": []map[string]any{}},
		map[string]any{"artifacts": []map[string]any{}},
		map[string]any{"artifacts": []map[string]any{}},
		map[string]any{"ok": true, "label": "报告"},
		map[string]any{"active": true, "hasPrompt": false, "canSubmit": false, "hasDefaultOption": true},
		map[string]any{"active": true, "hasPrompt": false, "canSubmit": true, "hasDefaultOption": false},
		map[string]any{"artifacts": []map[string]any{{"type": "report", "title": "CLI Report", "details": "1 source", "state": "ready"}}},
	}}

	got, err := GenerateStudioArtifact(context.Background(), b,
		"https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532",
		"report", "", true, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != "report" || got.Title != "CLI Report" {
		t.Fatalf("generated = %#v", got)
	}
	if !hasCall(b.calls, "click:#notebooklm-studio-default-option") || !hasCall(b.calls, "click:#notebooklm-studio-submit") {
		t.Fatalf("dialog option flow missing: %#v", b.calls)
	}
}

func TestGenerateStudioArtifactTreatsDefaultOptionAutoStartAsSubmitted(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"artifacts": []map[string]any{}},
		map[string]any{"artifacts": []map[string]any{}},
		map[string]any{"artifacts": []map[string]any{}},
		map[string]any{"artifacts": []map[string]any{}},
		map[string]any{"ok": true, "label": "报告"},
		map[string]any{"active": true, "hasPrompt": false, "canSubmit": false, "hasDefaultOption": true},
		map[string]any{"active": false, "hasPrompt": false, "canSubmit": false, "hasDefaultOption": false},
		map[string]any{"artifacts": []map[string]any{{"type": "report", "title": "CLI Report", "details": "1 source", "state": "ready"}}},
	}}

	got, err := GenerateStudioArtifact(context.Background(), b,
		"https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532",
		"report", "", true, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Submitted {
		t.Fatalf("default-option autostart should be reported as submitted: %#v", got)
	}
}

func TestGenerateStudioArtifactFallsBackToKeyTypeWhenPromptFillFails(t *testing.T) {
	b := &scriptedBridge{fillErr: errScriptedFill, evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"artifacts": []map[string]any{}},
		map[string]any{"artifacts": []map[string]any{}},
		map[string]any{"artifacts": []map[string]any{}},
		map[string]any{"artifacts": []map[string]any{}},
		map[string]any{"ok": true, "label": "数据表格"},
		map[string]any{"active": true, "hasPrompt": true, "canSubmit": true},
		map[string]any{"artifacts": []map[string]any{{"type": "data_table", "title": "CLI Table", "details": "1 source", "state": "ready"}}},
	}}

	got, err := GenerateStudioArtifact(context.Background(), b,
		"https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532",
		"data_table", "CLI prompt", false, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != "data_table" || got.Title != "CLI Table" {
		t.Fatalf("generated = %#v", got)
	}
	if !hasCall(b.calls, "mouse_click:#notebooklm-studio-prompt") || !hasCall(b.calls, "key_type:CLI prompt") {
		t.Fatalf("prompt fallback missing: %#v", b.calls)
	}
	if !hasCall(b.calls, "click:#notebooklm-create-studio-artifact") || !hasCall(b.calls, "click:#notebooklm-studio-submit") {
		t.Fatalf("Studio generation must use WebBridge click action: %#v", b.calls)
	}
}
