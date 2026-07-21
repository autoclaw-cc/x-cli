package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

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
		t.Fatalf("invalid json: %v\n%s", err, out.String())
	}
	if envelope.OK || envelope.Error.Code != "not_logged_in" {
		t.Fatalf("envelope = %#v", envelope)
	}
}
