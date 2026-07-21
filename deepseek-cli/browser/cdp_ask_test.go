package browser

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestCDPClientAskSendsPromptAndReturnsStableLatestAssistant(t *testing.T) {
	var serverURL string
	var insertedPrompt string
	answerReads := 0
	mouseClicks := 0
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/json/list":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `[{"type":"page","url":"https://chat.deepseek.com/","title":"DeepSeek","webSocketDebuggerUrl":%q}]`, strings.Replace(serverURL, "http://", "ws://", 1)+"/devtools/page/ask")
		case "/devtools/page/ask":
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Errorf("upgrade websocket: %v", err)
				return
			}
			defer conn.Close()
			for {
				var req struct {
					ID     int            `json:"id"`
					Method string         `json:"method"`
					Params map[string]any `json:"params"`
				}
				if err := conn.ReadJSON(&req); err != nil {
					return
				}
				switch req.Method {
				case "Input.insertText":
					insertedPrompt, _ = req.Params["text"].(string)
					_ = conn.WriteJSON(map[string]any{"id": req.ID, "result": map[string]any{}})
				case "Input.dispatchMouseEvent":
					if req.Params["type"] == "mouseReleased" {
						mouseClicks++
					}
					_ = conn.WriteJSON(map[string]any{"id": req.ID, "result": map[string]any{}})
				case "Runtime.evaluate":
					expr, _ := req.Params["expression"].(string)
					if strings.Contains(expr, "ds-assistant-message-main-content") {
						answerReads++
						count := 0
						answer := ""
						if answerReads >= 2 {
							count = 1
							answer = "stable answer"
						}
						_ = conn.WriteJSON(map[string]any{
							"id": req.ID,
							"result": map[string]any{
								"result": map[string]any{
									"value": map[string]any{"count": count, "latest": answer},
								},
							},
						})
						continue
					}
					_ = conn.WriteJSON(map[string]any{
						"id": req.ID,
						"result": map[string]any{
							"result": map[string]any{
								"value": map[string]any{"ok": true, "x": 50, "y": 60},
							},
						},
					})
				default:
					_ = conn.WriteJSON(map[string]any{"id": req.ID, "result": map[string]any{}})
				}
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	serverURL = server.URL

	got, err := NewCDPClient(server.URL).Ask(context.Background(), "hello", AskOptions{
		PollInterval:  time.Millisecond,
		StableSamples: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if insertedPrompt != "hello" {
		t.Fatalf("insertedPrompt = %q", insertedPrompt)
	}
	if mouseClicks != 1 {
		t.Fatalf("mouseClicks = %d", mouseClicks)
	}
	if got.Answer != "stable answer" {
		t.Fatalf("answer = %q", got.Answer)
	}
}
