package browser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const DefaultDaemonURL = "http://127.0.0.1:10086"

type Client struct {
	baseURL string
	session string
	http    *http.Client
}

func NewClient(session string) *Client {
	return &Client{
		baseURL: DefaultDaemonURL,
		session: session,
		http:    &http.Client{Timeout: 90 * time.Second},
	}
}

type daemonResponse struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) Call(action string, args map[string]any) (json.RawMessage, error) {
	payload := map[string]any{
		"action":  action,
		"session": c.session,
	}
	if args != nil {
		payload["args"] = args
	}
	body, _ := json.Marshal(payload)
	resp, err := c.http.Post(c.baseURL+"/command", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("daemon unreachable at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	var result daemonResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode daemon response: %w", err)
	}
	if !result.OK {
		if result.Error != nil {
			return nil, fmt.Errorf("%s: %s", result.Error.Code, result.Error.Message)
		}
		return nil, fmt.Errorf("daemon returned ok=false without error payload")
	}
	return result.Data, nil
}

func (c *Client) Navigate(url string) error {
	_, err := c.Call("navigate", map[string]any{"url": url, "newTab": false})
	return err
}

// Evaluate runs JS in the active tab. The JS must be an expression (or IIFE)
// that returns a value — kimi-webbridge wraps the return into {type, value}.
func (c *Client) Evaluate(code string) (json.RawMessage, error) {
	raw, err := c.Call("evaluate", map[string]any{"code": code})
	if err != nil {
		return nil, err
	}
	var env struct {
		Type  string          `json:"type"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return raw, nil
	}
	if len(env.Value) == 0 {
		return raw, nil
	}
	return env.Value, nil
}

// EvaluateJSON runs `code` and decodes the stringified JSON it returns into v.
// `code` must end with a `JSON.stringify(...)` expression.
func (c *Client) EvaluateJSON(code string, v any) error {
	raw, err := c.Evaluate(code)
	if err != nil {
		return err
	}
	var env struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("decode evaluate envelope: %w", err)
	}
	if env.Type != "string" {
		return fmt.Errorf("expected evaluate type=string, got %q", env.Type)
	}
	if err := json.Unmarshal([]byte(env.Value), v); err != nil {
		return fmt.Errorf("decode evaluate value: %w", err)
	}
	return nil
}

// evaluateWithRetry retries transient CDP execution-context errors.
func evaluateWithRetry(client *Client, code string, out any) error {
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
