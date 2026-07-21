package notebooklm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

func TestListStudioArtifactsInfersGeneratingTypesFromTitles(t *testing.T) {
	items := []map[string]any{
		{"type": "unknown", "title": "正在生成信息图…", "details": "基于 2 个来源", "state": "generating"},
		{"type": "unknown", "title": "正在生成音频概览…", "details": "请过几分钟后再来查看", "state": "generating"},
		{"type": "unknown", "title": "正在生成全面解析视频概览…", "details": "这可能需要一些时间", "state": "generating"},
	}
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"artifacts": items},
		map[string]any{"artifacts": items},
		map[string]any{"artifacts": items},
		map[string]any{"artifacts": items},
	}}

	got, err := ListStudioArtifacts(context.Background(), b,
		"https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"infographic", "audio", "video"}
	for i, wantType := range want {
		if got.Artifacts[i].Type != wantType {
			t.Fatalf("artifact %d type = %q, want %q; all=%#v", i, got.Artifacts[i].Type, wantType, got.Artifacts)
		}
	}
}

func TestWaitGeneratedStudioArtifactInfersGeneratingTypeFromTitle(t *testing.T) {
	cases := []struct {
		kind  string
		title string
	}{
		{kind: "presentation", title: "正在生成演示文稿…"},
		{kind: "infographic", title: "正在生成信息图…"},
		{kind: "audio", title: "正在生成音频概览…"},
		{kind: "video", title: "正在生成视频概览…"},
	}
	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			b := &scriptedBridge{evals: []any{
				map[string]any{"artifacts": []map[string]any{
					{"type": "unknown", "title": tc.title, "details": "基于 2 个来源", "state": "generating"},
				}},
			}}

			got, err := waitGeneratedStudioArtifact(context.Background(), b, tc.kind, 0, false, true, time.Now().Add(2*time.Second))
			if err != nil {
				t.Fatal(err)
			}
			if got.Type != tc.kind || got.State != "generating" || got.ArtifactCount != 1 {
				t.Fatalf("generated = %#v", got)
			}
		})
	}
}

func TestWaitStudioArtifactReadyReturnsNewReadyArtifactAfterGenerating(t *testing.T) {
	oldReady := map[string]any{"type": "video", "title": "Old Video", "details": "6:00", "state": "ready", "playable": true}
	generating := map[string]any{"type": "video", "title": "正在生成视频概览…", "details": "这可能需要一些时间", "state": "generating"}
	newReady := map[string]any{"type": "video", "title": "New Video", "details": "7:00", "state": "ready", "playable": true}
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"artifacts": []map[string]any{generating, oldReady}},
		map[string]any{"artifacts": []map[string]any{generating, oldReady}},
		map[string]any{"artifacts": []map[string]any{newReady, oldReady}},
	}}

	got, err := WaitStudioArtifactReady(context.Background(), b,
		"https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532",
		"video", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "New Video" || !got.Playable {
		t.Fatalf("artifact = %#v", got)
	}
}

func TestExportStudioArtifactOpensUniqueReadyArtifactAndExtractsBody(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{
			"ok": true,
			"artifact": map[string]any{
				"type": "report", "title": "CLI Report", "details": "Briefing Doc", "state": "ready",
			},
		},
		map[string]any{
			"ready": true,
			"body":  "CLI Report\n\nExecutive summary",
		},
	}}

	got, err := ExportStudioArtifact(context.Background(), b,
		"https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532",
		"report", "CLI Report", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got.Artifact.Type != "report" || got.Artifact.Title != "CLI Report" || got.Body != "CLI Report\n\nExecutive summary" {
		t.Fatalf("export = %#v", got)
	}
	if got.BodyCharacters != len([]rune(got.Body)) {
		t.Fatalf("body characters = %d, want %d", got.BodyCharacters, len([]rune(got.Body)))
	}
	if !hasCallContaining(b.calls, "ARTIFACT-LIBRARY-") || !hasCallContaining(b.calls, "artifact-stretched-button") {
		t.Fatalf("export must use artifact library unique card activation: %#v", b.calls)
	}
	if !hasCallContaining(b.calls, "artifact-viewer") || !hasCallContaining(b.calls, "labs-tailwind-structural-element-view-v2") {
		t.Fatalf("export must read structured artifact viewer content: %#v", b.calls)
	}
}

func TestExportStudioArtifactPollsUntilArtifactListIsReady(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": false, "matches": 0},
		map[string]any{
			"ok": true,
			"artifact": map[string]any{
				"type": "report", "title": "CLI Report", "details": "Briefing Doc", "state": "ready",
			},
		},
		map[string]any{
			"ready": true,
			"body":  "CLI Report\n\nExecutive summary",
		},
	}}

	got, err := ExportStudioArtifact(context.Background(), b,
		"https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532",
		"report", "CLI Report", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got.Artifact.Title != "CLI Report" {
		t.Fatalf("export = %#v", got)
	}
}

