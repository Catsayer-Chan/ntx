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

func runHTTPRequest(ctx context.Context, client *http.Client, url string, headers map[string]string, outputFormat types.OutputFormat, noColor bool) {
	method := strings.ToUpper(httpMethod)

	logger.Info("发送 HTTP 请求",
		zap.String("method", method),
		zap.String("url", url))

	var body []byte
	if httpData != "" {
		body = []byte(httpData)
	}

	result, err := client.Request(ctx, method, url, body, headers)
	if err != nil {
		logger.Error("HTTP 请求失败", zap.Error(err))
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if outputFormat == types.OutputText || outputFormat == "" {
		printHTTPResultText(result, noColor)
	} else {
		f := formatter.NewFormatter(outputFormat, noColor)
		output, err := f.Format(result)
		if err != nil {
			logger.Error("格式化输出失败", zap.Error(err))
			fmt.Fprintf(os.Stderr, "错误: 格式化输出失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(output)
	}

	if result.StatusCode < 200 || result.StatusCode >= 300 {
		os.Exit(1)
	}
}
