package output

import (
	"encoding/json"
	"io"
)

func Success(w io.Writer, data any) error {
	return write(w, map[string]any{"ok": true, "data": data})
}

func Error(w io.Writer, code, message string) error {
	return write(w, map[string]any{
		"ok":    false,
		"error": map[string]any{"code": code, "message": message},
	})
}

func write(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(value)
}
