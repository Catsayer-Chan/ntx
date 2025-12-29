//go:build linux

// Package types 定义 Linux 平台相关常量
package types

const (
	// ProcNetTCP Linux TCP IPv4 表路径
	ProcNetTCP = "/proc/net/tcp"
	// ProcNetTCP6 Linux TCP IPv6 表路径
	ProcNetTCP6 = "/proc/net/tcp6"
	// ProcNetUDP Linux UDP IPv4 表路径
	ProcNetUDP = "/proc/net/udp"
	// ProcNetUDP6 Linux UDP IPv6 表路径
	ProcNetUDP6 = "/proc/net/udp6"
	// ProcNetDev Linux 网卡统计路径
	ProcNetDev = "/proc/net/dev"
	// ProcNetRoute Linux 路由表路径
	ProcNetRoute = "/proc/net/route"
)
