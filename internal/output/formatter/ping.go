// Package formatter 提供 Ping 结果格式化
//
// 作者: Catsayer
package formatter

import (
	"fmt"
	"strings"
	"time"

	"github.com/catsayer/ntx/pkg/termutil"
	"github.com/catsayer/ntx/pkg/types"
)

// FormatPingText 格式化 Ping 结果为文本
func FormatPingText(result *types.PingResult, noColor bool) string {
	var sb strings.Builder

	// 设置颜色函数
	printer := termutil.NewColorPrinter(noColor)
	green := printer.Success
	red := printer.Error
	yellow := printer.Warning
	cyan := printer.Info
	bold := printer.Bold

	// 标题
	sb.WriteString(bold(fmt.Sprintf("PING %s (%s) %s protocol\n",
		result.Target.Hostname,
		result.Target.IP,
		result.Protocol)))
	sb.WriteString(strings.Repeat("-", types.TableWidthPingText) + "\n")

	// 响应列表
	for _, reply := range result.Replies {
		switch reply.Status {
		case types.StatusSuccess:
			sb.WriteString(green(fmt.Sprintf("Reply from %s: bytes=%d time=%v ttl=%d seq=%d\n",
				reply.From,
				reply.Bytes,
				formatDuration(reply.RTT),
				reply.TTL,
				reply.Seq)))
		case types.StatusTimeout:
			sb.WriteString(red(fmt.Sprintf("Request timeout for seq=%d\n", reply.Seq)))
		case types.StatusFailure:
			sb.WriteString(red(fmt.Sprintf("Request failed for seq=%d: %s\n", reply.Seq, reply.Error)))
		default:
			sb.WriteString(yellow(fmt.Sprintf("Unknown status for seq=%d\n", reply.Seq)))
		}
	}

	// 统计信息
	if result.Statistics != nil {
		sb.WriteString("\n" + strings.Repeat("-", types.TableWidthPingText) + "\n")
		sb.WriteString(bold(fmt.Sprintf("--- %s ping statistics ---\n", result.Target.Hostname)))

		stats := result.Statistics
		sb.WriteString(fmt.Sprintf("%d packets transmitted, %d packets received, %.1f%% packet loss\n",
			stats.Sent,
			stats.Received,
			stats.LossRate))

		if stats.Received > 0 {
			sb.WriteString(cyan(fmt.Sprintf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
				formatDuration(stats.MinRTT),
				formatDuration(stats.AvgRTT),
				formatDuration(stats.MaxRTT),
				formatDuration(stats.StdDevRTT))))
		}

		if result.Context != nil {
			sb.WriteString(fmt.Sprintf("time %v\n", formatDuration(result.Context.Duration)))
		}
	}

	// 错误信息
	if result.Error != nil {
		sb.WriteString("\n" + red(fmt.Sprintf("Error: %v\n", result.Error)))
	}

	return sb.String()
}

// FormatPingTable 格式化 Ping 结果为表格
func FormatPingTable(result *types.PingResult, noColor bool) string {
	var sb strings.Builder

	// 设置颜色函数
	printer := termutil.NewColorPrinter(noColor)
	green := printer.Success
	red := printer.Error
	yellow := printer.Warning
	bold := printer.Bold

	// 标题
	sb.WriteString(bold(fmt.Sprintf("PING %s (%s)\n\n", result.Target.Hostname, result.Target.IP)))

	// 表头
	headerFormat := fmt.Sprintf("%%-%ds %%-%ds %%-%ds %%-%ds %%-%ds %%-%ds",
		types.ColumnWidthPingSeq,
		types.ColumnWidthPingFrom,
		types.ColumnWidthPingBytes,
		types.ColumnWidthPingTTL,
		types.ColumnWidthPingTime,
		types.ColumnWidthPingStatus)
	rowFormat := fmt.Sprintf("%%-%dd %%-%ds %%-%dd %%-%dd %%-%ds %%-%ds",
		types.ColumnWidthPingSeq,
		types.ColumnWidthPingFrom,
		types.ColumnWidthPingBytes,
		types.ColumnWidthPingTTL,
		types.ColumnWidthPingTime,
		types.ColumnWidthPingStatus)
	header := fmt.Sprintf(headerFormat,
		"SEQ",
		"FROM",
		"BYTES",
		"TTL",
		"TIME",
		"STATUS")
	sb.WriteString(bold(header) + "\n")
	sb.WriteString(strings.Repeat("-", types.TableWidthPingTable) + "\n")

	// 数据行
	for _, reply := range result.Replies {
		statusStr := ""
		switch reply.Status {
		case types.StatusSuccess:
			statusStr = green("OK")
		case types.StatusTimeout:
			statusStr = red("TIMEOUT")
		case types.StatusFailure:
			statusStr = red("FAILED")
		default:
			statusStr = yellow("UNKNOWN")
		}

		row := fmt.Sprintf(rowFormat,
			reply.Seq,
			reply.From,
			reply.Bytes,
			reply.TTL,
			formatDuration(reply.RTT),
			statusStr)

		sb.WriteString(row + "\n")
	}

	// 统计信息
	if result.Statistics != nil {
		sb.WriteString("\n" + strings.Repeat("-", types.TableWidthPingTable) + "\n")
		sb.WriteString(bold("Statistics:\n"))

		stats := result.Statistics
		sb.WriteString(fmt.Sprintf("  Sent:     %d\n", stats.Sent))
		sb.WriteString(fmt.Sprintf("  Received: %d\n", stats.Received))
		sb.WriteString(fmt.Sprintf("  Loss:     %d (%.1f%%)\n", stats.Loss, stats.LossRate))

		if stats.Received > 0 {
			sb.WriteString(fmt.Sprintf("  Min RTT:  %v\n", formatDuration(stats.MinRTT)))
			sb.WriteString(fmt.Sprintf("  Avg RTT:  %v\n", formatDuration(stats.AvgRTT)))
			sb.WriteString(fmt.Sprintf("  Max RTT:  %v\n", formatDuration(stats.MaxRTT)))
			sb.WriteString(fmt.Sprintf("  Std Dev:  %v\n", formatDuration(stats.StdDevRTT)))
		}

		if result.Context != nil {
			sb.WriteString(fmt.Sprintf("  Duration: %v\n", formatDuration(result.Context.Duration)))
		}
	}

	return sb.String()
}

// formatDuration 格式化时间间隔
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}

	// 转换为毫秒显示
	ms := float64(d.Microseconds()) / 1000.0
	if ms < 1 {
		// 小于 1ms，显示微秒
		return fmt.Sprintf("%.0fµs", float64(d.Microseconds()))
	} else if ms < 1000 {
		// 小于 1s，显示毫秒
		return fmt.Sprintf("%.3fms", ms)
	} else {
		// 大于 1s，显示秒
		return fmt.Sprintf("%.3fs", ms/1000.0)
	}
}
