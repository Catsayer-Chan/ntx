package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/catsayer/ntx/internal/output/formatter"
	"github.com/catsayer/ntx/pkg/termutil"
	"github.com/catsayer/ntx/pkg/types"
)

func printConnectionsText(connections []*types.Connection, noColor bool) {
	if len(connections) == 0 {
		fmt.Println("无网络连接")
		return
	}

	printer := termutil.NewColorPrinter(noColor)
	bold := printer.Bold

	var table *formatter.Table
	if connProcess {
		table = formatter.NewTable(
			[]string{bold("Proto"), bold("Local Address"), bold("Remote Address"), bold("State"), bold("PID"), bold("Process")},
			[]int{8, 23, 23, 12, 8, 20},
		)
	} else {
		table = formatter.NewTable(
			[]string{bold("Proto"), bold("Local Address"), bold("Remote Address"), bold("State")},
			[]int{8, 23, 23, 12},
		)
	}

	for _, conn := range connections {
		localAddr := fmt.Sprintf("%s:%d", conn.LocalAddr, conn.LocalPort)
		remoteAddr := fmt.Sprintf("%s:%d", conn.RemoteAddr, conn.RemotePort)

		if connProcess {
			processInfo := "-"
			if conn.PID > 0 {
				processInfo = fmt.Sprintf("%d/%s", conn.PID, conn.ProcessName)
			}
			table.AddRow(
				conn.Protocol,
				localAddr,
				remoteAddr,
				string(conn.State),
				fmt.Sprintf("%d", conn.PID),
				processInfo,
			)
		} else {
			table.AddRow(
				conn.Protocol,
				localAddr,
				remoteAddr,
				string(conn.State),
			)
		}
	}

	table.Render(os.Stdout)
	fmt.Printf("\nTotal: %d connections\n", len(connections))
}

func printListenersText(listeners []*types.Listener, noColor bool) {
	if len(listeners) == 0 {
		fmt.Println("无监听端口")
		return
	}

	printer := termutil.NewColorPrinter(noColor)
	bold := printer.Bold

	var table *formatter.Table
	if connProcess {
		table = formatter.NewTable(
			[]string{bold("Proto"), bold("Local Address"), bold("PID"), bold("Process")},
			[]int{8, 23, 8, 20},
		)
	} else {
		table = formatter.NewTable(
			[]string{bold("Proto"), bold("Local Address")},
			[]int{8, 23},
		)
	}

	for _, listener := range listeners {
		localAddr := fmt.Sprintf("%s:%d", listener.Addr, listener.Port)

		if connProcess {
			processInfo := "-"
			if listener.PID > 0 {
				processInfo = fmt.Sprintf("%d/%s", listener.PID, listener.ProcessName)
			}
			table.AddRow(
				listener.Protocol,
				localAddr,
				fmt.Sprintf("%d", listener.PID),
				processInfo,
			)
		} else {
			table.AddRow(
				listener.Protocol,
				localAddr,
			)
		}
	}

	table.Render(os.Stdout)
	fmt.Printf("\nTotal: %d listeners\n", len(listeners))
}

func printStatsText(stats *types.NetStatistics, noColor bool) {
	printer := termutil.NewColorPrinter(noColor)
	bold := printer.Bold
	green := printer.Success

	fmt.Println(bold("Network Connection Statistics"))
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("TCP Connections:\n")
	fmt.Printf("  ESTABLISHED: %s\n", green(stats.TCPEstablished))
	fmt.Printf("  LISTEN:      %s\n", green(stats.TCPListen))
	fmt.Printf("  TIME_WAIT:   %s\n", green(stats.TCPTimeWait))
	fmt.Printf("  CLOSE_WAIT:  %s\n", green(stats.TCPCloseWait))
	fmt.Printf("  Total:       %s\n", green(stats.TCPTotal))
	fmt.Println()
	fmt.Printf("UDP Connections: %s\n", green(stats.UDPTotal))
	fmt.Println()
	fmt.Printf("Total Connections: %s\n", bold(green(stats.TotalConnections)))
}
