// Package cmd 提供 Traceroute 命令实现
//
// 作者: Catsayer
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/catsayer/ntx/internal/app"
	"github.com/catsayer/ntx/internal/cmd/options"
	"github.com/catsayer/ntx/internal/core/trace"
	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/internal/output/formatter"
	"github.com/catsayer/ntx/pkg/errors"
	"github.com/catsayer/ntx/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

var (
	traceMaxHops  int
	traceTimeout  float64
	traceQueries  int
	tracePort     int
	traceIPv4     bool
	traceIPv6     bool
	traceFirstTTL int
)

// traceCmd 表示 trace 命令
var traceCmd = &cobra.Command{
	Use:     "trace <target>",
	Aliases: []string{"traceroute", "tr"},
	Short:   "追踪到目标主机的网络路径",
	Long: `追踪数据包到达目标主机所经过的路由路径。

Traceroute 通过发送具有递增 TTL (Time To Live) 值的数据包来追踪路由路径。
每个路由器在转发数据包时会将 TTL 减 1，当 TTL 为 0 时会返回 ICMP Time Exceeded 消息。
通过这种方式，可以逐跳发现数据包经过的路由器。

ICMP Traceroute:
  使用 ICMP Echo Request 数据包
  需要 root 权限或 CAP_NET_RAW 能力

示例:
  # 基本 traceroute
  ntx trace google.com

  # 指定最大跳数
  ntx trace google.com --max-hops 20

  # 每跳查询 5 次
  ntx trace google.com --queries 5

  # 从第 5 跳开始
  ntx trace google.com --first-ttl 5

  # 表格输出
  ntx trace google.com -o table

  # JSON 输出
  ntx trace google.com -o json`,
	Args: cobra.ExactArgs(1),
	Run:  runTrace,
}

func init() {
	rootCmd.AddCommand(traceCmd)

	// 基本选项
	traceCmd.Flags().IntVarP(&traceMaxHops, "max-hops", "m", 30,
		"最大跳数")
	traceCmd.Flags().Float64VarP(&traceTimeout, "timeout", "t", 3.0,
		"每跳超时时间（秒）")
	traceCmd.Flags().IntVarP(&traceQueries, "queries", "q", 3,
		"每跳查询次数")
	traceCmd.Flags().IntVarP(&tracePort, "port", "p", 33434,
		"起始端口号（UDP）")
	traceCmd.Flags().IntVar(&traceFirstTTL, "first-ttl", 1,
		"起始 TTL 值")

	// IP 版本选项
	traceCmd.Flags().BoolVarP(&traceIPv4, "ipv4", "4", false,
		"强制使用 IPv4")
	traceCmd.Flags().BoolVarP(&traceIPv6, "ipv6", "6", false,
		"强制使用 IPv6")
}

func runTrace(cmd *cobra.Command, args []string) {
	appCtx := mustAppContext(cmd)
	target := args[0]
	opts := buildTraceOptions(cmd, appCtx)

	logger.Info("开始 Traceroute",
		zap.String("target", target),
		zap.Int("max_hops", opts.MaxHops))

	// 验证参数
	if opts.MaxHops <= 0 || opts.MaxHops > 255 {
		fmt.Fprintf(os.Stderr, "错误: 无效的最大跳数 %d，必须在 1-255 之间\n", opts.MaxHops)
		os.Exit(1)
	}

	if opts.Queries <= 0 || opts.Queries > 10 {
		fmt.Fprintf(os.Stderr, "错误: 无效的查询次数 %d，必须在 1-10 之间\n", opts.Queries)
		os.Exit(1)
	}

	if opts.FirstTTL <= 0 || opts.FirstTTL > opts.MaxHops {
		fmt.Fprintf(os.Stderr, "错误: 无效的起始 TTL %d，必须在 1-%d 之间\n", opts.FirstTTL, opts.MaxHops)
		os.Exit(1)
	}

	// 创建 Tracer
	tracer, err := trace.NewICMPTracer()
	if err != nil {
		if errors.IsPermissionDenied(err) {
			logger.Error("ICMP Traceroute 需要 root 权限", zap.Error(err))
			fmt.Fprintf(os.Stderr, "错误: ICMP Traceroute 需要 root 权限或 CAP_NET_RAW 能力\n")
			fmt.Fprintf(os.Stderr, "提示: 请使用 sudo 运行命令\n")
			fmt.Fprintf(os.Stderr, "      sudo %s trace %s\n", os.Args[0], target)
		} else {
			logger.Error("创建 Tracer 失败", zap.Error(err))
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		}
		os.Exit(1)
	}
	defer tracer.Close()

	// 执行 Traceroute
	result, err := tracer.Trace(target, opts)
	if err != nil {
		logger.Error("Traceroute 失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	// 格式化输出
	outputFormat := types.OutputFormat(appCtx.Flags.Output)
	noColor := appCtx.Flags.NoColor

	f := formatter.NewFormatter(outputFormat, noColor)
	output, err := f.Format(result)
	if err != nil {
		logger.Error("格式化输出失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: 格式化输出失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(output)

	// 根据结果设置退出码
	if !result.ReachedDestination {
		os.Exit(1)
	}
}

func buildTraceOptions(cmd *cobra.Command, appCtx *app.Context) *types.TraceOptions {
	return options.NewBuilder(types.DefaultTraceOptions()).
		WithContext(appCtx).
		WithCommand(cmd).
		ApplyConfig(func(opts *types.TraceOptions, ctx *app.Context) {
			if ctx == nil || ctx.Config == nil {
				return
			}
			cfg := ctx.Config.Trace
			if cfg.Protocol != "" {
				opts.Protocol = cfg.Protocol
			}
			if cfg.MaxHops > 0 {
				opts.MaxHops = cfg.MaxHops
			}
			if cfg.Timeout > 0 {
				opts.Timeout = cfg.Timeout
			}
			if cfg.Queries > 0 {
				opts.Queries = cfg.Queries
			}
			if cfg.Port != 0 {
				opts.Port = cfg.Port
			}
			if cfg.PacketSize > 0 {
				opts.PacketSize = cfg.PacketSize
			}
			if cfg.IPVersion != 0 {
				opts.IPVersion = cfg.IPVersion
			}
			if cfg.FirstTTL > 0 {
				opts.FirstTTL = cfg.FirstTTL
			}
		}).
		ApplyFlags(func(opts *types.TraceOptions, flags *pflag.FlagSet) {
			if flags.Changed("max-hops") {
				opts.MaxHops = traceMaxHops
			}
			if flags.Changed("timeout") {
				opts.Timeout = time.Duration(traceTimeout * float64(time.Second))
			}
			if flags.Changed("queries") {
				opts.Queries = traceQueries
			}
			if flags.Changed("port") {
				opts.Port = tracePort
			}
			if flags.Changed("first-ttl") {
				opts.FirstTTL = traceFirstTTL
			}
			if flags.Changed("ipv4") && traceIPv4 {
				opts.IPVersion = types.IPv4
			} else if flags.Changed("ipv6") && traceIPv6 {
				opts.IPVersion = types.IPv6
			}
		}).
		Result()
}
