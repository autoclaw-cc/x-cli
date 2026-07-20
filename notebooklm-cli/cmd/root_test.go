package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpListsPrimaryCommands(t *testing.T) {
	var out, errOut bytes.Buffer
	code := Execute([]string{"--help"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("code = %d, stderr = %s", code, errOut.String())
	}
	for _, want := range []string{"login-status", "capabilities", "notebook", "source", "chat", "note", "studio"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("help missing %q:\n%s", want, out.String())
		}
	}
}

func TestNoteCreateRejectsUnknownNotebookBeforeBrowser(t *testing.T) {
	var out, errOut bytes.Buffer
	path := filepath.Join(t.TempDir(), "registry.json")
	code := Execute([]string{
		"--registry", path,
		"note", "create", "--notebook", "unknown", "--title", "CLI NOTE", "--text", "body",
	}, &out, &errOut)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(out.String(), `"code":"notebook_not_owned"`) {
		t.Fatalf("output = %s", out.String())
	}
}

func TestAuthorizeRequiresExplicitConfirmation(t *testing.T) {
	var out, errOut bytes.Buffer
	path := filepath.Join(t.TempDir(), "registry.json")
	code := Execute([]string{
		"--registry", path,
		"notebook", "authorize",
		"--url", "https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532",
		"--title", "CLI TEST",
	}, &out, &errOut)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	var envelope struct {
		OK    bool `json:"ok"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.OK || envelope.Error.Code != "confirmation_required" {
		t.Fatalf("envelope = %#v", envelope)
	}
}

func TestAuthorizeThenListOwnedNotebook(t *testing.T) {
	path := filepath.Join(t.TempDir(), "registry.json")
	var authorizeOut, authorizeErr bytes.Buffer
	code := Execute([]string{
		"--registry", path,
		"notebook", "authorize", "--confirm",
		"--url", "https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532",
		"--title", "CLI TEST",
	}, &authorizeOut, &authorizeErr)
	if code != 0 {
		t.Fatalf("authorize code = %d, out = %s, err = %s", code, authorizeOut.String(), authorizeErr.String())
	}

	var listOut, listErr bytes.Buffer
	code = Execute([]string{"--registry", path, "notebook", "list"}, &listOut, &listErr)
	if code != 0 {
		t.Fatalf("list code = %d, out = %s, err = %s", code, listOut.String(), listErr.String())
	}
	if !strings.Contains(listOut.String(), "7471c40e-b33c-4518-b952-3cd786a4e532") || !strings.Contains(listOut.String(), "CLI TEST") {
		t.Fatalf("list output = %s", listOut.String())
	}
}

func TestSourceRejectsUnknownNotebookBeforeBrowser(t *testing.T) {
	var out, errOut bytes.Buffer
	path := filepath.Join(t.TempDir(), "registry.json")
	code := Execute([]string{
		"--registry", path,
		"source", "add-text", "--notebook", "unknown", "--text", "hello",
	}, &out, &errOut)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(out.String(), `"code":"notebook_not_owned"`) {
		t.Fatalf("output = %s", out.String())
	}
}

func TestURLSourceRejectsUnknownNotebookBeforeBrowser(t *testing.T) {
	var out, errOut bytes.Buffer
	path := filepath.Join(t.TempDir(), "registry.json")
	code := Execute([]string{
		"--registry", path,
		"source", "add-url", "--notebook", "unknown", "--url", "https://example.com/",
	}, &out, &errOut)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(out.String(), `"code":"notebook_not_owned"`) {
		t.Fatalf("output = %s", out.String())
	}
}

func TestStudioListRejectsUnknownNotebookBeforeBrowser(t *testing.T) {
	var out, errOut bytes.Buffer
	path := filepath.Join(t.TempDir(), "registry.json")
	code := Execute([]string{
		"--registry", path,
		"studio", "list", "--notebook", "unknown",
	}, &out, &errOut)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(out.String(), `"code":"notebook_not_owned"`) {
		t.Fatalf("output = %s", out.String())
	}
}
