// Package cmd 提供 batch 命令实现
//
// 本文件实现批量任务命令，支持:
// - YAML 配置文件驱动
// - 多种任务类型（ping/dns/scan）
// - 并发执行
// - 结果汇总
//
// 使用示例:
//
//	ntx batch -f tasks.yaml
//	ntx batch --sample > tasks.yaml
//
// 作者: Catsayer
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/catsayer/ntx/internal/app"
	"github.com/catsayer/ntx/internal/core/batch"
	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/internal/output/formatter"
	"github.com/catsayer/ntx/pkg/types"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	batchFile   string
	batchSample bool
)

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "批量任务执行",
	Long: `通过 YAML 配置文件批量执行网络任务。

支持功能:
  • Ping 监控任务
  • DNS 查询任务
  • 端口扫描任务
  • 并发执行控制
  • 结果汇总

示例:
  ntx batch -f tasks.yaml           # 执行配置文件中的任务
  ntx batch --sample > tasks.yaml   # 生成示例配置
  ntx batch -f tasks.yaml -o json   # JSON 输出`,
	RunE: runBatch,
}

func init() {
	rootCmd.AddCommand(batchCmd)

	batchCmd.Flags().StringVarP(&batchFile, "file", "f", "", "任务配置文件路径（YAML 格式）")
	batchCmd.Flags().BoolVar(&batchSample, "sample", false, "生成示例配置文件")
}

func runBatch(cmd *cobra.Command, args []string) error {
	appCtx := mustAppContext(cmd)

	// 如果请求示例配置
	if batchSample {
		fmt.Println(batch.GenerateSampleConfig())
		return nil
	}

	// 检查配置文件参数
	if batchFile == "" {
		return fmt.Errorf("请使用 -f 参数指定任务配置文件")
	}

	logger.Info("开始执行批量任务", zap.String("file", batchFile))

	// 创建执行器
	executor := batch.NewExecutor()

	// 执行任务
	ctx := context.Background()
	result, err := executor.ExecuteFile(ctx, batchFile)
	if err != nil {
		return fmt.Errorf("执行批量任务失败: %w", err)
	}

	// 输出结果
	return outputBatchResult(result, appCtx)
}

// outputBatchResult 输出批量任务结果
func outputBatchResult(result *batch.BatchResult, appCtx *app.Context) error {
	output := types.OutputText
	verbose := false
	noColor := false
	if appCtx != nil {
		if appCtx.Flags.Output != "" {
			output = types.OutputFormat(appCtx.Flags.Output)
		}
		verbose = appCtx.Flags.Verbose
		noColor = appCtx.Flags.NoColor
	}

	if output == types.OutputText || output == "" {
		return outputBatchText(result, verbose, noColor)
	}

	f := formatter.NewFormatter(output, noColor)
	return f.FormatTo(os.Stdout, result)
}

// outputBatchText 文本格式输出
func outputBatchText(result *batch.BatchResult, verbose bool, noColor bool) error {
	color.NoColor = noColor
	// 打印标题
	fmt.Println()
	fmt.Println("================================================================================")
	fmt.Println("  批量任务执行报告")
	fmt.Println("================================================================================")
	fmt.Println()

	// 显示任务执行结果
	fmt.Println(">>> 任务执行结果")
	table := formatter.NewTable(
		[]string{"任务名称", "类型", "状态", "耗时"},
		[]int{30, 10, 12, 20},
	)
	for _, taskResult := range result.TaskResults {
		statusStr := "✓ 成功"
		statusColor := color.GreenString
		if !taskResult.Success {
			statusStr = "✗ 失败"
			statusColor = color.RedString
		}

		table.AddRow(
			taskResult.TaskName,
			string(taskResult.TaskType),
			statusColor(statusStr),
			taskResult.Duration.Round(100).String(),
		)

		// 显示详细信息（如果是 verbose 模式）
		if verbose {
			if taskResult.Error != nil {
				fmt.Printf("  错误: %s\n", taskResult.Error)
			} else {
				fmt.Printf("  结果数: %d\n", len(taskResult.Results))
			}
		}
	}
	table.Render(os.Stdout)

	fmt.Println()

	// 显示统计信息
	fmt.Println(">>> 统计信息")
	fmt.Printf("总任务数:   %d\n", result.TotalTasks)
	fmt.Printf("成功任务:   %s\n", color.GreenString("%d", result.SuccessTasks))
	fmt.Printf("失败任务:   %s\n", color.RedString("%d", result.FailedTasks))
	fmt.Printf("总耗时:     %s\n", result.TotalDuration.Round(100))
	fmt.Println()

	// 显示任务详情
	if verbose {
		fmt.Println(">>> 任务详情")
		fmt.Println()

		for _, taskResult := range result.TaskResults {
			fmt.Println()
			fmt.Printf("任务: %s\n", color.CyanString(taskResult.TaskName))
			fmt.Println(strings.Repeat("-", 60))

			switch taskResult.TaskType {
			case batch.TaskTypePing:
				outputPingTaskDetails(taskResult)
			case batch.TaskTypeDNS:
				outputDNSTaskDetails(taskResult)
			case batch.TaskTypeScan:
				outputScanTaskDetails(taskResult)
			}
		}
	}

	return nil
}

// outputPingTaskDetails 输出 Ping 任务详情
func outputPingTaskDetails(taskResult *batch.TaskResult) {
	for _, r := range taskResult.Results {
		if pingResult, ok := r.(*types.PingResult); ok {
			fmt.Printf("  目标: %s (%s)\n", pingResult.Target.Hostname, pingResult.Target.IP)
			fmt.Printf("    平均延迟: %.2fms\n", float64(pingResult.Statistics.AvgRTT.Microseconds())/1000)
			fmt.Printf("    丢包率:   %.1f%%\n", pingResult.Statistics.LossRate)
		}
	}
}

// outputDNSTaskDetails 输出 DNS 任务详情
func outputDNSTaskDetails(taskResult *batch.TaskResult) {
	for _, r := range taskResult.Results {
		if dnsResult, ok := r.(*types.DNSResult); ok {
			fmt.Printf("  域名: %s\n", dnsResult.Domain)
			fmt.Printf("    记录数: %d\n", len(dnsResult.Records))
			if len(dnsResult.Records) > 0 {
				for _, record := range dnsResult.Records {
					fmt.Printf("      %s: %s\n", record.Type, record.Value)
				}
			}
		}
	}
}

// outputScanTaskDetails 输出扫描任务详情
func outputScanTaskDetails(taskResult *batch.TaskResult) {
	for _, r := range taskResult.Results {
		if scanResult, ok := r.(*types.ScanResult); ok {
			fmt.Printf("  目标: %s (%s)\n", scanResult.Target, scanResult.IP)
			fmt.Printf("    开放端口: %d\n", scanResult.Summary.OpenPorts)
			if scanResult.Summary.OpenPorts > 0 {
				fmt.Printf("    端口列表: ")
				openPorts := make([]int, 0)
				for _, port := range scanResult.Ports {
					if port.State == types.PortOpen {
						openPorts = append(openPorts, port.Port)
					}
				}
				fmt.Printf("%v\n", openPorts)
			}
		}
	}
}
