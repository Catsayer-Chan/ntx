// Package cmd 提供 whois 命令实现
//
// 本文件实现 Whois 查询命令，支持:
// - 域名 Whois 查询
// - IP Whois 查询
// - AS 号查询
// - 批量查询
//
// 使用示例:
//
//	ntx whois google.com
//	ntx whois 8.8.8.8
//	ntx whois AS15169
//
// 作者: Catsayer
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/catsayer/ntx/internal/app"
	"github.com/catsayer/ntx/internal/core/whois"
	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/internal/output/formatter"
	"github.com/catsayer/ntx/pkg/types"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	whoisServer string
	whoisRaw    bool
)

var whoisCmd = &cobra.Command{
	Use:   "whois <domain|ip|as>",
	Short: "Whois 查询",
	Long: `查询域名、IP 或 AS 号的 Whois 信息。

支持功能:
  • 域名 Whois 查询
  • IP Whois 查询
  • AS 号查询
  • 批量查询
  • 自定义 Whois 服务器

示例:
  ntx whois google.com                 # 域名查询
  ntx whois 8.8.8.8                   # IP 查询
  ntx whois AS15169                   # AS 号查询
  ntx whois google.com --server whois.verisign-grs.com
  ntx whois google.com baidu.com      # 批量查询
  ntx whois google.com --raw          # 显示原始响应
  ntx whois google.com -o json        # JSON 输出`,
	Args: cobra.MinimumNArgs(1),
	RunE: runWhois,
}

func init() {
	rootCmd.AddCommand(whoisCmd)

	whoisCmd.Flags().StringVar(&whoisServer, "server", "", "指定 Whois 服务器")
	whoisCmd.Flags().BoolVar(&whoisRaw, "raw", false, "显示原始响应")
}

func runWhois(cmd *cobra.Command, args []string) error {
	appCtx := mustAppContext(cmd)
	queries := args

	logger.Info("开始 Whois 查询", zap.Int("queries", len(queries)))

	// 构建查询选项
	opts := types.DefaultWhoisOptions()
	opts.Server = whoisServer

	// 创建 Whois 客户端
	client := whois.NewClient()
	ctx := context.Background()

	// 如果是单个查询
	if len(queries) == 1 {
		result, err := client.Query(ctx, queries[0], opts)
		if err != nil {
			return fmt.Errorf("Whois 查询失败: %w", err)
		}
		return outputWhoisResult(result, appCtx.Flags)
	}

	// 批量查询
	results, err := client.QueryBatch(ctx, queries, opts)
	if err != nil {
		return fmt.Errorf("批量查询失败: %w", err)
	}

	// 输出批量结果
	for i, result := range results {
		if i > 0 {
			fmt.Println()
			fmt.Println(strings.Repeat("=", 70))
			fmt.Println()
		}
		if err := outputWhoisResult(result, appCtx.Flags); err != nil {
			logger.Error("输出结果失败",
				zap.String("query", result.Query),
				zap.Error(err),
			)
		}
	}

	return nil
}

// outputWhoisResult 输出 Whois 查询结果
func outputWhoisResult(result *types.WhoisResult, flags app.GlobalFlags) error {
	outputFormat := types.OutputFormat(flags.Output)
	if outputFormat == types.OutputText || outputFormat == "" {
		return outputWhoisText(result, flags)
	}

	f := formatter.NewFormatter(outputFormat, flags.NoColor)
	return f.FormatTo(os.Stdout, result)
}

// outputWhoisText 文本格式输出
func outputWhoisText(result *types.WhoisResult, flags app.GlobalFlags) error {
	f := formatter.NewTextFormatter(!flags.NoColor)

	// 显示查询信息
	f.PrintHeader(fmt.Sprintf("Whois 查询: %s", result.Query))
	fmt.Printf("查询服务器: %s\n", result.Server)
	fmt.Printf("查询耗时:   %s\n", result.QueryTime)
	fmt.Println()

	// 如果显示原始响应
	if whoisRaw {
		f.PrintSubHeader("原始响应")
		fmt.Println(result.RawResponse)
		return nil
	}

	// 显示解析后的数据
	data := result.ParsedData
	if data == nil {
		color.NoColor = flags.NoColor
		fmt.Println(color.YellowString("无法解析响应数据"))
		return nil
	}

	switch result.Type {
	case types.WhoisDomain:
		outputDomainInfo(data, flags)
	case types.WhoisIP:
		outputIPInfo(data, flags)
	case types.WhoisAS:
		outputASInfo(data, flags)
	}

	return nil
}

// outputDomainInfo 输出域名信息
func outputDomainInfo(data *types.WhoisData, flags app.GlobalFlags) {
	f := formatter.NewTextFormatter(!flags.NoColor)
	color.NoColor = flags.NoColor

	f.PrintSubHeader("域名信息")
	if data.Domain != "" {
		fmt.Printf("域名:         %s\n", color.CyanString(data.Domain))
	}
	if data.Registrar != "" {
		fmt.Printf("注册商:       %s\n", data.Registrar)
	}
	if data.RegistrantOrg != "" {
		fmt.Printf("注册组织:     %s\n", data.RegistrantOrg)
	}
	if data.RegistrantName != "" {
		fmt.Printf("注册人:       %s\n", data.RegistrantName)
	}
	if !data.CreationDate.IsZero() {
		fmt.Printf("创建日期:     %s\n", data.CreationDate.Format("2006-01-02"))
	}
	if !data.ExpirationDate.IsZero() {
		fmt.Printf("过期日期:     %s\n", data.ExpirationDate.Format("2006-01-02"))
	}
	if !data.UpdatedDate.IsZero() {
		fmt.Printf("更新日期:     %s\n", data.UpdatedDate.Format("2006-01-02"))
	}

	if len(data.NameServers) > 0 {
		fmt.Println()
		f.PrintSubHeader("域名服务器")
		for _, ns := range data.NameServers {
			fmt.Printf("  • %s\n", ns)
		}
	}

	if len(data.Status) > 0 {
		fmt.Println()
		f.PrintSubHeader("域名状态")
		for _, status := range data.Status {
			fmt.Printf("  • %s\n", status)
		}
	}
}

// outputIPInfo 输出 IP 信息
func outputIPInfo(data *types.WhoisData, flags app.GlobalFlags) {
	f := formatter.NewTextFormatter(!flags.NoColor)
	color.NoColor = flags.NoColor

	f.PrintSubHeader("IP 信息")
	if data.IPRange != "" {
		fmt.Printf("IP 范围:      %s\n", color.CyanString(data.IPRange))
	}
	if data.Organization != "" {
		fmt.Printf("组织:         %s\n", data.Organization)
	}
	if data.Country != "" {
		fmt.Printf("国家:         %s\n", data.Country)
	}
	if data.City != "" {
		fmt.Printf("城市:         %s\n", data.City)
	}

	if len(data.Address) > 0 {
		fmt.Println()
		f.PrintSubHeader("地址信息")
		for _, addr := range data.Address {
			fmt.Printf("  %s\n", addr)
		}
	}
}

// outputASInfo 输出 AS 信息
func outputASInfo(data *types.WhoisData, flags app.GlobalFlags) {
	f := formatter.NewTextFormatter(!flags.NoColor)
	color.NoColor = flags.NoColor

	f.PrintSubHeader("AS 信息")
	if data.ASName != "" {
		fmt.Printf("AS 名称:      %s\n", color.CyanString(data.ASName))
	}
	if data.Organization != "" {
		fmt.Printf("组织:         %s\n", data.Organization)
	}
}
