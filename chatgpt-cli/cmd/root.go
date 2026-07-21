package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"chatgpt-cli/browser"
	"chatgpt-cli/chatgpt"
	"chatgpt-cli/output"
)

type app struct {
	out          io.Writer
	errOut       io.Writer
	webbridgeURL string
	timeout      time.Duration
}

type commandError struct {
	code    string
	message string
}

func (e *commandError) Error() string { return e.message }

func Execute(args []string, stdout, stderr io.Writer) int {
	a := &app{
		out:          stdout,
		errOut:       stderr,
		webbridgeURL: os.Getenv("KIMI_WEBBRIDGE_URL"),
		timeout:      10 * time.Minute,
	}
	root := a.rootCommand()
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)
	if err := root.Execute(); err != nil {
		code, message := classifyError(err)
		_ = output.Error(stdout, code, message)
		return 1
	}
	return 0
}

func (a *app) rootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "chatgpt-cli",
		Short:         "Drive ChatGPT Web through an already signed-in Chrome session",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.PersistentFlags().StringVar(&a.webbridgeURL, "webbridge-url", a.webbridgeURL, "Kimi WebBridge daemon URL")
	root.PersistentFlags().DurationVar(&a.timeout, "timeout", a.timeout, "maximum workflow wait")
	root.AddCommand(a.loginStatusCommand())
	root.AddCommand(a.capabilitiesCommand())
	return root
}

func (a *app) loginStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login-status",
		Short: "Check ChatGPT login state without exposing account identity",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.outputPageState()
		},
	}
}

func (a *app) capabilitiesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "capabilities",
		Short: "List ChatGPT capabilities observed in the isolated Chrome page",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.outputPageState()
		},
	}
}

func (a *app) outputPageState() error {
	client, err := a.readyBridgeClient()
	if err != nil {
		return err
	}
	defer func() { _ = client.CloseSession() }()
	state, err := a.inspectPage(client)
	if err != nil {
		return err
	}
	return output.Success(a.out, map[string]any{
		"browser":          "webbridge",
		"authenticated":    state.Authenticated,
		"prompt_available": state.PromptAvailable,
		"locale":           state.Locale,
		"capabilities":     state.Capabilities,
		"url":              state.URL,
	})
}

func (a *app) readyBridgeClient() (*browser.Client, error) {
	client := browser.NewClient(a.webbridgeURL, fmt.Sprintf("chatgpt-cli-%d", time.Now().UnixNano()))
	status, err := client.Status()
	if err != nil {
		return nil, &commandError{code: "daemon_unreachable", message: err.Error()}
	}
	if !status.Running {
		return nil, &commandError{code: "daemon_not_running", message: "Kimi WebBridge daemon is not running."}
	}
	if !status.ExtensionConnected {
		return nil, &commandError{code: "extension_not_connected", message: "Kimi WebBridge extension is not connected to the configured daemon."}
	}
	return client, nil
}

func (a *app) inspectPage(client *browser.Client) (chatgpt.PageState, error) {
	if err := a.ensureChatGPTTab(client); err != nil {
		return chatgpt.PageState{}, err
	}
	var snapshot chatgpt.PageSnapshot
	if err := client.EvaluateValue(chatgpt.SnapshotScript, &snapshot); err != nil {
		return chatgpt.PageState{}, &commandError{code: "chatgpt_page_unavailable", message: err.Error()}
	}
	return chatgpt.AnalyzePage(snapshot), nil
}

func (a *app) ensureChatGPTTab(client *browser.Client) error {
	if err := client.Navigate("https://chatgpt.com/", true); err != nil {
		return &commandError{code: "chatgpt_page_unavailable", message: err.Error()}
	}
	deadline := time.Now().Add(a.timeout)
	for time.Now().Before(deadline) {
		var readiness struct {
			Ready bool `json:"ready"`
		}
		if err := client.EvaluateValue(chatgpt.ReadyScript, &readiness); err != nil {
			return &commandError{code: "chatgpt_page_unavailable", message: err.Error()}
		}
		if readiness.Ready {
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return &commandError{code: "chatgpt_page_unavailable", message: "timed out waiting for ChatGPT home page"}
}

func classifyError(err error) (string, string) {
	if err == nil {
		return "unknown_error", "unknown error"
	}
	if typed, ok := err.(*commandError); ok {
		return typed.code, typed.message
	}
	return "command_failed", err.Error()
}
