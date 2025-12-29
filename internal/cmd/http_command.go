package cmd

import (
	"context"
	"time"

	"github.com/catsayer/ntx/internal/core/http"
	"github.com/catsayer/ntx/pkg/types"
	"github.com/spf13/cobra"
)

var (
	httpMethod      string
	httpData        string
	httpHeaders     []string
	httpTimeout     float64
	httpNoRedirect  bool
	httpIncludeHead bool
	httpHeadOnly    bool
	httpBench       bool
	httpBenchCount  int
)

var httpCmd = &cobra.Command{
	Use:   "http <url>",
	Short: "发送 HTTP 请求",
	Long: `发送 HTTP 请求并显示响应。

类似 curl 工具,但提供更友好的输出格式。

支持的方法:
  GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS

示例:
  # GET 请求
  ntx http https://api.github.com

  # POST JSON 数据
  ntx http https://httpbin.org/post -X POST -d '{"key":"value"}'

  # 添加请求头
  ntx http https://api.github.com -H "Authorization: token xxx"

  # 显示响应头
  ntx http https://api.github.com --include

  # 仅显示响应头
  ntx http https://api.github.com --head

  # 不跟随重定向
  ntx http https://example.com --no-redirect

  # 性能测试（发送 100 次请求）
  ntx http https://api.github.com --bench -n 100

  # JSON 输出
  ntx http https://api.github.com -o json`,
	Args: cobra.ExactArgs(1),
	Run:  runHTTP,
}

func init() {
	rootCmd.AddCommand(httpCmd)

	httpCmd.Flags().StringVarP(&httpMethod, "method", "X", "GET",
		"HTTP 方法 (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)")
	httpCmd.Flags().StringVarP(&httpData, "data", "d", "",
		"请求体数据")
	httpCmd.Flags().StringSliceVarP(&httpHeaders, "header", "H", nil,
		"请求头 (可多次指定)")
	httpCmd.Flags().Float64Var(&httpTimeout, "timeout", 30.0,
		"请求超时时间（秒）")
	httpCmd.Flags().BoolVar(&httpNoRedirect, "no-redirect", false,
		"不跟随重定向")
	httpCmd.Flags().BoolVarP(&httpIncludeHead, "include", "i", false,
		"在输出中包含响应头")
	httpCmd.Flags().BoolVarP(&httpHeadOnly, "head", "I", false,
		"仅显示响应头（HEAD 请求）")
	httpCmd.Flags().BoolVar(&httpBench, "bench", false,
		"性能测试模式")
	httpCmd.Flags().IntVarP(&httpBenchCount, "count", "n", 10,
		"性能测试请求次数")
}

func runHTTP(cmd *cobra.Command, args []string) {
	appCtx := mustAppContext(cmd)
	url := args[0]
	opts := buildHTTPOptions(cmd, appCtx)

	headers := buildHTTPHeaders(httpHeaders, httpData)

	if httpHeadOnly {
		httpMethod = "HEAD"
	}

	client := http.NewClient(opts)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout+5*time.Second)
	defer cancel()

	outputFormat := types.OutputFormat(appCtx.Flags.Output)
	noColor := appCtx.Flags.NoColor

	if httpBench {
		runHTTPBenchmark(ctx, client, url, headers, outputFormat, noColor)
		return
	}
	runHTTPRequest(ctx, client, url, headers, outputFormat, noColor)
}
