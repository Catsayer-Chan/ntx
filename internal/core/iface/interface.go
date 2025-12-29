// Package iface 提供网卡信息查询功能
//
// 本模块实现了网卡信息查询功能,包括:
// - 网卡列表
// - IP 地址（IPv4/IPv6）
// - MAC 地址
// - 网卡状态和标志
// - MTU、速率等详细信息
// - 网络统计信息
//
// 依赖:
// - net: Go 标准库网络接口
//
// 使用示例:
//
//	reader := iface.NewInterfaceReader()
//	interfaces, err := reader.GetInterfaces()
//
// 作者: Catsayer
package iface

import (
	"fmt"
	"net"

	"github.com/catsayer/ntx/pkg/types"
)

// InterfaceReader 网卡信息读取器
type InterfaceReader struct{}

// NewInterfaceReader 创建新的网卡信息读取器
func NewInterfaceReader() *InterfaceReader {
	return &InterfaceReader{}
}

// GetInterfaces 获取所有网卡信息
func (r *InterfaceReader) GetInterfaces() ([]*types.Interface, error) {
	// 获取系统所有网卡
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("获取网卡列表失败: %w", err)
	}

	interfaces := make([]*types.Interface, 0, len(netInterfaces))

	for _, netIface := range netInterfaces {
		iface, err := r.parseInterface(&netIface)
		if err != nil {
			// 跳过解析失败的网卡
			continue
		}
		interfaces = append(interfaces, iface)
	}

	return interfaces, nil
}

// GetInterface 获取指定网卡的信息
func (r *InterfaceReader) GetInterface(name string) (*types.Interface, error) {
	netIface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, fmt.Errorf("获取网卡 %s 失败: %w", name, err)
	}

	return r.parseInterface(netIface)
}

// parseInterface 解析网卡信息
func (r *InterfaceReader) parseInterface(netIface *net.Interface) (*types.Interface, error) {
	// 获取网卡地址列表
	addrs, err := netIface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("获取网卡地址失败: %w", err)
	}

	// 解析标志
	flags := r.parseFlags(netIface.Flags)

	// 分离 IPv4 和 IPv6 地址
	var ipv4Addrs, ipv6Addrs []string
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}

		if ipv4 := ipNet.IP.To4(); ipv4 != nil {
			ipv4Addrs = append(ipv4Addrs, ipNet.String())
		} else if ipv6 := ipNet.IP.To16(); ipv6 != nil {
			ipv6Addrs = append(ipv6Addrs, ipNet.String())
		}
	}

	iface := &types.Interface{
		Name:         netIface.Name,
		Index:        netIface.Index,
		MTU:          netIface.MTU,
		Flags:        flags,
		HardwareAddr: netIface.HardwareAddr.String(),
		IPv4Addrs:    ipv4Addrs,
		IPv6Addrs:    ipv6Addrs,
	}

	// 尝试获取网卡统计信息（依赖平台）
	stats, err := r.getInterfaceStats(netIface.Name)
	if err == nil {
		iface.Stats = stats
	}

	return iface, nil
}

// parseFlags 解析网卡标志
func (r *InterfaceReader) parseFlags(flags net.Flags) []string {
	var flagStrs []string

	if flags&net.FlagUp != 0 {
		flagStrs = append(flagStrs, "UP")
	}
	if flags&net.FlagBroadcast != 0 {
		flagStrs = append(flagStrs, "BROADCAST")
	}
	if flags&net.FlagLoopback != 0 {
		flagStrs = append(flagStrs, "LOOPBACK")
	}
	if flags&net.FlagPointToPoint != 0 {
		flagStrs = append(flagStrs, "POINTTOPOINT")
	}
	if flags&net.FlagMulticast != 0 {
		flagStrs = append(flagStrs, "MULTICAST")
	}
	if flags&net.FlagRunning != 0 {
		flagStrs = append(flagStrs, "RUNNING")
	}

	return flagStrs
}

// GetRoutes 获取路由表信息
// 注意: 此功能需要平台特定实现
func (r *InterfaceReader) GetRoutes() ([]*types.Route, error) {
	return r.getRoutesImpl()
}

// Close 关闭读取器 (当前无需实际操作)
func (r *InterfaceReader) Close() error {
	return nil
}
