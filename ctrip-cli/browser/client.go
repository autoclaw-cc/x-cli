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

func (c *Client) Call(action string, args map[string]any) (json.RawMessage, error) {
	body, _ := json.Marshal(map[string]any{
		"action":  action,
		"session": c.session,
		"args":    args,
	})
	resp, err := c.http.Post(c.baseURL+"/command", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("daemon unreachable: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		OK    bool            `json:"ok"`
		Data  json.RawMessage `json:"data"`
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if !result.OK {
		if result.Error != nil {
			return nil, fmt.Errorf("%s: %s", result.Error.Code, result.Error.Message)
		}
		return nil, fmt.Errorf("unknown error")
	}
	return result.Data, nil
}

func (c *Client) Navigate(url string) error {
	_, err := c.Call("navigate", map[string]any{"url": url})
	return err
}

func (c *Client) NavigateNewTab(url string) error {
	_, err := c.Call("navigate", map[string]any{"url": url, "newTab": true})
	return err
}

func (c *Client) Evaluate(code string) (json.RawMessage, error) {
	return c.Call("evaluate", map[string]any{"code": code})
}

func (c *Client) Snapshot() (json.RawMessage, error) {
	return c.Call("snapshot", nil)
}

// EvaluateJSON runs JS that returns a JSON string via evaluate,
// unwraps the daemon's {"type":"string","value":"..."} envelope,
// and returns the inner JSON as RawMessage.
func (c *Client) EvaluateJSON(code string) (json.RawMessage, error) {
	data, err := c.Evaluate(code)
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Type  string          `json:"type"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(data, &envelope); err == nil && envelope.Type == "string" {
		var s string
		if err := json.Unmarshal(envelope.Value, &s); err == nil {
			return json.RawMessage(s), nil
		}
	}
	return data, nil
}