func TestNormalizeStudioBlocksDropsContainerDuplicate(t *testing.T) {
	blocks := []string{
		"CLI Report",
		"属性详细描述主要用途专供文档示例使用，旨在提供直观的参考。许可要求在文档中使用此域名无需获取额外许可。操作限制严禁将此类域名应用于实际的生产运营或业务流程中。",
		"属性",
		"详细描述",
		"主要用途",
		"专供文档示例使用，旨在提供直观的参考。",
		"许可要求",
		"在文档中使用此域名无需获取额外许可。",
		"操作限制",
		"严禁将此类域名应用于实际的生产运营或业务流程中。",
	}
	got := normalizeStudioBlocks(blocks)
	if strings.Contains(got, "属性详细描述主要用途") {
		t.Fatalf("container duplicate was not removed:\n%s", got)
	}
	if !strings.Contains(got, "CLI Report") || !strings.Contains(got, "许可要求") {
		t.Fatalf("normalized body lost leaf content:\n%s", got)
	}
}

func TestInspectStudioAttributionReturnsPromptAndSources(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true, "artifact": map[string]any{"type": "video", "title": "CLI Video", "details": "2 sources", "state": "ready"}},
		map[string]any{"ok": true},
		map[string]any{"ready": true, "prompt": "Create a video", "sources": []string{"Source A", "Source B"}},
	}}
	got, err := InspectStudioAttribution(context.Background(), b,
		"https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532",
		"video", "CLI Video", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got.Prompt != "Create a video" || len(got.Sources) != 2 || got.SourceCount != 2 {
		t.Fatalf("attribution = %#v", got)
	}
	if !hasCallContaining(b.calls, "查看提示和来源") || !hasCallContaining(b.calls, "SOURCE-ATTRIBUTION-DIALOG") {
		t.Fatalf("attribution flow missing: %#v", b.calls)
	}
}

func TestRenameStudioArtifactUsesInlineTitleInput(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true, "artifact": map[string]any{"type": "report", "title": "Old", "details": "Briefing", "state": "ready"}},
		map[string]any{"ok": true},
		map[string]any{"ok": true, "value": "Old"},
		map[string]any{"ready": true},
	}}
	got, err := RenameStudioArtifact(context.Background(), b,
		"https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532",
		"report", "Old", "New", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Renamed || got.OldTitle != "Old" || got.NewTitle != "New" {
		t.Fatalf("rename = %#v", got)
	}
	if !hasCallContaining(b.calls, "artifact-title-input") || !hasCallContaining(b.calls, "HTMLInputElement.prototype") {
		t.Fatalf("rename must use inline native input setter: %#v", b.calls)
	}
	if !hasCall(b.calls, "send_keys:Enter") {
		t.Fatalf("rename must commit the inline title with Enter: %#v", b.calls)
	}
}

func TestDeleteStudioArtifactRequiresConfirmationAndWaitsRemoval(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true},
		map[string]any{"ready": true},
		map[string]any{"ok": true, "artifact": map[string]any{"type": "report", "title": "Disposable", "details": "Briefing", "state": "ready"}},
		map[string]any{"ok": true},
		map[string]any{"ok": true},
		map[string]any{"removed": true, "artifactCount": 3},
	}}
	got, err := DeleteStudioArtifact(context.Background(), b,
		"https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532",
		"report", "Disposable", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Deleted || got.ArtifactCount != 3 {
		t.Fatalf("delete = %#v", got)
	}
	if !hasCallContaining(b.calls, "确认删除") || !hasCallContaining(b.calls, "Delete") {
		t.Fatalf("delete confirmation flow missing: %#v", b.calls)
	}
	if !hasCallContaining(b.calls, "notebooklm-confirm-delete-dialog") {
		t.Fatalf("delete confirmation must be scoped to the visible dialog: %#v", b.calls)
	}
}

