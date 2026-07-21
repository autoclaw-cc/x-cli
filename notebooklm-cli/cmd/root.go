package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"notebooklm-cli/browser"
	"notebooklm-cli/config"
	"notebooklm-cli/notebooklm"
	"notebooklm-cli/output"
	"notebooklm-cli/registry"
)

type app struct {
	out          io.Writer
	errOut       io.Writer
	webbridgeURL string
	registryPath string
	timeout      time.Duration
}

type commandError struct {
	code    string
	message string
}

func (e *commandError) Error() string { return e.message }

func Execute(args []string, stdout, stderr io.Writer) int {
	a := &app{out: stdout, errOut: stderr, registryPath: registry.DefaultPath(), timeout: 5 * time.Minute}
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
		Use:           "notebooklm-cli",
		Short:         "Drive NotebookLM through an already signed-in Chrome session",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.PersistentFlags().StringVar(&a.webbridgeURL, "webbridge-url", "", "Kimi WebBridge daemon URL")
	root.PersistentFlags().StringVar(&a.registryPath, "registry", a.registryPath, "owned notebook registry path")
	root.PersistentFlags().DurationVar(&a.timeout, "timeout", a.timeout, "maximum workflow wait")
	root.AddCommand(a.loginStatusCommand(), a.capabilitiesCommand())
	root.AddCommand(a.notebookCommand(), a.sourceCommand(), a.chatCommand(), a.noteCommand(), a.researchCommand(), a.studioCommand())
	return root
}

func (a *app) loginStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login-status",
		Short: "Check the current NotebookLM login without returning account identity",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				result, err := notebooklm.CheckLogin(cmd.Context(), client)
				if err != nil {
					return err
				}
				return output.Success(a.out, result)
			})
		},
	}
}

func (a *app) capabilitiesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "capabilities",
		Short: "Inspect account-level controls visible in the current session",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				result, err := notebooklm.InspectAccountCapabilities(cmd.Context(), client)
				if err != nil {
					return err
				}
				return output.Success(a.out, result)
			})
		},
	}
}

func (a *app) notebookCommand() *cobra.Command {
	parent := &cobra.Command{Use: "notebook", Short: "Manage CLI-owned notebooks"}
	parent.AddCommand(a.notebookListCommand(), a.notebookAuthorizeCommand(), a.notebookCreateCommand())
	return parent
}

func (a *app) notebookListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List notebooks in the local ownership registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			return output.Success(a.out, r.Notebooks)
		},
	}
}

func (a *app) notebookAuthorizeCommand() *cobra.Command {
	var rawURL, title string
	var confirm bool
	command := &cobra.Command{
		Use:   "authorize",
		Short: "Explicitly mark one notebook as CLI-owned",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return &commandError{code: "confirmation_required", message: "pass --confirm to authorize this exact notebook URL"}
			}
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			notebook, err := r.Authorize(rawURL, title)
			if err != nil {
				return &commandError{code: "invalid_args", message: err.Error()}
			}
			if err := r.Save(); err != nil {
				return err
			}
			return output.Success(a.out, notebook)
		},
	}
	command.Flags().StringVar(&rawURL, "url", "", "exact NotebookLM notebook URL")
	command.Flags().StringVar(&title, "title", "", "non-secret local title")
	command.Flags().BoolVar(&confirm, "confirm", false, "confirm this exact notebook is authorized for CLI writes")
	_ = command.MarkFlagRequired("url")
	_ = command.MarkFlagRequired("title")
	return command
}

func (a *app) notebookCreateCommand() *cobra.Command {
	var title string
	command := &cobra.Command{
		Use:   "create",
		Short: "Create and register a new CLI-owned notebook",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				created, err := notebooklm.CreateNotebook(cmd.Context(), client, title, a.timeout)
				if err != nil {
					return err
				}
				notebook, err := r.Authorize(created.URL, created.Title)
				if err != nil {
					return err
				}
				if err := r.Save(); err != nil {
					return err
				}
				return output.Success(a.out, notebook)
			})
		},
	}
	command.Flags().StringVar(&title, "title", "", "title for the new notebook")
	_ = command.MarkFlagRequired("title")
	return command
}

