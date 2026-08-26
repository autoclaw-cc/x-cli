package bing

import (
	"strings"
	"time"

	"bing-cli/browser"
)

// evaluateWithRetry calls client.EvaluateJSON and retries on transient CDP
// context errors with a three-level backoff: 300ms → 700ms → 1500ms.
// Non-transient errors are returned immediately.
func evaluateWithRetry(client *browser.Client, code string, v any) error {
	delays := []time.Duration{300 * time.Millisecond, 700 * time.Millisecond, 1500 * time.Millisecond}
	var lastErr error
	for i, d := range delays {
		if i > 0 {
			time.Sleep(d)
		}
		err := client.EvaluateJSON(code, v)
		if err == nil {
			return nil
		}
		lastErr = err
		if !isTransientContextError(err) {
			return err
		}
	}
	return lastErr
}

// isTransientContextError returns true when the error message indicates a
// temporary CDP execution-context race (tab is still loading / has navigated).
func isTransientContextError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "Cannot find default execution context") ||
		strings.Contains(msg, "Execution context was destroyed")
}
