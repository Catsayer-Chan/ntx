package config

import (
	"time"

	"github.com/catsayer/ntx/pkg/buildinfo"
	"github.com/catsayer/ntx/pkg/types"
)

// Config 定义应用的完整配置结构
type Config struct {
	Global GlobalConfig `yaml:"global" json:"global"`
	Ping   PingConfig   `yaml:"ping" json:"ping"`
	DNS    DNSConfig    `yaml:"dns" json:"dns"`
	HTTP   HTTPConfig   `yaml:"http" json:"http"`
	Scan   ScanConfig   `yaml:"scan" json:"scan"`
	Trace  TraceConfig  `yaml:"trace" json:"trace"`
}

// GlobalConfig 全局配置
type GlobalConfig struct {
	Verbose  bool   `yaml:"verbose" json:"verbose"`
	Output   string `yaml:"output" json:"output"`
	NoColor  bool   `yaml:"no_color" json:"no_color"`
	LogLevel string `yaml:"log_level" json:"log_level"`
	LogFile  string `yaml:"log_file" json:"log_file"`
}

// PingConfig Ping 相关配置
type PingConfig struct {
	Protocol  types.Protocol  `yaml:"protocol" json:"protocol"`
	Count     int             `yaml:"count" json:"count"`
	Interval  time.Duration   `yaml:"interval" json:"interval"`
	Timeout   time.Duration   `yaml:"timeout" json:"timeout"`
	Size      int             `yaml:"size" json:"size"`
	TTL       int             `yaml:"ttl" json:"ttl"`
	Port      int             `yaml:"port" json:"port"`
	IPVersion types.IPVersion `yaml:"ip_version" json:"ip_version"`
}

// DNSConfig DNS 相关配置
type DNSConfig struct {
	Server          string        `yaml:"server" json:"server"`
	Timeout         time.Duration `yaml:"timeout" json:"timeout"`
	FallbackServers []string      `yaml:"fallback_servers" json:"fallback_servers"`
}

// HTTPConfig HTTP 相关配置
type HTTPConfig struct {
	Timeout        time.Duration `yaml:"timeout" json:"timeout"`
	FollowRedirect bool          `yaml:"follow_redirect" json:"follow_redirect"`
	MaxRedirects   int           `yaml:"max_redirects" json:"max_redirects"`
	UserAgent      string        `yaml:"user_agent" json:"user_agent"`
}

// ScanConfig 扫描相关配置
type ScanConfig struct {
	Timeout       time.Duration `yaml:"timeout" json:"timeout"`
	Concurrency   int           `yaml:"concurrency" json:"concurrency"`
	ServiceDetect bool          `yaml:"service_detect" json:"service_detect"`
}

// TraceConfig 路由追踪配置
type TraceConfig struct {
	Protocol   types.Protocol  `yaml:"protocol" json:"protocol"`
	MaxHops    int             `yaml:"max_hops" json:"max_hops"`
	Timeout    time.Duration   `yaml:"timeout" json:"timeout"`
	Queries    int             `yaml:"queries" json:"queries"`
	Port       int             `yaml:"port" json:"port"`
	PacketSize int             `yaml:"packet_size" json:"packet_size"`
	IPVersion  types.IPVersion `yaml:"ip_version" json:"ip_version"`
	FirstTTL   int             `yaml:"first_ttl" json:"first_ttl"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Global: GlobalConfig{
			Verbose:  false,
			Output:   "text",
			NoColor:  false,
			LogLevel: "info",
			LogFile:  "",
		},
		Ping: PingConfig{
			Protocol:  types.ProtocolICMP,
			Count:     4,
			Interval:  time.Second,
			Timeout:   5 * time.Second,
			Size:      64,
			TTL:       64,
			Port:      0,
			IPVersion: types.IPvAny,
		},
		DNS: DNSConfig{
			Server:  types.DefaultDNSServer,
			Timeout: types.DefaultDNSTimeout,
			FallbackServers: []string{
				types.GoogleDNSServer2,
				types.CloudflareDNSServer,
				types.AliDNSServer,
				types.DNSPodServer,
			},
		},
		HTTP: HTTPConfig{
			Timeout:        types.DefaultHTTPTimeout,
			FollowRedirect: true,
			MaxRedirects:   10,
			UserAgent:      buildinfo.UserAgent(),
		},
		Scan: ScanConfig{
			Timeout:       types.DefaultScanTimeout,
			Concurrency:   100,
			ServiceDetect: false,
		},
		Trace: TraceConfig{
			Protocol:   types.ProtocolICMP,
			MaxHops:    30,
			Timeout:    types.DefaultTraceTimeout,
			Queries:    3,
			Port:       types.DefaultTraceroutePort,
			PacketSize: 60,
			IPVersion:  types.IPvAny,
			FirstTTL:   1,
		},
	}
}
