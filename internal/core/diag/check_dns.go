package diag

import (
	"context"
	"time"

	"github.com/catsayer/ntx/pkg/types"
)

// checkDNSResolution 检查 DNS 解析
func (s *Service) checkDNSResolution(ctx context.Context) *CheckResult {
	startTime := time.Now()

	testDomains := []string{"google.com", "baidu.com"}

	successCount := 0
	for _, domain := range testDomains {
		result, err := s.resolver.Query(ctx, domain, types.DNSTypeA)
		if err == nil && len(result.Records) > 0 {
			successCount++
		}
	}

	if successCount == 0 {
		return &CheckResult{
			Name:     "DNS 解析检查",
			Category: "DNS",
			Status:   StatusCritical,
			Message:  "DNS 解析失败",
			Duration: time.Since(startTime),
		}
	}

	if successCount < len(testDomains) {
		return &CheckResult{
			Name:     "DNS 解析检查",
			Category: "DNS",
			Status:   StatusWarning,
			Message:  "部分域名解析失败",
			Duration: time.Since(startTime),
		}
	}

	return &CheckResult{
		Name:     "DNS 解析检查",
		Category: "DNS",
		Status:   StatusHealthy,
		Message:  "DNS 解析正常",
		Duration: time.Since(startTime),
	}
}
