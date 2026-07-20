package notebooklm

import (
	"context"
	"testing"
)

func TestLoginStatusReturnsAllowlistedState(t *testing.T) {
	b := &scriptedBridge{evals: []any{map[string]any{
		"loggedIn": true, "locale": "zh-CN", "planLabel": "", "captchaVisible": false,
	}}}
	got, err := CheckLogin(context.Background(), b)
	if err != nil {
		t.Fatal(err)
	}
	if !got.LoggedIn || got.Locale != "zh-CN" || got.CaptchaVisible {
		t.Fatalf("status = %#v", got)
	}
	if !hasCall(b.calls, "navigate:https://notebooklm.google.com/:true:NotebookLM CLI") {
		t.Fatalf("calls = %#v", b.calls)
	}
}

func TestAccountCapabilitiesReturnsObservedNames(t *testing.T) {
	b := &scriptedBridge{evals: []any{map[string]any{
		"loggedIn": true,
		"controls": []string{"new_notebook", "fast_research", "deep_research"},
	}}}
	got, err := InspectAccountCapabilities(context.Background(), b)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Controls) != 3 || got.Controls[2] != "deep_research" {
		t.Fatalf("capabilities = %#v", got)
	}
}
