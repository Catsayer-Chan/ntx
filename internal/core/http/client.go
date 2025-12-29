// Package http 提供 HTTP 客户端功能
//
// 本模块实现了 HTTP 客户端功能,包括:
// - GET/POST/PUT/DELETE/PATCH 等 HTTP 方法
// - 自定义请求头
// - 请求体支持（JSON、Form 等）
// - 超时控制
// - 重定向控制
// - 响应详情显示
//
// 依赖:
// - net/http: Go 标准库 HTTP 客户端
//
// 使用示例:
//
//	client := http.NewClient(&http.Options{Timeout: types.DefaultHTTPTimeout})
//	result, err := client.Request("GET", "https://api.github.com", nil)
//
// 作者: Catsayer
package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/catsayer/ntx/pkg/buildinfo"
	"github.com/catsayer/ntx/pkg/types"
)

// Client HTTP 客户端
type Client struct {
	client  *http.Client
	options *types.HTTPOptions
}

// NewClient 创建新的 HTTP 客户端
func NewClient(opts *types.HTTPOptions) *Client {
	if opts == nil {
		opts = &types.HTTPOptions{
			Timeout:        types.DefaultHTTPTimeout,
			FollowRedirect: true,
			MaxRedirects:   10,
		}
	}

	if opts.Timeout == 0 {
		opts.Timeout = types.DefaultHTTPTimeout
	}
	if opts.UserAgent == "" {
		opts.UserAgent = buildinfo.UserAgent()
	}

	httpClient := &http.Client{
		Timeout: opts.Timeout,
	}

	// 配置重定向策略
	if !opts.FollowRedirect {
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else if opts.MaxRedirects > 0 {
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= opts.MaxRedirects {
				return fmt.Errorf("超过最大重定向次数: %d", opts.MaxRedirects)
			}
			return nil
		}
	}

	return &Client{
		client:  httpClient,
		options: opts,
	}
}

// Request 执行 HTTP 请求
func (c *Client) Request(ctx context.Context, method, url string, body []byte, headers map[string]string) (*types.HTTPResult, error) {
	startTime := time.Now()

	// 创建请求
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 如果没有设置 User-Agent，使用默认值
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.options.UserAgent)
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	endTime := time.Now()

	// 构建结果
	result := &types.HTTPResult{
		Method:          method,
		URL:             url,
		StatusCode:      resp.StatusCode,
		Status:          resp.Status,
		Proto:           resp.Proto,
		ContentLength:   resp.ContentLength,
		Headers:         make(map[string][]string),
		Body:            respBody,
		StartTime:       startTime,
		EndTime:         endTime,
		Duration:        endTime.Sub(startTime),
		TLSUsed:         resp.TLS != nil,
		Uncompressed:    resp.Uncompressed,
		ContentType:     resp.Header.Get("Content-Type"),
		TransferredSize: int64(len(respBody)),
	}

	// 复制响应头
	for key, values := range resp.Header {
		result.Headers[key] = values
	}

	return result, nil
}

// Get 执行 GET 请求
func (c *Client) Get(ctx context.Context, url string, headers map[string]string) (*types.HTTPResult, error) {
	return c.Request(ctx, "GET", url, nil, headers)
}

// Post 执行 POST 请求
func (c *Client) Post(ctx context.Context, url string, body []byte, headers map[string]string) (*types.HTTPResult, error) {
	return c.Request(ctx, "POST", url, body, headers)
}

// Put 执行 PUT 请求
func (c *Client) Put(ctx context.Context, url string, body []byte, headers map[string]string) (*types.HTTPResult, error) {
	return c.Request(ctx, "PUT", url, body, headers)
}

// Delete 执行 DELETE 请求
func (c *Client) Delete(ctx context.Context, url string, headers map[string]string) (*types.HTTPResult, error) {
	return c.Request(ctx, "DELETE", url, nil, headers)
}

// Patch 执行 PATCH 请求
func (c *Client) Patch(ctx context.Context, url string, body []byte, headers map[string]string) (*types.HTTPResult, error) {
	return c.Request(ctx, "PATCH", url, body, headers)
}

// Head 执行 HEAD 请求
func (c *Client) Head(ctx context.Context, url string, headers map[string]string) (*types.HTTPResult, error) {
	return c.Request(ctx, "HEAD", url, nil, headers)
}

// Options 执行 OPTIONS 请求
func (c *Client) Options(ctx context.Context, url string, headers map[string]string) (*types.HTTPResult, error) {
	return c.Request(ctx, "OPTIONS", url, nil, headers)
}

// Benchmark 执行性能测试
func (c *Client) Benchmark(ctx context.Context, method, url string, body []byte, headers map[string]string, count int) (*types.HTTPBenchmarkResult, error) {
	if count <= 0 {
		count = 1
	}

	results := make([]*types.HTTPResult, 0, count)
	var totalDuration time.Duration
	var successCount, failureCount int

	startTime := time.Now()

	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		result, err := c.Request(ctx, method, url, body, headers)
		if err != nil {
			failureCount++
			continue
		}

		results = append(results, result)
		totalDuration += result.Duration

		if result.StatusCode >= 200 && result.StatusCode < 300 {
			successCount++
		} else {
			failureCount++
		}
	}

	endTime := time.Now()

	// 计算统计信息
	var minDuration, maxDuration, avgDuration time.Duration
	if len(results) > 0 {
		minDuration = results[0].Duration
		maxDuration = results[0].Duration
		avgDuration = totalDuration / time.Duration(len(results))

		for _, result := range results {
			if result.Duration < minDuration {
				minDuration = result.Duration
			}
			if result.Duration > maxDuration {
				maxDuration = result.Duration
			}
		}
	}

	benchResult := &types.HTTPBenchmarkResult{
		Method:         method,
		URL:            url,
		TotalRequests:  count,
		SuccessCount:   successCount,
		FailureCount:   failureCount,
		TotalDuration:  endTime.Sub(startTime),
		MinDuration:    minDuration,
		MaxDuration:    maxDuration,
		AvgDuration:    avgDuration,
		RequestsPerSec: float64(count) / endTime.Sub(startTime).Seconds(),
	}

	return benchResult, nil
}

// Close 关闭客户端 (当前无需实际操作)
func (c *Client) Close() error {
	return nil
}
