# Go CLI Layout Reference

适用于在**任意目录**新建独立 Go CLI 项目（不依赖任何外部私有 repo）。

## 初始化

```bash
mkdir {platform}-cli && cd {platform}-cli
go mod init {platform}-cli
go get github.com/spf13/cobra
```

## 目录结构

```
{platform}-cli/
├── go.mod
├── main.go                  # ~10 行：入口，组装命令并执行
├── browser/
│   └── client.go            # 轻量 daemon HTTP 客户端
├── output/
│   └── output.go            # JSON 输出标准化
├── {platform}/
│   ├── login.go             # 登录状态检测（需要登录时）
│   └── {feature}.go         # 每个功能一个文件
└── cmd/
    └── root.go              # cobra 命令注册
```

## main.go

```go
package main

import (
    "os"
    "{platform}-cli/cmd"
)

func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

## browser/client.go（自己实现，不依赖外部包）

```go
package browser

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

const DefaultDaemonURL = "http://127.0.0.1:10086"

type Client struct {
    baseURL string
    session string
    http    *http.Client
}

func NewClient(session string) *Client {
    return &Client{
        baseURL: DefaultDaemonURL,
        session: session,
        http:    &http.Client{Timeout: 90 * time.Second},
    }
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
        return nil, fmt.Errorf("%s: %s", result.Error.Code, result.Error.Message)
    }
    return result.Data, nil
}

// 常用快捷方法
func (c *Client) Navigate(url string) error {
    _, err := c.Call("navigate", map[string]any{"url": url, "newTab": true})
    return err
}

func (c *Client) Evaluate(code string) (json.RawMessage, error) {
    return c.Call("evaluate", map[string]any{"code": code})
}

func (c *Client) NetworkStart() error { _, err := c.Call("network", map[string]any{"cmd": "start"}); return err }
func (c *Client) NetworkStop() error  { _, err := c.Call("network", map[string]any{"cmd": "stop"}); return err }
func (c *Client) NetworkList() (json.RawMessage, error) {
    return c.Call("network", map[string]any{"cmd": "list"})
}
```

## output/output.go

```go
package output

import (
    "encoding/json"
    "fmt"
    "os"
)

func Success(data any) {
    printJSON(map[string]any{"ok": true, "data": data})
}

func Error(code, message string) {
    printJSON(map[string]any{"ok": false, "error": map[string]any{"code": code, "message": message}})
}

func printJSON(v any) {
    enc := json.NewEncoder(os.Stdout)
    enc.SetEscapeHTML(false)
    enc.SetIndent("", "  ")
    if err := enc.Encode(v); err != nil {
        fmt.Fprintf(os.Stderr, "output error: %v\n", err)
    }
}
```

## cmd/root.go（cobra 命令注册示例）

```go
package cmd

import (
    "os"

    "github.com/spf13/cobra"
    "{platform}-cli/browser"
    "{platform}-cli/output"
    "{platform}-cli/{platform}"
)

var rootCmd = &cobra.Command{
    Use:   "{platform}-cli",
    Short: "{Platform} automation CLI",
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    hotCmd := &cobra.Command{
        Use:   "hot",
        Short: "List hot topics",
        Run: func(cmd *cobra.Command, args []string) {
            limit, _ := cmd.Flags().GetInt("limit")
            client := browser.NewClient("{platform}")
            items, err := {platform}.FetchHot(client, limit)
            if err != nil {
                output.Error("hot_error", err.Error())
                os.Exit(1)
            }
            output.Success(items)
        },
    }
    hotCmd.Flags().Int("limit", 20, "max items to return")
    rootCmd.AddCommand(hotCmd)
}
```

## 构建与运行

```bash
go build -o {platform}-cli .
./{platform}-cli --help
./{platform}-cli hot --limit 5
```
