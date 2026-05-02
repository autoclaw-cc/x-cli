package browser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
	// Unwrap the {type, value} envelope so callers get the raw JS return value.
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
