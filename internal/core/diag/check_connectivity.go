package diag

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/catsayer/ntx/pkg/types"
)

// checkLocalConnectivity 检查本地连通性（Ping 网关）
func (s *Service) checkLocalConnectivity(ctx context.Context) *CheckResult {
	startTime := time.Now()

	gateway, err := getDefaultGateway()
	if err != nil {
		return &CheckResult{
			Name:     "本地连通性检查",
			Category: "连通性",
			Status:   StatusWarning,
			Message:  "无法获取默认网关",
			Duration: time.Since(startTime),
		}
	}

	pingOpts := types.DefaultPingOptions()
	pingOpts.Count = 3
	pingOpts.Timeout = types.DiagnosticGatewayTimeout

	result, err := s.pinger.Ping(ctx, gateway, pingOpts)
	if err != nil || result.Statistics.Received == 0 {
		return &CheckResult{
			Name:     "本地连通性检查",
			Category: "连通性",
			Status:   StatusCritical,
			Message:  fmt.Sprintf("无法连接到网关 %s", gateway),
			Duration: time.Since(startTime),
		}
	}

	return &CheckResult{
		Name:     "本地连通性检查",
		Category: "连通性",
		Status:   StatusHealthy,
		Message:  fmt.Sprintf("网关 %s 可达，延迟 %.2fms", gateway, float64(result.Statistics.AvgRTT.Microseconds())/1000),
		Duration: time.Since(startTime),
		Details: map[string]interface{}{
			"gateway": gateway,
			"avg_rtt": result.Statistics.AvgRTT.String(),
		},
	}
}

// checkInternetConnectivity 检查互联网连通性
func (s *Service) checkInternetConnectivity(ctx context.Context) *CheckResult {
	startTime := time.Now()

	publicDNS := types.DNSServerList()

	pingOpts := types.DefaultPingOptions()
	pingOpts.Count = 2
	pingOpts.Timeout = types.DiagnosticTargetTimeout
	pingOpts.Protocol = types.ProtocolTCP
	pingOpts.Port = types.DefaultDNSPort

	successCount := 0
	for _, dnsAddr := range publicDNS {
		host := dnsAddr
		if h, _, err := net.SplitHostPort(dnsAddr); err == nil {
			host = h
		}
		result, err := s.pinger.Ping(ctx, host, pingOpts)
		if err == nil && result.Statistics.Received > 0 {
			successCount++
		}
	}

	if successCount == 0 {
		return &CheckResult{
			Name:     "互联网连通性检查",
			Category: "连通性",
			Status:   StatusCritical,
			Message:  "无法连接到互联网",
			Duration: time.Since(startTime),
		}
	}

	if successCount < len(publicDNS) {
		return &CheckResult{
			Name:     "互联网连通性检查",
			Category: "连通性",
			Status:   StatusWarning,
			Message:  "互联网连接不稳定",
			Duration: time.Since(startTime),
		}
	}

	return &CheckResult{
		Name:     "互联网连通性检查",
		Category: "连通性",
		Status:   StatusHealthy,
		Message:  "互联网连接正常",
		Duration: time.Since(startTime),
	}
}
