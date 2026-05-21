package gaokao

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	baseURL    = "https://static-data.gaokao.cn/www/2.0"
	queryParam = "?a=www.gaokao.cn"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

// APIResponse is the common envelope for all gaokao.cn JSON responses.
type APIResponse struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// FetchJSON fetches a URL and decodes the response JSON into target.
// If the response has the standard {code, message, data} envelope and code != "0000", returns an error.
func FetchJSON(url string, target any) error {
	resp, err := httpClient.Get(url + queryParam)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("http %d from %s", resp.StatusCode, url)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

// FetchData fetches a URL, unwraps the {code, data} envelope, and decodes data into target.
func FetchData(url string, target any) error {
	var envelope APIResponse
	if err := FetchJSON(url, &envelope); err != nil {
		return err
	}
	if envelope.Code != "0000" {
		return fmt.Errorf("api error: %s - %s", envelope.Code, envelope.Message)
	}
	return json.Unmarshal(envelope.Data, target)
}
