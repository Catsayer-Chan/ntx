// Package types 提供 Ping 相关类型定义
//
// 作者: Catsayer
package types

import (
	"math"
	"time"
)

// PingOptions Ping 配置选项
type PingOptions struct {
	// Protocol 协议类型 (icmp/tcp/http)
	Protocol Protocol `json:"protocol" yaml:"protocol"`
	// Count 发送次数，0 表示无限次
	Count int `json:"count" yaml:"count"`
	// Interval 发送间隔
	Interval time.Duration `json:"interval" yaml:"interval"`
	// Timeout 超时时间
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
	// Size 数据包大小（字节）
	Size int `json:"size" yaml:"size"`
	// TTL Time To Live
	TTL int `json:"ttl" yaml:"ttl"`
	// Port 端口号（TCP/HTTP）
	Port int `json:"port,omitempty" yaml:"port,omitempty"`
	// IPVersion IP 版本
	IPVersion IPVersion `json:"ip_version" yaml:"ip_version"`
	// Source 源地址
	Source string `json:"source,omitempty" yaml:"source,omitempty"`
	// DontFragment 不分片标志
	DontFragment bool `json:"dont_fragment" yaml:"dont_fragment"`
	// HTTPMethod HTTP 方法（HTTP Ping）
	HTTPMethod string `json:"http_method,omitempty" yaml:"http_method,omitempty"`
	// HTTPPath HTTP 路径（HTTP Ping）
	HTTPPath string `json:"http_path,omitempty" yaml:"http_path,omitempty"`
}

// DefaultPingOptions 返回默认 Ping 选项
func DefaultPingOptions() *PingOptions {
	return &PingOptions{
		Protocol:     ProtocolICMP,
		Count:        4,
		Interval:     time.Second,
		Timeout:      5 * time.Second,
		Size:         64,
		TTL:          64,
		IPVersion:    IPvAny,
		DontFragment: false,
		HTTPMethod:   "GET",
		HTTPPath:     "/",
	}
}

// PingReply Ping 响应
type PingReply struct {
	// Seq 序列号
	Seq int `json:"seq" yaml:"seq"`
	// From 响应来源
	From string `json:"from" yaml:"from"`
	// Bytes 响应字节数
	Bytes int `json:"bytes" yaml:"bytes"`
	// TTL Time To Live
	TTL int `json:"ttl" yaml:"ttl"`
	// RTT 往返时间
	RTT time.Duration `json:"rtt" yaml:"rtt"`
	// Time 时间戳
	Time time.Time `json:"time" yaml:"time"`
	// Status 状态
	Status Status `json:"status" yaml:"status"`
	// Error 错误信息
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
}

// PingResult Ping 结果
type PingResult struct {
	// Target 目标主机信息
	Target *Host `json:"target" yaml:"target"`
	// Protocol 使用的协议
	Protocol Protocol `json:"protocol" yaml:"protocol"`
	// Replies 所有响应
	Replies []*PingReply `json:"replies" yaml:"replies"`
	// Statistics 统计信息
	Statistics *Statistics `json:"statistics" yaml:"statistics"`
	// Context 执行上下文
	Context *ExecutionContext `json:"context" yaml:"context"`
	// Status 总体状态
	Status Status `json:"status" yaml:"status"`
	// Error 错误信息
	Error error `json:"error,omitempty" yaml:"error,omitempty"`
}

// GetStatus 实现 Result 接口
func (r *PingResult) GetStatus() Status {
	return r.Status
}

// GetError 实现 Result 接口
func (r *PingResult) GetError() error {
	return r.Error
}

// GetStatistics 实现 Result 接口
func (r *PingResult) GetStatistics() *Statistics {
	return r.Statistics
}

// AddReply 添加响应
func (r *PingResult) AddReply(reply *PingReply) {
	r.Replies = append(r.Replies, reply)
}

// UpdateStatistics 更新统计信息
func (r *PingResult) UpdateStatistics() {
	if r.Statistics == nil {
		r.Statistics = &Statistics{}
	}

	stats := r.Statistics
	stats.Sent = len(r.Replies)
	stats.Received = 0

	var totalRTT time.Duration
	stats.MinRTT = time.Duration(0)
	stats.MaxRTT = time.Duration(0)

	for _, reply := range r.Replies {
		if reply.Status == StatusSuccess {
			stats.Received++
			totalRTT += reply.RTT

			if stats.MinRTT == 0 || reply.RTT < stats.MinRTT {
				stats.MinRTT = reply.RTT
			}
			if reply.RTT > stats.MaxRTT {
				stats.MaxRTT = reply.RTT
			}
		}
	}

	stats.Loss = stats.Sent - stats.Received
	if stats.Sent > 0 {
		stats.LossRate = float64(stats.Loss) / float64(stats.Sent) * 100
		if stats.Received > 0 {
			stats.AvgRTT = totalRTT / time.Duration(stats.Received)
		}
	}

	// 计算标准差
	if stats.Received > 0 {
		var variance float64
		avgNano := float64(stats.AvgRTT.Nanoseconds())
		for _, reply := range r.Replies {
			if reply.Status == StatusSuccess {
				diff := float64(reply.RTT.Nanoseconds()) - avgNano
				variance += diff * diff
			}
		}
		variance /= float64(stats.Received)
		stddev := math.Sqrt(variance)
		stats.StdDevRTT = time.Duration(int64(stddev))
	}
}

// Pinger Ping 执行器接口
type Pinger interface {
	// Ping 执行 Ping 操作
	Ping(target string, opts *PingOptions) (*PingResult, error)
	// Close 关闭资源
	Close() error
}