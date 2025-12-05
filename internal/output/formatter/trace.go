// Package formatter 提供 Traceroute 结果格式化
//
// 作者: Catsayer
package formatter

import (
	"fmt"
	"strings"

	"github.com/catsayer/ntx/pkg/types"
	"github.com/fatih/color"
)

// FormatTraceText 格式化 Traceroute 结果为文本
func FormatTraceText(result *types.TraceResult, noColor bool) string {
	var sb strings.Builder

	// 设置颜色函数
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()

	if noColor {
		color.NoColor = true
		green = fmt.Sprint
		red = fmt.Sprint
		yellow = fmt.Sprint
		cyan = fmt.Sprint
		bold = fmt.Sprint
		gray = fmt.Sprint
	}

	// 标题
	sb.WriteString(bold(fmt.Sprintf("traceroute to %s (%s), %d hops max, %s protocol\n",
		result.Target.Hostname,
		result.Target.IP,
		len(result.Hops),
		result.Protocol)))
	sb.WriteString(strings.Repeat("-", 70) + "\n")

	// 跳信息
	for _, hop := range result.Hops {
		// TTL
		sb.WriteString(cyan(fmt.Sprintf("%2d  ", hop.TTL)))

		// 主机名和 IP
		if hop.IP != "" {
			if hop.Hostname != "" && hop.Hostname != hop.IP {
				sb.WriteString(fmt.Sprintf("%-40s (%s)", hop.Hostname, hop.IP))
			} else {
				sb.WriteString(fmt.Sprintf("%-40s", hop.IP))
			}
		} else {
			sb.WriteString(gray("* * *"))
			sb.WriteString("\n")
			continue
		}

		// 探测结果
		sb.WriteString("  ")
		for i, probe := range hop.Probes {
			if i > 0 {
				sb.WriteString("  ")
			}
			switch probe.Status {
			case types.StatusSuccess:
				sb.WriteString(green(formatDuration(probe.RTT)))
			case types.StatusTimeout:
				sb.WriteString(gray("*"))
			case types.StatusFailure:
				sb.WriteString(red("!"))
			default:
				sb.WriteString(yellow("?"))
			}
		}

		// 目标标记
		if hop.IsDestination {
			sb.WriteString(bold(green("  [DEST]")))
		}

		sb.WriteString("\n")
	}

	// 统计信息
	sb.WriteString("\n" + strings.Repeat("-", 70) + "\n")
	if result.ReachedDestination {
		sb.WriteString(green(fmt.Sprintf("Trace complete: reached %s in %d hops\n",
			result.Target.Hostname,
			result.HopCount)))
	} else {
		sb.WriteString(yellow(fmt.Sprintf("Trace incomplete: did not reach %s after %d hops\n",
			result.Target.Hostname,
			result.HopCount)))
	}

	if result.Context != nil {
		sb.WriteString(fmt.Sprintf("Time: %v\n", formatDuration(result.Context.Duration)))
	}

	// 错误信息
	if result.Error != nil {
		sb.WriteString("\n" + red(fmt.Sprintf("Error: %v\n", result.Error)))
	}

	return sb.String()
}

// FormatTraceTable 格式化 Traceroute 结果为表格
func FormatTraceTable(result *types.TraceResult, noColor bool) string {
	var sb strings.Builder

	// 设置颜色函数
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()

	if noColor {
		color.NoColor = true
		green = fmt.Sprint
		red = fmt.Sprint
		yellow = fmt.Sprint
		cyan = fmt.Sprint
		bold = fmt.Sprint
		gray = fmt.Sprint
	}

	// 标题
	sb.WriteString(bold(fmt.Sprintf("TRACEROUTE %s (%s)\n\n", result.Target.Hostname, result.Target.IP)))

	// 表头
	header := fmt.Sprintf("%-4s %-40s %-12s %-12s %-12s %-8s",
		"HOP", "HOSTNAME (IP)", "PROBE 1", "PROBE 2", "PROBE 3", "AVG RTT")
	sb.WriteString(bold(header) + "\n")
	sb.WriteString(strings.Repeat("-", 95) + "\n")

	// 数据行
	for _, hop := range result.Hops {
		hostname := hop.Hostname
		if hostname == "" {
			hostname = gray("*")
		} else if hop.IsDestination {
			hostname = green(hostname)
		}

		// 格式化探测时间
		probeStrs := make([]string, 3)
		for i := 0; i < 3; i++ {
			if i < len(hop.Probes) {
				probe := hop.Probes[i]
				switch probe.Status {
				case types.StatusSuccess:
					probeStrs[i] = green(formatDuration(probe.RTT))
				case types.StatusTimeout:
					probeStrs[i] = gray("*")
				case types.StatusFailure:
					probeStrs[i] = red("!")
				default:
					probeStrs[i] = yellow("?")
				}
			} else {
				probeStrs[i] = "-"
			}
		}

		// 平均 RTT
		avgRTT := hop.GetAvgRTT()
		avgStr := "-"
		if avgRTT > 0 {
			avgStr = cyan(formatDuration(avgRTT))
		}

		row := fmt.Sprintf("%-4d %-40s %-12s %-12s %-12s %s",
			hop.TTL,
			hostname,
			probeStrs[0],
			probeStrs[1],
			probeStrs[2],
			avgStr)

		sb.WriteString(row + "\n")
	}

	// 统计信息
	sb.WriteString("\n" + strings.Repeat("-", 95) + "\n")
	sb.WriteString(bold("Summary:\n"))
	sb.WriteString(fmt.Sprintf("  Total Hops:     %d\n", result.HopCount))
	if result.ReachedDestination {
		sb.WriteString(fmt.Sprintf("  Destination:    %s\n", green("Reached")))
	} else {
		sb.WriteString(fmt.Sprintf("  Destination:    %s\n", yellow("Not Reached")))
	}
	if result.Context != nil {
		sb.WriteString(fmt.Sprintf("  Duration:       %v\n", formatDuration(result.Context.Duration)))
	}

	return sb.String()
}
