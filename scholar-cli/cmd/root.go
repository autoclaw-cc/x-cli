package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"scholar-cli/browser"
	"scholar-cli/download"
	"scholar-cli/output"
	"scholar-cli/paper"
	"scholar-cli/search"
	"scholar-cli/store"
)

var rootCmd = &cobra.Command{
	Use:   "scholar-cli",
	Short: "Academic paper search CLI — multi-source parallel search, dedup, BibTeX export",
}

func Execute() error {
	return rootCmd.Execute()
}

// checkDaemon emits a structured error and exits if the kimi-webbridge daemon
// or the Chrome extension isn't reachable. Used by the WebBridge-backed
// subcommands (search-google, search-cnki, search-wos, login-status) so their
// failure modes stay inside the {ok:false, error:{code,message}} contract.
func checkDaemon(client *browser.Client) {
	st, err := client.Status()
	if err != nil {
		output.Error("daemon_unreachable", err.Error())
		os.Exit(1)
	}
	if !st.Running {
		output.Error("daemon_not_running", "kimi-webbridge daemon is not running (open the Kimi Desktop App)")
		os.Exit(1)
	}
	if !st.ExtensionConnected {
		output.Error("extension_not_connected", "Chrome WebBridge extension is not connected (see https://www.kimi.com/features/webbridge)")
		os.Exit(1)
	}
}

// emitSearch prints a single-source search result, persisting it into the
// workspace first when one was given.
//
// The browser-backed sources used to print and stop, which quietly cut them out
// of the rest of the pipeline: `export` reads only from the workspace, so a
// Google Scholar hit could never reach BibTeX without the user retyping the
// title into search-en. Sharing the workspace also lets dedup merge a paper
// found in two places — arXiv contributes the PDF link, Google Scholar the
// citation count.
func emitSearch(papers []paper.Paper, source string, workspace string) {
	if workspace == "" {
		output.Success(map[string]any{
			"papers": papers,
			"total":  len(papers),
			"source": source,
		})
		return
	}

	s, err := store.Open(workspace)
	if err != nil {
		output.Error("store_error", err.Error())
		os.Exit(1)
	}
	added := s.AddPapers(papers)
	if err := s.Save(); err != nil {
		output.Error("store_error", err.Error())
		os.Exit(1)
	}
	output.Success(map[string]any{
		"papers":       papers,
		"total":        len(papers),
		"source":       source,
		"workspace":    workspace,
		"papers_added": added,
		"total_stored": s.Count(),
	})
}

