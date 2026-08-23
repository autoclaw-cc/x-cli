package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"apartments-cli/browser"
	"apartments-cli/output"
	"apartments-cli/property"
)

var rootCmd = &cobra.Command{
	Use:   "apartments-cli",
	Short: "Apartments.com rental search CLI via kimi-webbridge",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search rental listings on Apartments.com",
		Run: func(cmd *cobra.Command, args []string) {
			location, _ := cmd.Flags().GetString("location")
			minBeds, _ := cmd.Flags().GetInt("min-beds")
			maxBeds, _ := cmd.Flags().GetInt("max-beds")
			minPrice, _ := cmd.Flags().GetInt("min-price")
			maxPrice, _ := cmd.Flags().GetInt("max-price")
			limit, _ := cmd.Flags().GetInt("limit")
			page, _ := cmd.Flags().GetInt("page")

			if location == "" {
				output.Error("missing_param", "--location is required (e.g. new-york-ny, san-francisco-ca, chicago-il)")
				os.Exit(1)
			}

			client := withActivateMode(browser.NewClient("apartments"))
			result, err := property.Search(client, property.SearchParams{
				Location: location,
				MinBeds:  minBeds,
				MaxBeds:  maxBeds,
				MinPrice: minPrice,
				MaxPrice: maxPrice,
				Limit:    limit,
				Page:     page,
			})
			if err != nil {
				output.Error("search_error", err.Error())
				os.Exit(1)
			}
			output.Success(result)
		},
	}
	searchCmd.Flags().String("location", "", "Location slug (e.g. new-york-ny, los-angeles-ca)")
	searchCmd.Flags().Int("min-beds", 0, "Minimum bedrooms")
	searchCmd.Flags().Int("max-beds", 0, "Maximum bedrooms")
	searchCmd.Flags().Int("min-price", 0, "Minimum monthly rent ($)")
	searchCmd.Flags().Int("max-price", 0, "Maximum monthly rent ($)")
	searchCmd.Flags().Int("limit", 20, "Max results to return")
	searchCmd.Flags().Int("page", 1, "Page number")

	detailCmd := &cobra.Command{
		Use:   "detail",
		Short: "Get property details from Apartments.com",
		Run: func(cmd *cobra.Command, args []string) {
			urlFlag, _ := cmd.Flags().GetString("url")
			if urlFlag == "" {
				output.Error("missing_param", "--url is required")
				os.Exit(1)
			}

			client := withActivateMode(browser.NewClient("apartments"))
			result, err := property.GetDetail(client, urlFlag)
			if err != nil {
				output.Error("detail_error", err.Error())
				os.Exit(1)
			}
			output.Success(result)
		},
	}
	detailCmd.Flags().String("url", "", "Property URL on Apartments.com")

	rootCmd.AddCommand(searchCmd, detailCmd)
}
