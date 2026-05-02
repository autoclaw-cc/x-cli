package browser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
		http:    &http.Client{Timeout: 120 * time.Second},
	}
}

type Status struct {
	Running            bool   `json:"running"`
	ExtensionConnected bool   `json:"extension_connected"`
	ExtensionVersion   string `json:"extension_version"`
	Version            string `json:"version"`
}

func (c *Client) Status() (*Status, error) {
	resp, err := c.http.Get(c.baseURL + "/status")
	if err != nil {
		return nil, fmt.Errorf("daemon unreachable at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var s Status
	if err := json.Unmarshal(body, &s); err != nil {
		return nil, fmt.Errorf("parse status: %w (body=%s)", err, string(body))
	}
	return &s, nil
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
		return nil, fmt.Errorf("daemon returned ok=false with no error body")
	}
	return result.Data, nil
}

func (c *Client) Navigate(url string, newTab bool) error {
	_, err := c.Call("navigate", map[string]any{"url": url, "newTab": newTab})
	return err
}

// Evaluate runs JS and returns the wrapped {type, value} payload.
// Caller typically parses .value into the expected shape.
func (c *Client) Evaluate(code string) (json.RawMessage, error) {
	return c.Call("evaluate", map[string]any{"code": code})
}

// EvaluateValue runs JS and unmarshals the returned expression's value into v.
// The daemon wraps evaluate results as {"type": "...", "value": <json>}.
func (c *Client) EvaluateValue(code string, v any) error {
	raw, err := c.Evaluate(code)
	if err != nil {
		return err
	}
	var wrap struct {
		Type  string          `json:"type"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return fmt.Errorf("parse evaluate wrapper: %w", err)
	}
	if len(wrap.Value) == 0 {
		return fmt.Errorf("evaluate returned no value (type=%s)", wrap.Type)
	}
	if err := json.Unmarshal(wrap.Value, v); err != nil {
		return fmt.Errorf("parse evaluate value: %w (raw=%s)", err, string(wrap.Value))
	}
	return nil
}