func (a *app) sourceCommand() *cobra.Command {
	parent := &cobra.Command{Use: "source", Short: "Manage sources in CLI-owned notebooks"}
	var notebookID, text, file string
	command := &cobra.Command{
		Use:   "add-text",
		Short: "Add pasted text from --text or a local UTF-8 file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if (text == "") == (file == "") {
				return &commandError{code: "invalid_args", message: "provide exactly one of --text or --file"}
			}
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			owned, err := r.RequireOwned(notebookID)
			if err != nil {
				return &commandError{code: "notebook_not_owned", message: err.Error()}
			}
			content := text
			if file != "" {
				body, err := os.ReadFile(file)
				if err != nil {
					return &commandError{code: "invalid_args", message: fmt.Sprintf("read source file: %v", err)}
				}
				if len(body) > 5<<20 {
					return &commandError{code: "invalid_args", message: "pasted text file exceeds 5 MiB"}
				}
				content = string(body)
			}
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				result, err := notebooklm.AddTextSource(cmd.Context(), client, owned.URL, content, a.timeout)
				if err != nil {
					return err
				}
				return output.Success(a.out, result)
			})
		},
	}
	command.Flags().StringVar(&notebookID, "notebook", "", "owned notebook ID")
	command.Flags().StringVar(&text, "text", "", "text to paste as a source")
	command.Flags().StringVar(&file, "file", "", "UTF-8 text file to paste as a source")
	_ = command.MarkFlagRequired("notebook")
	parent.AddCommand(command, a.sourceAddURLCommand())
	return parent
}

func (a *app) sourceAddURLCommand() *cobra.Command {
	var notebookID, rawURL string
	command := &cobra.Command{
		Use:   "add-url",
		Short: "Add a public website or YouTube URL as a source",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			owned, err := r.RequireOwned(notebookID)
			if err != nil {
				return &commandError{code: "notebook_not_owned", message: err.Error()}
			}
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				result, err := notebooklm.AddURLSource(cmd.Context(), client, owned.URL, rawURL, a.timeout)
				if err != nil {
					return err
				}
				return output.Success(a.out, result)
			})
		},
	}
	command.Flags().StringVar(&notebookID, "notebook", "", "owned notebook ID")
	command.Flags().StringVar(&rawURL, "url", "", "absolute public website or YouTube URL")
	_ = command.MarkFlagRequired("notebook")
	_ = command.MarkFlagRequired("url")
	return command
}

func (a *app) chatCommand() *cobra.Command {
	parent := &cobra.Command{Use: "chat", Short: "Ask grounded questions in CLI-owned notebooks"}
	var notebookID, question string
	command := &cobra.Command{
		Use:   "ask",
		Short: "Ask a question and return the grounded answer and citation count",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			owned, err := r.RequireOwned(notebookID)
			if err != nil {
				return &commandError{code: "notebook_not_owned", message: err.Error()}
			}
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				result, err := notebooklm.Ask(cmd.Context(), client, owned.URL, question, a.timeout)
				if err != nil {
					return err
				}
				return output.Success(a.out, result)
			})
		},
	}
	command.Flags().StringVar(&notebookID, "notebook", "", "owned notebook ID")
	command.Flags().StringVar(&question, "question", "", "grounded question")
	_ = command.MarkFlagRequired("notebook")
	_ = command.MarkFlagRequired("question")
	parent.AddCommand(command)
	return parent
}

type researchEvidence struct {
	NotebookID string                    `json:"notebook_id"`
	Mode       string                    `json:"mode"`
	ObservedAt string                    `json:"observed_at"`
	OutputPath string                    `json:"output_path,omitempty"`
	Result     notebooklm.ResearchResult `json:"result"`
}

