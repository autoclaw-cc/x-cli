package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"idealista-cli/browser"
	"idealista-cli/output"
	"idealista-cli/property"
)

const sessionName = "idealista"

var validCountries = []string{"spain", "italy", "portugal"}

var rootCmd = &cobra.Command{
	Use:   "idealista-cli",
	Short: "Idealista European property rental CLI via kimi-webbridge",
}

// --- search command ---

var searchCountry string
var searchCity string
var searchLimit int
var searchPage int

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search rental properties on Idealista",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !isValidCountry(searchCountry) {
			output.Error("invalid_country", fmt.Sprintf("country must be one of: %v", validCountries))
			return fmt.Errorf("invalid country")
		}

		client := browser.NewClient(sessionName)
		params := property.SearchParams{
			Country: searchCountry,
			City:    searchCity,
			Limit:   searchLimit,
			Page:    searchPage,
		}
		result, err := property.Search(client, params)
		if err != nil {
			output.Error("search_error", err.Error())
			return err
		}
		output.Success(result)
		return nil
	},
}

// --- detail command ---

var detailURL string

var detailCmd = &cobra.Command{
	Use:   "detail",
	Short: "Get detailed info for a single property listing",
	RunE: func(cmd *cobra.Command, args []string) error {
		if detailURL == "" {
			output.Error("missing_param", "--url is required")
			return fmt.Errorf("missing url")
		}

		client := browser.NewClient(sessionName)
		detail, err := property.GetDetail(client, detailURL)
		if err != nil {
			output.Error("detail_error", err.Error())
			return err
		}
		output.Success(detail)
		return nil
	},
}

func isValidCountry(c string) bool {
	for _, v := range validCountries {
		if v == c {
			return true
		}
	}
	return false
}

func init() {
	// search flags
	searchCmd.Flags().StringVar(&searchCountry, "country", "", "Country: spain, italy, portugal (required)")
	searchCmd.Flags().StringVar(&searchCity, "city", "", "City slug, e.g. madrid-madrid, barcelona-barcelona, roma, lisboa (required)")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 20, "Max results to return")
	searchCmd.Flags().IntVar(&searchPage, "page", 1, "Page number")
	_ = searchCmd.MarkFlagRequired("country")
	_ = searchCmd.MarkFlagRequired("city")

	// detail flags
	detailCmd.Flags().StringVar(&detailURL, "url", "", "Full URL of the property listing (required)")
	_ = detailCmd.MarkFlagRequired("url")

	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(detailCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
