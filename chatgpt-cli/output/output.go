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
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func printJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("write json: %w", err)
	}
	return nil
}