func (a *app) researchCommand() *cobra.Command {
	parent := &cobra.Command{Use: "research", Short: "Run NotebookLM Fast or Deep Research in CLI-owned notebooks"}
	var notebookID, mode, query, outPath string
	var importResults bool
	command := &cobra.Command{
		Use:   "run",
		Short: "Run Fast Research or Deep Research and optionally import discovered sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			owned, err := r.RequireOwned(notebookID)
			if err != nil {
				return &commandError{code: "notebook_not_owned", message: err.Error()}
			}
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				result, err := notebooklm.RunResearch(cmd.Context(), client, owned.URL, mode, query, importResults, a.timeout)
				if err != nil {
					return err
				}
				evidence := researchEvidence{
					NotebookID: notebookID,
					Mode:       result.Mode,
					ObservedAt: time.Now().UTC().Format(time.RFC3339),
					OutputPath: outPath,
					Result:     *result,
				}
				if outPath != "" {
					if err := writeJSONEvidence(outPath, evidence); err != nil {
						return &commandError{code: "write_failed", message: err.Error()}
					}
				}
				return output.Success(a.out, evidence)
			})
		},
	}
	command.Flags().StringVar(&notebookID, "notebook", "", "owned notebook ID")
	command.Flags().StringVar(&mode, "mode", "fast", "fast or deep")
	command.Flags().StringVar(&query, "query", "", "research query")
	command.Flags().BoolVar(&importResults, "import", false, "import discovered sources after research completes")
	command.Flags().StringVar(&outPath, "out", "", "optional local JSON evidence path")
	_ = command.MarkFlagRequired("notebook")
	_ = command.MarkFlagRequired("query")
	parent.AddCommand(command)
	return parent
}

func (a *app) noteCommand() *cobra.Command {
	parent := &cobra.Command{Use: "note", Short: "Manage editable notes in CLI-owned notebooks"}
	parent.AddCommand(a.noteCreateCommand(), a.noteListCommand(), a.noteToSourceCommand())
	return parent
}

func (a *app) noteCreateCommand() *cobra.Command {
	var notebookID, title, text, file string
	command := &cobra.Command{
		Use:   "create",
		Short: "Create an editable note and verify it after reopening",
		RunE: func(cmd *cobra.Command, args []string) error {
			if (text == "") == (file == "") {
				return &commandError{code: "invalid_args", message: "provide exactly one of --text or --file"}
			}
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			owned, err := r.RequireOwned(notebookID)
			if err != nil {
				return &commandError{code: "notebook_not_owned", message: err.Error()}
			}
			content := text
			if file != "" {
				body, err := os.ReadFile(file)
				if err != nil {
					return &commandError{code: "invalid_args", message: fmt.Sprintf("read note file: %v", err)}
				}
				if len(body) > 5<<20 {
					return &commandError{code: "invalid_args", message: "note file exceeds 5 MiB"}
				}
				content = string(body)
			}
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				result, err := notebooklm.CreateNote(cmd.Context(), client, owned.URL, title, content, a.timeout)
				if err != nil {
					return err
				}
				return output.Success(a.out, result)
			})
		},
	}
	command.Flags().StringVar(&notebookID, "notebook", "", "owned notebook ID")
	command.Flags().StringVar(&title, "title", "", "unique note title")
	command.Flags().StringVar(&text, "text", "", "plain text note body")
	command.Flags().StringVar(&file, "file", "", "UTF-8 file to use as the note body")
	_ = command.MarkFlagRequired("notebook")
	_ = command.MarkFlagRequired("title")
	return command
}

func (a *app) noteListCommand() *cobra.Command {
	var notebookID string
	command := &cobra.Command{
		Use:   "list",
		Short: "List editable and saved-answer note titles",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			owned, err := r.RequireOwned(notebookID)
			if err != nil {
				return &commandError{code: "notebook_not_owned", message: err.Error()}
			}
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				result, err := notebooklm.ListNotes(cmd.Context(), client, owned.URL, a.timeout)
				if err != nil {
					return err
				}
				return output.Success(a.out, result)
			})
		},
	}
	command.Flags().StringVar(&notebookID, "notebook", "", "owned notebook ID")
	_ = command.MarkFlagRequired("notebook")
	return command
}

