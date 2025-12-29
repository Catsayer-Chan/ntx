// Package types 提供 DNS 相关类型定义
//
// 作者: Catsayer
package types

import (
	"fmt"
	"strings"
	"time"
)

var (
	// GoogleDNSServer1 Google 公共 DNS (主)
	GoogleDNSServer1 = FormatDNSServer("8.8.8.8")
	// GoogleDNSServer2 Google 公共 DNS (备)
	GoogleDNSServer2 = FormatDNSServer("8.8.4.4")
	// CloudflareDNSServer Cloudflare 公共 DNS
	CloudflareDNSServer = FormatDNSServer("1.1.1.1")
	// AliDNSServer 阿里云 DNS (适用于中国大陆)
	AliDNSServer = FormatDNSServer("223.5.5.5")
	// DNSPodServer 腾讯 DNSPod (适用于中国大陆)
	DNSPodServer = FormatDNSServer("119.29.29.29")
	// DefaultDNSServer 默认 DNS 服务器
	DefaultDNSServer = GoogleDNSServer1
)

// DNSRecordType DNS 记录类型
type DNSRecordType uint16

const (
	DNSTypeA     DNSRecordType = 1   // IPv4 地址
	DNSTypeNS    DNSRecordType = 2   // 名称服务器
	DNSTypeCNAME DNSRecordType = 5   // 规范名称
	DNSTypeSOA   DNSRecordType = 6   // 授权起始
	DNSTypePTR   DNSRecordType = 12  // 指针记录
	DNSTypeMX    DNSRecordType = 15  // 邮件交换
	DNSTypeTXT   DNSRecordType = 16  // 文本记录
	DNSTypeAAAA  DNSRecordType = 28  // IPv6 地址
	DNSTypeSRV   DNSRecordType = 33  // 服务记录
	DNSTypeANY   DNSRecordType = 255 // 所有记录
)

// String 返回记录类型的字符串表示
func (t DNSRecordType) String() string {
	switch t {
	case DNSTypeA:
		return "A"
	case DNSTypeNS:
		return "NS"
	case DNSTypeCNAME:
		return "CNAME"
	case DNSTypeSOA:
		return "SOA"
	case DNSTypePTR:
		return "PTR"
	case DNSTypeMX:
		return "MX"
	case DNSTypeTXT:
		return "TXT"
	case DNSTypeAAAA:
		return "AAAA"
	case DNSTypeSRV:
		return "SRV"
	default:
		return "UNKNOWN"
	}
}

// FormatDNSServer 将裸地址转换为 host:port 形式
func FormatDNSServer(host string) string {
	if host == "" {
		return ""
	}
	if strings.Contains(host, ":") {
		return host
	}
	return fmt.Sprintf("%s:%d", host, DefaultDNSPort)
}

// DNSServerList 返回常用 DNS 服务器列表
func DNSServerList() []string {
	return []string{
		GoogleDNSServer1,
		GoogleDNSServer2,
		CloudflareDNSServer,
		AliDNSServer,
		DNSPodServer,
	}
}

// DNSOptions DNS 查询选项
type DNSOptions struct {
	// Server DNS 服务器地址 (使用 host 或 host:port，默认为 DefaultDNSServer)
	Server string `json:"server" yaml:"server"`

	// Timeout 查询超时时间
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// Recursive 是否递归查询
	Recursive bool `json:"recursive" yaml:"recursive"`
}

// DNSResult DNS 查询结果
type DNSResult struct {
	// Domain 查询的域名
	Domain string `json:"domain" yaml:"domain"`

	// RecordType 记录类型
	RecordType DNSRecordType `json:"record_type" yaml:"record_type"`

	// Server 使用的 DNS 服务器
	Server string `json:"server" yaml:"server"`

	// QueryTime 查询耗时
	QueryTime time.Duration `json:"query_time" yaml:"query_time"`

	// StartTime 查询开始时间
	StartTime time.Time `json:"start_time" yaml:"start_time"`

	// EndTime 查询结束时间
	EndTime time.Time `json:"end_time" yaml:"end_time"`

	// Records 答案记录
	Records []*DNSRecord `json:"records,omitempty" yaml:"records,omitempty"`

	// Authority 权威记录
	Authority []*DNSRecord `json:"authority,omitempty" yaml:"authority,omitempty"`

	// Additional 附加记录
	Additional []*DNSRecord `json:"additional,omitempty" yaml:"additional,omitempty"`

	// Error 错误信息
	Error error `json:"error,omitempty" yaml:"error,omitempty"`
}

// DNSRecord DNS 记录
type DNSRecord struct {
	// Name 记录名称
	Name string `json:"name" yaml:"name"`

	// Type 记录类型
	Type DNSRecordType `json:"type" yaml:"type"`

	// TTL 生存时间（秒）
	TTL int `json:"ttl" yaml:"ttl"`

	// Value 记录值
	Value string `json:"value" yaml:"value"`
}
