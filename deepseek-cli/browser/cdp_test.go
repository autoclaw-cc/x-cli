package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestCDPClientReadsDeepSeekSnapshot(t *testing.T) {
	var serverURL string
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/json/list":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `[{"type":"page","url":"https://chat.deepseek.com/","title":"DeepSeek","webSocketDebuggerUrl":%q}]`, strings.Replace(serverURL, "http://", "ws://", 1)+"/devtools/page/1")
		case "/devtools/page/1":
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Errorf("upgrade websocket: %v", err)
				return
			}
			defer conn.Close()
			for {
				var req struct {
					ID     int    `json:"id"`
					Method string `json:"method"`
				}
				if err := conn.ReadJSON(&req); err != nil {
					return
				}
				if req.Method == "Runtime.evaluate" {
					_ = conn.WriteJSON(map[string]any{
						"id": req.ID,
						"result": map[string]any{
							"result": map[string]any{
								"value": map[string]any{
									"href":           "https://chat.deepseek.com/",
									"title":          "DeepSeek - 探索未至之境",
									"bodyText":       "默认模式 深度思考 联网搜索",
									"hasPromptInput": true,
									"inputs": []map[string]any{{
										"tag":         "TEXTAREA",
										"placeholder": "给 DeepSeek 发送消息",
									}},
									"controls": []map[string]any{{
										"text": "深度思考",
									}},
								},
							},
						},
					})
					continue
				}
				_ = conn.WriteJSON(map[string]any{"id": req.ID, "result": map[string]any{}})
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	serverURL = server.URL

	got, err := NewCDPClient(server.URL).DeepSeekSnapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.Href != "https://chat.deepseek.com/" || !got.HasPromptInput {
		encoded, _ := json.Marshal(got)
		t.Fatalf("snapshot = %s", encoded)
	}
}
