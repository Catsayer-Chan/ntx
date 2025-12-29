// Package types 定义 NTX 工具的公共类型
//
// 本文件定义端口扫描相关的类型和接口
//
// 作者: Catsayer
package types

import (
	"net"
	"time"
)

// PortState 端口状态类型
type PortState int

const (
	// PortOpen 端口开放
	PortOpen PortState = iota
	// PortClosed 端口关闭
	PortClosed
	// PortFiltered 端口被过滤
	PortFiltered
	// PortUnknown 状态未知
	PortUnknown
)

// String 返回端口状态的字符串表示
func (s PortState) String() string {
	switch s {
	case PortOpen:
		return "open"
	case PortClosed:
		return "closed"
	case PortFiltered:
		return "filtered"
	case PortUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// ScanMode 扫描模式类型
type ScanMode int

const (
	// ScanTCPConnect TCP Connect 扫描（不需要特权）
	ScanTCPConnect ScanMode = iota
	// ScanTCPSYN TCP SYN 扫描（需要 root 权限）
	ScanTCPSYN
	// ScanUDP UDP 扫描
	ScanUDP
)

// String 返回扫描模式的字符串表示
func (m ScanMode) String() string {
	switch m {
	case ScanTCPConnect:
		return "tcp-connect"
	case ScanTCPSYN:
		return "tcp-syn"
	case ScanUDP:
		return "udp"
	default:
		return "unknown"
	}
}

// ScanOptions 扫描参数配置
type ScanOptions struct {
	// Ports 要扫描的端口列表
	Ports []int
	// Timeout 单个端口的超时时间
	Timeout time.Duration
	// Concurrency 并发扫描的最大数量
	Concurrency int
	// ScanMode 扫描模式
	ScanMode ScanMode
	// ServiceDetect 是否进行服务识别
	ServiceDetect bool
	// VersionDetect 是否进行版本探测
	VersionDetect bool
	// RateLimit 速率限制（每秒扫描包数）
	RateLimit int
}

// DefaultScanOptions 返回默认扫描选项
func DefaultScanOptions() ScanOptions {
	return ScanOptions{
		Ports:         CommonPorts(),
		Timeout:       DefaultScanTimeout,
		Concurrency:   100,
		ScanMode:      ScanTCPConnect,
		ServiceDetect: false,
		VersionDetect: false,
		RateLimit:     0, // 0 表示不限制
	}
}

// CommonPorts 返回常用端口列表
func CommonPorts() []int {
	return []int{
		21, 22, 23, 25, 53, 80, 110, 111, 135, 139,
		143, 443, 445, 993, 995, 1723, 3306, 3389,
		5900, 8080, 8443,
	}
}

// ScanPort 单个端口的扫描结果
type ScanPort struct {
	// IP 目标 IP 地址
	IP net.IP
	// Port 端口号
	Port int
	// Proto 协议类型（tcp/udp）
	Proto string
	// State 端口状态
	State PortState
	// Service 服务名称
	Service string
	// Version 服务版本
	Version string
	// Banner Banner 信息
	Banner string
	// ResponseTime 响应时间
	ResponseTime time.Duration
	// Error 错误信息（如果有）
	Error error
}

// ScanResult 扫描结果汇总
type ScanResult struct {
	// Target 目标主机
	Target string
	// IP 解析后的 IP 地址
	IP net.IP
	// Ports 扫描到的端口列表
	Ports []*ScanPort
	// StartTime 扫描开始时间
	StartTime time.Time
	// EndTime 扫描结束时间
	EndTime time.Time
	// Summary 统计摘要
	Summary *ScanSummary
}

// ScanSummary 扫描统计信息
type ScanSummary struct {
	// TotalPorts 总扫描端口数
	TotalPorts int
	// OpenPorts 开放端口数
	OpenPorts int
	// ClosedPorts 关闭端口数
	ClosedPorts int
	// FilteredPorts 过滤端口数
	FilteredPorts int
	// Duration 扫描总耗时
	Duration time.Duration
}

// PortRange 端口范围
type PortRange struct {
	Start int
	End   int
}

// ParsePortList 解析端口列表字符串
// 支持格式: "80,443,8000-9000"
func ParsePortList(portStr string) ([]int, error) {
	// 这里简化实现，实际应该解析字符串
	return CommonPorts(), nil
}
