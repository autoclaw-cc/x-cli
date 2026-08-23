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
	baseURL    string
	session    string
	groupTitle string
	grouped    bool
	activate   ActivateMode
	http       *http.Client
}

func NewClient(session, groupTitle string) *Client {
	return &Client{
		baseURL:    DefaultDaemonURL,
		session:    session,
		groupTitle: groupTitle,
		activate:   defaultActivateMode(),
		http:       &http.Client{Timeout: 90 * time.Second},
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
		return nil, fmt.Errorf("daemon returned ok=false without an error payload")
	}
	return result.Data, nil
}

// Navigate points this session's tab at url, creating the tab on the first
// call and reusing it afterwards.
//
// It deliberately does not pass newTab:true. WebBridge reserves that for pages
// that must coexist; a read-one-page-at-a-time CLI would otherwise leave one
// orphaned tab per invocation in the user's browser. Reusing the session tab is
// also self-healing — if the user closes it, the daemon just makes a new one.
func (c *Client) Navigate(url string) error {
	args := map[string]any{"url": url, "newTab": false}
	if !c.grouped {
		// Label the tab group on the first navigate so the user can tell which
		// group belongs to this CLI and close it whenever they like.
		args["group_title"] = c.groupTitle
		c.grouped = true
	}
	if _, err := c.Call("navigate", args); err != nil {
		return err
	}
	c.Activate()
	return nil
}

// Evaluate runs a JS expression (or IIFE) in the session tab and returns the
// unwrapped return value — kimi-webbridge wraps it in a {type, value} envelope.
//
// Note: the code must be an *expression*. A bare `return` or a top-level
// `await` is a syntax error in the page realm; wrap async work in
// `(async () => { ... })()` and the daemon will await the promise.
func (c *Client) Evaluate(code string) (json.RawMessage, error) {
	raw, err := c.Call("evaluate", map[string]any{"code": code})
	if err != nil {
		return nil, err
	}
	var env struct {
		Type  string          `json:"type"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(raw, &env); err != nil || len(env.Value) == 0 {
		return raw, nil
	}
	return env.Value, nil
}

// CloseSession closes every tab this session opened. Callers should defer it so
// long-running shells don't accumulate Unsplash tabs.
func (c *Client) CloseSession() error {
	_, err := c.Call("close_session", nil)
	return err
}
