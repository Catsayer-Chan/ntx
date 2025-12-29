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

func runDNSStandard(ctx context.Context, resolver *dns.Resolver, domains []string, outputFormat types.OutputFormat, server string, noColor bool) {
	recordType := parseRecordType(dnsType)

	logger.Info("开始 DNS 查询",
		zap.Strings("domains", domains),
		zap.String("type", dnsType),
		zap.String("server", server))

	if len(domains) == 1 {
		result, err := resolver.Query(ctx, domains[0], recordType)
		if err != nil {
			logger.Error("DNS 查询失败", zap.Error(err))
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		printDNSResult(result, outputFormat, noColor)
		return
	}

	results, err := resolver.QueryBatch(ctx, domains, recordType)
	if err != nil {
		logger.Error("批量 DNS 查询失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	printDNSBatchResults(results, outputFormat, noColor)
}
