// Package cmd 中 DNS 命令的入口与标志定义
//
// 将命令定义、参数解析与执行调度从其它逻辑中拆出，便于维护。
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/catsayer/ntx/internal/core/dns"
	"github.com/catsayer/ntx/pkg/types"
	"github.com/spf13/cobra"
)

var (
	dnsServer  string
	dnsType    string
	dnsTimeout float64
	dnsReverse bool
	dnsAll     bool
)

// dnsCmd 表示 dns 命令
var dnsCmd = &cobra.Command{
	Use:   "dns <domain...>",
	Short: "查询 DNS 记录",
	Long: `查询域名的 DNS 记录。

支持的记录类型:
  A      - IPv4 地址记录
  AAAA   - IPv6 地址记录
  CNAME  - 规范名称记录
  MX     - 邮件交换记录
  NS     - 名称服务器记录
  TXT    - 文本记录
  SOA    - 授权起始记录
  PTR    - 指针记录 (用于反向查询)
  SRV    - 服务记录

示例:
  # 查询 A 记录 (默认)
  ntx dns google.com

  # 查询 AAAA 记录
  ntx dns google.com --type AAAA

  # 查询 MX 记录
  ntx dns google.com --type MX

  # 查询所有常见记录
  ntx dns google.com --all

  # 使用自定义 DNS 服务器
  ntx dns google.com --server 1.1.1.1

  # 反向 DNS 查询
  ntx dns --reverse 8.8.8.8

  # 批量查询多个域名
  ntx dns google.com baidu.com github.com

  # JSON 输出
  ntx dns google.com --type A -o json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if dnsReverse {
			if len(args) < 1 {
				return fmt.Errorf("反向查询需要至少一个 IP 地址")
			}
			return nil
		}
		if len(args) < 1 {
			return fmt.Errorf("需要至少一个域名")
		}
		return nil
	},
	Run: runDNS,
}

func init() {
	rootCmd.AddCommand(dnsCmd)

	dnsCmd.Flags().StringVarP(&dnsServer, "server", "s", types.DefaultDNSServer,
		"DNS 服务器地址")
	dnsCmd.Flags().StringVarP(&dnsType, "type", "t", "A",
		"记录类型 (A, AAAA, CNAME, MX, NS, TXT, SOA, PTR, SRV)")
	dnsCmd.Flags().Float64Var(&dnsTimeout, "timeout", types.DefaultDNSTimeout.Seconds(),
		"查询超时时间（秒）")
	dnsCmd.Flags().BoolVarP(&dnsReverse, "reverse", "r", false,
		"反向 DNS 查询 (IP 到域名)")
	dnsCmd.Flags().BoolVarP(&dnsAll, "all", "a", false,
		"查询所有常见记录类型")
}

func runDNS(cmd *cobra.Command, args []string) {
	appCtx := mustAppContext(cmd)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := buildDNSOptions(cmd, appCtx)
	resolver := dns.NewResolver(opts)
	defer resolver.Close()

	outputFormat := types.OutputFormat(appCtx.Flags.Output)
	noColor := appCtx.Flags.NoColor

	switch {
	case dnsReverse:
		runDNSReverse(ctx, resolver, args, outputFormat, noColor)
	case dnsAll:
		runDNSAll(ctx, resolver, args, outputFormat, noColor)
	default:
		runDNSStandard(ctx, resolver, args, outputFormat, opts.Server, noColor)
	}
}
