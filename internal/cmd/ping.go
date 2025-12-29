// Package cmd 提供 Ping 命令实现
//
// 作者: Catsayer
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/catsayer/ntx/internal/app"
	"github.com/catsayer/ntx/internal/cmd/options"
	pingcmd "github.com/catsayer/ntx/internal/cmd/ping"
	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

var (
	pingProtocol string
	pingCount    int
	pingInterval float64
	pingTimeout  float64
	pingSize     int
	pingTTL      int
	pingPort     int
	pingIPv4     bool
	pingIPv6     bool
	pingMonitor  bool
)

// pingCmd 表示 ping 命令
var pingCmd = &cobra.Command{
	Use:   "ping <target...>",
	Short: "Ping one or more hosts",
	Long: `Use ICMP, TCP, or HTTP to ping one or more hosts.

ICMP Ping (default, recommended):
  Tests connectivity using ICMP Echo Request/Reply.
  Requires root privileges or the CAP_NET_RAW capability.
  Automatically falls back to TCP Ping if permissions are insufficient.

TCP Ping:
  Tests port reachability by establishing a TCP connection.
  Does not require special privileges but can only test specific ports.

HTTP Ping:
  Tests a web service by sending an HTTP request.
  Does not require special privileges.

Examples:
  # ICMP Ping a single host (default)
  ntx ping google.com

  # Ping multiple hosts sequentially
  ntx ping google.com baidu.com

  # TCP Ping a specific port
  ntx ping google.com --protocol tcp --port 443

  # HTTP Ping
  ntx ping https://www.google.com --protocol http

  # Specify count and interval
  ntx ping google.com -c 10 -i 0.5

  # Real-time monitoring chart
  ntx ping google.com --monitor

  # JSON output for multiple hosts (executed concurrently)
  ntx ping google.com baidu.com -c 3 -o json`,
	Args: cobra.MinimumNArgs(1),
	Run:  runPing,
}

func init() {
	rootCmd.AddCommand(pingCmd)

	// 协议选项
	pingCmd.Flags().StringVarP(&pingProtocol, "protocol", "p", "icmp",
		"协议类型: icmp, tcp, http (默认: icmp, 无权限时自动降级到 tcp)")

	// 基本选项
	pingCmd.Flags().IntVarP(&pingCount, "count", "c", 4,
		"发送次数，0 表示无限次")
	pingCmd.Flags().Float64VarP(&pingInterval, "interval", "i", 1.0,
		"发送间隔（秒）")
	pingCmd.Flags().Float64VarP(&pingTimeout, "timeout", "t", 5.0,
		"超时时间（秒）")

	// 模式选项
	pingCmd.Flags().BoolVar(&pingMonitor, "monitor", false, "显示实时延迟图表")

	// ICMP 选项
	pingCmd.Flags().IntVarP(&pingSize, "size", "s", 64,
		"数据包大小（字节）")
	pingCmd.Flags().IntVar(&pingTTL, "ttl", 64,
		"Time To Live")

	// TCP/HTTP 选项
	pingCmd.Flags().IntVar(&pingPort, "port", 0,
		"端口号（TCP/HTTP）")

	// IP 版本选项
	pingCmd.Flags().BoolVarP(&pingIPv4, "ipv4", "4", false,
		"强制使用 IPv4")
	pingCmd.Flags().BoolVarP(&pingIPv6, "ipv6", "6", false,
		"强制使用 IPv6")
}

func runPing(cmd *cobra.Command, args []string) {
	appCtx := mustAppContext(cmd)

	// 1. 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听 Ctrl+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
	}()

	// 2. 解析和验证选项
	opts := buildPingOptions(cmd, appCtx)
	protocol := opts.Protocol
	if protocol != types.ProtocolICMP && protocol != types.ProtocolTCP && protocol != types.ProtocolHTTP {
		logger.Error("无效的协议", zap.String("protocol", string(protocol)))
		fmt.Fprintf(os.Stderr, "错误: 无效的协议 '%s'，支持的协议: tcp, icmp, http\n", protocol)
		os.Exit(1)
	}

	// 3. 根据输出格式选择执行模式
	outputFormat := types.OutputFormat(appCtx.Flags.Output)
	mode := pingcmd.ModeStream
	if pingMonitor {
		mode = pingcmd.ModeMonitor
		if len(args) > 1 {
			fmt.Fprintln(os.Stderr, "警告: 监控模式当前只支持单个目标")
		}
		opts.Count = 0
	}
	if outputFormat != types.OutputText && outputFormat != "" && mode != pingcmd.ModeMonitor {
		mode = pingcmd.ModeBatch
	}

	runner := pingcmd.NewRunner(pingcmd.Config{
		Mode:         mode,
		OutputFormat: outputFormat,
		NoColor:      appCtx.Flags.NoColor,
	}, appCtx.PingFactory)

	if err := runner.Run(ctx, args, opts); err != nil {
		if errors.Is(err, pingcmd.ErrPartialFailure) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

func buildPingOptions(cmd *cobra.Command, appCtx *app.Context) *types.PingOptions {
	return options.NewBuilder(types.DefaultPingOptions()).
		WithContext(appCtx).
		WithCommand(cmd).
		ApplyConfig(func(opts *types.PingOptions, ctx *app.Context) {
			if ctx == nil || ctx.Config == nil {
				return
			}
			cfg := ctx.Config.Ping
			if cfg.Protocol != "" {
				opts.Protocol = cfg.Protocol
			}
			opts.Count = cfg.Count
			if cfg.Interval > 0 {
				opts.Interval = cfg.Interval
			}
			if cfg.Timeout > 0 {
				opts.Timeout = cfg.Timeout
			}
			if cfg.Size > 0 {
				opts.Size = cfg.Size
			}
			if cfg.TTL > 0 {
				opts.TTL = cfg.TTL
			}
			opts.Port = cfg.Port
			if cfg.IPVersion != 0 {
				opts.IPVersion = cfg.IPVersion
			}
		}).
		ApplyFlags(func(opts *types.PingOptions, flags *pflag.FlagSet) {
			if flags.Changed("protocol") && pingProtocol != "" {
				opts.Protocol = types.Protocol(strings.ToLower(pingProtocol))
			}
			if flags.Changed("count") {
				opts.Count = pingCount
			}
			if flags.Changed("interval") {
				opts.Interval = time.Duration(pingInterval * float64(time.Second))
			}
			if flags.Changed("timeout") {
				opts.Timeout = time.Duration(pingTimeout * float64(time.Second))
			}
			if flags.Changed("size") {
				opts.Size = pingSize
			}
			if flags.Changed("ttl") {
				opts.TTL = pingTTL
			}
			if flags.Changed("port") {
				opts.Port = pingPort
			}
			if flags.Changed("ipv4") && pingIPv4 {
				opts.IPVersion = types.IPv4
			} else if flags.Changed("ipv6") && pingIPv6 {
				opts.IPVersion = types.IPv6
			}
		}).
		Result()
}
