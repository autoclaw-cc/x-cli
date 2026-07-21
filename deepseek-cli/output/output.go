package output

import (
	"encoding/json"
	"fmt"
	"io"
)

func Success(w io.Writer, data any) error {
	return printJSON(w, map[string]any{"ok": true, "data": data})
}

func Error(w io.Writer, code, message string) error {
	return printJSON(w, map[string]any{
		"ok": false,
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	})
}

func printJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("write json: %w", err)
	}
	return nil
}
