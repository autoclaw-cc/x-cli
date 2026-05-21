package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"rightmove-cli/browser"
	"rightmove-cli/output"
	"rightmove-cli/property"
)

var rootCmd = &cobra.Command{
	Use:   "rightmove-cli",
	Short: "Rightmove UK property rental CLI via kimi-webbridge",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search Rightmove rental listings",
		Run: func(cmd *cobra.Command, args []string) {
			location, _ := cmd.Flags().GetString("location")
			minPrice, _ := cmd.Flags().GetInt("min-price")
			maxPrice, _ := cmd.Flags().GetInt("max-price")
			minBeds, _ := cmd.Flags().GetInt("min-beds")
			maxBeds, _ := cmd.Flags().GetInt("max-beds")
			radius, _ := cmd.Flags().GetFloat64("radius")
			limit, _ := cmd.Flags().GetInt("limit")
			page, _ := cmd.Flags().GetInt("page")
			propType, _ := cmd.Flags().GetString("type")

			if location == "" {
				output.Error("missing_param", "--location is required (e.g. \"London\", \"Manchester\")")
				os.Exit(1)
			}

			client := browser.NewClient("rightmove")
			result, err := property.Search(client, property.SearchParams{
				Location: location,
				MinPrice: minPrice,
				MaxPrice: maxPrice,
				MinBeds:  minBeds,
				MaxBeds:  maxBeds,
				Radius:   radius,
				Limit:    limit,
				Page:     page,
				Type:     propType,
			})
			if err != nil {
				output.Error("search_error", err.Error())
				os.Exit(1)
			}
			output.Success(result)
		},
	}
	searchCmd.Flags().String("location", "", "Location to search (e.g. \"London\", \"Manchester\")")
	searchCmd.Flags().Int("min-price", 0, "Minimum price per calendar month (pcm)")
	searchCmd.Flags().Int("max-price", 0, "Maximum price per calendar month (pcm)")
	searchCmd.Flags().Int("min-beds", 0, "Minimum number of bedrooms")
	searchCmd.Flags().Int("max-beds", 0, "Maximum number of bedrooms")
	searchCmd.Flags().Float64("radius", 0, "Search radius in miles")
	searchCmd.Flags().Int("limit", 20, "Maximum number of results to return")
	searchCmd.Flags().Int("page", 1, "Page number")
	searchCmd.Flags().String("type", "rent", "Listing type: rent or flatshare")

	detailCmd := &cobra.Command{
		Use:   "detail",
		Short: "Get details for a specific Rightmove property listing",
		Run: func(cmd *cobra.Command, args []string) {
			urlFlag, _ := cmd.Flags().GetString("url")

			if urlFlag == "" {
				output.Error("missing_param", "--url is required (e.g. \"https://www.rightmove.co.uk/properties/12345\")")
				os.Exit(1)
			}

			client := browser.NewClient("rightmove")
			result, err := property.GetDetail(client, urlFlag)
			if err != nil {
				output.Error("detail_error", err.Error())
				os.Exit(1)
			}
			output.Success(result)
		},
	}
	detailCmd.Flags().String("url", "", "Full Rightmove property URL")

	rootCmd.AddCommand(searchCmd, detailCmd)
}
