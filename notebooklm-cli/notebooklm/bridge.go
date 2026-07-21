package notebooklm

type Bridge interface {
	Navigate(url string, newTab bool, groupTitle string) error
	EvaluateValue(code string, dst any) error
	Click(selector string) error
	MouseClick(selector string) error
	KeyType(text string) error
	Fill(selector, value string) error
	SendKeys(keys string) error
	CDP(method string, params map[string]any) error
	NetworkValue(cmd, filter, requestID string, dst any) error
	CloseSession() error
}
