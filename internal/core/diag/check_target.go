package diag

import (
	"context"
	"fmt"
	"time"

	"github.com/catsayer/ntx/pkg/types"
)

// checkTargetReachability 检查目标主机可达性
func (s *Service) checkTargetReachability(ctx context.Context, target string) *CheckResult {
	startTime := time.Now()

	pingOpts := types.DefaultPingOptions()
	pingOpts.Count = 5
	pingOpts.Timeout = types.DiagnosticTargetTimeout

	result, err := s.pinger.Ping(ctx, target, pingOpts)
	if err != nil || result.Statistics.Received == 0 {
		return &CheckResult{
			Name:     fmt.Sprintf("目标主机检查 (%s)", target),
			Category: "连通性",
			Status:   StatusCritical,
			Message:  fmt.Sprintf("无法连接到 %s", target),
			Duration: time.Since(startTime),
		}
	}

	lossRate := result.Statistics.LossRate
	avgRTT := result.Statistics.AvgRTT

	status := StatusHealthy
	message := fmt.Sprintf("目标主机 %s 可达，延迟 %.2fms", target, float64(avgRTT.Microseconds())/1000)

	if lossRate > 20 {
		status = StatusWarning
		message = fmt.Sprintf("目标主机 %s 可达但丢包率较高 (%.1f%%)", target, lossRate)
	}

	return &CheckResult{
		Name:     fmt.Sprintf("目标主机检查 (%s)", target),
		Category: "连通性",
		Status:   status,
		Message:  message,
		Duration: time.Since(startTime),
		Details: map[string]interface{}{
			"avg_rtt":   avgRTT.String(),
			"loss_rate": lossRate,
		},
	}
}
