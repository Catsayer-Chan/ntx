// Package types 提供 Traceroute 相关类型定义
//
// 作者: Catsayer
package types

import (
	"time"
)

// TraceOptions Traceroute 配置选项
type TraceOptions struct {
	// Protocol 协议类型 (icmp/udp/tcp)
	Protocol Protocol `json:"protocol" yaml:"protocol"`
	// MaxHops 最大跳数
	MaxHops int `json:"max_hops" yaml:"max_hops"`
	// Timeout 每跳超时时间
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
	// Queries 每跳查询次数
	Queries int `json:"queries" yaml:"queries"`
	// Port 起始端口（UDP/TCP）
	Port int `json:"port,omitempty" yaml:"port,omitempty"`
	// PacketSize 数据包大小
	PacketSize int `json:"packet_size" yaml:"packet_size"`
	// IPVersion IP 版本
	IPVersion IPVersion `json:"ip_version" yaml:"ip_version"`
	// FirstTTL 起始 TTL
	FirstTTL int `json:"first_ttl" yaml:"first_ttl"`
	// DontFragment 不分片标志
	DontFragment bool `json:"dont_fragment" yaml:"dont_fragment"`
}

// DefaultTraceOptions 返回默认 Traceroute 选项
func DefaultTraceOptions() *TraceOptions {
	return &TraceOptions{
		Protocol:     ProtocolICMP,
		MaxHops:      30,
		Timeout:      DefaultTraceTimeout,
		Queries:      3,
		Port:         DefaultTraceroutePort,
		PacketSize:   60,
		IPVersion:    IPvAny,
		FirstTTL:     1,
		DontFragment: false,
	}
}

// TraceHop 单跳信息
type TraceHop struct {
	// TTL Time To Live
	TTL int `json:"ttl" yaml:"ttl"`
	// Probes 探测结果列表
	Probes []*TraceProbe `json:"probes" yaml:"probes"`
	// IP 响应的 IP 地址
	IP string `json:"ip,omitempty" yaml:"ip,omitempty"`
	// Hostname 响应的主机名
	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	// IsDestination 是否为目标主机
	IsDestination bool `json:"is_destination" yaml:"is_destination"`
}

// TraceProbe 单次探测结果
type TraceProbe struct {
	// Seq 序列号
	Seq int `json:"seq" yaml:"seq"`
	// IP 响应 IP
	IP string `json:"ip,omitempty" yaml:"ip,omitempty"`
	// RTT 往返时间
	RTT time.Duration `json:"rtt,omitempty" yaml:"rtt,omitempty"`
	// Status 状态
	Status Status `json:"status" yaml:"status"`
	// Error 错误信息
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
}

// GetBestProbe 获取最佳探测结果（最快响应）
func (h *TraceHop) GetBestProbe() *TraceProbe {
	var best *TraceProbe
	for _, probe := range h.Probes {
		if probe.Status == StatusSuccess {
			if best == nil || probe.RTT < best.RTT {
				best = probe
			}
		}
	}
	return best
}

// GetAvgRTT 获取平均 RTT
func (h *TraceHop) GetAvgRTT() time.Duration {
	var total time.Duration
	count := 0
	for _, probe := range h.Probes {
		if probe.Status == StatusSuccess {
			total += probe.RTT
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / time.Duration(count)
}

// GetSuccessCount 获取成功探测数
func (h *TraceHop) GetSuccessCount() int {
	count := 0
	for _, probe := range h.Probes {
		if probe.Status == StatusSuccess {
			count++
		}
	}
	return count
}

// TraceResult Traceroute 结果
type TraceResult struct {
	// Target 目标主机信息
	Target *Host `json:"target" yaml:"target"`
	// Protocol 使用的协议
	Protocol Protocol `json:"protocol" yaml:"protocol"`
	// Hops 所有跳信息
	Hops []*TraceHop `json:"hops" yaml:"hops"`
	// ReachedDestination 是否到达目标
	ReachedDestination bool `json:"reached_destination" yaml:"reached_destination"`
	// HopCount 跳数
	HopCount int `json:"hop_count" yaml:"hop_count"`
	// Context 执行上下文
	Context *ExecutionContext `json:"context" yaml:"context"`
	// Status 总体状态
	Status Status `json:"status" yaml:"status"`
	// Error 错误信息
	Error error `json:"error,omitempty" yaml:"error,omitempty"`
}

// GetStatus 实现 Result 接口
func (r *TraceResult) GetStatus() Status {
	return r.Status
}

// GetError 实现 Result 接口
func (r *TraceResult) GetError() error {
	return r.Error
}

// GetStatistics 实现 Result 接口
func (r *TraceResult) GetStatistics() *Statistics {
	// Traceroute 不提供统计信息
	return nil
}

// AddHop 添加跳信息
func (r *TraceResult) AddHop(hop *TraceHop) {
	r.Hops = append(r.Hops, hop)
	r.HopCount = len(r.Hops)
}

// GetLastHop 获取最后一跳
func (r *TraceResult) GetLastHop() *TraceHop {
	if len(r.Hops) == 0 {
		return nil
	}
	return r.Hops[len(r.Hops)-1]
}

// Tracer Traceroute 执行器接口
type Tracer interface {
	// Trace 执行 Traceroute 操作
	Trace(target string, opts *TraceOptions) (*TraceResult, error)
	// Close 关闭资源
	Close() error
}
