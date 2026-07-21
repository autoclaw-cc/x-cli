package browser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type AskOptions struct {
	PollInterval  time.Duration
	StableSamples int
}

type ChatResult struct {
	Prompt        string `json:"prompt"`
	Answer        string `json:"answer"`
	StableSamples int    `json:"stable_samples"`
}

type answerSnapshot struct {
	Count  int    `json:"count"`
	Latest string `json:"latest"`
}

type pointResult struct {
	OK    bool    `json:"ok"`
	Error string  `json:"error"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
}

func (c *CDPClient) Ask(ctx context.Context, prompt string, options AskOptions) (ChatResult, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return ChatResult{}, fmt.Errorf("prompt is required")
	}
	if options.PollInterval <= 0 {
		options.PollInterval = time.Second
	}
	if options.StableSamples <= 0 {
		options.StableSamples = 2
	}
	page, err := c.deepSeekPage(ctx)
	if err != nil {
		return ChatResult{}, err
	}
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, page.WebSocketDebuggerURL, nil)
	if err != nil {
		return ChatResult{}, fmt.Errorf("connect Chrome CDP websocket: %w", err)
	}
	defer conn.Close()
	if _, err := c.callCDP(conn, "Runtime.enable", nil); err != nil {
		return ChatResult{}, err
	}
	_, _ = c.callCDP(conn, "Page.bringToFront", nil)
	baseline, err := c.answerSnapshot(conn)
	if err != nil {
		return ChatResult{}, err
	}
	if err := c.focusPrompt(conn); err != nil {
		return ChatResult{}, err
	}
	if _, err := c.callCDP(conn, "Input.insertText", map[string]any{"text": prompt}); err != nil {
		return ChatResult{}, err
	}
	sendPoint, err := c.sendButtonPoint(conn)
	if err != nil {
		return ChatResult{}, err
	}
	if err := c.clickPoint(conn, sendPoint.X, sendPoint.Y); err != nil {
		return ChatResult{}, err
	}
	return c.waitForStableAnswer(ctx, conn, baseline.Count, prompt, options)
}

func (c *CDPClient) focusPrompt(conn *websocket.Conn) error {
	var point pointResult
	if err := c.evaluateValue(conn, focusPromptScript, &point); err != nil {
		return err
	}
	if !point.OK {
		if point.Error == "" {
			point.Error = "prompt input unavailable"
		}
		return errors.New(point.Error)
	}
	return nil
}

func (c *CDPClient) sendButtonPoint(conn *websocket.Conn) (pointResult, error) {
	var point pointResult
	if err := c.evaluateValue(conn, sendButtonPointScript, &point); err != nil {
		return pointResult{}, err
	}
	if !point.OK {
		if point.Error == "" {
			point.Error = "send button unavailable"
		}
		return pointResult{}, errors.New(point.Error)
	}
	return point, nil
}

func (c *CDPClient) clickPoint(conn *websocket.Conn, x, y float64) error {
	events := []map[string]any{
		{"type": "mouseMoved", "x": x, "y": y, "button": "none"},
		{"type": "mousePressed", "x": x, "y": y, "button": "left", "clickCount": 1},
		{"type": "mouseReleased", "x": x, "y": y, "button": "left", "clickCount": 1},
	}
	for _, event := range events {
		if _, err := c.callCDP(conn, "Input.dispatchMouseEvent", event); err != nil {
			return err
		}
	}
	return nil
}

func (c *CDPClient) waitForStableAnswer(ctx context.Context, conn *websocket.Conn, baselineCount int, prompt string, options AskOptions) (ChatResult, error) {
	ticker := time.NewTicker(options.PollInterval)
	defer ticker.Stop()
	var previous string
	stable := 0
	for {
		select {
		case <-ctx.Done():
			return ChatResult{}, fmt.Errorf("wait for DeepSeek answer: %w", ctx.Err())
		case <-ticker.C:
			current, err := c.answerSnapshot(conn)
			if err != nil {
				return ChatResult{}, err
			}
			answer := strings.TrimSpace(current.Latest)
			if current.Count <= baselineCount || answer == "" {
				previous = ""
				stable = 0
				continue
			}
			if answer == previous {
				stable++
			} else {
				previous = answer
				stable = 1
			}
			if stable >= options.StableSamples {
				return ChatResult{Prompt: prompt, Answer: answer, StableSamples: stable}, nil
			}
		}
	}
}

func (c *CDPClient) answerSnapshot(conn *websocket.Conn) (answerSnapshot, error) {
	var snapshot answerSnapshot
	if err := c.evaluateValue(conn, answerSnapshotScript, &snapshot); err != nil {
		return answerSnapshot{}, err
	}
	return snapshot, nil
}

func (c *CDPClient) evaluateValue(conn *websocket.Conn, expression string, v any) error {
	raw, err := c.callCDP(conn, "Runtime.evaluate", map[string]any{
		"expression":    expression,
		"returnByValue": true,
		"awaitPromise":  true,
	})
	if err != nil {
		return err
	}
	var envelope struct {
		Result struct {
			Value json.RawMessage `json:"value"`
		} `json:"result"`
		ExceptionDetails any `json:"exceptionDetails"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return fmt.Errorf("parse Chrome CDP evaluate result: %w", err)
	}
	if envelope.ExceptionDetails != nil {
		return fmt.Errorf("Chrome CDP evaluate returned exception")
	}
	if len(envelope.Result.Value) == 0 {
		return fmt.Errorf("Chrome CDP evaluate returned no value")
	}
	if err := json.Unmarshal(envelope.Result.Value, v); err != nil {
		return fmt.Errorf("parse Chrome CDP evaluate value: %w", err)
	}
	return nil
}

const focusPromptScript = `
(() => {
  const textarea=document.querySelector('textarea');
  if(!textarea) return {ok:false,error:'textarea_missing'};
  textarea.focus();
  const setter=Object.getOwnPropertyDescriptor(HTMLTextAreaElement.prototype,'value')?.set;
  if(setter) setter.call(textarea,''); else textarea.value='';
  textarea.dispatchEvent(new InputEvent('input',{bubbles:true,inputType:'deleteContentBackward',data:''}));
  const r=textarea.getBoundingClientRect();
  return {ok:true,x:r.x+r.width/2,y:r.y+r.height/2};
})()
`

const sendButtonPointScript = `
(() => {
  const textarea=document.querySelector('textarea');
  if(!textarea) return {ok:false,error:'textarea_missing'};
  const tr=textarea.getBoundingClientRect();
  const candidates=[...document.querySelectorAll('[role=button],button')]
    .map(e=>({e,r:e.getBoundingClientRect(),cls:String(e.className),text:(e.innerText||e.textContent||'').trim()}))
    .filter(x=>x.r.width>20&&x.r.height>20&&x.r.y>=tr.y&&x.r.x>tr.x+tr.width*0.65)
    .sort((a,b)=>b.r.x-a.r.x);
  const c=candidates[0];
  if(!c) return {ok:false,error:'send_button_missing'};
  return {ok:true,x:c.r.x+c.r.width/2,y:c.r.y+c.r.height/2};
})()
`

const answerSnapshotScript = `
(() => {
  const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));
  const txt=s=>(s||'').replace(/\s+/g,' ').trim();
  const answers=[...document.querySelectorAll('.ds-assistant-message-main-content')]
    .filter(visible).map(e=>txt(e.innerText||e.textContent)).filter(Boolean);
  return {count:answers.length,latest:answers[answers.length-1]||''};
})()
`
