package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"gaokao-cli/gaokao"
	"gaokao-cli/output"
)

var rootCmd = &cobra.Command{
	Use:   "gaokao-cli",
	Short: "高考信息查询 CLI — 省控线、一分一段表、院校查询",
	Long: `gaokao-cli 基于掌上高考 (gaokao.cn) 的公开数据，提供高考相关信息的
结构化查询，包括各省批次线、一分一段表、院校信息等。

数据来源：中国教育在线 (eol.cn) 旗下掌上高考平台。`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// provinces command
	provincesCmd := &cobra.Command{
		Use:   "provinces",
		Short: "列出所有省份及其 ID",
		Run: func(cmd *cobra.Command, args []string) {
			output.Success(gaokao.ListProvinces())
		},
	}
	rootCmd.AddCommand(provincesCmd)

	// score-line command
	scoreLineCmd := &cobra.Command{
		Use:   "score-line",
		Short: "查询省控线/批次线",
		Long: `查询各省高考录取控制分数线（批次线）。

Examples:
  gaokao-cli score-line --province 北京
  gaokao-cli score-line --province 北京 --year 2025
  gaokao-cli score-line --province 河南 --year 2025 --type 物理类`,
		Run: func(cmd *cobra.Command, args []string) {
			province, _ := cmd.Flags().GetString("province")
			year, _ := cmd.Flags().GetString("year")
			typeName, _ := cmd.Flags().GetString("type")

			if province == "" {
				output.Error("missing_args", "--province is required")
				os.Exit(1)
			}

			id, name, err := gaokao.ResolveProvince(province)
			if err != nil {
				output.Error("invalid_province", err.Error())
				os.Exit(1)
			}

			result, err := gaokao.FetchScoreLine(id, year, typeName)
			if err != nil {
				output.Error("fetch_failed", err.Error())
				os.Exit(1)
			}

			output.Success(map[string]any{
				"province":    name,
				"province_id": id,
				"year":        result.Year,
				"lines":       result.Lines,
			})
		},
	}
	scoreLineCmd.Flags().StringP("province", "p", "", "省份名称或 ID (required)")
	scoreLineCmd.Flags().StringP("year", "y", "", "年份 (default: latest)")
	scoreLineCmd.Flags().StringP("type", "t", "", "科类: 综合/理科/文科/物理类/历史类")
	rootCmd.AddCommand(scoreLineCmd)

	// score-section command
	sectionCmd := &cobra.Command{
		Use:   "score-section",
		Short: "查询一分一段表",
		Long: `查询各省高考一分一段表（分数段统计），支持按分数精确查询排名。

Examples:
  gaokao-cli score-section --province 北京 --year 2025
  gaokao-cli score-section --province 天津 --year 2025 --score 650
  gaokao-cli score-section --province 河南 --year 2024 --type 理科 --score 600`,
		Run: func(cmd *cobra.Command, args []string) {
			province, _ := cmd.Flags().GetString("province")
			year, _ := cmd.Flags().GetString("year")
			typeName, _ := cmd.Flags().GetString("type")
			level, _ := cmd.Flags().GetString("level")
			score, _ := cmd.Flags().GetString("score")

			if province == "" {
				output.Error("missing_args", "--province is required")
				os.Exit(1)
			}

			id, name, err := gaokao.ResolveProvince(province)
			if err != nil {
				output.Error("invalid_province", err.Error())
				os.Exit(1)
			}

			result, err := gaokao.FetchScoreSection(id, year, typeName, level, score)
			if err != nil {
				output.Error("fetch_failed", err.Error())
				os.Exit(1)
			}

			output.Success(map[string]any{
				"province":    name,
				"province_id": id,
				"year":        result.Year,
				"type":        result.TypeName,
				"level":       result.Level,
				"count":       len(result.Entries),
				"entries":     result.Entries,
			})
		},
	}
	sectionCmd.Flags().StringP("province", "p", "", "省份名称或 ID (required)")
	sectionCmd.Flags().StringP("year", "y", "", "年份 (default: latest)")
	sectionCmd.Flags().StringP("type", "t", "", "科类: 综合/理科/文科/物理类/历史类")
	sectionCmd.Flags().StringP("level", "l", "", "层次: 本科/专科")
	sectionCmd.Flags().StringP("score", "s", "", "指定分数查排名")
	rootCmd.AddCommand(sectionCmd)

	// school command
	schoolCmd := &cobra.Command{
		Use:   "school",
		Short: "查询院校信息",
		Long: `查询全国高校信息，支持按名称、省份、985/211/双一流等筛选。

Examples:
  gaokao-cli school --name 清华
  gaokao-cli school --985
  gaokao-cli school --211 --province 北京
  gaokao-cli school --dual-class --province 上海`,
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			province, _ := cmd.Flags().GetString("province")
			is985, _ := cmd.Flags().GetBool("985")
			is211, _ := cmd.Flags().GetBool("211")
			dualClass, _ := cmd.Flags().GetBool("dual-class")
			level, _ := cmd.Flags().GetString("level")

			filter := gaokao.SchoolFilter{
				Name: name, Province: province,
				Is985: is985, Is211: is211, DualClass: dualClass,
				Level: level,
			}

			schools, err := gaokao.FetchSchools(filter)
			if err != nil {
				output.Error("fetch_failed", err.Error())
				os.Exit(1)
			}
			output.Success(map[string]any{
				"count":   len(schools),
				"schools": schools,
			})
		},
	}
	schoolCmd.Flags().StringP("name", "n", "", "模糊搜索校名")
	schoolCmd.Flags().StringP("province", "p", "", "按省份筛选")
	schoolCmd.Flags().Bool("985", false, "只看 985 院校")
	schoolCmd.Flags().Bool("211", false, "只看 211 院校")
	schoolCmd.Flags().Bool("dual-class", false, "只看双一流院校")
	schoolCmd.Flags().StringP("level", "l", "", "本科 or 专科")
	rootCmd.AddCommand(schoolCmd)

	// batch-history command
	batchHistoryCmd := &cobra.Command{
		Use:   "batch-history",
		Short: "查询批次改革历史",
		Long: `查询各省高考批次合并/改革历史，帮助理解历史分数线变化的背景。

Examples:
  gaokao-cli batch-history --province 北京
  gaokao-cli batch-history`,
		Run: func(cmd *cobra.Command, args []string) {
			province, _ := cmd.Flags().GetString("province")

			var provinceID string
			if province != "" {
				id, _, err := gaokao.ResolveProvince(province)
				if err != nil {
					output.Error("invalid_province", err.Error())
					os.Exit(1)
				}
				provinceID = id
			}

			results, err := gaokao.FetchBatchHistory(provinceID)
			if err != nil {
				output.Error("fetch_failed", err.Error())
				os.Exit(1)
			}
			output.Success(results)
		},
	}
	batchHistoryCmd.Flags().StringP("province", "p", "", "省份名称或 ID (optional)")
	rootCmd.AddCommand(batchHistoryCmd)
}
