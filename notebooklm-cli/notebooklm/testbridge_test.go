package notebooklm

import (
	"encoding/json"
	"fmt"
)

type scriptedBridge struct {
	calls     []string
	evals     []any
	evalIndex int
}

func (b *scriptedBridge) Navigate(url string, newTab bool, groupTitle string) error {
	b.calls = append(b.calls, fmt.Sprintf("navigate:%s:%t:%s", url, newTab, groupTitle))
	return nil
}

func (b *scriptedBridge) EvaluateValue(code string, dst any) error {
	b.calls = append(b.calls, "evaluate:"+code)
	if b.evalIndex >= len(b.evals) {
		return fmt.Errorf("unexpected evaluate call %d", b.evalIndex)
	}
	body, err := json.Marshal(b.evals[b.evalIndex])
	b.evalIndex++
	if err != nil {
		return err
	}
	return json.Unmarshal(body, dst)
}

func (b *scriptedBridge) MouseClick(selector string) error {
	b.calls = append(b.calls, "mouse_click:"+selector)
	return nil
}

func (b *scriptedBridge) KeyType(text string) error {
	b.calls = append(b.calls, "key_type:"+text)
	return nil
}

func (b *scriptedBridge) SendKeys(keys string) error {
	b.calls = append(b.calls, "send_keys:"+keys)
	return nil
}

func (b *scriptedBridge) CDP(method string, params map[string]any) error {
	b.calls = append(b.calls, "cdp:"+method)
	return nil
}

func (b *scriptedBridge) CloseSession() error {
	b.calls = append(b.calls, "close_session")
	return nil
}

func hasCall(calls []string, want string) bool {
	for _, call := range calls {
		if call == want {
			return true
		}
	}
	return false
}
