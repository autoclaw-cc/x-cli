package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestSuccessEnvelopeIsMachineReadable(t *testing.T) {
	var out bytes.Buffer
	if err := Success(&out, map[string]string{"answer": "ok"}); err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		OK   bool `json:"ok"`
		Data struct {
			Answer string `json:"answer"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if !envelope.OK || envelope.Data.Answer != "ok" {
		t.Fatalf("envelope = %#v", envelope)
	}
}

func TestErrorEnvelopeIsMachineReadable(t *testing.T) {
	var out bytes.Buffer
	if err := Error(&out, "not_logged_in", "Please log in manually."); err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		OK    bool `json:"ok"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if envelope.OK || envelope.Error.Code != "not_logged_in" {
		t.Fatalf("envelope = %#v", envelope)
	}
}
