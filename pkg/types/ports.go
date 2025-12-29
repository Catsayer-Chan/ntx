// Package types 提供端口相关的默认值与辅助方法
package types

import "strings"

const (
	// DefaultHTTPPort HTTP 默认端口
	DefaultHTTPPort = 80
	// DefaultHTTPSPort HTTPS 默认端口
	DefaultHTTPSPort = 443
	// DefaultTCPPort TCP 默认端口（用于普通端口可达性测试）
	DefaultTCPPort = 80
	// DefaultDNSPort DNS 默认端口
	DefaultDNSPort = 53
	// DefaultTraceroutePort Traceroute 默认起始端口
	DefaultTraceroutePort = 33434
	// MinPort 最小有效端口
	MinPort = 1
	// MaxPort 最大有效端口
	MaxPort = 65535
)

// GetDefaultPort 根据协议和目标智能返回默认端口
func GetDefaultPort(protocol Protocol, target string) int {
	switch protocol {
	case ProtocolTCP:
		return DefaultTCPPort
	case ProtocolHTTP, ProtocolHTTPS:
		if strings.HasPrefix(target, "https://") || protocol == ProtocolHTTPS {
			return DefaultHTTPSPort
		}
		return DefaultHTTPPort
	default:
		return 0
	}
}

// EnsurePort 设置默认端口（如果尚未指定）
func (opts *PingOptions) EnsurePort(target string) {
	if opts == nil || opts.Port != 0 {
		return
	}
	opts.Port = GetDefaultPort(opts.Protocol, target)
}
