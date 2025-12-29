package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/internal/output/formatter"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
)

func printDNSResult(result *types.DNSResult, outputFormat types.OutputFormat, noColor bool) {
	if outputFormat == types.OutputText || outputFormat == "" {
		fmt.Printf("; <<>> NTX DNS Query <<>> %s %s\n", result.Domain, result.RecordType)
		fmt.Printf(";; SERVER: %s\n", result.Server)
		fmt.Printf(";; WHEN: %s\n", result.StartTime.Format("Mon Jan 2 15:04:05 MST 2006"))
		fmt.Printf(";; Query time: %v\n\n", result.QueryTime)

		if len(result.Records) > 0 {
			fmt.Println(";; ANSWER SECTION:")
			printDNSTable(result.Records)
		}

		if len(result.Authority) > 0 {
			fmt.Println("\n;; AUTHORITY SECTION:")
			printDNSTable(result.Authority)
		}

		if len(result.Additional) > 0 {
			fmt.Println("\n;; ADDITIONAL SECTION:")
			printDNSTable(result.Additional)
		}
		return
	}

	f := formatter.NewFormatter(outputFormat, noColor)
	output, err := f.Format(result)
	if err != nil {
		logger.Error("格式化输出失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: 格式化输出失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}

func printDNSBatchResults(results []*types.DNSResult, outputFormat types.OutputFormat, noColor bool) {
	if outputFormat == types.OutputText || outputFormat == "" {
		for i, result := range results {
			if i > 0 {
				fmt.Println()
			}
			if result.Error != nil {
				fmt.Printf("%s: 查询失败: %v\n", result.Domain, result.Error)
				continue
			}
			printDNSResult(result, outputFormat, noColor)
		}
		return
	}

	f := formatter.NewFormatter(outputFormat, noColor)
	output, err := f.Format(results)
	if err != nil {
		logger.Error("格式化输出失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: 格式化输出失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}

func parseRecordType(typeStr string) types.DNSRecordType {
	switch strings.ToUpper(typeStr) {
	case "A":
		return types.DNSTypeA
	case "AAAA":
		return types.DNSTypeAAAA
	case "CNAME":
		return types.DNSTypeCNAME
	case "MX":
		return types.DNSTypeMX
	case "NS":
		return types.DNSTypeNS
	case "TXT":
		return types.DNSTypeTXT
	case "SOA":
		return types.DNSTypeSOA
	case "PTR":
		return types.DNSTypePTR
	case "SRV":
		return types.DNSTypeSRV
	default:
		return types.DNSTypeA
	}
}

func printDNSTable(records []*types.DNSRecord) {
	if len(records) == 0 {
		return
	}

	table := formatter.NewTable(
		[]string{"Name", "TTL", "Type", "Value"},
		[]int{30, 6, 10, 30},
	)
	for _, record := range records {
		table.AddRow(
			record.Name,
			fmt.Sprintf("%d", record.TTL),
			record.Type.String(),
			record.Value,
		)
	}
	table.Render(os.Stdout)
}
