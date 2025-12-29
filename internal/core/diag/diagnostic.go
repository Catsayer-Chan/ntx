// Package diag 提供智能网络诊断功能
//
// 本模块实现了自动化网络诊断:
// - 本机网络配置检查
// - 连通性测试
// - DNS 解析测试
// - 路由路径测试
// - 问题分析和修复建议
//
// 依赖:
// - internal/core/ping: Ping 功能
// - internal/core/dns: DNS 功能
// - internal/core/iface: 网卡信息
//
// 使用示例:
//
//	diagService := diag.NewService()
//	result, err := diagService.Diagnose(ctx, opts)
//
// 作者: Catsayer
package diag

import (
	"context"
	"fmt"
	"time"

	"github.com/catsayer/ntx/internal/core/dns"
	"github.com/catsayer/ntx/internal/core/iface"
	"github.com/catsayer/ntx/internal/core/ping"
	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
)

// DiagnosticLevel 诊断级别
type DiagnosticLevel int

const (
	// DiagLevelFast 快速诊断（基础检查）
	DiagLevelFast DiagnosticLevel = iota
	// DiagLevelNormal 标准诊断
	DiagLevelNormal
	// DiagLevelFull 完整诊断（包含性能测试）
	DiagLevelFull
)

// DiagnosticOptions 诊断选项
type DiagnosticOptions struct {
	Level  DiagnosticLevel
	Target string // 可选的目标主机
}

// DiagnosticResult 诊断结果
type DiagnosticResult struct {
	Timestamp   time.Time
	Duration    time.Duration
	Status      DiagnosticStatus
	Checks      []*CheckResult
	Issues      []*Issue
	Suggestions []string
}

// DiagnosticStatus 诊断状态
type DiagnosticStatus int

const (
	StatusHealthy DiagnosticStatus = iota
	StatusWarning
	StatusCritical
)

func (s DiagnosticStatus) String() string {
	switch s {
	case StatusHealthy:
		return "HEALTHY"
	case StatusWarning:
		return "WARNING"
	case StatusCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// CheckResult 单项检查结果
type CheckResult struct {
	Name     string
	Category string
	Status   DiagnosticStatus
	Message  string
	Duration time.Duration
	Details  map[string]interface{}
}

// Issue 发现的问题
type Issue struct {
	Severity    DiagnosticStatus
	Category    string
	Description string
	Suggestion  string
}

// Service 诊断服务
type Service struct {
	pinger   types.Pinger
	resolver *dns.Resolver
	ifReader *iface.InterfaceReader
}

// NewService 创建新的诊断服务
func NewService() *Service {
	// 创建 ICMP Pinger (如果失败则回退到 TCP)
	opts := types.DefaultPingOptions()
	var pinger types.Pinger
	icmpPinger, err := ping.NewICMPPinger(opts)
	if err != nil {
		// 如果 ICMP 失败,使用 TCP Pinger
		pinger = ping.NewTCPPinger(opts)
	} else {
		pinger = icmpPinger
	}

	// 创建 DNS Resolver
	dnsOpts := &types.DNSOptions{
		Server:  types.DefaultDNSServer,
		Timeout: types.DefaultDNSTimeout,
	}

	return &Service{
		pinger:   pinger,
		resolver: dns.NewResolver(dnsOpts),
		ifReader: iface.NewInterfaceReader(),
	}
}

// Diagnose 执行网络诊断
//
// 参数:
//
//	ctx: 上下文
//	opts: 诊断选项
//
// 返回:
//
//	*DiagnosticResult: 诊断结果
//	error: 错误信息
func (s *Service) Diagnose(ctx context.Context, opts DiagnosticOptions) (*DiagnosticResult, error) {
	startTime := time.Now()

	logger.Info("开始网络诊断", zap.Int("level", int(opts.Level)))

	result := &DiagnosticResult{
		Timestamp:   startTime,
		Checks:      make([]*CheckResult, 0),
		Issues:      make([]*Issue, 0),
		Suggestions: make([]string, 0),
	}

	// 1. 检查网络接口配置
	if check := s.checkNetworkInterfaces(ctx); check != nil {
		result.Checks = append(result.Checks, check)
		if check.Status != StatusHealthy {
			result.Issues = append(result.Issues, &Issue{
				Severity:    check.Status,
				Category:    "网络配置",
				Description: check.Message,
				Suggestion:  "检查网络接口配置和状态",
			})
		}
	}

	// 2. 检查本地连通性（网关）
	if check := s.checkLocalConnectivity(ctx); check != nil {
		result.Checks = append(result.Checks, check)
		if check.Status != StatusHealthy {
			result.Issues = append(result.Issues, &Issue{
				Severity:    check.Status,
				Category:    "本地连通性",
				Description: check.Message,
				Suggestion:  "检查网关配置和本地网络连接",
			})
		}
	}

	// 3. 检查公网连通性
	if check := s.checkInternetConnectivity(ctx); check != nil {
		result.Checks = append(result.Checks, check)
		if check.Status != StatusHealthy {
			result.Issues = append(result.Issues, &Issue{
				Severity:    check.Status,
				Category:    "互联网连通性",
				Description: check.Message,
				Suggestion:  "检查路由器配置和 ISP 连接",
			})
		}
	}

	// 4. 检查 DNS 解析
	if check := s.checkDNSResolution(ctx); check != nil {
		result.Checks = append(result.Checks, check)
		if check.Status != StatusHealthy {
			result.Issues = append(result.Issues, &Issue{
				Severity:    check.Status,
				Category:    "DNS 解析",
				Description: check.Message,
				Suggestion:  fmt.Sprintf("检查 DNS 服务器配置，尝试使用公共 DNS（如 %s）", types.DefaultDNSServer),
			})
		}
	}

	// 5. 如果指定了目标，进行额外测试
	if opts.Target != "" {
		if check := s.checkTargetReachability(ctx, opts.Target); check != nil {
			result.Checks = append(result.Checks, check)
		}
	}

	// 计算整体状态
	result.Status = s.calculateOverallStatus(result.Checks)
	result.Duration = time.Since(startTime)

	// 生成建议
	result.Suggestions = s.generateSuggestions(result)

	logger.Info("网络诊断完成",
		zap.String("status", result.Status.String()),
		zap.Duration("duration", result.Duration),
		zap.Int("issues", len(result.Issues)),
	)

	return result, nil
}

// calculateOverallStatus 计算整体状态
func (s *Service) calculateOverallStatus(checks []*CheckResult) DiagnosticStatus {
	hasCritical := false
	hasWarning := false

	for _, check := range checks {
		if check.Status == StatusCritical {
			hasCritical = true
		} else if check.Status == StatusWarning {
			hasWarning = true
		}
	}

	if hasCritical {
		return StatusCritical
	}
	if hasWarning {
		return StatusWarning
	}
	return StatusHealthy
}

// generateSuggestions 生成诊断建议
func (s *Service) generateSuggestions(result *DiagnosticResult) []string {
	suggestions := make([]string, 0)

	if result.Status == StatusHealthy {
		suggestions = append(suggestions, "网络配置正常，所有检查通过")
		return suggestions
	}

	for _, issue := range result.Issues {
		if issue.Suggestion != "" {
			suggestions = append(suggestions, issue.Suggestion)
		}
	}

	return suggestions
}
