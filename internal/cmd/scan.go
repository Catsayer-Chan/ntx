// Package cmd 提供 scan 命令实现
//
// 本文件实现端口扫描命令，支持:
// - TCP Connect 扫描
// - 自定义端口列表
// - 服务识别
// - 多种输出格式
//
// 使用示例:
//
//	ntx scan 192.168.1.1 -p 1-1024
//	ntx scan example.com --service
//	ntx scan 10.0.0.1 -o json
//
// 作者: Catsayer
package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/catsayer/ntx/internal/app"
	"github.com/catsayer/ntx/internal/cmd/options"
	"github.com/catsayer/ntx/internal/core/scan"
	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/internal/output/formatter"
	"github.com/catsayer/ntx/pkg/types"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

var (
	scanPorts       string
	scanTimeout     int
	scanConcurrency int
	scanService     bool
	scanFast        bool
)

var scanCmd = &cobra.Command{
	Use:   "scan <target>",
	Short: "端口扫描",
	Long: `对目标主机执行端口扫描。

支持功能:
  • TCP Connect 扫描（无需特权）
  • 自定义端口列表
  • 服务识别
  • 并发扫描
  • 多种输出格式

示例:
  ntx scan 192.168.1.1                  # 扫描常用端口
  ntx scan 192.168.1.1 -p 1-1024        # 扫描端口范围
  ntx scan 192.168.1.1 -p 80,443,8080   # 扫描指定端口
  ntx scan example.com --service        # 启用服务识别
  ntx scan 192.168.1.1 --fast           # 快速扫描
  ntx scan 192.168.1.1 -o json          # JSON 输出`,
	Args: cobra.ExactArgs(1),
	RunE: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// 扫描参数
	scanCmd.Flags().StringVarP(&scanPorts, "ports", "p", "", "端口列表 (如: 80,443 或 1-1024)")
	scanCmd.Flags().IntVarP(&scanTimeout, "timeout", "t", 3, "单端口超时时间（秒）")
	scanCmd.Flags().IntVarP(&scanConcurrency, "concurrency", "c", 100, "并发扫描数量")
	scanCmd.Flags().BoolVar(&scanService, "service", false, "启用服务识别")
	scanCmd.Flags().BoolVar(&scanFast, "fast", false, "快速扫描模式（仅检测开放端口）")
}

func runScan(cmd *cobra.Command, args []string) error {
	appCtx := mustAppContext(cmd)
	target := args[0]

	logger.Info("开始端口扫描", zap.String("target", target))

	// 显示安全警告
	color.NoColor = appCtx.Flags.NoColor
	fmt.Println(color.YellowString("⚠️  安全提示:"))
	fmt.Println(color.YellowString("   端口扫描功能仅用于合法授权场景"))
	fmt.Println(color.YellowString("   未经授权扫描他人系统属于非法行为"))
	fmt.Println()

	// 构建扫描选项
	opts := buildScanOptions(cmd, appCtx)

	// 解析端口列表
	if scanPorts != "" {
		ports, err := parsePortList(scanPorts)
		if err != nil {
			return fmt.Errorf("解析端口列表失败: %w", err)
		}
		opts.Ports = ports
	}

	// 快速扫描模式
	if scanFast {
		opts.Timeout = 1 * time.Second
		opts.Concurrency = 200
	}

	// 创建扫描器
	scanner := scan.NewTCPScanner()

	// 执行扫描
	ctx := context.Background()
	result, err := scanner.Scan(ctx, target, opts)
	if err != nil {
		return fmt.Errorf("扫描失败: %w", err)
	}

	// 输出结果
	return outputScanResult(result, appCtx.Flags)
}

