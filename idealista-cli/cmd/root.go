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
var searchMinPrice int
var searchMaxPrice int
var searchMinRooms int
var searchMaxRooms int
var searchLimit int
var searchPage int

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search rental properties on Idealista",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !isValidCountry(searchCountry) {
			output.Error("INVALID_COUNTRY", fmt.Sprintf("country must be one of: %v", validCountries))
			return fmt.Errorf("invalid country")
		}

		client := browser.NewClient(sessionName)
		params := property.SearchParams{
			Country:  searchCountry,
			City:     searchCity,
			MinPrice: searchMinPrice,
			MaxPrice: searchMaxPrice,
			MinRooms: searchMinRooms,
			MaxRooms: searchMaxRooms,
			Limit:    searchLimit,
			Page:     searchPage,
		}
		result, err := property.Search(client, params)
		if err != nil {
			output.Error("SEARCH_ERROR", err.Error())
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
			output.Error("MISSING_URL", "--url is required")
			return fmt.Errorf("missing url")
		}

		client := browser.NewClient(sessionName)
		detail, err := property.Detail(client, detailURL)
		if err != nil {
			output.Error("DETAIL_ERROR", err.Error())
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
	searchCmd.Flags().StringVar(&searchCity, "city", "", "City name, e.g. madrid, barcelona, rome, lisbon (required)")
	searchCmd.Flags().IntVar(&searchMinPrice, "min-price", 0, "Minimum monthly rent in EUR")
	searchCmd.Flags().IntVar(&searchMaxPrice, "max-price", 0, "Maximum monthly rent in EUR")
	searchCmd.Flags().IntVar(&searchMinRooms, "min-rooms", 0, "Minimum number of rooms")
	searchCmd.Flags().IntVar(&searchMaxRooms, "max-rooms", 0, "Maximum number of rooms")
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
