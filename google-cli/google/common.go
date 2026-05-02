package google

import (
	"strings"
	"time"

	"google-cli/browser"
)

// evaluateWithRetry retries transient CDP execution-context errors. Navigate
// returns as soon as the request is routed (the new V8 context may not be
// attached yet), and in-flight redirects can destroy the current context.
// Both manifest as distinct error messages from Chrome DevTools Protocol.
func evaluateWithRetry(client *browser.Client, code string, out any) error {
	delays := []time.Duration{300 * time.Millisecond, 700 * time.Millisecond, 1500 * time.Millisecond}
	var lastErr error
	for _, d := range delays {
		time.Sleep(d)
		err := client.EvaluateJSON(code, out)
		if err == nil {
			return nil
		}
		if !isTransientContextError(err) {
			return err
		}
		lastErr = err
	}
	return lastErr
}

func isTransientContextError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "Cannot find default execution context") ||
		strings.Contains(msg, "Execution context was destroyed")
}
