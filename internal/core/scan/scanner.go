// Package scan 提供端口扫描功能
//
// 本模块实现了多种端口扫描方式:
// - TCP Connect 扫描（不需要特权）
// - TCP SYN 扫描（需要 root 权限）
// - UDP 扫描
// - 服务识别和版本探测
//
// 依赖:
// - net: 标准网络库
// - pkg/types: 类型定义
//
// 使用示例:
//
//	scanner := scan.NewTCPScanner()
//	result, err := scanner.Scan(ctx, "192.168.1.1", opts)
//
// 作者: Catsayer
package scan

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"net"
	"sync"
	"time"

	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/pkg/errors"
	"github.com/catsayer/ntx/pkg/types"
	"golang.org/x/sync/semaphore"
)

// Scanner 定义端口扫描器接口
type Scanner interface {
	// Scan 执行端口扫描
	Scan(ctx context.Context, target string, opts types.ScanOptions) (*types.ScanResult, error)

	// ScanStream 返回实时扫描结果的 Channel
	ScanStream(ctx context.Context, target string, opts types.ScanOptions) (<-chan *types.ScanPort, error)
}

// TCPScanner TCP Connect 扫描器实现
type TCPScanner struct {
	timeout time.Duration
}

// NewTCPScanner 创建新的 TCP 扫描器
func NewTCPScanner() *TCPScanner {
	return &TCPScanner{
		timeout: types.DefaultScanTimeout,
	}
}

// Scan 执行 TCP Connect 扫描
//
// 参数:
//
//	ctx: 上下文，用于超时控制和取消操作
//	target: 目标主机（域名或 IP）
//	opts: 扫描选项
//
// 返回:
//
//	*types.ScanResult: 扫描结果
//	error: 错误信息
//
// 错误类型:
//   - ErrInvalidTarget: 目标地址无效
//   - ErrTimeout: 扫描超时
func (s *TCPScanner) Scan(ctx context.Context, target string, opts types.ScanOptions) (*types.ScanResult, error) {
	logger.Info("开始 TCP 扫描",
		zap.String("target", target),
		zap.Int("ports", len(opts.Ports)),
		zap.Int("concurrency", opts.Concurrency),
	)

	startTime := time.Now()

	// 解析目标主机
	ip, err := resolveTarget(target)
	if err != nil {
		return nil, fmt.Errorf("解析目标失败: %w", err)
	}

	result := &types.ScanResult{
		Target:    target,
		IP:        ip,
		Ports:     make([]*types.ScanPort, 0),
		StartTime: startTime,
	}

	// 创建扫描结果 channel
	portCh := make(chan *types.ScanPort, len(opts.Ports))

	// 使用信号量控制并发
	sem := semaphore.NewWeighted(int64(opts.Concurrency))
	var wg sync.WaitGroup

	// 并发扫描所有端口
	for _, port := range opts.Ports {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()

			// 获取信号量
			if err := sem.Acquire(ctx, 1); err != nil {
				logger.Error("获取信号量失败", zap.Error(err))
				return
			}
			defer sem.Release(1)

			// 扫描单个端口
			scanPort := s.scanPort(ctx, ip, p, opts.Timeout)

			// 服务识别
			if opts.ServiceDetect && scanPort.State == types.PortOpen {
				scanPort.Service = identifyService(p)
			}

			portCh <- scanPort
		}(port)
	}

	// 等待所有扫描完成
	go func() {
		wg.Wait()
		close(portCh)
	}()

	// 收集结果
	for scanPort := range portCh {
		result.Ports = append(result.Ports, scanPort)
	}

	result.EndTime = time.Now()
	result.Summary = calculateSummary(result)

	logger.Info("TCP 扫描完成",
		zap.String("target", target),
		zap.Int("total", result.Summary.TotalPorts),
		zap.Int("open", result.Summary.OpenPorts),
		zap.Duration("duration", result.Summary.Duration),
	)

	return result, nil
}

