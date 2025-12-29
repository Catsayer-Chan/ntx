package config

import (
	"fmt"
	"strings"

	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/multierr"
)

// Validate 检查配置合法性
func Validate(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("配置为空")
	}

	var result error
	result = multierr.Append(result, validateGlobal(cfg.Global))
	result = multierr.Append(result, validatePing(cfg.Ping))
	result = multierr.Append(result, validateDNS(cfg.DNS))
	result = multierr.Append(result, validateHTTP(cfg.HTTP))
	result = multierr.Append(result, validateScan(cfg.Scan))
	result = multierr.Append(result, validateTrace(cfg.Trace))

	return result
}

func validateGlobal(cfg GlobalConfig) error {
	var err error
	output := strings.ToLower(cfg.Output)
	switch output {
	case "", "text", "json", "yaml", "table":
	default:
		err = multierr.Append(err, fmt.Errorf("global.output 不支持的值: %s", cfg.Output))
	}

	switch strings.ToLower(cfg.LogLevel) {
	case "debug", "info", "warn", "error":
	default:
		err = multierr.Append(err, fmt.Errorf("global.log_level 不支持的值: %s", cfg.LogLevel))
	}

	return err
}

func validatePing(cfg PingConfig) error {
	var err error
	if cfg.Count < 0 {
		err = multierr.Append(err, fmt.Errorf("ping.count 不能为负数"))
	}
	if cfg.Interval <= 0 {
		err = multierr.Append(err, fmt.Errorf("ping.interval 必须大于 0"))
	}
	if cfg.Timeout <= 0 {
		err = multierr.Append(err, fmt.Errorf("ping.timeout 必须大于 0"))
	}
	if cfg.Size <= 0 {
		err = multierr.Append(err, fmt.Errorf("ping.size 必须大于 0"))
	}
	if cfg.TTL <= 0 || cfg.TTL > 255 {
		err = multierr.Append(err, fmt.Errorf("ping.ttl 必须在 1-255 之间"))
	}

	switch cfg.Protocol {
	case "", types.ProtocolICMP, types.ProtocolTCP, types.ProtocolHTTP:
	default:
		err = multierr.Append(err, fmt.Errorf("ping.protocol 不支持的值: %s", cfg.Protocol))
	}

	return err
}

func validateDNS(cfg DNSConfig) error {
	var err error
	if strings.TrimSpace(cfg.Server) == "" {
		err = multierr.Append(err, fmt.Errorf("dns.server 不能为空"))
	}
	if cfg.Timeout <= 0 {
		err = multierr.Append(err, fmt.Errorf("dns.timeout 必须大于 0"))
	}
	return err
}

func validateHTTP(cfg HTTPConfig) error {
	var err error
	if cfg.Timeout <= 0 {
		err = multierr.Append(err, fmt.Errorf("http.timeout 必须大于 0"))
	}
	if cfg.MaxRedirects < 0 {
		err = multierr.Append(err, fmt.Errorf("http.max_redirects 不能为负数"))
	}
	return err
}

func validateScan(cfg ScanConfig) error {
	var err error
	if cfg.Timeout <= 0 {
		err = multierr.Append(err, fmt.Errorf("scan.timeout 必须大于 0"))
	}
	if cfg.Concurrency <= 0 {
		err = multierr.Append(err, fmt.Errorf("scan.concurrency 必须大于 0"))
	}
	return err
}

func validateTrace(cfg TraceConfig) error {
	var err error
	if cfg.MaxHops <= 0 || cfg.MaxHops > 255 {
		err = multierr.Append(err, fmt.Errorf("trace.max_hops 必须在 1-255 之间"))
	}
	if cfg.Timeout <= 0 {
		err = multierr.Append(err, fmt.Errorf("trace.timeout 必须大于 0"))
	}
	if cfg.Queries <= 0 {
		err = multierr.Append(err, fmt.Errorf("trace.queries 必须大于 0"))
	}
	if cfg.FirstTTL <= 0 {
		err = multierr.Append(err, fmt.Errorf("trace.first_ttl 必须大于 0"))
	}
	switch cfg.Protocol {
	case "", types.ProtocolICMP, types.ProtocolUDP, types.ProtocolTCP:
	default:
		err = multierr.Append(err, fmt.Errorf("trace.protocol 不支持的值: %s", cfg.Protocol))
	}
	return err
}
