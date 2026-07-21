package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestSuccessWritesEnvelope(t *testing.T) {
	var buf bytes.Buffer
	if err := Success(&buf, map[string]any{"logged_in": true}); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["ok"] != true {
		t.Fatalf("ok = %#v", got["ok"])
	}
}

func TestErrorWritesStableError(t *testing.T) {
	var buf bytes.Buffer
	if err := Error(&buf, "not_logged_in", "sign in manually"); err != nil {
		t.Fatal(err)
	}
	var got struct {
		OK    bool `json:"ok"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.OK || got.Error.Code != "not_logged_in" || got.Error.Message != "sign in manually" {
		t.Fatalf("unexpected envelope: %#v", got)
	}
}
