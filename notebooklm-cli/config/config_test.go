package config

import "testing"

func TestResolveWebBridgeURLPrecedence(t *testing.T) {
	tests := []struct {
		name string
		flag string
		env  string
		want string
	}{
		{"flag", "http://127.0.0.1:10400", "http://127.0.0.1:10087", "http://127.0.0.1:10400"},
		{"env", "", "http://127.0.0.1:10087", "http://127.0.0.1:10087"},
		{"default", "", "", "http://127.0.0.1:10086"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveWebBridgeURL(tt.flag, func(key string) string {
				if key != "KIMI_WEBBRIDGE_URL" {
					t.Fatalf("unexpected key %q", key)
				}
				return tt.env
			})
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
