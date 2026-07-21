package browser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	session string
	http    *http.Client
}

type Status struct {
	Running            bool   `json:"running"`
	ExtensionConnected bool   `json:"extension_connected"`
	ExtensionVersion   string `json:"extension_version"`
	Version            string `json:"version"`
}

func NewClient(baseURL, session string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		session: session,
		http:    &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *Client) Status() (*Status, error) {
	resp, err := c.http.Get(c.baseURL + "/status")
	if err != nil {
		return nil, fmt.Errorf("daemon unreachable at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read status: %w", err)
	}
	var status Status
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("parse status: %w", err)
	}
	return &status, nil
}

func (c *Client) Call(action string, args map[string]any) (json.RawMessage, error) {
	payload, err := json.Marshal(map[string]any{
		"action": action, "session": c.session, "args": args,
	})
	if err != nil {
		return nil, fmt.Errorf("encode command: %w", err)
	}
	resp, err := c.http.Post(c.baseURL+"/command", "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("daemon unreachable at %s: %w", c.baseURL, err)
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
		return nil, fmt.Errorf("decode command response: %w", err)
	}
	if !result.OK {
		if result.Error != nil {
			return nil, fmt.Errorf("%s: %s", result.Error.Code, result.Error.Message)
		}
		return nil, fmt.Errorf("daemon returned ok=false")
	}
	return result.Data, nil
}

func (c *Client) Navigate(url string, newTab bool, groupTitle string) error {
	args := map[string]any{"url": url, "newTab": newTab}
	if groupTitle != "" {
		args["group_title"] = groupTitle
	}
	_, err := c.Call("navigate", args)
	return err
}

func (c *Client) EvaluateValue(code string, dst any) error {
	raw, err := c.Call("evaluate", map[string]any{"code": code})
	if err != nil {
		return err
	}
	var result struct {
		Type  string          `json:"type"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return fmt.Errorf("parse evaluate wrapper: %w", err)
	}
	if len(result.Value) == 0 {
		return fmt.Errorf("evaluate returned no value")
	}
	if err := json.Unmarshal(result.Value, dst); err != nil {
		return fmt.Errorf("parse evaluate value: %w", err)
	}
	return nil
}

func (c *Client) Click(selector string) error {
	_, err := c.Call("click", map[string]any{"selector": selector})
	return err
}

func (c *Client) MouseClick(selector string) error {
	_, err := c.Call("mouse_click", map[string]any{"selector": selector})
	return err
}

func (c *Client) KeyType(text string) error {
	_, err := c.Call("key_type", map[string]any{"text": text})
	return err
}

func (c *Client) Fill(selector, value string) error {
	_, err := c.Call("fill", map[string]any{"selector": selector, "value": value})
	return err
}

func (c *Client) SendKeys(keys string) error {
	_, err := c.Call("send_keys", map[string]any{"keys": keys})
	return err
}

func (c *Client) CDP(method string, params map[string]any) error {
	_, err := c.Call("cdp", map[string]any{"method": method, "params": params})
	return err
}

func (c *Client) NetworkValue(cmd, filter, requestID string, dst any) error {
	args := map[string]any{"cmd": cmd}
	if filter != "" {
		args["filter"] = filter
	}
	if requestID != "" {
		args["requestId"] = requestID
	}
	raw, err := c.Call("network", args)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("parse network value: %w", err)
	}
	return nil
}

func (c *Client) CloseSession() error {
	_, err := c.Call("close_session", map[string]any{})
	return err
}
