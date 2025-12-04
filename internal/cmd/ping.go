// Package cmd 提供 Ping 命令实现
//
// 作者: Catsayer
package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/catsayer/ntx/internal/core/ping"
	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/internal/output/formatter"
	"github.com/catsayer/ntx/pkg/errors"
	"github.com/catsayer/ntx/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
)

// pingCmd 表示 ping 命令
var pingCmd = &cobra.Command{
	Use:   "ping <target>",
	Short: "Ping 一个主机",
	Long: `使用 ICMP、TCP 或 HTTP 协议 Ping 一个主机。

ICMP Ping（默认）:
  使用 ICMP Echo Request/Reply 测试网络连通性
  需要 root 权限或 CAP_NET_RAW 能力

TCP Ping:
  通过建立 TCP 连接来测试端口可达性
  不需要特殊权限

HTTP Ping:
  通过发送 HTTP 请求来测试 Web 服务
  不需要特殊权限

示例:
  # ICMP Ping (默认)
  ntx ping google.com

  # TCP Ping
  ntx ping google.com --protocol tcp --port 443

  # HTTP Ping
  ntx ping https://www.google.com --protocol http

  # 指定次数和间隔
  ntx ping google.com -c 10 -i 0.5

  # JSON 输出
  ntx ping google.com -c 3 -o json`,
	Args: cobra.ExactArgs(1),
	Run:  runPing,
}

func init() {
	rootCmd.AddCommand(pingCmd)

	// 协议选项
	pingCmd.Flags().StringVarP(&pingProtocol, "protocol", "p", "icmp",
		"协议类型: icmp, tcp, http")

	// 基本选项
	pingCmd.Flags().IntVarP(&pingCount, "count", "c", 4,
		"发送次数，0 表示无限次")
	pingCmd.Flags().Float64VarP(&pingInterval, "interval", "i", 1.0,
		"发送间隔（秒）")
	pingCmd.Flags().Float64VarP(&pingTimeout, "timeout", "t", 5.0,
		"超时时间（秒）")

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
	target := args[0]

	logger.Info("开始 Ping",
		zap.String("target", target),
		zap.String("protocol", pingProtocol))

	// 解析协议
	protocol := types.Protocol(strings.ToLower(pingProtocol))
	if protocol != types.ProtocolICMP && protocol != types.ProtocolTCP && protocol != types.ProtocolHTTP {
		logger.Error("无效的协议", zap.String("protocol", pingProtocol))
		fmt.Fprintf(os.Stderr, "错误: 无效的协议 '%s'，支持的协议: icmp, tcp, http\n", pingProtocol)
		os.Exit(1)
	}

	// 构建选项
	opts := &types.PingOptions{
		Protocol:  protocol,
		Count:     pingCount,
		Interval:  time.Duration(pingInterval * float64(time.Second)),
		Timeout:   time.Duration(pingTimeout * float64(time.Second)),
		Size:      pingSize,
		TTL:       pingTTL,
		Port:      pingPort,
		IPVersion: types.IPvAny,
	}

	// 设置 IP 版本
	if pingIPv4 {
		opts.IPVersion = types.IPv4
	} else if pingIPv6 {
		opts.IPVersion = types.IPv6
	}

	// 自动设置端口
	if opts.Port == 0 {
		switch protocol {
		case types.ProtocolTCP:
			opts.Port = 80
		case types.ProtocolHTTP:
			if strings.HasPrefix(target, "https://") {
				opts.Port = 443
			} else {
				opts.Port = 80
			}
		}
	}

	// 创建 Pinger
	var pinger types.Pinger
	var err error

	switch protocol {
	case types.ProtocolICMP:
		pinger, err = ping.NewICMPPinger()
		if err != nil {
			if errors.IsPermissionDenied(err) {
				logger.Warn("ICMP Ping 需要 root 权限，自动切换到 TCP Ping",
					zap.Error(err))
				fmt.Fprintf(os.Stderr, "警告: ICMP Ping 需要 root 权限，自动切换到 TCP Ping\n\n")
				pinger = ping.NewTCPPinger()
				opts.Protocol = types.ProtocolTCP
				if opts.Port == 0 {
					opts.Port = 80
				}
			} else {
				logger.Error("创建 ICMP Pinger 失败", zap.Error(err))
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		}
	case types.ProtocolTCP:
		pinger = ping.NewTCPPinger()
	case types.ProtocolHTTP:
		pinger = ping.NewHTTPPinger()
	}

	defer pinger.Close()

	// 执行 Ping
	result, err := pinger.Ping(target, opts)
	if err != nil {
		logger.Error("Ping 失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	// 格式化输出
	outputFormat := types.OutputFormat(viper.GetString("output"))
	noColor := viper.GetBool("no-color")

	f := formatter.NewFormatter(outputFormat, noColor)
	output, err := f.Format(result)
	if err != nil {
		logger.Error("格式化输出失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: 格式化输出失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(output)

	// 根据结果设置退出码
	if result.Statistics.Received == 0 {
		os.Exit(1)
	}
}