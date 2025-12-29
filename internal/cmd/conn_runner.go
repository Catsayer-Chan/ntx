package cmd

import (
	"fmt"
	"os"

	"github.com/catsayer/ntx/internal/core/netstat"
	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/internal/output/formatter"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
)

func runConnConnections(reader *netstat.NetStatReader, opts *types.NetStatOptions, outputFormat types.OutputFormat, noColor bool) {
	logger.Info("查询网络连接")

	connections, err := reader.GetConnections(opts)
	if err != nil {
		logger.Error("获取连接列表失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if outputFormat == types.OutputText || outputFormat == "" {
		printConnectionsText(connections, noColor)
		return
	}

	f := formatter.NewFormatter(outputFormat, noColor)
	output, err := f.Format(connections)
	if err != nil {
		logger.Error("格式化输出失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: 格式化输出失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}

func runConnListeners(reader *netstat.NetStatReader, opts *types.NetStatOptions, outputFormat types.OutputFormat, noColor bool) {
	logger.Info("查询监听端口")

	listeners, err := reader.GetListeners(opts)
	if err != nil {
		logger.Error("获取监听端口失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if outputFormat == types.OutputText || outputFormat == "" {
		printListenersText(listeners, noColor)
		return
	}

	f := formatter.NewFormatter(outputFormat, noColor)
	output, err := f.Format(listeners)
	if err != nil {
		logger.Error("格式化输出失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: 格式化输出失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}

func runConnStats(reader *netstat.NetStatReader, outputFormat types.OutputFormat, noColor bool) {
	logger.Info("查询连接统计")

	stats, err := reader.GetStatistics()
	if err != nil {
		logger.Error("获取统计信息失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if outputFormat == types.OutputText || outputFormat == "" {
		printStatsText(stats, noColor)
		return
	}

	f := formatter.NewFormatter(outputFormat, noColor)
	output, err := f.Format(stats)
	if err != nil {
		logger.Error("格式化输出失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: 格式化输出失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}
