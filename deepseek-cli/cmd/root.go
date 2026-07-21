package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	root.AddCommand(a.capabilitiesCommand())
	root.AddCommand(a.chatCommand())
	return root
}

func (a *app) loginStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login-status",
		Short: "Check DeepSeek login state without exposing account identity",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.readyBridgeClient()
			if err != nil {
				return err
			}
			state, err := a.inspectDeepSeekPage(client)
			if err != nil {
				return err
			}
			return output.Success(a.out, map[string]any{
				"browser":          "webbridge",
				"authenticated":    state.Authenticated,
				"prompt_available": state.PromptAvailable,
				"capabilities":     state.Capabilities,
				"url":              state.URL,
			})
		},
	}
}

func (a *app) capabilitiesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "capabilities",
		Short: "List DeepSeek UI capabilities visible in the isolated Chrome page",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.readyBridgeClient()
			if err != nil {
				return err
			}
			state, err := a.inspectDeepSeekPage(client)
			if err != nil {
				return err
			}
			return output.Success(a.out, map[string]any{
				"browser":          "webbridge",
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
			client, err := a.readyBridgeClient()
			if err != nil {
				return err
			}
			result, err := a.askDeepSeek(client, prompt)
			if err != nil {
				return &commandError{code: "deepseek_chat_failed", message: err.Error()}
			}
			data := map[string]any{
				"browser":        "webbridge",
				"prompt":         result.prompt,
				"answer":         result.answer,
				"stable_samples": result.stableSamples,
			}
			if outputPath != "" {
				if err := writeTextOutput(outputPath, result.answer); err != nil {
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

type askResult struct {
	prompt        string
	answer        string
	stableSamples int
}

func (a *app) readyBridgeClient() (*browser.Client, error) {
	client := browser.NewClient(a.webbridgeURL, fmt.Sprintf("deepseek-cli-%d", time.Now().UnixNano()))
	status, err := client.Status()
	if err != nil {
		return nil, &commandError{code: "daemon_unreachable", message: err.Error()}
	}
	if !status.Running {
		return nil, &commandError{code: "daemon_not_running", message: "Kimi WebBridge daemon is not running."}
	}
	if !status.ExtensionConnected {
		return nil, &commandError{
			code:    "extension_not_connected",
			message: "Kimi WebBridge extension is not connected. Open the isolated Chrome profile with the Kimi WebBridge extension, connect it to the configured daemon, then run this command again.",
		}
	}
	return client, nil
}

func (a *app) inspectDeepSeekPage(client *browser.Client) (deepseek.PageState, error) {
	if err := a.ensureDeepSeekTab(client); err != nil {
		return deepseek.PageState{}, err
	}
	var snapshot deepseek.PageSnapshot
	if err := client.EvaluateValue(deepseek.SnapshotScript, &snapshot); err != nil {
		return deepseek.PageState{}, &commandError{code: "deepseek_page_unavailable", message: err.Error()}
	}
	return deepseek.AnalyzePage(snapshot), nil
}

func (a *app) ensureDeepSeekTab(client *browser.Client) error {
	if err := client.FindTab("https://chat.deepseek.com/", true); err == nil {
		return nil
	}
	if err := client.Navigate("https://chat.deepseek.com/", true); err != nil {
		return &commandError{code: "deepseek_page_unavailable", message: err.Error()}
	}
	return nil
}

func (a *app) askDeepSeek(client *browser.Client, prompt string) (askResult, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return askResult{}, fmt.Errorf("prompt is required")
	}
	if err := a.ensureDeepSeekTab(client); err != nil {
		return askResult{}, err
	}
	var baseline deepseek.AnswerSnapshot
	if err := client.EvaluateValue(deepseek.AnswerSnapshotScript, &baseline); err != nil {
		return askResult{}, err
	}
	if err := client.Fill("textarea", prompt); err != nil {
		return askResult{}, err
	}
	var submit deepseek.SubmitResult
	if err := client.EvaluateValue(deepseek.SubmitPromptScript, &submit); err != nil {
		return askResult{}, err
	}
	if !submit.OK {
		if submit.Error == "" {
			submit.Error = "send_button_missing"
		}
		return askResult{}, fmt.Errorf("%s", submit.Error)
	}
	answer, stable, err := a.waitForStableAnswer(client, baseline.Count)
	if err != nil {
		return askResult{}, err
	}
	return askResult{prompt: prompt, answer: answer, stableSamples: stable}, nil
}

func (a *app) waitForStableAnswer(client *browser.Client, baselineCount int) (string, int, error) {
	deadline := time.Now().Add(a.timeout)
	var previous string
	stable := 0
	for time.Now().Before(deadline) {
		var current deepseek.AnswerSnapshot
		if err := client.EvaluateValue(deepseek.AnswerSnapshotScript, &current); err != nil {
			return "", 0, err
		}
		answer := strings.TrimSpace(current.Latest)
		if current.Count > baselineCount && answer != "" {
			if answer == previous {
				stable++
			} else {
				previous = answer
				stable = 1
			}
			if stable >= 2 {
				return answer, stable, nil
			}
		} else {
			previous = ""
			stable = 0
		}
		time.Sleep(time.Second)
	}
	return "", 0, fmt.Errorf("timed out waiting for DeepSeek answer")
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
