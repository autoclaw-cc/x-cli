package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"58-cli/browser"
	"58-cli/house"
	"58-cli/output"
)

var rootCmd = &cobra.Command{
	Use:   "58-cli",
	Short: "58同城租房信息 CLI via kimi-webbridge",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "搜索 58同城 租房列表",
		Run: func(cmd *cobra.Command, args []string) {
			city, _ := cmd.Flags().GetString("city")
			keyword, _ := cmd.Flags().GetString("keyword")
			minPrice, _ := cmd.Flags().GetInt("min-price")
			maxPrice, _ := cmd.Flags().GetInt("max-price")
			rooms, _ := cmd.Flags().GetInt("rooms")
			limit, _ := cmd.Flags().GetInt("limit")
			page, _ := cmd.Flags().GetInt("page")

			if city == "" {
				output.Error("missing_param", "--city is required (e.g. sz, bj, sh, gz, hz, cd)")
				os.Exit(1)
			}
			if _, ok := house.CityNames[city]; !ok {
				valid := ""
				for k, v := range house.CityNames {
					valid += fmt.Sprintf("  %s = %s\n", k, v)
				}
				output.Error("invalid_city", fmt.Sprintf("unknown city code %q. Valid codes:\n%s", city, valid))
				os.Exit(1)
			}

			client := browser.NewClient("58")
			result, err := house.Search(client, house.SearchParams{
				City:     city,
				Keyword:  keyword,
				MinPrice: minPrice,
				MaxPrice: maxPrice,
				Rooms:    rooms,
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
	searchCmd.Flags().String("city", "", "城市代码 (sz/bj/sh/gz/hz/cd)")
	searchCmd.Flags().String("keyword", "", "搜索关键词（区域/小区名）")
	searchCmd.Flags().Int("min-price", 0, "最低月租")
	searchCmd.Flags().Int("max-price", 0, "最高月租")
	searchCmd.Flags().Int("rooms", 0, "几室 (1/2/3)")
	searchCmd.Flags().Int("limit", 20, "最大返回数量")
	searchCmd.Flags().Int("page", 1, "页码")

	detailCmd := &cobra.Command{
		Use:   "detail",
		Short: "获取单条 58同城 房源详情",
		Run: func(cmd *cobra.Command, args []string) {
			urlFlag, _ := cmd.Flags().GetString("url")
			city, _ := cmd.Flags().GetString("city")
			id, _ := cmd.Flags().GetString("id")

			detailURL := urlFlag
			if detailURL == "" && city != "" && id != "" {
				detailURL = fmt.Sprintf("https://%s.58.com/zufang/%s.shtml", city, id)
			}
			if detailURL == "" {
				output.Error("missing_param", "需要 --url 或 --city + --id")
				os.Exit(1)
			}

			client := browser.NewClient("58")
			result, err := house.GetDetail(client, detailURL)
			if err != nil {
				output.Error("detail_error", err.Error())
				os.Exit(1)
			}
			output.Success(result)
		},
	}
	detailCmd.Flags().String("url", "", "房源完整 URL")
	detailCmd.Flags().String("city", "", "城市代码")
	detailCmd.Flags().String("id", "", "房源 ID")

	rootCmd.AddCommand(searchCmd, detailCmd)
}
