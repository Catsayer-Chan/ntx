// Package cmd 提供 Interface 命令实现
//
// 作者: Catsayer
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/catsayer/ntx/internal/core/iface"
	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/internal/output/formatter"
	"github.com/catsayer/ntx/pkg/termutil"
	"github.com/catsayer/ntx/pkg/types"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	ifaceDetail bool
	ifaceStats  bool
	ifaceRoutes bool
)

// ifaceCmd 表示 iface 命令
var ifaceCmd = &cobra.Command{
	Use:     "iface [interface]",
	Aliases: []string{"interface", "if"},
	Short:   "显示网卡信息",
	Long: `显示网络接口（网卡）的信息。

包括:
  • 网卡名称和索引
  • MAC 地址
  • IPv4/IPv6 地址
  • MTU
  • 网卡状态标志
  • 流量统计信息

示例:
  # 显示所有网卡
  ntx iface

  # 显示特定网卡
  ntx iface eth0

  # 显示详细信息
  ntx iface --detail

  # 显示统计信息
  ntx iface --stats

  # 显示路由表
  ntx iface --routes

  # JSON 输出
  ntx iface -o json`,
	Args: cobra.MaximumNArgs(1),
	Run:  runIface,
}

func init() {
	rootCmd.AddCommand(ifaceCmd)

	ifaceCmd.Flags().BoolVarP(&ifaceDetail, "detail", "d", false,
		"显示详细信息")
	ifaceCmd.Flags().BoolVar(&ifaceStats, "stats", false,
		"显示流量统计信息")
	ifaceCmd.Flags().BoolVar(&ifaceRoutes, "routes", false,
		"显示路由表")
}

func runIface(cmd *cobra.Command, args []string) {
	appCtx := mustAppContext(cmd)
	reader := iface.NewInterfaceReader()
	defer reader.Close()

	outputFormat := types.OutputFormat(appCtx.Flags.Output)
	noColor := appCtx.Flags.NoColor

	if ifaceRoutes {
		// 显示路由表
		runIfaceRoutes(reader, outputFormat, noColor)
		return
	}

	if len(args) == 1 {
		// 显示特定网卡
		runIfaceSingle(reader, args[0], outputFormat, noColor)
	} else {
		// 显示所有网卡
		runIfaceAll(reader, outputFormat, noColor)
	}
}

func runIfaceSingle(reader *iface.InterfaceReader, name string, outputFormat types.OutputFormat, noColor bool) {
	logger.Info("查询网卡信息", zap.String("interface", name))

	iface, err := reader.GetInterface(name)
	if err != nil {
		logger.Error("获取网卡信息失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if outputFormat == types.OutputText || outputFormat == "" {
		printInterfaceText(iface, ifaceDetail || ifaceStats, noColor)
	} else {
		f := formatter.NewFormatter(outputFormat, noColor)
		output, err := f.Format(iface)
		if err != nil {
			logger.Error("格式化输出失败", zap.Error(err))
			fmt.Fprintf(os.Stderr, "错误: 格式化输出失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(output)
	}
}

func runIfaceAll(reader *iface.InterfaceReader, outputFormat types.OutputFormat, noColor bool) {
	logger.Info("查询所有网卡信息")

	interfaces, err := reader.GetInterfaces()
	if err != nil {
		logger.Error("获取网卡列表失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if outputFormat == types.OutputText || outputFormat == "" {
		for i, iface := range interfaces {
			if i > 0 {
				fmt.Println()
			}
			printInterfaceText(iface, ifaceDetail || ifaceStats, noColor)
		}
	} else {
		f := formatter.NewFormatter(outputFormat, noColor)
		output, err := f.Format(interfaces)
		if err != nil {
			logger.Error("格式化输出失败", zap.Error(err))
			fmt.Fprintf(os.Stderr, "错误: 格式化输出失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(output)
	}
}

func runIfaceRoutes(reader *iface.InterfaceReader, outputFormat types.OutputFormat, noColor bool) {
	logger.Info("查询路由表")

	routes, err := reader.GetRoutes()
	if err != nil {
		logger.Error("获取路由表失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if outputFormat == types.OutputText || outputFormat == "" {
		printRoutesText(routes)
	} else {
		f := formatter.NewFormatter(outputFormat, noColor)
		output, err := f.Format(routes)
		if err != nil {
			logger.Error("格式化输出失败", zap.Error(err))
			fmt.Fprintf(os.Stderr, "错误: 格式化输出失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(output)
	}
}

func printInterfaceText(iface *types.Interface, showStats bool, noColor bool) {
	// 设置颜色
	printer := termutil.NewColorPrinter(noColor)
	bold := printer.Bold
	green := printer.Success
	yellow := printer.Warning

	// 显示网卡基本信息
	fmt.Printf("%s: %s\n", bold(iface.Name), strings.Join(iface.Flags, ","))

	if iface.HardwareAddr != "" {
		fmt.Printf("    Link/ether %s\n", green(iface.HardwareAddr))
	}

	// IPv4 地址
	if len(iface.IPv4Addrs) > 0 {
		for _, addr := range iface.IPv4Addrs {
			fmt.Printf("    inet %s\n", green(addr))
		}
	}

	// IPv6 地址
	if len(iface.IPv6Addrs) > 0 {
		for _, addr := range iface.IPv6Addrs {
			fmt.Printf("    inet6 %s\n", green(addr))
		}
	}

	// MTU
	fmt.Printf("    MTU: %s\n", yellow(fmt.Sprintf("%d", iface.MTU)))

	// 统计信息
	if showStats && iface.Stats != nil {
		stats := iface.Stats
		fmt.Printf("    RX: %s packets %s bytes (%s errors, %s dropped)\n",
			formatNumber(stats.RxPackets),
			formatBytes(stats.RxBytes),
			formatNumber(stats.RxErrors),
			formatNumber(stats.RxDropped))
		fmt.Printf("    TX: %s packets %s bytes (%s errors, %s dropped)\n",
			formatNumber(stats.TxPackets),
			formatBytes(stats.TxBytes),
			formatNumber(stats.TxErrors),
			formatNumber(stats.TxDropped))
	}
}

func printRoutesText(routes []*types.Route) {
	if len(routes) == 0 {
		fmt.Println("无路由信息")
		return
	}

	table := formatter.NewTable(
		[]string{"Destination", "Gateway", "Interface", "Metric", "Flags"},
		[]int{20, 16, 10, 10, 20},
	)
	for _, route := range routes {
		table.AddRow(
			route.Destination,
			route.Gateway,
			route.Interface,
			route.Metric,
			strings.Join(route.Flags, ","),
		)
	}
	table.Render(os.Stdout)
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatNumber(n uint64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	if n < 1000000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	return fmt.Sprintf("%.1fG", float64(n)/1000000000)
}
