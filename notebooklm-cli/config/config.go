package config

const DefaultWebBridgeURL = "http://127.0.0.1:10086"

func ResolveWebBridgeURL(flagValue string, getenv func(string) string) string {
	if flagValue != "" {
		return flagValue
	}
	if value := getenv("KIMI_WEBBRIDGE_URL"); value != "" {
		return value
	}
	return DefaultWebBridgeURL
}
