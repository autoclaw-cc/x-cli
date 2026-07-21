package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"deepseek-cli/browser"
	"deepseek-cli/output"
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
		timeout:      5 * time.Minute,
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
		Use:           "deepseek-cli",
		Short:         "Drive DeepSeek Web through an already signed-in Chrome session",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.PersistentFlags().StringVar(&a.webbridgeURL, "webbridge-url", a.webbridgeURL, "Kimi WebBridge daemon URL")
	root.PersistentFlags().DurationVar(&a.timeout, "timeout", a.timeout, "maximum workflow wait")
	root.AddCommand(a.loginStatusCommand())
	return root
}

func (a *app) loginStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login-status",
		Short: "Check Kimi WebBridge readiness before DeepSeek account inspection",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := browser.NewClient(a.webbridgeURL, fmt.Sprintf("deepseek-cli-%d", time.Now().UnixNano()))
			status, err := client.Status()
			if err != nil {
				return &commandError{code: "daemon_unreachable", message: err.Error()}
			}
			if !status.Running {
				return &commandError{code: "daemon_not_running", message: "Kimi WebBridge daemon is not running."}
			}
			if !status.ExtensionConnected {
				return &commandError{
					code:    "extension_not_connected",
					message: "Kimi WebBridge extension is not connected. Open Chrome with the Kimi WebBridge extension, connect it to the configured daemon, then run this command again.",
				}
			}
			return output.Success(a.out, map[string]any{
				"bridge_ready":      true,
				"daemon_version":    status.Version,
				"extension_version": status.ExtensionVersion,
				"login_state":       "unchecked",
			})
		},
	}
}

func classifyError(err error) (string, string) {
	if err == nil {
		return "unknown_error", "unknown error"
	}
	if e, ok := err.(*commandError); ok {
		return e.code, e.message
	}
	return "command_failed", err.Error()
}
