package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/catsayer/ntx/internal/core/http"
	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/internal/output/formatter"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
)

func runHTTPBenchmark(ctx context.Context, client *http.Client, url string, headers map[string]string, outputFormat types.OutputFormat, noColor bool) {
	method := strings.ToUpper(httpMethod)

	logger.Info("开始 HTTP 性能测试",
		zap.String("method", method),
		zap.String("url", url),
		zap.Int("count", httpBenchCount))

	var body []byte
	if httpData != "" {
		body = []byte(httpData)
	}

	result, err := client.Benchmark(ctx, method, url, body, headers, httpBenchCount)
	if err != nil {
		logger.Error("HTTP 性能测试失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if outputFormat == types.OutputText || outputFormat == "" {
		printHTTPBenchmarkText(result, noColor)
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
