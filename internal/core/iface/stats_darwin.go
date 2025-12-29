//go:build darwin
// +build darwin

package iface

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"unsafe"

	"github.com/catsayer/ntx/pkg/types"
	"golang.org/x/sys/unix"
)

// getInterfaceStats 获取网卡统计信息 (macOS)
func (r *InterfaceReader) getInterfaceStats(name string) (*types.InterfaceStats, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, fmt.Errorf("获取网卡 %s 失败: %w", name, err)
	}

	// 读取 iflist2，包含每个接口的 if_msghdr2 数据
	data, err := unix.SysctlRaw("net.route.0.0.iflist2")
	if err != nil {
		return nil, fmt.Errorf("读取 macOS 接口统计失败: %w", err)
	}

	for len(data) > 0 {
		if len(data) < unix.SizeofIfMsghdr2 {
			break
		}

		ifm := (*unix.IfMsghdr2)(unsafe.Pointer(&data[0]))
		msgLen := int(ifm.Msglen)
		if msgLen <= 0 || msgLen > len(data) {
			break
		}

		if int(ifm.Index) == iface.Index {
			stats := &types.InterfaceStats{
				RxBytes:   ifm.Data.Ibytes,
				RxPackets: ifm.Data.Ipackets,
				RxErrors:  ifm.Data.Ierrors,
				RxDropped: ifm.Data.Iqdrops,
				TxBytes:   ifm.Data.Obytes,
				TxPackets: ifm.Data.Opackets,
				TxErrors:  ifm.Data.Oerrors,
				// macOS 内核未提供明确的 TxDropped, 置零
				TxDropped: 0,
			}
			return stats, nil
		}

		data = data[msgLen:]
	}

	return nil, fmt.Errorf("未找到网卡 %s 的统计信息", name)
}

// getRoutesImpl 获取路由表信息 (macOS)
func (r *InterfaceReader) getRoutesImpl() ([]*types.Route, error) {
	cmd := exec.Command("netstat", "-rn")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("执行 netstat -rn 失败: %w", err)
	}

	return parseDarwinRoutes(string(output)), nil
}

// parseDarwinRoutes 解析 macOS netstat 输出
func parseDarwinRoutes(raw string) []*types.Route {
	scanner := bufio.NewScanner(strings.NewReader(raw))
	routes := make([]*types.Route, 0)

	inIPv4 := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "Internet6"):
			// 仅解析 IPv4 表
			inIPv4 = false
		case strings.HasPrefix(line, "Internet:"):
			inIPv4 = true
			continue
		case strings.HasPrefix(line, "Routing tables"):
			continue
		}

		if !inIPv4 {
			continue
		}

		if strings.HasPrefix(line, "Destination") {
			// 表头
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		flags := make([]string, 0, len(fields[2]))
		for _, ch := range fields[2] {
			flags = append(flags, string(ch))
		}

		metric := ""
		if len(fields) > 4 {
			metric = fields[len(fields)-1]
		}

		route := &types.Route{
			Destination: fields[0],
			Gateway:     fields[1],
			Flags:       flags,
			Interface:   fields[3],
			Metric:      metric,
		}
		routes = append(routes, route)
	}

	return routes
}
