package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/catsayer/ntx/pkg/types"
	"gopkg.in/yaml.v3"
)

// Loader 负责加载配置文件并应用环境变量
type Loader struct {
	searchPaths []string
}

// NewLoader 创建配置加载器并使用默认搜索路径
func NewLoader() *Loader {
	paths := []string{".ntx.yaml"}

	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths,
			filepath.Join(homeDir, ".ntx.yaml"),
			filepath.Join(homeDir, ".config", "ntx", "config.yaml"),
		)
	}

	paths = append(paths, "/etc/ntx/config.yaml")

	return &Loader{
		searchPaths: uniquePaths(paths),
	}
}

// Load 从配置文件加载配置，如果 configPath 为空则按照默认搜索路径查找
func (l *Loader) Load(configPath string) (*Config, string, error) {
	cfg := DefaultConfig()
	candidates := l.candidatePaths(configPath)

	for _, path := range candidates {
		if path == "" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, "", fmt.Errorf("读取配置文件 %s 失败: %w", path, err)
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, "", fmt.Errorf("解析配置文件 %s 失败: %w", path, err)
		}
		return cfg, path, nil
	}

	// 未找到配置文件，返回默认配置
	return cfg, "", nil
}

// LoadWithEnv 加载配置并应用环境变量覆盖
func (l *Loader) LoadWithEnv(configPath string) (*Config, string, error) {
	cfg, path, err := l.Load(configPath)
	if err != nil {
		return nil, "", err
	}

	applyEnvOverrides(cfg)
	return cfg, path, nil
}

func (l *Loader) candidatePaths(configPath string) []string {
	if configPath == "" {
		return l.searchPaths
	}
	return append([]string{configPath}, l.searchPaths...)
}

func uniquePaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	result := make([]string, 0, len(paths))
	for _, p := range paths {
		if p == "" {
			continue
		}
		abs := p
		if !filepath.IsAbs(p) {
			abs = filepath.Clean(p)
		}
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		result = append(result, p)
	}
	return result
}

func applyEnvOverrides(cfg *Config) {
	if cfg == nil {
		return
	}

	if v := os.Getenv("NTX_VERBOSE"); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			cfg.Global.Verbose = parsed
		}
	}
	if v := os.Getenv("NTX_OUTPUT"); v != "" {
		cfg.Global.Output = strings.ToLower(v)
	}
	if v := os.Getenv("NTX_NO_COLOR"); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			cfg.Global.NoColor = parsed
		}
	}
	if v := os.Getenv("NTX_LOG_LEVEL"); v != "" {
		cfg.Global.LogLevel = strings.ToLower(v)
	}

	if v := os.Getenv("NTX_PING_COUNT"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			cfg.Ping.Count = parsed
		}
	}
	if v := os.Getenv("NTX_PING_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Ping.Interval = d
		}
	}
	if v := os.Getenv("NTX_PING_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Ping.Timeout = d
		}
	}
	if v := os.Getenv("NTX_PING_PROTOCOL"); v != "" {
		cfg.Ping.Protocol = types.Protocol(strings.ToLower(v))
	}

	if v := os.Getenv("NTX_DNS_SERVER"); v != "" {
		cfg.DNS.Server = v
	}
	if v := os.Getenv("NTX_DNS_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.DNS.Timeout = d
		}
	}

	if v := os.Getenv("NTX_HTTP_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.HTTP.Timeout = d
		}
	}
	if v := os.Getenv("NTX_HTTP_MAX_REDIRECTS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			cfg.HTTP.MaxRedirects = parsed
		}
	}
	if v := os.Getenv("NTX_HTTP_USER_AGENT"); v != "" {
		cfg.HTTP.UserAgent = v
	}

	if v := os.Getenv("NTX_SCAN_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Scan.Timeout = d
		}
	}
	if v := os.Getenv("NTX_SCAN_CONCURRENCY"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			cfg.Scan.Concurrency = parsed
		}
	}
	if v := os.Getenv("NTX_SCAN_SERVICE_DETECT"); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			cfg.Scan.ServiceDetect = parsed
		}
	}

	if v := os.Getenv("NTX_TRACE_MAX_HOPS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			cfg.Trace.MaxHops = parsed
		}
	}
	if v := os.Getenv("NTX_TRACE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Trace.Timeout = d
		}
	}
	if v := os.Getenv("NTX_TRACE_QUERIES"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			cfg.Trace.Queries = parsed
		}
	}
	if v := os.Getenv("NTX_TRACE_PROTOCOL"); v != "" {
		cfg.Trace.Protocol = types.Protocol(strings.ToLower(v))
	}
}