func init() {
	searchEnCmd := &cobra.Command{
		Use:   "search-en",
		Short: "Search English academic sources (OpenAlex, Semantic Scholar, CrossRef, arXiv, PubMed)",
		Run: func(cmd *cobra.Command, args []string) {
			query, _ := cmd.Flags().GetString("query")
			limit, _ := cmd.Flags().GetInt("limit")
			sourcesFlag, _ := cmd.Flags().GetString("sources")
			workspace, _ := cmd.Flags().GetString("workspace")

			if query == "" {
				output.Error("missing_param", "--query is required")
				os.Exit(1)
			}

			var sourceNames []string
			if sourcesFlag != "" {
				sourceNames = strings.Split(sourcesFlag, ",")
			}

			result, err := search.SearchEnglish(cmd.Context(), query, limit, sourceNames)
			if err != nil {
				output.Error("search_error", err.Error())
				os.Exit(1)
			}

			if workspace != "" {
				s, err := store.Open(workspace)
				if err != nil {
					output.Error("store_error", err.Error())
					os.Exit(1)
				}
				added := s.AddPapers(result.Papers)
				if err := s.Save(); err != nil {
					output.Error("store_error", err.Error())
					os.Exit(1)
				}
				output.Success(map[string]any{
					"search":       result,
					"workspace":    workspace,
					"papers_added": added,
					"total_stored": s.Count(),
				})
			} else {
				output.Success(result)
			}
		},
	}
	searchEnCmd.Flags().String("query", "", "Search query (required)")
	searchEnCmd.Flags().Int("limit", 10, "Max results per source")
	searchEnCmd.Flags().String("sources", "", "Comma-separated source list (openalex,semantic,crossref,arxiv,pubmed). Default: all")
	searchEnCmd.Flags().String("workspace", "", "Workspace directory to save results (auto-dedup)")

	searchGoogleCmd := &cobra.Command{
		Use:   "search-google",
		Short: "Search Google Scholar via WebBridge (requires browser)",
		Run: func(cmd *cobra.Command, args []string) {
			query, _ := cmd.Flags().GetString("query")
			limit, _ := cmd.Flags().GetInt("limit")
			workspace, _ := cmd.Flags().GetString("workspace")

			if query == "" {
				output.Error("missing_param", "--query is required")
				os.Exit(1)
			}

			client := withActivateMode(browser.NewClient("scholar-cli"))
			checkDaemon(client)
			src := search.NewGoogleScholar(client)
			papers, err := src.Search(cmd.Context(), query, limit)
			if err != nil {
				output.Error("search_error", err.Error())
				os.Exit(1)
			}
			emitSearch(papers, "google_scholar", workspace)
		},
	}
	searchGoogleCmd.Flags().String("query", "", "Search query (required)")
	searchGoogleCmd.Flags().Int("limit", 10, "Max results")
	searchGoogleCmd.Flags().String("workspace", "", "Workspace directory to save results (auto-dedup)")

	searchCnkiCmd := &cobra.Command{
		Use:   "search-cnki",
		Short: "Search CNKI (知网) via WebBridge (requires login)",
		Run: func(cmd *cobra.Command, args []string) {
			query, _ := cmd.Flags().GetString("query")
			limit, _ := cmd.Flags().GetInt("limit")
			workspace, _ := cmd.Flags().GetString("workspace")

			if query == "" {
				output.Error("missing_param", "--query is required")
				os.Exit(1)
			}

			client := withActivateMode(browser.NewClient("scholar-cli"))
			checkDaemon(client)
			src := search.NewCNKI(client)
			papers, err := src.Search(cmd.Context(), query, limit)
			if err != nil {
				output.Error("search_error", err.Error())
				os.Exit(1)
			}
			emitSearch(papers, "cnki", workspace)
		},
	}
	searchCnkiCmd.Flags().String("query", "", "Search query in Chinese (required)")
	searchCnkiCmd.Flags().String("workspace", "", "Workspace directory to save results (auto-dedup)")
	searchCnkiCmd.Flags().Int("limit", 20, "Max results")

	detailCmd := &cobra.Command{
		Use:   "detail",
		Short: "Enrich a paper by DOI (fetches from CrossRef + Semantic Scholar)",
		Run: func(cmd *cobra.Command, args []string) {
			doi, _ := cmd.Flags().GetString("doi")
			if doi == "" {
				output.Error("missing_param", "--doi is required")
				os.Exit(1)
			}

			result, err := search.EnrichByDOI(cmd.Context(), doi)
			if err != nil {
				output.Error("detail_error", err.Error())
				os.Exit(1)
			}
			output.Success(result)
		},
	}
	detailCmd.Flags().String("doi", "", "DOI to look up (e.g. 10.1145/3442188.3445922)")

	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export workspace papers to BibTeX format",
		Run: func(cmd *cobra.Command, args []string) {
			workspace, _ := cmd.Flags().GetString("workspace")
			outputFile, _ := cmd.Flags().GetString("output")

			s, err := store.Open(workspace)
			if err != nil {
				output.Error("store_error", err.Error())
				os.Exit(1)
			}

			if len(s.Papers) == 0 {
				output.Error("no_papers", "No papers in workspace. Run search first.")
				os.Exit(1)
			}

			bib := paper.ExportBibTeX(s.Papers)

			if outputFile != "" {
				if err := os.WriteFile(outputFile, []byte(bib), 0644); err != nil {
					output.Error("write_error", err.Error())
					os.Exit(1)
				}
				output.Success(map[string]any{
					"file":   outputFile,
					"count":  len(s.Papers),
					"format": "bibtex",
				})
			} else {
				fmt.Print(bib)
			}
		},
	}
	exportCmd.Flags().String("workspace", "", "Workspace directory (default: ~/.cache/scholar-cli/default)")
	exportCmd.Flags().String("output", "", "Output .bib file path (prints to stdout if not set)")

	searchWosCmd := &cobra.Command{
		Use:   "search-wos",
		Short: "Search Web of Science via WebBridge (requires institutional login)",
		Run: func(cmd *cobra.Command, args []string) {
			query, _ := cmd.Flags().GetString("query")
			limit, _ := cmd.Flags().GetInt("limit")
			workspace, _ := cmd.Flags().GetString("workspace")

			if query == "" {
				output.Error("missing_param", "--query is required")
				os.Exit(1)
			}

			client := withActivateMode(browser.NewClient("scholar-cli"))
			checkDaemon(client)
			src := search.NewWoS(client)
			papers, err := src.Search(cmd.Context(), query, limit)
			if err != nil {
				output.Error("search_error", err.Error())
				os.Exit(1)
			}
			emitSearch(papers, "wos", workspace)
		},
	}
	searchWosCmd.Flags().String("query", "", "Search query (required)")
	searchWosCmd.Flags().String("workspace", "", "Workspace directory to save results (auto-dedup)")
	searchWosCmd.Flags().Int("limit", 10, "Max results")

	loginStatusCmd := &cobra.Command{
		Use:   "login-status",
		Short: "Check login status for WebBridge-based sources (cnki, wos)",
		Run: func(cmd *cobra.Command, args []string) {
			platform, _ := cmd.Flags().GetString("platform")

			switch platform {
			case "wos":
				client := withActivateMode(browser.NewClient("scholar-cli"))
				checkDaemon(client)
				ok, err := search.CheckWoSLogin(client)
				if err != nil {
					output.Error("check_error", err.Error())
					os.Exit(1)
				}
				output.Success(map[string]any{
					"platform":  "wos",
					"logged_in": ok,
				})
			default:
				output.Error("invalid_platform", "Use --platform wos")
				os.Exit(1)
			}
		},
	}
	loginStatusCmd.Flags().String("platform", "", "Platform to check (wos)")

	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Download paper PDF (tries open access → Unpaywall → Sci-Hub)",
		Run: func(cmd *cobra.Command, args []string) {
			doi, _ := cmd.Flags().GetString("doi")
			pdfURL, _ := cmd.Flags().GetString("url")
			outputDir, _ := cmd.Flags().GetString("output-dir")
			scihub, _ := cmd.Flags().GetString("scihub")

			if doi == "" && pdfURL == "" {
				output.Error("missing_param", "--doi or --url is required")
				os.Exit(1)
			}

			if outputDir == "" {
				outputDir = "."
			}

			p := &paper.Paper{DOI: doi, PDFURL: pdfURL, Title: doi}

			// If we have a DOI, enrich first to get title and pdf_url
			if doi != "" && pdfURL == "" {
				enriched, err := search.EnrichByDOI(cmd.Context(), doi)
				if err == nil {
					p = enriched
				}
			}

			result, err := download.Download(cmd.Context(), p, outputDir, scihub)
			if err != nil {
				output.Error("download_error", err.Error())
				os.Exit(1)
			}
			output.Success(result)
		},
	}
	downloadCmd.Flags().String("doi", "", "DOI of paper to download")
	downloadCmd.Flags().String("url", "", "Direct PDF URL to download")
	downloadCmd.Flags().String("output-dir", ".", "Directory to save PDF")
	downloadCmd.Flags().String("scihub", "", "Sci-Hub domain to use as last-resort fallback (off by default; pass e.g. 'sci-hub.se' to enable, only where legal)")

	subcommands := []*cobra.Command{
		searchEnCmd, searchGoogleCmd, searchCnkiCmd, searchWosCmd,
		detailCmd, exportCmd, downloadCmd, loginStatusCmd,
	}
	for _, c := range subcommands {
		// Keep error output in the JSON contract on stdout; don't let cobra
		// dump usage text to stderr on flag-validation failures.
		c.SilenceUsage = true
		c.SilenceErrors = true
		rootCmd.AddCommand(c)
	}
}
