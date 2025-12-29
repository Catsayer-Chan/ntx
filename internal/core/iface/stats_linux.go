//go:build linux
// +build linux

// Package iface 提供网卡统计信息的 Linux 实现
package iface

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/catsayer/ntx/pkg/types"
)

// getInterfaceStats 获取网卡统计信息 (Linux)
func (r *InterfaceReader) getInterfaceStats(name string) (*types.InterfaceStats, error) {
	// 从 /proc/net/dev 读取统计信息
	file, err := os.Open(types.ProcNetDev)
	if err != nil {
		return nil, fmt.Errorf("打开 /proc/net/dev 失败: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// 跳过前两行（表头）
	scanner.Scan()
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 17 {
			continue
		}

		// 网卡名称在第一列，格式为 "eth0:"
		ifaceName := strings.TrimSuffix(parts[0], ":")
		if ifaceName != name {
			continue
		}

		// 解析统计信息
		stats := &types.InterfaceStats{}

		// RX: bytes packets errs drop fifo frame compressed multicast
		stats.RxBytes, _ = strconv.ParseUint(parts[1], 10, 64)
		stats.RxPackets, _ = strconv.ParseUint(parts[2], 10, 64)
		stats.RxErrors, _ = strconv.ParseUint(parts[3], 10, 64)
		stats.RxDropped, _ = strconv.ParseUint(parts[4], 10, 64)

		// TX: bytes packets errs drop fifo colls carrier compressed
		stats.TxBytes, _ = strconv.ParseUint(parts[9], 10, 64)
		stats.TxPackets, _ = strconv.ParseUint(parts[10], 10, 64)
		stats.TxErrors, _ = strconv.ParseUint(parts[11], 10, 64)
		stats.TxDropped, _ = strconv.ParseUint(parts[12], 10, 64)

		return stats, nil
	}

	return nil, fmt.Errorf("未找到网卡 %s 的统计信息", name)
}

// getRoutesImpl 获取路由表信息 (Linux)
func (r *InterfaceReader) getRoutesImpl() ([]*types.Route, error) {
	// 从 /proc/net/route 读取路由表
	file, err := os.Open(types.ProcNetRoute)
	if err != nil {
		return nil, fmt.Errorf("打开 /proc/net/route 失败: %w", err)
	}
	defer file.Close()

	var routes []*types.Route
	scanner := bufio.NewScanner(file)

	// 跳过表头
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 11 {
			continue
		}

		// 解析路由信息
		// 格式: Iface Destination Gateway Flags RefCnt Use Metric Mask MTU Window IRTT
		route := &types.Route{
			Interface:   parts[0],
			Destination: parseHexIP(parts[1]),
			Gateway:     parseHexIP(parts[2]),
			Flags:       parseRouteFlags(parts[3]),
			Metric:      parts[6],
		}

		routes = append(routes, route)
	}

	return routes, nil
}

// parseHexIP 解析十六进制 IP 地址
func parseHexIP(hexIP string) string {
	if len(hexIP) != 8 {
		return "0.0.0.0"
	}

	// Linux /proc/net/route 中 IP 以小端序存储
	var ip [4]byte
	for i := 0; i < 4; i++ {
		val, _ := strconv.ParseUint(hexIP[i*2:(i+1)*2], 16, 8)
		ip[3-i] = byte(val)
	}

	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

// parseRouteFlags 解析路由标志
func parseRouteFlags(hexFlags string) []string {
	flags := make([]string, 0)
	val, _ := strconv.ParseUint(hexFlags, 16, 32)

	if val&0x0001 != 0 {
		flags = append(flags, "UP")
	}
	if val&0x0002 != 0 {
		flags = append(flags, "GATEWAY")
	}
	if val&0x0004 != 0 {
		flags = append(flags, "HOST")
	}

	return flags
}
