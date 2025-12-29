// Package types 提供常用超时配置
package types

import "time"

const (
	// DefaultPingTimeout Ping 默认超时时间
	DefaultPingTimeout = 5 * time.Second
	// DefaultScanTimeout 扫描默认超时时间
	DefaultScanTimeout = 3 * time.Second
	// DefaultDNSTimeout DNS 查询默认超时时间
	DefaultDNSTimeout = 5 * time.Second
	// DefaultTraceTimeout Traceroute 默认超时时间
	DefaultTraceTimeout = 3 * time.Second
	// DefaultHTTPTimeout HTTP 客户端默认超时时间
	DefaultHTTPTimeout = 30 * time.Second

	// DiagnosticGatewayTimeout 本地网关检查超时时间
	DiagnosticGatewayTimeout = 2 * time.Second
	// DiagnosticTargetTimeout 目标可达性检查超时时间
	DiagnosticTargetTimeout = 3 * time.Second
)
