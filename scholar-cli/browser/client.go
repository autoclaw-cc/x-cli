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
	http       *http.Client
}

func NewClient(session string) *Client {
	return &Client{
		baseURL:    DefaultDaemonURL,
		session:    session,
		groupTitle: "scholar-cli 文献检索",
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

// NavigateNewTab points this session's tab at url, creating it on the first
// call and reusing it afterwards.
//
// Despite the name it does not pass newTab. WebBridge reserves newTab for pages
// that must coexist (comparing, cross-referencing); a search CLI only ever
// reads one page at a time, so requesting a fresh tab per call just piles up
// tabs in the user's browser — one per search — that nothing ever closes.
// Reusing the session's tab is also self-healing: if the user closes it by
// hand, the daemon creates a new one instead of erroring.
func (c *Client) NavigateNewTab(url string) error {
	args := map[string]any{"url": url, "newTab": false}
	if !c.grouped {
		// Label the tab group on the first navigate so the user can see which
		// group is this task's, and close it themselves whenever they want.
		args["group_title"] = c.groupTitle
		c.grouped = true
	}
	_, err := c.Call("navigate", args)
	return err
}

func (c *Client) Evaluate(code string) (json.RawMessage, error) {
	return c.Call("evaluate", map[string]any{"code": code})
}

func (c *Client) Snapshot() (json.RawMessage, error) {
	return c.Call("snapshot", nil)
}

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
