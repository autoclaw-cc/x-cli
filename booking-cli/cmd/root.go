package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"booking-cli/booking"
	"booking-cli/browser"
	"booking-cli/output"
)

var rootCmd = &cobra.Command{
	Use:   "booking-cli",
	Short: "Booking.com hotel search CLI via kimi-webbridge",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	searchCmd := &cobra.Command{
		Use:   "search-hotels",
		Short: "Search hotels on Booking.com",
		Run: func(cmd *cobra.Command, args []string) {
			dest, _ := cmd.Flags().GetString("destination")
			checkin, _ := cmd.Flags().GetString("checkin")
			checkout, _ := cmd.Flags().GetString("checkout")
			adults, _ := cmd.Flags().GetInt("adults")
			rooms, _ := cmd.Flags().GetInt("rooms")
			limit, _ := cmd.Flags().GetInt("limit")

			if dest == "" {
				output.Error("missing_param", "--destination is required (e.g. Kyoto, Paris, Bangkok)")
				os.Exit(1)
			}

			client := browser.NewClient("booking")
			result, err := booking.SearchHotels(client, dest, checkin, checkout, adults, rooms, limit)
			if err != nil {
				output.Error("search_error", err.Error())
				os.Exit(1)
			}
			output.Success(result)
		},
	}
	searchCmd.Flags().String("destination", "", "Destination name (e.g. Kyoto, Paris, Bangkok, 京都)")
	searchCmd.Flags().String("checkin", "", "Check-in date (YYYY-MM-DD), defaults to 7 days from now")
	searchCmd.Flags().String("checkout", "", "Check-out date (YYYY-MM-DD), defaults to 2 days after checkin")
	searchCmd.Flags().Int("adults", 2, "Number of adults")
	searchCmd.Flags().Int("rooms", 1, "Number of rooms")
	searchCmd.Flags().Int("limit", 10, "Max hotels to return")

	rootCmd.AddCommand(searchCmd)
}