func TestDownloadStudioArtifactFetchesVisibleMediaSource(t *testing.T) {
	media := []byte("media-bytes")
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var start, end int
		if _, err := fmt.Sscanf(r.Header.Get("Range"), "bytes=%d-%d", &start, &end); err != nil ||
			r.Header.Get("Referer") != "https://notebooklm.google.com/" {
			t.Fatalf("media request headers missing: Range=%q Referer=%q", r.Header.Get("Range"), r.Header.Get("Referer"))
		}
		if end >= len(media) {
			end = len(media) - 1
		}
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(media)))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(media[start : end+1])
	}))
	defer server.Close()

	originalAllow := allowStudioMediaHost
	originalClient := studioMediaHTTPClient
	originalChunkSize := studioMediaRangeChunkSize
	originalConcurrency := studioMediaRangeConcurrency
	allowStudioMediaHost = func(string) bool { return true }
	studioMediaHTTPClient = server.Client()
	studioMediaRangeChunkSize = 6
	studioMediaRangeConcurrency = 2
	defer func() {
		allowStudioMediaHost = originalAllow
		studioMediaHTTPClient = originalClient
		studioMediaRangeChunkSize = originalChunkSize
		studioMediaRangeConcurrency = originalConcurrency
	}()

	b := &scriptedBridge{
		evals: []any{
			map[string]any{"ok": true},
			map[string]any{"ready": true},
			map[string]any{"ok": true},
			map[string]any{"ready": true},
			map[string]any{"ok": true, "playbackClicked": true, "artifact": map[string]any{"type": "video", "title": "CLI Video", "details": "6:50", "state": "ready", "playable": true}},
		},
		networks: []any{
			map[string]any{"success": true},
			map[string]any{"requests": []map[string]any{}},
			map[string]any{"requests": []map[string]any{
				{"url": server.URL + "/videoplayback", "method": "GET", "status": 206, "mimeType": "video/mp4"},
			}},
			map[string]any{"success": true},
		},
	}
	outPath := filepath.Join(t.TempDir(), "video.mp4")
	got, err := DownloadStudioArtifact(context.Background(), b,
		"https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532",
		"video", "CLI Video", outPath, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got.OutputPath != outPath || got.BytesWritten != int64(len("media-bytes")) || got.ContentType != "video/mp4" || !got.DownloadStarted {
		t.Fatalf("download = %#v", got)
	}
	body, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "media-bytes" {
		t.Fatalf("media file = %q", string(body))
	}
	if !hasCallContaining(b.calls, "artifact-stretched-button") ||
		!hasCall(b.calls, "click:#notebooklm-download-playback") ||
		!hasCallContaining(b.calls, "network:start") ||
		!hasCallContaining(b.calls, "network:list") {
		t.Fatalf("download must open the scoped artifact playback and inspect network capture: %#v", b.calls)
	}
}

func TestOpenStudioMediaArtifactPlaybackUsesScopedCardPlayButton(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ok": true, "playbackClicked": true, "artifact": map[string]any{"type": "audio", "title": "CLI Audio", "details": "15:21", "state": "ready", "playable": true}},
	}}

	got, playbackClicked, err := openStudioMediaArtifactPlayback(context.Background(), b, "audio", "CLI Audio", "notebooklm-download-playback", time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != "audio" || got.Title != "CLI Audio" || !playbackClicked {
		t.Fatalf("artifact = %#v playbackClicked=%t", got, playbackClicked)
	}
	if !hasCall(b.calls, "click:#notebooklm-download-playback") {
		t.Fatalf("scoped playback click missing: %#v", b.calls)
	}
	if !hasCallContaining(b.calls, "selected.node.querySelectorAll('button')") ||
		!hasCallContaining(b.calls, "play_arrow") {
		t.Fatalf("playback opener should inspect buttons only within the matched artifact: %#v", b.calls)
	}
}

func TestStartStudioMediaPlaybackClicksAudioPlayerButton(t *testing.T) {
	b := &scriptedBridge{evals: []any{
		map[string]any{"ready": true, "tag": "audio", "clickButton": true},
	}}

	if err := startStudioMediaPlayback(context.Background(), b, "audio", time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if !hasCall(b.calls, "click:#notebooklm-media-play-button") {
		t.Fatalf("audio playback button click missing: %#v", b.calls)
	}
	if !hasCallContaining(b.calls, "播放音频") || !hasCallContaining(b.calls, "Play audio") {
		t.Fatalf("audio fallback should target the visible player button: %#v", b.calls)
	}
	if !hasCallContaining(b.calls, "!inLibrary") {
		t.Fatalf("audio fallback should ignore Studio card playback buttons: %#v", b.calls)
	}
}

func TestDownloadStudioArtifactRejectsNonMediaArtifacts(t *testing.T) {
	b := &scriptedBridge{}
	_, err := DownloadStudioArtifact(context.Background(), b,
		"https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532",
		"report", "CLI Report", filepath.Join(t.TempDir(), "report.bin"), 2*time.Second)
	if err == nil || !strings.Contains(err.Error(), "download_unavailable") {
		t.Fatalf("err = %v", err)
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
