package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"ctrip-cli/browser"
	"ctrip-cli/ctrip"
	"ctrip-cli/output"
)

var rootCmd = &cobra.Command{
	Use:   "ctrip-cli",
	Short: "Ctrip (携程) automation CLI via kimi-webbridge",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	searchHotelsCmd := &cobra.Command{
		Use:   "search-hotels",
		Short: "Search hotels by destination keyword",
		Run: func(cmd *cobra.Command, args []string) {
			keyword, _ := cmd.Flags().GetString("keyword")
			checkin, _ := cmd.Flags().GetString("checkin")
			checkout, _ := cmd.Flags().GetString("checkout")
			cityID, _ := cmd.Flags().GetInt("city-id")
			countryID, _ := cmd.Flags().GetInt("country-id")
			limit, _ := cmd.Flags().GetInt("limit")

			if keyword == "" {
				output.Error("missing_param", "--keyword is required")
				os.Exit(1)
			}

			client := withActivateMode(browser.NewClient("ctrip"))
			result, err := ctrip.SearchHotels(client, keyword, checkin, checkout, cityID, countryID, limit)
			if err != nil {
				output.Error("search_hotels_error", err.Error())
				os.Exit(1)
			}
			output.Success(result)
		},
	}
	searchHotelsCmd.Flags().String("keyword", "", "Destination keyword (e.g. 上海, 京都)")
	searchHotelsCmd.Flags().String("checkin", "", "Check-in date (YYYY/MM/DD), defaults to 7 days from now")
	searchHotelsCmd.Flags().String("checkout", "", "Check-out date (YYYY/MM/DD), defaults to day after checkin")
	searchHotelsCmd.Flags().Int("city-id", 0, "Ctrip city ID (e.g. 734 for Kyoto). Get from 'destination' command's hotel_url")
	searchHotelsCmd.Flags().Int("country-id", 0, "Ctrip country ID (e.g. 78 for Japan, 1 for China)")
	searchHotelsCmd.Flags().Int("limit", 0, "Max hotels to return (0 = all)")

	searchFlightsCmd := &cobra.Command{
		Use:   "search-flights",
		Short: "Search flights between cities",
		Run: func(cmd *cobra.Command, args []string) {
			from, _ := cmd.Flags().GetString("from")
			to, _ := cmd.Flags().GetString("to")
			date, _ := cmd.Flags().GetString("date")
			limit, _ := cmd.Flags().GetInt("limit")

			if from == "" || to == "" {
				output.Error("missing_param", "--from and --to are required (use IATA city codes, e.g. SHA, BJS, TYO)")
				os.Exit(1)
			}

			client := withActivateMode(browser.NewClient("ctrip-flight"))
			result, err := ctrip.SearchFlights(client, from, to, date, limit)
			if err != nil {
				output.Error("search_flights_error", err.Error())
				os.Exit(1)
			}
			output.Success(result)
		},
	}
	searchFlightsCmd.Flags().String("from", "", "Departure city IATA code (e.g. SHA, PEK, CAN)")
	searchFlightsCmd.Flags().String("to", "", "Arrival city IATA code (e.g. BJS, NRT, ICN)")
	searchFlightsCmd.Flags().String("date", "", "Departure date (YYYY-MM-DD), defaults to 14 days from now")
	searchFlightsCmd.Flags().Int("limit", 0, "Max flights to return (0 = all)")

	searchAttractionsCmd := &cobra.Command{
		Use:   "search-attractions",
		Short: "Search attractions and tickets for a destination",
		Run: func(cmd *cobra.Command, args []string) {
			dest, _ := cmd.Flags().GetString("destination")
			limit, _ := cmd.Flags().GetInt("limit")

			if dest == "" {
				output.Error("missing_param", "--destination is required (e.g. shanghai2, kyoto430, tokyo294)")
				os.Exit(1)
			}

			client := withActivateMode(browser.NewClient("ctrip-sight"))
			result, err := ctrip.SearchAttractions(client, dest, limit)
			if err != nil {
				output.Error("search_attractions_error", err.Error())
				os.Exit(1)
			}
			output.Success(result)
		},
	}
	searchAttractionsCmd.Flags().String("destination", "", "Destination slug (e.g. shanghai2, kyoto430)")
	searchAttractionsCmd.Flags().Int("limit", 0, "Max attractions to return (0 = all)")

	destinationCmd := &cobra.Command{
		Use:   "destination",
		Short: "Get destination overview (must-do attractions, travel notes)",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")

			if name == "" {
				output.Error("missing_param", "--name is required (e.g. kyoto430, shanghai2, tokyo294)")
				os.Exit(1)
			}

			client := withActivateMode(browser.NewClient("ctrip-guide"))
			result, err := ctrip.GetDestination(client, name)
			if err != nil {
				output.Error("destination_error", err.Error())
				os.Exit(1)
			}
			output.Success(result)
		},
	}
	destinationCmd.Flags().String("name", "", "Destination slug (e.g. kyoto430, shanghai2)")

	rootCmd.AddCommand(searchHotelsCmd)
	rootCmd.AddCommand(searchFlightsCmd)
	rootCmd.AddCommand(searchAttractionsCmd)
	rootCmd.AddCommand(destinationCmd)
}
