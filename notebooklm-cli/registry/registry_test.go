package registry

import (
	"path/filepath"
	"testing"
)

func TestAuthorizeAndRequireOwned(t *testing.T) {
	path := filepath.Join(t.TempDir(), "registry.json")
	r, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	n, err := r.Authorize("https://notebooklm.google.com/notebook/7471c40e-b33c-4518-b952-3cd786a4e532", "CLI TEST")
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Save(); err != nil {
		t.Fatal(err)
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := reloaded.RequireOwned(n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "CLI TEST" || got.ID != "7471c40e-b33c-4518-b952-3cd786a4e532" {
		t.Fatalf("notebook = %#v", got)
	}
}

func TestAuthorizeRejectsNonNotebookLMURL(t *testing.T) {
	r, err := Load(filepath.Join(t.TempDir(), "registry.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Authorize("https://example.com/notebook/abc", "bad"); err == nil {
		t.Fatal("expected invalid URL")
	}
}

func TestRequireOwnedRejectsUnknownID(t *testing.T) {
	r, err := Load(filepath.Join(t.TempDir(), "registry.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.RequireOwned("unknown"); err == nil {
		t.Fatal("expected ownership error")
	}
}