func (a *app) noteToSourceCommand() *cobra.Command {
	var notebookID, title string
	command := &cobra.Command{
		Use:   "to-source",
		Short: "Convert a unique note title into a NotebookLM source",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			owned, err := r.RequireOwned(notebookID)
			if err != nil {
				return &commandError{code: "notebook_not_owned", message: err.Error()}
			}
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				result, err := notebooklm.ConvertNoteToSource(cmd.Context(), client, owned.URL, title, a.timeout)
				if err != nil {
					return err
				}
				return output.Success(a.out, result)
			})
		},
	}
	command.Flags().StringVar(&notebookID, "notebook", "", "owned notebook ID")
	command.Flags().StringVar(&title, "title", "", "exact unique note title")
	_ = command.MarkFlagRequired("notebook")
	_ = command.MarkFlagRequired("title")
	return command
}

func (a *app) studioCommand() *cobra.Command {
	parent := &cobra.Command{Use: "studio", Short: "Inspect Studio in CLI-owned notebooks"}
	var notebookID string
	command := &cobra.Command{
		Use:   "capabilities",
		Short: "List Studio output types visible for an owned notebook",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			owned, err := r.RequireOwned(notebookID)
			if err != nil {
				return &commandError{code: "notebook_not_owned", message: err.Error()}
			}
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				result, err := notebooklm.InspectStudio(cmd.Context(), client, owned.URL)
				if err != nil {
					return err
				}
				return output.Success(a.out, result)
			})
		},
	}
	command.Flags().StringVar(&notebookID, "notebook", "", "owned notebook ID")
	_ = command.MarkFlagRequired("notebook")
	parent.AddCommand(command, a.studioListCommand(), a.studioGenerateCommand(), a.studioWaitCommand(), a.studioExportCommand(), a.studioInspectCommand(), a.studioRenameCommand(), a.studioDeleteCommand(), a.studioDownloadCommand())
	return parent
}

func (a *app) studioListCommand() *cobra.Command {
	var notebookID string
	command := &cobra.Command{
		Use:   "list",
		Short: "List typed Studio artifacts and their visible state",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			owned, err := r.RequireOwned(notebookID)
			if err != nil {
				return &commandError{code: "notebook_not_owned", message: err.Error()}
			}
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				result, err := notebooklm.ListStudioArtifacts(cmd.Context(), client, owned.URL, a.timeout)
				if err != nil {
					return err
				}
				return output.Success(a.out, result)
			})
		},
	}
	command.Flags().StringVar(&notebookID, "notebook", "", "owned notebook ID")
	_ = command.MarkFlagRequired("notebook")
	return command
}

func (a *app) studioGenerateCommand() *cobra.Command {
	var notebookID, kind, prompt, waitMode string
	command := &cobra.Command{
		Use:   "generate",
		Short: "Start a Studio output for an owned notebook",
		RunE: func(cmd *cobra.Command, args []string) error {
			waitReady := false
			switch waitMode {
			case "started":
				waitReady = false
			case "ready":
				waitReady = true
			default:
				return &commandError{code: "invalid_args", message: "wait must be started or ready"}
			}
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			owned, err := r.RequireOwned(notebookID)
			if err != nil {
				return &commandError{code: "notebook_not_owned", message: err.Error()}
			}
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				result, err := notebooklm.GenerateStudioArtifact(cmd.Context(), client, owned.URL, kind, prompt, waitReady, a.timeout)
				if err != nil {
					return err
				}
				return output.Success(a.out, result)
			})
		},
	}
	command.Flags().StringVar(&notebookID, "notebook", "", "owned notebook ID")
	command.Flags().StringVar(&kind, "type", "", "Studio output type: audio, presentation, video, mind_map, report, flashcards, quiz, infographic, data_table")
	command.Flags().StringVar(&prompt, "prompt", "", "optional prompt or topic when the Studio dialog exposes a text field")
	command.Flags().StringVar(&waitMode, "wait", "started", "started or ready")
	_ = command.MarkFlagRequired("notebook")
	_ = command.MarkFlagRequired("type")
	return command
}

