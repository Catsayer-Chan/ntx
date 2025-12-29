package batch

import (
	"context"

	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
)

func (e *Executor) executeDNSTask(ctx context.Context, task Task, result *TaskResult) error {
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

	for _, target := range task.Targets {
		dnsResult, err := e.resolver.Query(ctx, target, recordType)
		if err != nil {
			logger.Error("DNS 查询失败", zap.String("domain", target), zap.Error(err))
			continue
		}

		result.Results = append(result.Results, dnsResult)

		logger.Info("DNS 查询完成",
			zap.String("domain", target),
			zap.Int("records", len(dnsResult.Records)),
		)
	}

	return nil
}
