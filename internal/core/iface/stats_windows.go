//go:build windows
// +build windows

package iface

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/catsayer/ntx/pkg/types"
	"golang.org/x/sys/windows"
)

// getInterfaceStats 获取网卡统计信息 (Windows)
func (r *InterfaceReader) getInterfaceStats(name string) (*types.InterfaceStats, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, fmt.Errorf("获取网卡 %s 失败: %w", name, err)
	}

	row := windows.MibIfRow2{
		InterfaceIndex: uint32(iface.Index),
	}
	if err := windows.GetIfEntry2Ex(windows.MibIfEntryNormal, &row); err == nil {
		return convertMibIfRow2(&row), nil
	}

	legacy := windows.MibIfRow{
		Index: uint32(iface.Index),
	}
	if err := windows.GetIfEntry(&legacy); err != nil {
		return nil, fmt.Errorf("GetIfEntry 失败: %w", err)
	}
	return convertMibIfRow(&legacy), nil
}

// getRoutesImpl 获取路由表信息 (Windows)
func (r *InterfaceReader) getRoutesImpl() ([]*types.Route, error) {
	cmd := exec.Command("route", "PRINT", "-4")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("执行 route PRINT 失败: %w (%s)", err, strings.TrimSpace(string(output)))
	}

	return parseWindowsRoutes(string(output)), nil
}

func convertMibIfRow2(row *windows.MibIfRow2) *types.InterfaceStats {
	return &types.InterfaceStats{
		RxBytes:   row.InOctets,
		RxPackets: row.InUcastPkts + row.InNUcastPkts,
		RxErrors:  row.InErrors,
		RxDropped: row.InDiscards,
		TxBytes:   row.OutOctets,
		TxPackets: row.OutUcastPkts + row.OutNUcastPkts,
		TxErrors:  row.OutErrors,
		TxDropped: row.OutDiscards,
	}
}

func convertMibIfRow(row *windows.MibIfRow) *types.InterfaceStats {
	return &types.InterfaceStats{
		RxBytes:   uint64(row.InOctets),
		RxPackets: uint64(row.InUcastPkts + row.InNUcastPkts),
		RxErrors:  uint64(row.InErrors),
		RxDropped: uint64(row.InDiscards),
		TxBytes:   uint64(row.OutOctets),
		TxPackets: uint64(row.OutUcastPkts + row.OutNUcastPkts),
		TxErrors:  uint64(row.OutErrors),
		TxDropped: uint64(row.OutDiscards),
	}
}

// parseWindowsRoutes 解析 route print 输出
func parseWindowsRoutes(output string) []*types.Route {
	scanner := bufio.NewScanner(strings.NewReader(output))
	routes := make([]*types.Route, 0)

	inIPv4Table := false
	inActive := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			if inActive && len(routes) > 0 {
				break
			}
			continue
		}

		switch {
		case strings.HasPrefix(line, "IPv4 Route Table"):
			inIPv4Table = true
			inActive = false
			continue
		case strings.HasPrefix(line, "IPv6 Route Table"):
			inIPv4Table = false
			continue
		case strings.HasPrefix(line, "Active Routes:"):
			inActive = true
			continue
		case strings.HasPrefix(line, "Persistent Routes"):
			break
		}

		if !inIPv4Table || !inActive {
			continue
		}

		if strings.HasPrefix(line, "Network Destination") {
			// 表头
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		dest := fmt.Sprintf("%s/%s", fields[0], fields[1])
		route := &types.Route{
			Destination: dest,
			Gateway:     fields[2],
			Interface:   fields[3],
			Metric:      fields[4],
			Flags:       []string{"U"},
		}
		routes = append(routes, route)
	}

	return routes
}
