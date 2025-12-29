package diag

import (
	"context"
	"fmt"
	"time"
)

// checkNetworkInterfaces 检查网络接口配置
func (s *Service) checkNetworkInterfaces(ctx context.Context) *CheckResult {
	startTime := time.Now()

	interfaces, err := s.ifReader.GetInterfaces()
	if err != nil {
		return &CheckResult{
			Name:     "网络接口检查",
			Category: "网络配置",
			Status:   StatusCritical,
			Message:  fmt.Sprintf("获取网络接口失败: %v", err),
			Duration: time.Since(startTime),
		}
	}

	hasActiveInterface := false
	hasIPv4Address := false

	for _, iface := range interfaces {
		if contains(iface.Flags, "UP") && !contains(iface.Flags, "LOOPBACK") {
			hasActiveInterface = true
			if len(iface.IPv4Addrs) > 0 {
				hasIPv4Address = true
				break
			}
		}
	}

	if !hasActiveInterface {
		return &CheckResult{
			Name:     "网络接口检查",
			Category: "网络配置",
			Status:   StatusCritical,
			Message:  "没有活动的网络接口",
			Duration: time.Since(startTime),
		}
	}

	if !hasIPv4Address {
		return &CheckResult{
			Name:     "网络接口检查",
			Category: "网络配置",
			Status:   StatusWarning,
			Message:  "网络接口没有配置 IPv4 地址",
			Duration: time.Since(startTime),
		}
	}

	return &CheckResult{
		Name:     "网络接口检查",
		Category: "网络配置",
		Status:   StatusHealthy,
		Message:  "网络接口配置正常",
		Duration: time.Since(startTime),
		Details: map[string]interface{}{
			"interface_count": len(interfaces),
		},
	}
}
