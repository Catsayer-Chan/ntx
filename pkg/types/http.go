// Package types 提供 HTTP 相关类型定义
//
// 作者: Catsayer
package types

import (
	"time"
)

// HTTPOptions HTTP 请求选项
type HTTPOptions struct {
	// Timeout 请求超时时间
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// FollowRedirect 是否跟随重定向
	FollowRedirect bool `json:"follow_redirect" yaml:"follow_redirect"`

	// MaxRedirects 最大重定向次数
	MaxRedirects int `json:"max_redirects" yaml:"max_redirects"`

	// InsecureSkipVerify 是否跳过 TLS 验证
	InsecureSkipVerify bool `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`

	// UserAgent 默认 User-Agent
	UserAgent string `json:"user_agent" yaml:"user_agent"`
}

// HTTPResult HTTP 请求结果
type HTTPResult struct {
	// Method 请求方法
	Method string `json:"method" yaml:"method"`

	// URL 请求 URL
	URL string `json:"url" yaml:"url"`

	// StatusCode HTTP 状态码
	StatusCode int `json:"status_code" yaml:"status_code"`

	// Status HTTP 状态描述
	Status string `json:"status" yaml:"status"`

	// Proto HTTP 协议版本 (如 "HTTP/1.1", "HTTP/2.0")
	Proto string `json:"proto" yaml:"proto"`

	// ContentLength 响应内容长度
	ContentLength int64 `json:"content_length" yaml:"content_length"`

	// Headers 响应头
	Headers map[string][]string `json:"headers,omitempty" yaml:"headers,omitempty"`

	// Body 响应体
	Body []byte `json:"body,omitempty" yaml:"body,omitempty"`

	// StartTime 请求开始时间
	StartTime time.Time `json:"start_time" yaml:"start_time"`

	// EndTime 请求结束时间
	EndTime time.Time `json:"end_time" yaml:"end_time"`

	// Duration 请求耗时
	Duration time.Duration `json:"duration" yaml:"duration"`

	// TLSUsed 是否使用 TLS
	TLSUsed bool `json:"tls_used" yaml:"tls_used"`

	// Uncompressed 是否解压缩
	Uncompressed bool `json:"uncompressed" yaml:"uncompressed"`

	// ContentType 内容类型
	ContentType string `json:"content_type" yaml:"content_type"`

	// TransferredSize 实际传输大小
	TransferredSize int64 `json:"transferred_size" yaml:"transferred_size"`

	// Error 错误信息
	Error error `json:"error,omitempty" yaml:"error,omitempty"`
}

// HTTPBenchmarkResult HTTP 性能测试结果
type HTTPBenchmarkResult struct {
	// Method 请求方法
	Method string `json:"method" yaml:"method"`

	// URL 请求 URL
	URL string `json:"url" yaml:"url"`

	// TotalRequests 总请求数
	TotalRequests int `json:"total_requests" yaml:"total_requests"`

	// SuccessCount 成功请求数
	SuccessCount int `json:"success_count" yaml:"success_count"`

	// FailureCount 失败请求数
	FailureCount int `json:"failure_count" yaml:"failure_count"`

	// TotalDuration 总耗时
	TotalDuration time.Duration `json:"total_duration" yaml:"total_duration"`

	// MinDuration 最小耗时
	MinDuration time.Duration `json:"min_duration" yaml:"min_duration"`

	// MaxDuration 最大耗时
	MaxDuration time.Duration `json:"max_duration" yaml:"max_duration"`

	// AvgDuration 平均耗时
	AvgDuration time.Duration `json:"avg_duration" yaml:"avg_duration"`

	// RequestsPerSec 每秒请求数
	RequestsPerSec float64 `json:"requests_per_sec" yaml:"requests_per_sec"`
}
