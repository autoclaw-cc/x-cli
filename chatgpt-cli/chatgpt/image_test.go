package chatgpt

import (
	"encoding/base64"
	"testing"
)

func TestSelectNewGeneratedImageSkipsBaselineAndThumbnail(t *testing.T) {
	baseline := map[string]bool{"file_old": true}
	candidates := []imageCandidate{
		{FileID: "file_old", Width: 1024, Height: 1024, Complete: true},
		{FileID: "file_thumb", Width: 128, Height: 128, Complete: true},
		{FileID: "file_new", Width: 1024, Height: 1024, Complete: true, Alt: "Generated image"},
	}
	got, ok := selectNewGeneratedImage(candidates, baseline)
	if !ok || got.FileID != "file_new" {
		t.Fatalf("got=%#v ok=%v", got, ok)
	}
}

func TestDecodeImagePayloadRejectsNonImageContent(t *testing.T) {
	_, err := decodeImagePayload(downloadPayload{
		OK:          true,
		ContentType: "text/html",
		Base64:      base64.StdEncoding.EncodeToString([]byte("not an image")),
	})
	if err == nil {
		t.Fatal("expected content-type error")
	}
}

func TestDecodeImagePayloadReturnsBytes(t *testing.T) {
	want := []byte{0x89, 'P', 'N', 'G'}
	got, err := decodeImagePayload(downloadPayload{
		OK:          true,
		ContentType: "image/png",
		Base64:      base64.StdEncoding.EncodeToString(want),
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
}