type studioWaitEvidence struct {
	NotebookID string                    `json:"notebook_id"`
	Type       string                    `json:"type"`
	ObservedAt string                    `json:"observed_at"`
	OutputPath string                    `json:"output_path,omitempty"`
	Artifact   notebooklm.StudioArtifact `json:"artifact"`
}

type studioExportEvidence struct {
	NotebookID string                        `json:"notebook_id"`
	Type       string                        `json:"type"`
	ObservedAt string                        `json:"observed_at"`
	OutputPath string                        `json:"output_path,omitempty"`
	Result     notebooklm.StudioExportResult `json:"result"`
}

func (a *app) studioWaitCommand() *cobra.Command {
	var notebookID, kind, outPath string
	command := &cobra.Command{
		Use:   "wait",
		Short: "Wait for a Studio artifact to become ready and optionally write evidence",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			owned, err := r.RequireOwned(notebookID)
			if err != nil {
				return &commandError{code: "notebook_not_owned", message: err.Error()}
			}
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				artifact, err := notebooklm.WaitStudioArtifactReady(cmd.Context(), client, owned.URL, kind, a.timeout)
				if err != nil {
					return err
				}
				evidence := studioWaitEvidence{
					NotebookID: notebookID,
					Type:       artifact.Type,
					ObservedAt: time.Now().UTC().Format(time.RFC3339),
					OutputPath: outPath,
					Artifact:   *artifact,
				}
				if outPath != "" {
					if err := writeStudioWaitEvidence(outPath, evidence); err != nil {
						return &commandError{code: "write_failed", message: err.Error()}
					}
				}
				return output.Success(a.out, evidence)
			})
		},
	}
	command.Flags().StringVar(&notebookID, "notebook", "", "owned notebook ID")
	command.Flags().StringVar(&kind, "type", "", "Studio output type to wait for")
	command.Flags().StringVar(&outPath, "out", "", "optional local JSON evidence path")
	_ = command.MarkFlagRequired("notebook")
	_ = command.MarkFlagRequired("type")
	return command
}

func (a *app) studioExportCommand() *cobra.Command {
	var notebookID, kind, title, outPath string
	command := &cobra.Command{
		Use:   "export",
		Short: "Export a unique ready Studio artifact body and metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			owned, err := r.RequireOwned(notebookID)
			if err != nil {
				return &commandError{code: "notebook_not_owned", message: err.Error()}
			}
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				result, err := notebooklm.ExportStudioArtifact(cmd.Context(), client, owned.URL, kind, title, a.timeout)
				if err != nil {
					return err
				}
				evidence := studioExportEvidence{
					NotebookID: notebookID,
					Type:       result.Artifact.Type,
					ObservedAt: time.Now().UTC().Format(time.RFC3339),
					OutputPath: outPath,
					Result:     *result,
				}
				if outPath != "" {
					if err := writeStudioExportEvidence(outPath, evidence); err != nil {
						return &commandError{code: "write_failed", message: err.Error()}
					}
				}
				return output.Success(a.out, evidence)
			})
		},
	}
	command.Flags().StringVar(&notebookID, "notebook", "", "owned notebook ID")
	command.Flags().StringVar(&kind, "type", "", "Studio output type to export")
	command.Flags().StringVar(&title, "title", "", "exact unique Studio artifact title")
	command.Flags().StringVar(&outPath, "out", "", "optional local JSON export path")
	_ = command.MarkFlagRequired("notebook")
	_ = command.MarkFlagRequired("type")
	_ = command.MarkFlagRequired("title")
	return command
}

