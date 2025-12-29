// Package types 提供 Ping 相关类型定义
//
// 作者: Catsayer
package types

import (
	"context"
	"time"

	"github.com/catsayer/ntx/pkg/stats"
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

	// TOS 服务类型 (Type of Service)

	TOS int `json:"tos,omitempty" yaml:"tos,omitempty"`

	// HTTPMethod HTTP 方法（HTTP Ping）

	HTTPMethod string `json:"http_method,omitempty" yaml:"http_method,omitempty"`

	// HTTPPath HTTP 路径（HTTP Ping）

	HTTPPath string `json:"http_path,omitempty" yaml:"http_path,omitempty"`
}

// DefaultPingOptions 返回默认 Ping 选项

func DefaultPingOptions() *PingOptions {

	return &PingOptions{

		Protocol: ProtocolICMP,

		Count: 4,

		Interval: time.Second,

		Timeout: DefaultPingTimeout,

		Size: 64,

		TTL: 64,

		IPVersion: IPvAny,

		DontFragment: false,

		HTTPMethod: "GET",

		HTTPPath: "/",
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

	statsData := r.Statistics
	statsData.Sent = len(r.Replies)
	statsData.Received = 0

	rtts := make([]time.Duration, 0, len(r.Replies))
	for _, reply := range r.Replies {
		if reply.Status == StatusSuccess {
			statsData.Received++
			rtts = append(rtts, reply.RTT)
		}
	}

	statsData.Loss = statsData.Sent - statsData.Received
	if statsData.Sent > 0 {
		statsData.LossRate = float64(statsData.Loss) / float64(statsData.Sent) * 100
	}

	if len(rtts) > 0 {
		min, max, avg, stddev := stats.ComputeRTTStats(rtts)
		statsData.MinRTT = min
		statsData.MaxRTT = max
		statsData.AvgRTT = avg
		statsData.StdDevRTT = stddev
	} else {
		statsData.MinRTT = 0
		statsData.MaxRTT = 0
		statsData.AvgRTT = 0
		statsData.StdDevRTT = 0
	}
}

// Pinger Ping 执行器接口

type Pinger interface {

	// Ping 执行 Ping 操作 (用于批量获取结果)

	Ping(ctx context.Context, target string, opts *PingOptions) (*PingResult, error)

	// PingStream 执行实时 Ping 操作 (用于流式获取结果)

	PingStream(ctx context.Context, target string, opts *PingOptions) (<-chan *PingReply, error)

	// Close 关闭资源

	Close() error
}
