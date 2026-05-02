package output

import (
	"encoding/json"
	"fmt"
	"os"
)

func Success(data any) {
	printJSON(map[string]any{"ok": true, "data": data})
}

func Error(code, message string) {
	printJSON(map[string]any{
		"ok":    false,
		"error": map[string]any{"code": code, "message": message},
	})
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(os.Stderr, "output error: %v\n", err)
	}
}
