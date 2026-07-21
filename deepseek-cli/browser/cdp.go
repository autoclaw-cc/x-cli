package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"deepseek-cli/deepseek"
)

const DefaultCDPURL = "http://127.0.0.1:9223"

type CDPClient struct {
	baseURL string
	http    *http.Client
	nextID  int64
}

func NewCDPClient(baseURL string) *CDPClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = DefaultCDPURL
	}
	return &CDPClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *CDPClient) DeepSeekSnapshot(ctx context.Context) (deepseek.PageSnapshot, error) {
	page, err := c.deepSeekPage(ctx)
	if err != nil {
		return deepseek.PageSnapshot{}, err
	}
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, page.WebSocketDebuggerURL, nil)
	if err != nil {
		return deepseek.PageSnapshot{}, fmt.Errorf("connect Chrome CDP websocket: %w", err)
	}
	defer conn.Close()
	if _, err := c.callCDP(conn, "Runtime.enable", nil); err != nil {
		return deepseek.PageSnapshot{}, err
	}
	raw, err := c.callCDP(conn, "Runtime.evaluate", map[string]any{
		"expression":    deepSeekSnapshotScript,
		"returnByValue": true,
		"awaitPromise":  true,
	})
	if err != nil {
		return deepseek.PageSnapshot{}, err
	}
	var envelope struct {
		Result struct {
			Value deepseek.PageSnapshot `json:"value"`
		} `json:"result"`
		ExceptionDetails any `json:"exceptionDetails"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return deepseek.PageSnapshot{}, fmt.Errorf("parse Chrome CDP evaluate result: %w", err)
	}
	if envelope.ExceptionDetails != nil {
		return deepseek.PageSnapshot{}, fmt.Errorf("Chrome CDP evaluate returned exception")
	}
	return envelope.Result.Value, nil
}

type cdpPage struct {
	Type                 string `json:"type"`
	URL                  string `json:"url"`
	Title                string `json:"title"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

func (c *CDPClient) deepSeekPage(ctx context.Context) (cdpPage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/json/list", nil)
	if err != nil {
		return cdpPage{}, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return cdpPage{}, fmt.Errorf("Chrome CDP unreachable at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return cdpPage{}, fmt.Errorf("Chrome CDP /json/list returned %s", resp.Status)
	}
	var pages []cdpPage
	if err := json.NewDecoder(resp.Body).Decode(&pages); err != nil {
		return cdpPage{}, fmt.Errorf("parse Chrome CDP page list: %w", err)
	}
	for _, page := range pages {
		if page.Type == "page" && strings.Contains(page.URL, "chat.deepseek.com") && page.WebSocketDebuggerURL != "" {
			return page, nil
		}
	}
	return cdpPage{}, fmt.Errorf("no DeepSeek page found in Chrome CDP target list")
}

func (c *CDPClient) callCDP(conn *websocket.Conn, method string, params map[string]any) (json.RawMessage, error) {
	id := int(atomic.AddInt64(&c.nextID, 1))
	if params == nil {
		params = map[string]any{}
	}
	if err := conn.WriteJSON(map[string]any{"id": id, "method": method, "params": params}); err != nil {
		return nil, fmt.Errorf("send Chrome CDP %s: %w", method, err)
	}
	for {
		var msg struct {
			ID     int             `json:"id"`
			Result json.RawMessage `json:"result"`
			Error  *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := conn.ReadJSON(&msg); err != nil {
			return nil, fmt.Errorf("read Chrome CDP %s: %w", method, err)
		}
		if msg.ID != id {
			continue
		}
		if msg.Error != nil {
			return nil, fmt.Errorf("Chrome CDP %s failed: %s", method, msg.Error.Message)
		}
		return msg.Result, nil
	}
}

const deepSeekSnapshotScript = `
(() => {
  const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));
  const txt=s=>(s||'').replace(/\s+/g,' ').trim();
  const body=txt(document.body?.innerText||'');
  const inputs=[...document.querySelectorAll('textarea,[contenteditable=true],input')]
    .filter(visible).slice(0,12).map(e=>({
      tag:e.tagName,
      role:e.getAttribute('role')||'',
      type:e.getAttribute('type')||'',
      aria:e.getAttribute('aria-label')||'',
      placeholder:e.getAttribute('placeholder')||'',
      text:txt(e.innerText||e.textContent||e.value).slice(0,120),
      disabled:!!e.disabled,
      cls:String(e.className).slice(0,120)
    }));
  const re=/DeepThink|R1|深度思考|联网|Search|搜索|上传|Attach|发送|Send|新对话|New chat|设置|Settings|模型|model|文件|识图|vision/i;
  const controls=[...document.querySelectorAll('button,[role=button],[role=radio],a,div,span')]
    .filter(e=>visible(e)&&re.test(txt(e.innerText||e.textContent)+' '+(e.getAttribute('aria-label')||'')))
    .slice(0,40).map(e=>({
      tag:e.tagName,
      role:e.getAttribute('role')||'',
      aria:e.getAttribute('aria-label')||'',
      text:txt(e.innerText||e.textContent).slice(0,120),
      disabled:!!e.disabled,
      cls:String(e.className).slice(0,120)
    }));
  return {href:location.href,title:document.title,bodyText:body.slice(0,600),hasPromptInput:inputs.length>0,inputs,controls};
})()
`
