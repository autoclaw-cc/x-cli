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
	if !hasCall(b.calls, "mouse_click:#notebooklm-studio-tab") {
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
}