type studioInspectEvidence struct {
	NotebookID string                             `json:"notebook_id"`
	Type       string                             `json:"type"`
	ObservedAt string                             `json:"observed_at"`
	OutputPath string                             `json:"output_path,omitempty"`
	Result     notebooklm.StudioAttributionResult `json:"result"`
}

func (a *app) studioInspectCommand() *cobra.Command {
	var notebookID, kind, title, outPath string
	command := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect prompt and source attribution for a Studio artifact",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			owned, err := r.RequireOwned(notebookID)
			if err != nil {
				return &commandError{code: "notebook_not_owned", message: err.Error()}
			}
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				result, err := notebooklm.InspectStudioAttribution(cmd.Context(), client, owned.URL, kind, title, a.timeout)
				if err != nil {
					return err
				}
				evidence := studioInspectEvidence{
					NotebookID: notebookID,
					Type:       result.Artifact.Type,
					ObservedAt: time.Now().UTC().Format(time.RFC3339),
					OutputPath: outPath,
					Result:     *result,
				}
				if outPath != "" {
					if err := writeJSONEvidence(outPath, evidence); err != nil {
						return &commandError{code: "write_failed", message: err.Error()}
					}
				}
				return output.Success(a.out, evidence)
			})
		},
	}
	command.Flags().StringVar(&notebookID, "notebook", "", "owned notebook ID")
	command.Flags().StringVar(&kind, "type", "", "Studio output type to inspect")
	command.Flags().StringVar(&title, "title", "", "exact unique Studio artifact title")
	command.Flags().StringVar(&outPath, "out", "", "optional local JSON evidence path")
	_ = command.MarkFlagRequired("notebook")
	_ = command.MarkFlagRequired("type")
	_ = command.MarkFlagRequired("title")
	return command
}

func (a *app) studioRenameCommand() *cobra.Command {
	var notebookID, kind, title, newTitle string
	command := &cobra.Command{
		Use:   "rename",
		Short: "Rename a unique Studio artifact",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			owned, err := r.RequireOwned(notebookID)
			if err != nil {
				return &commandError{code: "notebook_not_owned", message: err.Error()}
			}
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				result, err := notebooklm.RenameStudioArtifact(cmd.Context(), client, owned.URL, kind, title, newTitle, a.timeout)
				if err != nil {
					return err
				}
				return output.Success(a.out, result)
			})
		},
	}
	command.Flags().StringVar(&notebookID, "notebook", "", "owned notebook ID")
	command.Flags().StringVar(&kind, "type", "", "Studio output type to rename")
	command.Flags().StringVar(&title, "title", "", "exact current Studio artifact title")
	command.Flags().StringVar(&newTitle, "new-title", "", "new artifact title")
	_ = command.MarkFlagRequired("notebook")
	_ = command.MarkFlagRequired("type")
	_ = command.MarkFlagRequired("title")
	_ = command.MarkFlagRequired("new-title")
	return command
}

func (a *app) studioDeleteCommand() *cobra.Command {
	var notebookID, kind, title string
	var confirm bool
	command := &cobra.Command{
		Use:   "delete",
		Short: "Delete a unique Studio artifact after explicit confirmation",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return &commandError{code: "confirmation_required", message: "pass --confirm to delete this exact Studio artifact"}
			}
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			owned, err := r.RequireOwned(notebookID)
			if err != nil {
				return &commandError{code: "notebook_not_owned", message: err.Error()}
			}
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				result, err := notebooklm.DeleteStudioArtifact(cmd.Context(), client, owned.URL, kind, title, a.timeout)
				if err != nil {
					return err
				}
				return output.Success(a.out, result)
			})
		},
	}
	command.Flags().StringVar(&notebookID, "notebook", "", "owned notebook ID")
	command.Flags().StringVar(&kind, "type", "", "Studio output type to delete")
	command.Flags().StringVar(&title, "title", "", "exact unique Studio artifact title")
	command.Flags().BoolVar(&confirm, "confirm", false, "confirm deletion of this exact artifact")
	_ = command.MarkFlagRequired("notebook")
	_ = command.MarkFlagRequired("type")
	_ = command.MarkFlagRequired("title")
	return command
}

