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
	chat.AddCommand(a.chatNewCommand())
	return chat
}

func (a *app) chatAskCommand() *cobra.Command {
	var prompt string
	var outputPath string
	var deepthink bool
	var search bool
	var files []string
	var images []string
	ask := &cobra.Command{
		Use:   "ask",
		Short: "Ask DeepSeek and wait for a stable assistant answer",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.readyBridgeClient()
			if err != nil {
				return err
			}
			result, err := a.askDeepSeek(client, prompt, askOptions{
				deepthink: deepthink,
				search:    search,
				files:     files,
				images:    images,
			})
			if err != nil {
				return &commandError{code: "deepseek_chat_failed", message: err.Error()}
			}
			data := map[string]any{
				"browser":        "webbridge",
				"prompt":         result.prompt,
				"answer":         result.answer,
				"stable_samples": result.stableSamples,
				"modes":          result.modes,
				"files":          result.files,
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
	ask.Flags().BoolVar(&deepthink, "deepthink", false, "enable DeepSeek deep thinking mode before sending")
	ask.Flags().BoolVar(&search, "search", false, "enable DeepSeek web search mode before sending")
	ask.Flags().StringArrayVar(&files, "file", nil, "local file path to attach before sending; repeatable")
	ask.Flags().StringArrayVar(&images, "image", nil, "local image path to attach before sending; repeatable")
	_ = ask.MarkFlagRequired("prompt")
	return ask
}

func (a *app) chatNewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "new",
		Short: "Start a new DeepSeek chat through the visible page",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.readyBridgeClient()
			if err != nil {
				return err
			}
			if err := a.ensureDeepSeekTab(client); err != nil {
				return err
			}
			if err := client.BringToFront(); err != nil {
				return &commandError{code: "deepseek_new_chat_failed", message: err.Error()}
			}
			if err := client.Navigate("https://chat.deepseek.com/", false); err != nil {
				return &commandError{code: "deepseek_new_chat_failed", message: err.Error()}
			}
			return output.Success(a.out, map[string]any{
				"browser": "webbridge",
				"started": true,
			})
		},
	}
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
	modes         []string
	files         []string
}

type askOptions struct {
	deepthink bool
	search    bool
	files     []string
	images    []string
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

func (a *app) askDeepSeek(client *browser.Client, prompt string, options askOptions) (askResult, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return askResult{}, fmt.Errorf("prompt is required")
	}
	if err := a.ensureDeepSeekTab(client); err != nil {
		return askResult{}, err
	}
	if err := client.BringToFront(); err != nil {
		return askResult{}, err
	}
	modes, err := a.setDeepSeekModes(client, options)
	if err != nil {
		return askResult{}, err
	}
	attached, err := a.uploadDeepSeekFiles(client, options)
	if err != nil {
		return askResult{}, err
	}
	var baseline deepseek.AnswerSnapshot
	if err := client.EvaluateValue(deepseek.AnswerSnapshotScript, &baseline); err != nil {
		return askResult{}, err
	}
	if err := client.Fill("textarea", prompt); err != nil {
		return askResult{}, err
	}
	if err := a.waitForPromptSubmit(client); err != nil {
		return askResult{}, err
	}
	answer, stable, err := a.waitForStableAnswer(client, baseline.Count)
	if err != nil {
		return askResult{}, err
	}
	return askResult{prompt: prompt, answer: answer, stableSamples: stable, modes: modes, files: attached}, nil
}

func (a *app) setDeepSeekModes(client *browser.Client, options askOptions) ([]string, error) {
	vision := len(options.images) > 0
	if !options.deepthink && !options.search && !vision {
		return []string{}, nil
	}
	requested := []struct {
		name                      string
		deepthink, search, vision bool
	}{
		{name: "deepthink", deepthink: options.deepthink},
		{name: "web_search", search: options.search},
		{name: "vision", vision: vision},
	}
	enabled := []string{}
	for _, mode := range requested {
		if !mode.deepthink && !mode.search && !mode.vision {
			continue
		}
		var result deepseek.ModeResult
		if err := client.EvaluateValue(deepseek.SetModesScript(mode.deepthink, mode.search, mode.vision), &result); err != nil {
			return nil, err
		}
		if !result.OK {
			if result.Error == "" {
				result.Error = "mode_set_failed"
			}
			return nil, fmt.Errorf("%s", result.Error)
		}
		if err := a.waitForMode(client, mode.name); err != nil {
			return nil, err
		}
		enabled = append(enabled, result.Enabled...)
	}
	return enabled, nil
}

func (a *app) waitForMode(client *browser.Client, mode string) error {
	deadline := time.Now().Add(a.timeout)
	for time.Now().Before(deadline) {
		var state deepseek.ModeReadyResult
		if err := client.EvaluateValue(deepseek.ModeReadyScript(mode), &state); err != nil {
			return err
		}
		if state.Ready {
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for DeepSeek %s mode", mode)
}

func (a *app) uploadDeepSeekFiles(client *browser.Client, options askOptions) ([]string, error) {
	paths := append([]string{}, options.files...)
	paths = append(paths, options.images...)
	if len(paths) == 0 {
		return []string{}, nil
	}
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			return nil, fmt.Errorf("empty attachment path")
		}
		if _, err := os.Stat(path); err != nil {
			return nil, fmt.Errorf("attachment unavailable %s: %w", path, err)
		}
	}
	var target deepseek.UploadTargetResult
	if err := client.EvaluateValue(deepseek.PrepareUploadScript, &target); err != nil {
		return nil, err
	}
	if !target.OK {
		if target.Error == "" {
			target.Error = "file_input_missing"
		}
		return nil, fmt.Errorf("%s", target.Error)
	}
	if target.Selector == "" {
		target.Selector = "input[type=file]"
	}
	if err := client.Upload(target.Selector, paths); err != nil {
		return nil, err
	}
	return paths, nil
}

func (a *app) waitForPromptSubmit(client *browser.Client) error {
	deadline := time.Now().Add(a.timeout)
	for time.Now().Before(deadline) {
		var submit deepseek.SubmitResult
		if err := client.EvaluateValue(deepseek.SubmitPromptScript, &submit); err != nil {
			return err
		}
		if submit.OK {
			return nil
		}
		if submit.Error != "send_button_not_ready" {
			if submit.Error == "" {
				submit.Error = "send_button_missing"
			}
			return fmt.Errorf("%s", submit.Error)
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("timed out waiting for DeepSeek send button")
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
