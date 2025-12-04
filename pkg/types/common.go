// Package types 提供 NTX 的通用类型定义
//
// 本模块定义了项目中使用的所有公共类型，包括：
// - 基础数据结构
// - 配置选项
// - 结果类型
// - 接口定义
//
// 作者: Catsayer
package types

import (
	"time"
)

// Protocol 网络协议类型
type Protocol string

const (
	// ProtocolICMP ICMP 协议
	ProtocolICMP Protocol = "icmp"
	// ProtocolTCP TCP 协议
	ProtocolTCP Protocol = "tcp"
	// ProtocolUDP UDP 协议
	ProtocolUDP Protocol = "udp"
	// ProtocolHTTP HTTP 协议
	ProtocolHTTP Protocol = "http"
	// ProtocolHTTPS HTTPS 协议
	ProtocolHTTPS Protocol = "https"
)

// OutputFormat 输出格式类型
type OutputFormat string

const (
	// OutputText 文本格式
	OutputText OutputFormat = "text"
	// OutputJSON JSON 格式
	OutputJSON OutputFormat = "json"
	// OutputYAML YAML 格式
	OutputYAML OutputFormat = "yaml"
	// OutputTable 表格格式
	OutputTable OutputFormat = "table"
)

// Status 状态类型
type Status string

const (
	// StatusSuccess 成功
	StatusSuccess Status = "success"
	// StatusFailure 失败
	StatusFailure Status = "failure"
	// StatusTimeout 超时
	StatusTimeout Status = "timeout"
	// StatusUnknown 未知
	StatusUnknown Status = "unknown"
)

// IPVersion IP 版本
type IPVersion int

const (
	// IPv4 IPv4
	IPv4 IPVersion = 4
	// IPv6 IPv6
	IPv6 IPVersion = 6
	// IPvAny 自动选择
	IPvAny IPVersion = 0
)

// Statistics 统计信息
type Statistics struct {
	// Sent 发送的数据包数量
	Sent int `json:"sent" yaml:"sent"`
	// Received 接收的数据包数量
	Received int `json:"received" yaml:"received"`
	// Loss 丢包数量
	Loss int `json:"loss" yaml:"loss"`
	// LossRate 丢包率 (0-100)
	LossRate float64 `json:"loss_rate" yaml:"loss_rate"`
	// MinRTT 最小往返时间
	MinRTT time.Duration `json:"min_rtt" yaml:"min_rtt"`
	// MaxRTT 最大往返时间
	MaxRTT time.Duration `json:"max_rtt" yaml:"max_rtt"`
	// AvgRTT 平均往返时间
	AvgRTT time.Duration `json:"avg_rtt" yaml:"avg_rtt"`
	// StdDevRTT RTT 标准差
	StdDevRTT time.Duration `json:"stddev_rtt" yaml:"stddev_rtt"`
	// TotalTime 总耗时
	TotalTime time.Duration `json:"total_time" yaml:"total_time"`
}

// Host 主机信息
type Host struct {
	// Hostname 主机名
	Hostname string `json:"hostname" yaml:"hostname"`
	// IP IP 地址
	IP string `json:"ip" yaml:"ip"`
	// IPVersion IP 版本
	IPVersion IPVersion `json:"ip_version" yaml:"ip_version"`
	// Port 端口号（如果适用）
	Port int `json:"port,omitempty" yaml:"port,omitempty"`
}

// ExecutionContext 执行上下文
type ExecutionContext struct {
	// StartTime 开始时间
	StartTime time.Time `json:"start_time" yaml:"start_time"`
	// EndTime 结束时间
	EndTime time.Time `json:"end_time" yaml:"end_time"`
	// Duration 持续时间
	Duration time.Duration `json:"duration" yaml:"duration"`
	// User 执行用户
	User string `json:"user,omitempty" yaml:"user,omitempty"`
	// Hostname 执行主机
	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	// CommandLine 命令行
	CommandLine string `json:"command_line,omitempty" yaml:"command_line,omitempty"`
}

// Result 通用结果接口
type Result interface {
	// GetStatus 获取状态
	GetStatus() Status
	// GetError 获取错误信息
	GetError() error
	// GetStatistics 获取统计信息
	GetStatistics() *Statistics
}