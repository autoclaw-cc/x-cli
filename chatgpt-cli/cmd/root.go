package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	root.AddCommand(a.chatCommand())
	root.AddCommand(a.imageCommand())
	return root
}

func (a *app) imageCommand() *cobra.Command {
	image := &cobra.Command{
		Use:   "image",
		Short: "Generate and save images through ChatGPT Web",
	}
	image.AddCommand(a.imageGenerateCommand())
	return image
}

func (a *app) imageGenerateCommand() *cobra.Command {
	var prompt string
	var outDir string
	var confirm bool
	generate := &cobra.Command{
		Use:   "generate",
		Short: "Generate one image and save it locally",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt = strings.TrimSpace(prompt)
			if prompt == "" {
				return &commandError{code: "prompt_required", message: "--prompt is required"}
			}
			if !confirm {
				return &commandError{code: "confirmation_required", message: "image generation creates a ChatGPT conversation and consumes included account allowance; pass --confirm"}
			}
			client, err := a.readyBridgeClient()
			if err != nil {
				return err
			}
			defer func() { _ = client.CloseSession() }()
			result, err := chatgpt.GenerateImage(client, chatgpt.ImageOptions{
				Prompt:  prompt,
				OutDir:  outDir,
				Timeout: a.timeout,
			})
			if err != nil {
				return &commandError{code: "chatgpt_image_failed", message: err.Error()}
			}
			return output.Success(a.out, result)
		},
	}
	generate.Flags().StringVar(&prompt, "prompt", "", "image prompt")
	generate.Flags().StringVarP(&outDir, "out", "o", ".", "output directory")
	generate.Flags().BoolVar(&confirm, "confirm", false, "confirm one image generation using the signed-in account")
	return generate
}

func (a *app) chatCommand() *cobra.Command {
	chat := &cobra.Command{
		Use:   "chat",
		Short: "Run ChatGPT chat and research workflows",
	}
	chat.AddCommand(a.chatNewCommand())
	chat.AddCommand(a.chatAskCommand())
	return chat
}

func (a *app) chatNewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "new",
		Short: "Open a fresh ChatGPT conversation tab",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.readyBridgeClient()
			if err != nil {
				return err
			}
			if err := a.ensureChatGPTTab(client); err != nil {
				return err
			}
			return output.Success(a.out, map[string]any{
				"browser": "webbridge",
				"started": true,
				"url":     "https://chatgpt.com/",
			})
		},
	}
}

func (a *app) chatAskCommand() *cobra.Command {
	var prompt string
	var outputPath string
	var search bool
	var deepResearch bool
	ask := &cobra.Command{
		Use:   "ask",
		Short: "Ask ChatGPT and wait for a stable answer",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt = strings.TrimSpace(prompt)
			if prompt == "" {
				return &commandError{code: "prompt_required", message: "--prompt is required"}
			}
			if search && deepResearch {
				return &commandError{code: "modes_conflict", message: "--search and --deep-research are mutually exclusive"}
			}
			client, err := a.readyBridgeClient()
			if err != nil {
				return err
			}
			defer func() { _ = client.CloseSession() }()
			result, err := a.askChatGPT(client, prompt, search, deepResearch)
			if err != nil {
				return &commandError{code: "chatgpt_chat_failed", message: err.Error()}
			}
			data := map[string]any{
				"browser":        "webbridge",
				"prompt":         prompt,
				"answer":         result.answer,
				"citations":      result.citations,
				"mode":           result.mode,
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
	ask.Flags().StringVar(&prompt, "prompt", "", "prompt text to send to ChatGPT")
	ask.Flags().StringVar(&outputPath, "output", "", "optional local path for the answer text")
	ask.Flags().BoolVar(&search, "search", false, "enable ChatGPT web search")
	ask.Flags().BoolVar(&deepResearch, "deep-research", false, "enable ChatGPT Deep Research")
	return ask
}

type chatResult struct {
	answer        string
	citations     []chatgpt.Citation
	mode          string
	stableSamples int
}

func (a *app) askChatGPT(client *browser.Client, prompt string, search, deepResearch bool) (chatResult, error) {
	if err := a.ensureChatGPTTab(client); err != nil {
		return chatResult{}, err
	}
	mode := "chat"
	if search {
		mode = "web_search"
	}
	if deepResearch {
		mode = "deep_research"
	}
	if mode != "chat" {
		var selected chatgpt.ModeResult
		if err := client.EvaluateValue(chatgpt.SelectModeScript(mode), &selected); err != nil {
			return chatResult{}, err
		}
		if !selected.OK {
			return chatResult{}, fmt.Errorf("select %s: %s", mode, selected.Error)
		}
	}
	var baseline chatgpt.AnswerSnapshot
	if err := client.EvaluateValue(chatgpt.AnswerSnapshotScript, &baseline); err != nil {
		return chatResult{}, err
	}
	if err := client.Fill("#prompt-textarea", prompt); err != nil {
		return chatResult{}, err
	}
	if err := a.waitForSubmit(client); err != nil {
		return chatResult{}, err
	}
	answer, stable, citations, err := a.waitForAnswer(client, baseline.Count)
	if err != nil {
		return chatResult{}, err
	}
	return chatResult{answer: answer, citations: citations, mode: mode, stableSamples: stable}, nil
}

func (a *app) waitForSubmit(client *browser.Client) error {
	deadline := time.Now().Add(a.timeout)
	for time.Now().Before(deadline) {
		var result chatgpt.SubmitResult
		if err := client.EvaluateValue(chatgpt.SubmitPromptScript, &result); err != nil {
			return err
		}
		if result.OK {
			return nil
		}
		if result.Error != "send_button_not_ready" {
			return fmt.Errorf("submit prompt: %s", result.Error)
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for ChatGPT send button")
}

func (a *app) waitForAnswer(client *browser.Client, baselineCount int) (string, int, []chatgpt.Citation, error) {
	deadline := time.Now().Add(a.timeout)
	previous := ""
	stable := 0
	var citations []chatgpt.Citation
	for time.Now().Before(deadline) {
		var snapshot chatgpt.AnswerSnapshot
		if err := client.EvaluateValue(chatgpt.AnswerSnapshotScript, &snapshot); err != nil {
			return "", 0, nil, err
		}
		answer := strings.TrimSpace(snapshot.Latest)
		if snapshot.Count > baselineCount && answer != "" && !snapshot.Streaming {
			if answer == previous {
				stable++
			} else {
				previous = answer
				stable = 1
			}
			citations = snapshot.Citations
			if stable >= 3 {
				return answer, stable, citations, nil
			}
		} else {
			previous = ""
			stable = 0
		}
		time.Sleep(500 * time.Millisecond)
	}
	return "", 0, nil, fmt.Errorf("timed out waiting for ChatGPT answer")
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
	if err := client.BringToFront(); err != nil {
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