func buildScanOptions(cmd *cobra.Command, appCtx *app.Context) types.ScanOptions {
	defaults := types.DefaultScanOptions()
	optsPtr := options.NewBuilder(&defaults).
		WithContext(appCtx).
		WithCommand(cmd).
		ApplyConfig(func(opts *types.ScanOptions, ctx *app.Context) {
			if ctx == nil || ctx.Config == nil {
				return
			}
			if ctx.Config.Scan.Timeout > 0 {
				opts.Timeout = ctx.Config.Scan.Timeout
			}
			if ctx.Config.Scan.Concurrency > 0 {
				opts.Concurrency = ctx.Config.Scan.Concurrency
			}
			opts.ServiceDetect = ctx.Config.Scan.ServiceDetect
		}).
		ApplyFlags(func(opts *types.ScanOptions, flags *pflag.FlagSet) {
			if flags.Changed("timeout") {
				opts.Timeout = time.Duration(scanTimeout) * time.Second
			}
			if flags.Changed("concurrency") {
				opts.Concurrency = scanConcurrency
			}
			if flags.Changed("service") {
				opts.ServiceDetect = scanService
			}
		}).
		Result()

	opts := *optsPtr

	// 快速扫描模式覆盖部分参数
	if scanFast {
		opts.Timeout = 1 * time.Second
		opts.Concurrency = 200
	}

	return opts
}

// parsePortList 解析端口列表字符串
// 支持格式: "80,443,8000-9000"
func parsePortList(portStr string) ([]int, error) {
	var ports []int
	parts := strings.Split(portStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// 检查是否是范围
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("无效的端口范围: %s", part)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("无效的起始端口: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("无效的结束端口: %s", rangeParts[1])
			}

			if start < 1 || end > 65535 || start > end {
				return nil, fmt.Errorf("端口范围无效: %d-%d", start, end)
			}

			for i := start; i <= end; i++ {
				ports = append(ports, i)
			}
		} else {
			// 单个端口
			port, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("无效的端口号: %s", part)
			}

			if port < 1 || port > 65535 {
				return nil, fmt.Errorf("端口号超出范围: %d", port)
			}

			ports = append(ports, port)
		}
	}

	return ports, nil
}

// outputScanResult 输出扫描结果
func outputScanResult(result *types.ScanResult, flags app.GlobalFlags) error {
	outputFormat := types.OutputFormat(flags.Output)
	if outputFormat == types.OutputText || outputFormat == "" {
		return outputScanText(result, flags)
	}

	f := formatter.NewFormatter(outputFormat, flags.NoColor)
	return f.FormatTo(os.Stdout, result)
}

// outputScanText 文本格式输出
func outputScanText(result *types.ScanResult, flags app.GlobalFlags) error {
	color.NoColor = flags.NoColor
	// 打印标题
	fmt.Println()
	fmt.Println("================================================================================")
	fmt.Printf("  扫描报告: %s (%s)\n", result.Target, result.IP.String())
	fmt.Println("================================================================================")
	fmt.Println()

	// 显示开放端口
	openPorts := make([]*types.ScanPort, 0)
	for _, port := range result.Ports {
		if port.State == types.PortOpen {
			openPorts = append(openPorts, port)
		}
	}

	if len(openPorts) > 0 {
		fmt.Println(color.GreenString("开放端口:"))
		table := formatter.NewTable(
			[]string{"端口", "状态", "服务", "响应时间"},
			[]int{10, 15, 20, 15},
		)
		for _, port := range openPorts {
			stateStr := color.GreenString(port.State.String())
			serviceStr := port.Service
			if serviceStr == "" || serviceStr == "unknown" {
				serviceStr = "-"
			}
			table.AddRow(
				fmt.Sprintf("%d", port.Port),
				stateStr,
				serviceStr,
				port.ResponseTime.Round(time.Millisecond).String(),
			)
		}
		table.Render(os.Stdout)
		fmt.Println()
	} else {
		fmt.Println(color.YellowString("未发现开放端口"))
		fmt.Println()
	}

	// 显示统计信息
	fmt.Println(">>> 统计信息")
	fmt.Printf("总端口数:   %d\n", result.Summary.TotalPorts)
	fmt.Printf("开放端口:   %s\n", color.GreenString("%d", result.Summary.OpenPorts))
	fmt.Printf("关闭端口:   %d\n", result.Summary.ClosedPorts)
	fmt.Printf("过滤端口:   %d\n", result.Summary.FilteredPorts)
	fmt.Printf("扫描耗时:   %s\n", result.Summary.Duration.Round(time.Millisecond))
	fmt.Println()

	return nil
}
