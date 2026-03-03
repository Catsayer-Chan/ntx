// Package cmd 提供 diag 命令实现
//
// 本文件实现智能网络诊断命令，支持:
// - 一键网络诊断
// - 多级别诊断（快速/标准/完整）
// - 问题分析和修复建议
// - 诊断报告生成
//
// 使用示例:
//
//	ntx diag
//	ntx diag --fast
//	ntx diag --target google.com
//
// 作者: Catsayer
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/catsayer/ntx/internal/app"
	"github.com/catsayer/ntx/internal/core/diag"
	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/internal/output/formatter"
	"github.com/catsayer/ntx/pkg/types"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	diagFast   bool
	diagFull   bool
	diagTarget string
	diagReport bool
)

var diagCmd = &cobra.Command{
	Use:   "diag",
	Short: "智能网络诊断",
	Long: `自动化网络问题诊断和分析。

诊断项目:
  • 网络接口配置检查
  • 本地连通性测试（网关）
  • 互联网连通性测试
  • DNS 解析测试
  • 目标主机可达性测试（可选）
  • 问题分析和修复建议

示例:
  ntx diag                          # 标准诊断
  ntx diag --fast                   # 快速诊断
  ntx diag --full                   # 完整诊断
  ntx diag --target google.com      # 包含目标主机测试
  ntx diag --report                 # 生成详细报告
  ntx diag -o json                  # JSON 输出`,
	RunE: runDiag,
}

func init() {
	rootCmd.AddCommand(diagCmd)

	diagCmd.Flags().BoolVar(&diagFast, "fast", false, "快速诊断模式")
	diagCmd.Flags().BoolVar(&diagFull, "full", false, "完整诊断模式")
	diagCmd.Flags().StringVar(&diagTarget, "target", "", "指定目标主机进行额外测试")
	diagCmd.Flags().BoolVar(&diagReport, "report", false, "生成详细报告")
}

func runDiag(cmd *cobra.Command, args []string) error {
	flags := mustAppContext(cmd).Flags
	outputFormat := types.OutputFormat(flags.Output)
	logger.Info("开始网络诊断")

	// 构建诊断选项
	opts := diag.DiagnosticOptions{
		Level:  diag.DiagLevelNormal,
		Target: diagTarget,
	}

	if diagFast {
		opts.Level = diag.DiagLevelFast
	} else if diagFull {
		opts.Level = diag.DiagLevelFull
	}

	// 创建诊断服务
	diagService := diag.NewService()

	// 仅文本模式显示 banner，避免污染结构化输出
	if outputFormat == types.OutputText || outputFormat == "" {
		fmt.Println(color.CyanString("🔍 NTX 网络诊断工具"))
		fmt.Println(strings.Repeat("=", 70))
		fmt.Println()
	}

	// 执行诊断
	ctx := context.Background()
	result, err := diagService.Diagnose(ctx, opts)
	if err != nil {
		return fmt.Errorf("诊断失败: %w", err)
	}

	// 输出结果
	return outputDiagResult(result, flags)
}

// outputDiagResult 输出诊断结果
func outputDiagResult(result *diag.DiagnosticResult, flags app.GlobalFlags) error {
	outputFormat := types.OutputFormat(flags.Output)
	if outputFormat == types.OutputText || outputFormat == "" {
		return outputDiagText(result, flags)
	}

	f := formatter.NewFormatter(outputFormat, flags.NoColor)
	return f.FormatTo(os.Stdout, result)
}

// outputDiagText 文本格式输出
func outputDiagText(result *diag.DiagnosticResult, flags app.GlobalFlags) error {
	color.NoColor = flags.NoColor
	f := formatter.NewTextFormatter(!flags.NoColor)

	// 显示检查结果
	f.PrintSubHeader("检查结果")
	fmt.Println()

	for _, check := range result.Checks {
		statusSymbol := getStatusSymbol(check.Status, flags.NoColor)
		statusColor := getStatusColor(check.Status, flags.NoColor)

		fmt.Printf("%s %-30s %s\n",
			statusSymbol,
			check.Name,
			statusColor(check.Message),
		)

		if flags.Verbose && check.Details != nil && len(check.Details) > 0 {
			for key, value := range check.Details {
				fmt.Printf("    %s: %v\n", key, value)
			}
		}
	}

	fmt.Println()

	// 显示发现的问题
	if len(result.Issues) > 0 {
		f.PrintSubHeader("发现的问题")
		fmt.Println()

		for i, issue := range result.Issues {
			fmt.Printf("%d. [%s] %s\n",
				i+1,
				getStatusColor(issue.Severity, flags.NoColor)(issue.Severity.String()),
				issue.Description,
			)
			if issue.Suggestion != "" {
				fmt.Printf("   建议: %s\n", color.YellowString(issue.Suggestion))
			}
			fmt.Println()
		}
	}

	// 显示修复建议
	if len(result.Suggestions) > 0 {
		f.PrintSubHeader("修复建议")
		fmt.Println()

		for i, suggestion := range result.Suggestions {
			fmt.Printf("%d. %s\n", i+1, suggestion)
		}
		fmt.Println()
	}

	// 显示总结
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("整体状态: %s\n", getStatusColorBold(result.Status, flags.NoColor)(result.Status.String()))
	fmt.Printf("诊断耗时: %s\n", result.Duration.Round(100))
	fmt.Printf("检查项目: %d 项\n", len(result.Checks))
	fmt.Printf("发现问题: %d 个\n", len(result.Issues))
	fmt.Println(strings.Repeat("=", 70))

	return nil
}

// getStatusSymbol 获取状态符号
func getStatusSymbol(status diag.DiagnosticStatus, noColor bool) string {
	color.NoColor = noColor
	switch status {
	case diag.StatusHealthy:
		return color.GreenString("✓")
	case diag.StatusWarning:
		return color.YellowString("⚠")
	case diag.StatusCritical:
		return color.RedString("✗")
	default:
		return "?"
	}
}

// getStatusColor 获取状态颜色函数
func getStatusColor(status diag.DiagnosticStatus, noColor bool) func(a ...interface{}) string {
	if noColor {
		return fmt.Sprint
	}

	switch status {
	case diag.StatusHealthy:
		return color.New(color.FgGreen).SprintFunc()
	case diag.StatusWarning:
		return color.New(color.FgYellow).SprintFunc()
	case diag.StatusCritical:
		return color.New(color.FgRed).SprintFunc()
	default:
		return color.New(color.FgWhite).SprintFunc()
	}
}

// getStatusColorBold 获取加粗的状态颜色函数
func getStatusColorBold(status diag.DiagnosticStatus, noColor bool) func(a ...interface{}) string {
	if noColor {
		return fmt.Sprint
	}

	switch status {
	case diag.StatusHealthy:
		return color.New(color.FgGreen, color.Bold).SprintFunc()
	case diag.StatusWarning:
		return color.New(color.FgYellow, color.Bold).SprintFunc()
	case diag.StatusCritical:
		return color.New(color.FgRed, color.Bold).SprintFunc()
	default:
		return color.New(color.FgWhite, color.Bold).SprintFunc()
	}
}
