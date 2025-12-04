// Package ping 提供 HTTP Ping 功能
//
// HTTP Ping 通过发送 HTTP 请求来测试 Web 服务的可达性和响应时间
//
// 作者: Catsayer
package ping

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/catsayer/ntx/pkg/errors"
	"github.com/catsayer/ntx/pkg/types"
)

// HTTPPinger HTTP Ping 实现
type HTTPPinger struct {
	client *http.Client
}

// NewHTTPPinger 创建 HTTP Pinger
func NewHTTPPinger() *HTTPPinger {
	return &HTTPPinger{
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// 不跟随重定向
				return http.ErrUseLastResponse
			},
		},
	}
}

// Ping 执行 HTTP Ping
func (p *HTTPPinger) Ping(target string, opts *types.PingOptions) (*types.PingResult, error) {
	// 验证参数
	if target == "" {
		return nil, errors.ErrInvalidHost
	}
	if opts == nil {
		opts = types.DefaultPingOptions()
	}

	// 解析 URL
	targetURL, err := p.parseURL(target, opts)
	if err != nil {
		return nil, err
	}

	// 解析主机信息
	hostInfo, err := p.resolveHost(targetURL.Hostname(), opts.IPVersion)
	if err != nil {
		return nil, errors.NewNetworkError("resolve", target, err)
	}

	// 设置客户端超时
	p.client.Timeout = opts.Timeout

	// 创建结果对象
	result := &types.PingResult{
		Target: &types.Host{
			Hostname:  targetURL.Hostname(),
			IP:        hostInfo.IP,
			IPVersion: hostInfo.IPVersion,
			Port:      p.getPort(targetURL),
		},
		Protocol: types.ProtocolHTTP,
		Replies:  make([]*types.PingReply, 0, opts.Count),
		Statistics: &types.Statistics{},
		Context: &types.ExecutionContext{
			StartTime: time.Now(),
		},
		Status: types.StatusSuccess,
	}

	// 如果是 HTTPS，更新协议
	if targetURL.Scheme == "https" {
		result.Protocol = types.ProtocolHTTPS
	}

	// 获取主机名
	hostname, _ := os.Hostname()
	result.Context.Hostname = hostname

	// 执行 Ping
	ctx := context.Background()
	for i := 0; i < opts.Count; i++ {
		reply := p.pingOnce(ctx, targetURL, i+1, opts)
		result.AddReply(reply)

		// 如果不是最后一次，等待间隔
		if i < opts.Count-1 {
			time.Sleep(opts.Interval)
		}
	}

	// 更新上下文
	result.Context.EndTime = time.Now()
	result.Context.Duration = result.Context.EndTime.Sub(result.Context.StartTime)

	// 更新统计信息
	result.UpdateStatistics()

	// 判断整体状态
	if result.Statistics.Received == 0 {
		result.Status = types.StatusFailure
	} else if result.Statistics.Received < result.Statistics.Sent {
		result.Status = types.StatusTimeout
	}

	return result, nil
}

// pingOnce 执行一次 HTTP Ping
func (p *HTTPPinger) pingOnce(ctx context.Context, targetURL *url.URL, seq int, opts *types.PingOptions) *types.PingReply {
	reply := &types.PingReply{
		Seq:    seq,
		From:   targetURL.Host,
		Bytes:  0,
		TTL:    0,
		Time:   time.Now(),
		Status: types.StatusSuccess,
	}

	// 创建请求
	method := opts.HTTPMethod
	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL.String(), nil)
	if err != nil {
		reply.Status = types.StatusFailure
		reply.Error = err.Error()
		return reply
	}

	// 设置 User-Agent
	req.Header.Set("User-Agent", "NTX/0.1.0")

	// 记录开始时间
	start := time.Now()

	// 发送请求
	resp, err := p.client.Do(req)

	// 记录往返时间
	reply.RTT = time.Since(start)

	if err != nil {
		// 判断错误类型
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			reply.Status = types.StatusTimeout
		} else {
			reply.Status = types.StatusFailure
			reply.Error = err.Error()
		}
		return reply
	}

	// 读取响应体大小
	bodyBytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		reply.Status = types.StatusFailure
		reply.Error = err.Error()
		return reply
	}

	// 设置响应信息
	reply.Bytes = len(bodyBytes)
	reply.TTL = opts.TTL
	reply.From = fmt.Sprintf("%s (status: %d)", targetURL.Host, resp.StatusCode)

	// 检查 HTTP 状态码
	if resp.StatusCode >= 400 {
		reply.Status = types.StatusFailure
		reply.Error = fmt.Sprintf("HTTP %d %s", resp.StatusCode, resp.Status)
	}

	return reply
}

// parseURL 解析 URL
func (p *HTTPPinger) parseURL(target string, opts *types.PingOptions) (*url.URL, error) {
	// 如果没有 scheme，添加默认的
	if !strings.Contains(target, "://") {
		if opts.Port == 443 {
			target = "https://" + target
		} else {
			target = "http://" + target
		}
	}

	// 解析 URL
	u, err := url.Parse(target)
	if err != nil {
		return nil, errors.ErrInvalidHost
	}

	// 设置路径
	if u.Path == "" || u.Path == "/" {
		if opts.HTTPPath != "" {
			u.Path = opts.HTTPPath
		}
	}

	return u, nil
}

// getPort 获取端口号
func (p *HTTPPinger) getPort(u *url.URL) int {
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			return 443
		}
		return 80
	}

	var portNum int
	fmt.Sscanf(port, "%d", &portNum)
	return portNum
}

// resolveHost 解析主机名
func (p *HTTPPinger) resolveHost(host string, ipVersion types.IPVersion) (*types.Host, error) {
	// 检查是否已经是 IP 地址
	if ip := net.ParseIP(host); ip != nil {
		ver := types.IPv4
		if ip.To4() == nil {
			ver = types.IPv6
		}
		return &types.Host{
			Hostname:  host,
			IP:        ip.String(),
			IPVersion: ver,
		}, nil
	}

	// 解析域名
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, errors.ErrDNSResolution
	}

	if len(ips) == 0 {
		return nil, errors.ErrNoAddress
	}

	// 选择合适的 IP
	var selectedIP net.IP
	for _, ip := range ips {
		if ipVersion == types.IPv4 && ip.To4() != nil {
			selectedIP = ip
			break
		} else if ipVersion == types.IPv6 && ip.To4() == nil {
			selectedIP = ip
			break
		} else if ipVersion == types.IPvAny {
			selectedIP = ip
			break
		}
	}

	if selectedIP == nil {
		return nil, errors.ErrNoAddress
	}

	ver := types.IPv4
	if selectedIP.To4() == nil {
		ver = types.IPv6
	}

	return &types.Host{
		Hostname:  host,
		IP:        selectedIP.String(),
		IPVersion: ver,
	}, nil
}

// Close 关闭资源
func (p *HTTPPinger) Close() error {
	// HTTP Pinger 不需要关闭资源
	return nil
}