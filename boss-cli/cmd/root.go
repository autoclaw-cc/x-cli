package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"boss-cli/boss"
	"boss-cli/browser"
	"boss-cli/output"
)

var rootCmd = &cobra.Command{
	Use:   "boss-cli",
	Short: "Boss直聘 (BOSS Zhipin) job search CLI via kimi-webbridge",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	loginStatusCmd := &cobra.Command{
		Use:   "login-status",
		Short: "Check if logged in to Boss直聘",
		Run: func(cmd *cobra.Command, args []string) {
			client := browser.NewClient("boss")
			status, err := boss.CheckLogin(client)
			if err != nil {
				output.Error("login_check_error", err.Error())
				os.Exit(1)
			}
			if !status.LoggedIn {
				output.Error("not_logged_in", "Not logged in to Boss直聘. Please open Chrome, navigate to https://www.zhipin.com, and log in manually.")
				os.Exit(1)
			}
			output.Success(status)
		},
	}

	searchJobsCmd := &cobra.Command{
		Use:   "search-jobs",
		Short: "Search jobs on Boss直聘",
		Run: func(cmd *cobra.Command, args []string) {
			query, _ := cmd.Flags().GetString("query")
			city, _ := cmd.Flags().GetString("city")
			salary, _ := cmd.Flags().GetString("salary")
			experience, _ := cmd.Flags().GetString("experience")
			degree, _ := cmd.Flags().GetString("degree")
			scale, _ := cmd.Flags().GetString("scale")
			stage, _ := cmd.Flags().GetString("stage")
			jobType, _ := cmd.Flags().GetString("job-type")
			page, _ := cmd.Flags().GetInt("page")
			limit, _ := cmd.Flags().GetInt("limit")

			if query == "" {
				output.Error("missing_param", "--query is required (e.g. 前端开发, Java, 产品经理)")
				os.Exit(1)
			}

			client := browser.NewClient("boss")
			result, err := boss.SearchJobs(client, query, city, salary, experience, degree, scale, stage, jobType, page, limit)
			if err != nil {
				output.Error("search_error", err.Error())
				os.Exit(1)
			}
			output.Success(result)
		},
	}
	searchJobsCmd.Flags().String("query", "", "Job search keyword (e.g. 前端开发, Java, 产品经理)")
	searchJobsCmd.Flags().String("city", "101010100", "City code (101010100=北京, 101020100=上海, 101280100=广州, 101280600=深圳, 101210100=杭州, 101030100=天津, 101190400=苏州, 101110100=西安)")
	searchJobsCmd.Flags().String("salary", "", "Salary filter (401=3K以下, 402=3-5K, 403=5-10K, 404=10-20K, 405=20-50K, 406=50K以上)")
	searchJobsCmd.Flags().String("experience", "", "Experience filter (101=在校生, 102=应届生, 103=经验不限, 104=1年以内, 105=1-3年, 106=3-5年, 107=5-10年, 108=10年以上)")
	searchJobsCmd.Flags().String("degree", "", "Degree filter (209=初中及以下, 208=中专/中技, 206=高中, 202=大专, 203=本科, 204=硕士, 205=博士)")
	searchJobsCmd.Flags().String("scale", "", "Company scale (301=0-20人, 302=20-99人, 303=100-499人, 304=500-999人, 305=1000-9999人, 306=10000人以上)")
	searchJobsCmd.Flags().String("stage", "", "Funding stage (801=未融资, 802=天使轮, 803=A轮, 804=B轮, 805=C轮, 806=D轮及以上, 807=已上市, 808=不需要融资)")
	searchJobsCmd.Flags().String("job-type", "", "Job type (1901=全职, 1903=兼职)")
	searchJobsCmd.Flags().Int("page", 1, "Page number (each page ~15 results)")
	searchJobsCmd.Flags().Int("limit", 0, "Max jobs to return (0 = all)")

	jobDetailCmd := &cobra.Command{
		Use:   "job-detail",
		Short: "Get detailed info for a specific job",
		Run: func(cmd *cobra.Command, args []string) {
			jobID, _ := cmd.Flags().GetString("id")
			if jobID == "" {
				output.Error("missing_param", "--id is required (encrypted job ID from search results)")
				os.Exit(1)
			}

			client := browser.NewClient("boss-detail")
			detail, err := boss.GetJobDetail(client, jobID)
			if err != nil {
				output.Error("detail_error", err.Error())
				os.Exit(1)
			}
			output.Success(detail)
		},
	}
	jobDetailCmd.Flags().String("id", "", "Encrypted job ID (from search-jobs results)")

	rootCmd.AddCommand(loginStatusCmd)
	rootCmd.AddCommand(searchJobsCmd)
	rootCmd.AddCommand(jobDetailCmd)
}
