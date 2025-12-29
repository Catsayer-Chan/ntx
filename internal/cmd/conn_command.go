package cmd

import (
	"strings"

	"github.com/catsayer/ntx/internal/core/netstat"
	"github.com/catsayer/ntx/pkg/types"
	"github.com/spf13/cobra"
)

var (
	connTCP     bool
	connUDP     bool
	connListen  bool
	connProcess bool
	connState   string
	connPort    int
	connStats   bool
)

var connCmd = &cobra.Command{
	Use:     "conn",
	Aliases: []string{"connections", "netstat"},
	Short:   "显示网络连接状态",
	Long: `显示网络连接状态，类似 netstat 或 ss 命令。

显示信息包括:
  • TCP/UDP 连接列表
  • 监听端口
  • 连接状态
  • 进程信息 (需要 root 权限)
  • 连接统计

示例:
  # 显示所有连接
  ntx conn

  # 仅显示 TCP 连接
  ntx conn --tcp

  # 仅显示 UDP 连接
  ntx conn --udp

  # 仅显示监听端口
  ntx conn --listen

  # 显示进程信息 (需要 root)
  ntx conn --process

  # 按状态过滤
  ntx conn --state ESTABLISHED

  # 按端口过滤
  ntx conn --port 80

  # 显示统计信息
  ntx conn --stats

  # JSON 输出
  ntx conn -o json`,
	Run: runConn,
}

func init() {
	rootCmd.AddCommand(connCmd)

	connCmd.Flags().BoolVar(&connTCP, "tcp", false,
		"仅显示 TCP 连接")
	connCmd.Flags().BoolVar(&connUDP, "udp", false,
		"仅显示 UDP 连接")
	connCmd.Flags().BoolVarP(&connListen, "listen", "l", false,
		"仅显示监听端口")
	connCmd.Flags().BoolVarP(&connProcess, "process", "p", false,
		"显示进程信息 (需要 root 权限)")
	connCmd.Flags().StringVar(&connState, "state", "",
		"按状态过滤 (ESTABLISHED, LISTEN, TIME_WAIT 等)")
	connCmd.Flags().IntVar(&connPort, "port", 0,
		"按端口过滤")
	connCmd.Flags().BoolVar(&connStats, "stats", false,
		"显示统计信息")
}

func runConn(cmd *cobra.Command, args []string) {
	reader := netstat.NewNetStatReader()
	defer reader.Close()

	outputFormat := outputFormatFromCmd(cmd)
	noColor := noColorFromCmd(cmd)

	if connStats {
		runConnStats(reader, outputFormat, noColor)
		return
	}

	opts := buildConnOptions()
	if connListen {
		runConnListeners(reader, opts, outputFormat, noColor)
	} else {
		runConnConnections(reader, opts, outputFormat, noColor)
	}
}

func buildConnOptions() *types.NetStatOptions {
	opts := &types.NetStatOptions{
		Protocol:       "all",
		IncludeProcess: connProcess,
		ListenOnly:     connListen,
	}

	if connTCP {
		opts.Protocol = "tcp"
	} else if connUDP {
		opts.Protocol = "udp"
	}

	if connPort > 0 {
		opts.LocalPort = connPort
	}

	if connState != "" {
		opts.State = []types.ConnectionState{
			types.ConnectionState(strings.ToUpper(connState)),
		}
	}

	return opts
}
