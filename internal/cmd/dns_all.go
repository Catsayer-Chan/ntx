package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/catsayer/ntx/internal/core/dns"
	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/internal/output/formatter"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
)

func runDNSAll(ctx context.Context, resolver *dns.Resolver, domains []string, outputFormat types.OutputFormat, noColor bool) {
	logger.Info("查询所有 DNS 记录类型", zap.Strings("domains", domains))

	for i, domain := range domains {
		if i > 0 {
			fmt.Println()
		}

		results, err := resolver.QueryAll(ctx, domain)
		if err != nil {
			logger.Error("DNS 查询失败", zap.Error(err), zap.String("domain", domain))
			fmt.Fprintf(os.Stderr, "错误: 查询 %s 失败: %v\n", domain, err)
			continue
		}

		if outputFormat == types.OutputText || outputFormat == "" {
			fmt.Printf("DNS records for %s:\n", domain)
			for recordType, result := range results {
				fmt.Printf("\n%s records:\n", recordType)
				for _, record := range result.Records {
					fmt.Printf("  %-30s %6d  %-10s %s\n",
						record.Name, record.TTL, recordType, record.Value)
				}
			}
			continue
		}

		f := formatter.NewFormatter(outputFormat, noColor)
		output, err := f.Format(results)
		if err != nil {
			logger.Error("格式化输出失败", zap.Error(err))
			fmt.Fprintf(os.Stderr, "错误: 格式化输出失败: %v\n", err)
			continue
		}
		fmt.Print(output)
	}
}
