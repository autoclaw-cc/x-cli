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

const DefaultDaemonURL = "http://127.0.0.1:10086"

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
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = DefaultDaemonURL
	}
	return &Client{
		baseURL: baseURL,
		session: session,
		http:    &http.Client{Timeout: 180 * time.Second},
	}
}

func (c *Client) Status() (*Status, error) {
	resp, err := c.http.Get(c.baseURL + "/status")
	if err != nil {
		return nil, fmt.Errorf("daemon unreachable at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("daemon status returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var status Status
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("parse daemon status: %w", err)
	}
	return &status, nil
}

func (c *Client) Call(action string, args map[string]any) (json.RawMessage, error) {
	body, err := json.Marshal(map[string]any{
		"action":  action,
		"session": c.session,
		"args":    args,
	})
	if err != nil {
		return nil, fmt.Errorf("encode command: %w", err)
	}
	resp, err := c.http.Post(c.baseURL+"/command", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("daemon unreachable at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("daemon command %s returned %s: %s", action, resp.Status, strings.TrimSpace(string(responseBody)))
	}
	var envelope struct {
		OK    bool            `json:"ok"`
		Data  json.RawMessage `json:"data"`
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decode daemon response: %w", err)
	}
	if !envelope.OK {
		if envelope.Error != nil {
			return nil, fmt.Errorf("%s: %s", envelope.Error.Code, envelope.Error.Message)
		}
		return nil, fmt.Errorf("daemon returned ok=false without error detail")
	}
	return envelope.Data, nil
}

func (c *Client) FindTab(url string, active bool) error {
	_, err := c.Call("find_tab", map[string]any{"url": url, "active": active})
	return err
}

func (c *Client) Navigate(url string, newTab bool) error {
	_, err := c.Call("navigate", map[string]any{
		"url":         url,
		"newTab":      newTab,
		"group_title": "chatgpt-cli",
	})
	return err
}

func (c *Client) Fill(selector, value string) error {
	_, err := c.Call("fill", map[string]any{"selector": selector, "value": value})
	return err
}

func (c *Client) Click(selector string) error {
	_, err := c.Call("click", map[string]any{"selector": selector})
	return err
}

func (c *Client) BringToFront() error {
	_, err := c.Call("cdp", map[string]any{
		"method": "Page.bringToFront",
		"params": map[string]any{},
	})
	return err
}

func (c *Client) CloseSession() error {
	_, err := c.Call("close_session", map[string]any{})
	return err
}

func (c *Client) EvaluateValue(code string, target any) error {
	raw, err := c.Call("evaluate", map[string]any{"code": code})
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
	if wrap.Type == "string" {
		var value string
		if err := json.Unmarshal(wrap.Value, &value); err != nil {
			return fmt.Errorf("parse evaluate string: %w", err)
		}
		if err := json.Unmarshal([]byte(value), target); err != nil {
			return fmt.Errorf("parse evaluate JSON string: %w", err)
		}
		return nil
	}
	if err := json.Unmarshal(wrap.Value, target); err != nil {
		return fmt.Errorf("parse evaluate value: %w", err)
	}
	return nil
}
