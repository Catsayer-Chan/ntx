// Package types 提供网卡相关类型定义
//
// 作者: Catsayer
package types

// Interface 网卡信息
type Interface struct {
	// Name 网卡名称
	Name string `json:"name" yaml:"name"`

	// Index 网卡索引
	Index int `json:"index" yaml:"index"`

	// MTU 最大传输单元
	MTU int `json:"mtu" yaml:"mtu"`

	// Flags 网卡标志
	Flags []string `json:"flags" yaml:"flags"`

	// HardwareAddr MAC 地址
	HardwareAddr string `json:"hardware_addr" yaml:"hardware_addr"`

	// IPv4Addrs IPv4 地址列表 (CIDR 格式)
	IPv4Addrs []string `json:"ipv4_addrs,omitempty" yaml:"ipv4_addrs,omitempty"`

	// IPv6Addrs IPv6 地址列表 (CIDR 格式)
	IPv6Addrs []string `json:"ipv6_addrs,omitempty" yaml:"ipv6_addrs,omitempty"`

	// Stats 网卡统计信息
	Stats *InterfaceStats `json:"stats,omitempty" yaml:"stats,omitempty"`
}

// InterfaceStats 网卡流量统计
type InterfaceStats struct {
	// RxBytes 接收字节数
	RxBytes uint64 `json:"rx_bytes" yaml:"rx_bytes"`

	// RxPackets 接收包数
	RxPackets uint64 `json:"rx_packets" yaml:"rx_packets"`

	// RxErrors 接收错误数
	RxErrors uint64 `json:"rx_errors" yaml:"rx_errors"`

	// RxDropped 接收丢弃数
	RxDropped uint64 `json:"rx_dropped" yaml:"rx_dropped"`

	// TxBytes 发送字节数
	TxBytes uint64 `json:"tx_bytes" yaml:"tx_bytes"`

	// TxPackets 发送包数
	TxPackets uint64 `json:"tx_packets" yaml:"tx_packets"`

	// TxErrors 发送错误数
	TxErrors uint64 `json:"tx_errors" yaml:"tx_errors"`

	// TxDropped 发送丢弃数
	TxDropped uint64 `json:"tx_dropped" yaml:"tx_dropped"`
}

// Route 路由信息
type Route struct {
	// Destination 目标网络
	Destination string `json:"destination" yaml:"destination"`

	// Gateway 网关
	Gateway string `json:"gateway" yaml:"gateway"`

	// Interface 网卡名称
	Interface string `json:"interface" yaml:"interface"`

	// Flags 路由标志
	Flags []string `json:"flags" yaml:"flags"`

	// Metric 路由度量值
	Metric string `json:"metric" yaml:"metric"`
}