// ScanStream 返回实时扫描结果的 Channel
func (s *TCPScanner) ScanStream(ctx context.Context, target string, opts types.ScanOptions) (<-chan *types.ScanPort, error) {
	// 解析目标主机
	ip, err := resolveTarget(target)
	if err != nil {
		return nil, fmt.Errorf("解析目标失败: %w", err)
	}

	portCh := make(chan *types.ScanPort, opts.Concurrency)

	// 启动扫描协程
	go func() {
		defer close(portCh)

		sem := semaphore.NewWeighted(int64(opts.Concurrency))
		var wg sync.WaitGroup

		for _, port := range opts.Ports {
			wg.Add(1)
			go func(p int) {
				defer wg.Done()

				if err := sem.Acquire(ctx, 1); err != nil {
					return
				}
				defer sem.Release(1)

				scanPort := s.scanPort(ctx, ip, p, opts.Timeout)

				if opts.ServiceDetect && scanPort.State == types.PortOpen {
					scanPort.Service = identifyService(p)
				}

				select {
				case portCh <- scanPort:
				case <-ctx.Done():
					return
				}
			}(port)
		}

		wg.Wait()
	}()

	return portCh, nil
}

// scanPort 扫描单个端口
func (s *TCPScanner) scanPort(ctx context.Context, ip net.IP, port int, timeout time.Duration) *types.ScanPort {
	startTime := time.Now()

	scanPort := &types.ScanPort{
		IP:    ip,
		Port:  port,
		Proto: "tcp",
		State: types.PortClosed,
	}

	// 设置超时
	d := net.Dialer{Timeout: timeout}

	// 尝试连接
	conn, err := d.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", ip.String(), port))

	scanPort.ResponseTime = time.Since(startTime)

	if err != nil {
		// 判断错误类型
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			scanPort.State = types.PortFiltered
		} else {
			scanPort.State = types.PortClosed
		}
		scanPort.Error = err
		return scanPort
	}

	// 连接成功，端口开放
	defer conn.Close()
	scanPort.State = types.PortOpen

	return scanPort
}

// resolveTarget 解析目标主机名到 IP 地址
func resolveTarget(target string) (net.IP, error) {
	// 尝试直接解析为 IP
	if ip := net.ParseIP(target); ip != nil {
		return ip, nil
	}

	// 尝试 DNS 解析
	ips, err := net.LookupIP(target)
	if err != nil {
		return nil, errors.ErrInvalidTarget
	}

	if len(ips) == 0 {
		return nil, errors.ErrInvalidTarget
	}

	// 返回第一个 IPv4 地址
	for _, ip := range ips {
		if ip.To4() != nil {
			return ip, nil
		}
	}

	// 如果没有 IPv4，返回第一个 IP
	return ips[0], nil
}

// identifyService 根据端口号识别常见服务
func identifyService(port int) string {
	serviceMap := map[int]string{
		21:    "ftp",
		22:    "ssh",
		23:    "telnet",
		25:    "smtp",
		53:    "dns",
		80:    "http",
		110:   "pop3",
		143:   "imap",
		443:   "https",
		445:   "smb",
		3306:  "mysql",
		3389:  "rdp",
		5432:  "postgresql",
		5900:  "vnc",
		6379:  "redis",
		8080:  "http-proxy",
		8443:  "https-alt",
		9200:  "elasticsearch",
		27017: "mongodb",
	}

	if service, ok := serviceMap[port]; ok {
		return service
	}
	return "unknown"
}

// calculateSummary 计算扫描统计信息
func calculateSummary(result *types.ScanResult) *types.ScanSummary {
	summary := &types.ScanSummary{
		TotalPorts: len(result.Ports),
		Duration:   result.EndTime.Sub(result.StartTime),
	}

	for _, port := range result.Ports {
		switch port.State {
		case types.PortOpen:
			summary.OpenPorts++
		case types.PortClosed:
			summary.ClosedPorts++
		case types.PortFiltered:
			summary.FilteredPorts++
		}
	}

	return summary
}
