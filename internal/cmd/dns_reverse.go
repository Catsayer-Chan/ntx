package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/catsayer/ntx/internal/core/dns"
	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
)

func runDNSReverse(ctx context.Context, resolver *dns.Resolver, ips []string, outputFormat types.OutputFormat, noColor bool) {
	logger.Info("开始反向 DNS 查询", zap.Strings("ips", ips))

	if len(ips) == 1 {
		result, err := resolver.Reverse(ctx, ips[0])
		if err != nil {
			logger.Error("反向 DNS 查询失败", zap.Error(err))
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		printDNSResult(result, outputFormat, noColor)
		return
	}

	results := make([]*types.DNSResult, 0, len(ips))
	for _, ip := range ips {
		result, err := resolver.Reverse(ctx, ip)
		if err != nil {
			result = &types.DNSResult{
				Domain:     ip,
				RecordType: types.DNSTypePTR,
				Error:      err,
			}
		}
		results = append(results, result)
	}

	printDNSBatchResults(results, outputFormat, noColor)
}
