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
		return nil, fmt.Errorf("unknown error from daemon")
	}
	return result.Data, nil
}

// Quick helpers

func (c *Client) Navigate(url string) error {
	_, err := c.Call("navigate", map[string]any{"url": url, "newTab": true})
	return err
}

func (c *Client) Evaluate(code string) (json.RawMessage, error) {
	return c.Call("evaluate", map[string]any{"code": code})
}

func (c *Client) Snapshot() (json.RawMessage, error) {
	return c.Call("snapshot", map[string]any{})
}

func (c *Client) NetworkStart() error {
	_, err := c.Call("network", map[string]any{"cmd": "start"})
	return err
}

func (c *Client) NetworkStop() error {
	_, err := c.Call("network", map[string]any{"cmd": "stop"})
	return err
}

func (c *Client) NetworkList() (json.RawMessage, error) {
	return c.Call("network", map[string]any{"cmd": "list"})
}

func (c *Client) NetworkDetail(requestID string) (json.RawMessage, error) {
	return c.Call("network", map[string]any{"cmd": "detail", "requestId": requestID})
}

// Status checks if the kimi-webbridge daemon is running.
func (c *Client) Status() (json.RawMessage, error) {
	resp, err := c.http.Get(c.baseURL + "/status")
	if err != nil {
		return nil, fmt.Errorf("daemon unreachable at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()
	var result json.RawMessage
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}
