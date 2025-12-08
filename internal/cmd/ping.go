// Package cmd 提供 Ping 命令实现
//
// 作者: Catsayer
package cmd

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/catsayer/ntx/internal/core/ping"
	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/internal/output/formatter"
	"github.com/catsayer/ntx/pkg/errors"
	"github.com/catsayer/ntx/pkg/types"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
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
	Long: `使用 TCP、ICMP 或 HTTP 协议 Ping 一个主机。

TCP Ping（默认）:
  通过建立 TCP 连接来测试端口可达性
  不需要特殊权限，推荐日常使用

ICMP Ping:
  使用 ICMP Echo Request/Reply 测试网络连通性
  需要 root 权限或 CAP_NET_RAW 能力

HTTP Ping:
  通过发送 HTTP 请求来测试 Web 服务
  不需要特殊权限

示例:
  # TCP Ping (默认，连接 80 端口)
  ntx ping google.com

  # TCP Ping 指定端口
  ntx ping google.com --port 443

  # ICMP Ping (需要 root 权限)
  sudo ntx ping google.com --protocol icmp

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
	pingCmd.Flags().StringVarP(&pingProtocol, "protocol", "p", "tcp",
		"协议类型: tcp, icmp, http (默认: tcp)")

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

func runPing(_ *cobra.Command, args []string) {
	target := args[0]

	// 检查输出格式，非text格式仍使用旧方式（批量输出）
	outputFormat := types.OutputFormat(Output)
	if outputFormat != types.OutputText && outputFormat != "" {
		runPingBatch(target)
		return
	}

	// Text格式使用实时输出
	runPingRealtime(target)
}

// runPingRealtime 实时输出模式
func runPingRealtime(target string) {
	logger.Info("开始 Ping",
		zap.String("target", target),
		zap.String("protocol", pingProtocol))

	// 设置颜色函数
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	if NoColor {
		color.NoColor = true
		green = fmt.Sprint
		red = fmt.Sprint
	}

	// 解析协议
	protocol := types.Protocol(strings.ToLower(pingProtocol))
	if protocol != types.ProtocolICMP && protocol != types.ProtocolTCP && protocol != types.ProtocolHTTP {
		logger.Error("无效的协议", zap.String("protocol", pingProtocol))
		fmt.Fprintf(os.Stderr, "错误: 无效的协议 '%s'，支持的协议: tcp, icmp, http\n", pingProtocol)
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
				logger.Warn("ICMP 需要权限，切换到 TCP",
					zap.Error(err))
				pinger = ping.NewTCPPinger()
				opts.Protocol = types.ProtocolTCP
				protocol = types.ProtocolTCP
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

	// 执行一次ping来获取目标信息
	singleOpt := *opts
	singleOpt.Count = 1
	result, err := pinger.Ping(target, &singleOpt)
	if err != nil {
		logger.Error("Ping 失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	// 打印头部信息
	targetHostname := result.Target.Hostname
	targetIP := result.Target.IP
	fmt.Printf("PING %s (%s): %d data bytes\n", targetHostname, targetIP, opts.Size)

	// 实时ping循环
	sent := 1 // 已经发送了一次
	received := 0
	var rtts []time.Duration

	// 打印第一次的结果
	if len(result.Replies) > 0 && result.Replies[0].Status == types.StatusSuccess {
		reply := result.Replies[0]
		fmt.Println(green(fmt.Sprintf("%d bytes from %s: icmp_seq=%d ttl=%d time=%.3f ms",
			reply.Bytes, reply.From, reply.Seq-1, reply.TTL, float64(reply.RTT.Microseconds())/1000.0)))
		received++
		rtts = append(rtts, reply.RTT)
	} else if len(result.Replies) > 0 {
		fmt.Println(red(fmt.Sprintf("Request timeout for icmp_seq=%d", 0)))
	}

	// 继续剩余的ping
	for i := 1; i < opts.Count; i++ {
		time.Sleep(opts.Interval)

		singleOpt.Count = 1
		result, err = pinger.Ping(target, &singleOpt)
		sent++

		if err == nil && len(result.Replies) > 0 && result.Replies[0].Status == types.StatusSuccess {
			reply := result.Replies[0]
			fmt.Println(green(fmt.Sprintf("%d bytes from %s: icmp_seq=%d ttl=%d time=%.3f ms",
				reply.Bytes, reply.From, i, reply.TTL, float64(reply.RTT.Microseconds())/1000.0)))
			received++
			rtts = append(rtts, reply.RTT)
		} else {
			fmt.Println(red(fmt.Sprintf("Request timeout for icmp_seq=%d", i)))
		}
	}

	// 打印统计信息
	lossRate := float64(sent-received) / float64(sent) * 100.0

	fmt.Printf("\n--- %s ping statistics ---\n", targetHostname)
	fmt.Printf("%d packets transmitted, %d packets received, %.1f%% packet loss\n",
		sent, received, lossRate)

	if len(rtts) > 0 {
		MeMin, MeMax, avg, stddev := calculateStats(rtts)
		fmt.Printf("round-trip MeMin/avg/MeMax/stddev = %.3f/%.3f/%.3f/%.3f ms\n",
			float64(MeMin.Microseconds())/1000.0,
			float64(avg.Microseconds())/1000.0,
			float64(MeMax.Microseconds())/1000.0,
			float64(stddev.Microseconds())/1000.0)
	}

	// 根据结果设置退出码
	if received == 0 {
		os.Exit(1)
	}
}

// runPingBatch 批量输出模式（JSON/YAML等格式）
func runPingBatch(target string) {
	logger.Info("开始 Ping",
		zap.String("target", target),
		zap.String("protocol", pingProtocol))

	// 解析协议
	protocol := types.Protocol(strings.ToLower(pingProtocol))
	if protocol != types.ProtocolICMP && protocol != types.ProtocolTCP && protocol != types.ProtocolHTTP {
		logger.Error("无效的协议", zap.String("protocol", pingProtocol))
		fmt.Fprintf(os.Stderr, "错误: 无效的协议 '%s'，支持的协议: tcp, icmp, http\n", pingProtocol)
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
	outputFormat := types.OutputFormat(Output)
	f := formatter.NewFormatter(outputFormat, NoColor)
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

// calculateStats 计算统计信息
func calculateStats(rtts []time.Duration) (min, max, avg, stddev time.Duration) {
	if len(rtts) == 0 {
		return
	}

	min = rtts[0]
	max = rtts[0]
	var sum time.Duration

	for _, rtt := range rtts {
		if rtt < min {
			min = rtt
		}
		if rtt > max {
			max = rtt
		}
		sum += rtt
	}

	avg = sum / time.Duration(len(rtts))

	// 计算标准差
	var variance float64
	avgFloat := float64(avg.Nanoseconds())
	for _, rtt := range rtts {
		diff := float64(rtt.Nanoseconds()) - avgFloat
		variance += diff * diff
	}
	variance /= float64(len(rtts))

	// 使用math.Sqrt计算标准差
	stddevFloat := math.Sqrt(variance)
	stddev = time.Duration(int64(stddevFloat))

	return
}
