package batch

import (
	"context"
	"fmt"

	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
)

func (e *Executor) executeDNSTask(ctx context.Context, task Task, result *TaskResult) error {
	if len(task.Targets) == 0 {
		return fmt.Errorf("dns 任务未配置目标")
	}

	recordType := types.DNSTypeA

	if task.Options != nil {
		if recordTypeStr, ok := task.Options["type"].(string); ok {
			switch recordTypeStr {
			case "A":
				recordType = types.DNSTypeA
			case "AAAA":
				recordType = types.DNSTypeAAAA
			case "CNAME":
				recordType = types.DNSTypeCNAME
			case "MX":
				recordType = types.DNSTypeMX
			case "NS":
				recordType = types.DNSTypeNS
			case "TXT":
				recordType = types.DNSTypeTXT
			case "SOA":
				recordType = types.DNSTypeSOA
			case "PTR":
				recordType = types.DNSTypePTR
			}
		}
	}

	failures := 0
	for _, target := range task.Targets {
		dnsResult, err := e.resolver.Query(ctx, target, recordType)
		if err != nil {
			logger.Error("DNS 查询失败", zap.String("domain", target), zap.Error(err))
			failures++
			continue
		}

		result.Results = append(result.Results, dnsResult)

		logger.Info("DNS 查询完成",
			zap.String("domain", target),
			zap.Int("records", len(dnsResult.Records)),
		)
	}

	if failures > 0 {
		return fmt.Errorf("dns 任务部分失败: %d/%d 个目标失败", failures, len(task.Targets))
	}
	if len(result.Results) == 0 {
		return fmt.Errorf("dns 任务未产生有效结果")
	}

	return nil
}