func (a *app) studioDownloadCommand() *cobra.Command {
	var notebookID, kind, title, outPath string
	command := &cobra.Command{
		Use:   "download",
		Short: "Attempt a managed raw Studio media download when WebBridge permits it",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := registry.Load(a.registryPath)
			if err != nil {
				return err
			}
			owned, err := r.RequireOwned(notebookID)
			if err != nil {
				return &commandError{code: "notebook_not_owned", message: err.Error()}
			}
			dir := filepath.Dir(outPath)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return &commandError{code: "write_failed", message: err.Error()}
			}
			return a.withClient(cmd.Context(), func(client *browser.Client) error {
				result, err := notebooklm.DownloadStudioArtifact(cmd.Context(), client, owned.URL, kind, title, outPath, a.timeout)
				if err != nil {
					return err
				}
				return output.Success(a.out, result)
			})
		},
	}
	command.Flags().StringVar(&notebookID, "notebook", "", "owned notebook ID")
	command.Flags().StringVar(&kind, "type", "", "Studio output type to download")
	command.Flags().StringVar(&title, "title", "", "exact unique Studio artifact title")
	command.Flags().StringVar(&outPath, "out", "", "local file path for the raw media bytes")
	_ = command.MarkFlagRequired("notebook")
	_ = command.MarkFlagRequired("type")
	_ = command.MarkFlagRequired("title")
	_ = command.MarkFlagRequired("out")
	return command
}

func writeStudioWaitEvidence(path string, evidence studioWaitEvidence) error {
	return writeJSONEvidence(path, evidence)
}

func writeStudioExportEvidence(path string, evidence studioExportEvidence) error {
	if evidence.Result.BodyCharacters == 0 && evidence.Result.Body != "" {
		evidence.Result.BodyCharacters = len([]rune(evidence.Result.Body))
	}
	return writeJSONEvidence(path, evidence)
}

func writeJSONEvidence(path string, value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode evidence: %w", err)
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create evidence directory: %w", err)
		}
	}
	if err := os.WriteFile(path, append(body, '\n'), 0o600); err != nil {
		return fmt.Errorf("write evidence: %w", err)
	}
	return nil
}

func (a *app) withClient(ctx context.Context, fn func(*browser.Client) error) error {
	baseURL := config.ResolveWebBridgeURL(a.webbridgeURL, os.Getenv)
	session := fmt.Sprintf("notebooklm-cli-%d", time.Now().UnixNano())
	client := browser.NewClient(baseURL, session)
	status, err := client.Status()
	if err != nil {
		return &commandError{code: "daemon_unreachable", message: err.Error()}
	}
	if !status.Running {
		return &commandError{code: "daemon_not_running", message: "Kimi WebBridge daemon is not running"}
	}
	if !status.ExtensionConnected {
		return &commandError{code: "extension_not_connected", message: "Kimi WebBridge extension is not connected"}
	}
	defer client.CloseSession()
	if err := ctx.Err(); err != nil {
		return err
	}
	return fn(client)
}

func classifyError(err error) (string, string) {
	var coded *commandError
	if ok := asCommandError(err, &coded); ok {
		return coded.code, coded.message
	}
	message := err.Error()
	for _, code := range []string{"not_logged_in", "captcha_required", "timeout", "notebook_not_owned", "download_unavailable", "research_result_present"} {
		if strings.HasPrefix(message, code+":") {
			return code, strings.TrimSpace(strings.TrimPrefix(message, code+":"))
		}
	}
	return "command_failed", message
}

func asCommandError(err error, target **commandError) bool {
	value, ok := err.(*commandError)
	if ok {
		*target = value
	}
	return ok
}
