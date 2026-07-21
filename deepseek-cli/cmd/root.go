package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"deepseek-cli/browser"
	"deepseek-cli/deepseek"
	"deepseek-cli/output"
)

type app struct {
	out          io.Writer
	errOut       io.Writer
	webbridgeURL string
	cdpURL       string
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
		cdpURL:       os.Getenv("DEEPSEEK_CDP_URL"),
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
	root.PersistentFlags().StringVar(&a.cdpURL, "cdp-url", a.cdpURL, "Chrome DevTools endpoint for the isolated DeepSeek profile")
	root.PersistentFlags().DurationVar(&a.timeout, "timeout", a.timeout, "maximum workflow wait")
	root.AddCommand(a.loginStatusCommand())
	root.AddCommand(a.capabilitiesCommand())
	root.AddCommand(a.chatCommand())
	return root
}

func (a *app) loginStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login-status",
		Short: "Check DeepSeek login state without exposing account identity",
		RunE: func(cmd *cobra.Command, args []string) error {
			if a.cdpURL != "" {
				state, err := a.inspectDeepSeekPage()
				if err != nil {
					return err
				}
				return output.Success(a.out, map[string]any{
					"browser":          "cdp",
					"authenticated":    state.Authenticated,
					"prompt_available": state.PromptAvailable,
					"capabilities":     state.Capabilities,
					"url":              state.URL,
				})
			}
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

func (a *app) capabilitiesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "capabilities",
		Short: "List DeepSeek UI capabilities visible in the isolated Chrome page",
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := a.inspectDeepSeekPage()
			if err != nil {
				return err
			}
			return output.Success(a.out, map[string]any{
				"browser":          "cdp",
				"authenticated":    state.Authenticated,
				"prompt_available": state.PromptAvailable,
				"capabilities":     state.Capabilities,
				"url":              state.URL,
			})
		},
	}
}

func (a *app) chatCommand() *cobra.Command {
	chat := &cobra.Command{
		Use:   "chat",
		Short: "Run DeepSeek chat workflows through the isolated Chrome page",
	}
	chat.AddCommand(a.chatAskCommand())
	return chat
}

func (a *app) chatAskCommand() *cobra.Command {
	var prompt string
	var outputPath string
	ask := &cobra.Command{
		Use:   "ask",
		Short: "Ask DeepSeek and wait for a stable assistant answer",
		RunE: func(cmd *cobra.Command, args []string) error {
			if a.cdpURL == "" {
				return &commandError{
					code:    "cdp_url_required",
					message: "Set --cdp-url to the isolated Chrome DevTools endpoint, for example http://127.0.0.1:9223.",
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
			defer cancel()
			result, err := browser.NewCDPClient(a.cdpURL).Ask(ctx, prompt, browser.AskOptions{
				PollInterval:  time.Second,
				StableSamples: 2,
			})
			if err != nil {
				return &commandError{code: "deepseek_chat_failed", message: err.Error()}
			}
			data := map[string]any{
				"browser":        "cdp",
				"prompt":         result.Prompt,
				"answer":         result.Answer,
				"stable_samples": result.StableSamples,
			}
			if outputPath != "" {
				if err := writeTextOutput(outputPath, result.Answer); err != nil {
					return &commandError{code: "write_output_failed", message: err.Error()}
				}
				data["output"] = outputPath
			}
			return output.Success(a.out, data)
		},
	}
	ask.Flags().StringVar(&prompt, "prompt", "", "prompt text to send to DeepSeek")
	ask.Flags().StringVar(&outputPath, "output", "", "optional local path for the assistant answer text")
	_ = ask.MarkFlagRequired("prompt")
	return ask
}

func writeTextOutput(path, text string) error {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, []byte(text), 0o644)
}

func (a *app) inspectDeepSeekPage() (deepseek.PageState, error) {
	if a.cdpURL == "" {
		return deepseek.PageState{}, &commandError{
			code:    "cdp_url_required",
			message: "Set --cdp-url to the isolated Chrome DevTools endpoint, for example http://127.0.0.1:9223.",
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()
	snapshot, err := browser.NewCDPClient(a.cdpURL).DeepSeekSnapshot(ctx)
	if err != nil {
		return deepseek.PageState{}, &commandError{code: "deepseek_page_unavailable", message: err.Error()}
	}
	return deepseek.AnalyzePage(snapshot), nil
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
