// Package ping 提供 HTTP Ping 功能
//
// # HTTP Ping 通过发送 HTTP 请求来测试 Web 服务的可达性和响应时间
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

	"github.com/catsayer/ntx/pkg/buildinfo"
	"github.com/catsayer/ntx/pkg/errors"
	"github.com/catsayer/ntx/pkg/netutil"
	"github.com/catsayer/ntx/pkg/types"
)

// HTTPPinger HTTP Ping 实现
type HTTPPinger struct {
	client *http.Client
}

// NewHTTPPinger 创建 HTTP Pinger
func NewHTTPPinger(opts *types.PingOptions) *HTTPPinger {
	return &HTTPPinger{
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// Ping 执行 HTTP Ping
func (p *HTTPPinger) Ping(ctx context.Context, target string, opts *types.PingOptions) (*types.PingResult, error) {
	if target == "" {
		return nil, errors.ErrInvalidHost
	}
	if opts == nil {
		opts = types.DefaultPingOptions()
	}

	targetURL, err := p.parseURL(target, opts)
	if err != nil {
		return nil, err
	}

	hostInfo, err := netutil.ResolveHost(targetURL.Hostname(), opts.IPVersion)
	if err != nil {
		return nil, errors.NewNetworkError("resolve", target, err)
	}

	p.client.Timeout = opts.Timeout

	result := &types.PingResult{
		Target: &types.Host{
			Hostname:  targetURL.Hostname(),
			IP:        hostInfo.IP,
			IPVersion: hostInfo.IPVersion,
			Port:      p.getPort(targetURL),
		},
		Protocol:   types.ProtocolHTTP,
		Replies:    make([]*types.PingReply, 0, opts.Count),
		Statistics: &types.Statistics{},
		Context: &types.ExecutionContext{
			StartTime: time.Now(),
		},
		Status: types.StatusSuccess,
	}

	if targetURL.Scheme == "https" {
		result.Protocol = types.ProtocolHTTPS
	}

	hostname, _ := os.Hostname()
	result.Context.Hostname = hostname

	for i := 0; i < opts.Count; i++ {
		select {
		case <-ctx.Done():
			result.Error = ctx.Err()
			result.Status = types.StatusFailure
			goto end
		default:
		}

		reply := p.pingOnce(ctx, targetURL, i+1, opts)
		result.AddReply(reply)

		if i < opts.Count-1 {
			select {
			case <-time.After(opts.Interval):
			case <-ctx.Done():
				result.Error = ctx.Err()
				result.Status = types.StatusFailure
				goto end
			}
		}
	}

end:
	result.Context.EndTime = time.Now()
	result.Context.Duration = result.Context.EndTime.Sub(result.Context.StartTime)
	result.UpdateStatistics()

	if result.Statistics.Received == 0 {
		result.Status = types.StatusFailure
		if result.Error == nil {
			result.Error = errors.ErrNoResponse
		}
	} else if result.Statistics.Received < result.Statistics.Sent {
		result.Status = types.StatusTimeout
	}

	return result, nil
}

// PingStream 执行实时 HTTP Ping
func (p *HTTPPinger) PingStream(ctx context.Context, target string, opts *types.PingOptions) (<-chan *types.PingReply, error) {
	if target == "" {
		return nil, errors.ErrInvalidHost
	}
	if opts == nil {
		opts = types.DefaultPingOptions()
	}

	targetURL, err := p.parseURL(target, opts)
	if err != nil {
		return nil, err
	}

	p.client.Timeout = opts.Timeout

	replyChan := make(chan *types.PingReply)

	go func() {
		defer close(replyChan)

		for i := 0; opts.Count <= 0 || i < opts.Count; i++ {
			select {
			case <-ctx.Done():
				return
			default:
			}

			reply := p.pingOnce(ctx, targetURL, i+1, opts)
			replyChan <- reply

			if opts.Count <= 0 || i < opts.Count-1 {
				select {
				case <-time.After(opts.Interval):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return replyChan, nil
}

// pingOnce 执行一次 HTTP Ping
func (p *HTTPPinger) pingOnce(ctx context.Context, targetURL *url.URL, seq int, opts *types.PingOptions) *types.PingReply {
	reply := &types.PingReply{
		Seq:    seq,
		From:   targetURL.Host,
		Time:   time.Now(),
		Status: types.StatusSuccess,
	}

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

	req.Header.Set("User-Agent", buildinfo.UserAgent())

	start := time.Now()
	resp, err := p.client.Do(req)
	reply.RTT = time.Since(start)

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			reply.Status = types.StatusTimeout
		} else {
			reply.Status = types.StatusFailure
			reply.Error = err.Error()
		}
		return reply
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		reply.Status = types.StatusFailure
		reply.Error = err.Error()
		return reply
	}

	reply.Bytes = len(bodyBytes)
	reply.From = fmt.Sprintf("%s (status: %d)", targetURL.Host, resp.StatusCode)

	if resp.StatusCode >= 400 {
		reply.Status = types.StatusFailure
		reply.Error = fmt.Sprintf("HTTP %d %s", resp.StatusCode, resp.Status)
	}

	return reply
}

// parseURL 解析 URL
func (p *HTTPPinger) parseURL(target string, opts *types.PingOptions) (*url.URL, error) {
	if !strings.Contains(target, "://") {
		if opts.Port == types.DefaultHTTPSPort {
			target = "https://" + target
		} else {
			target = "http://" + target
		}
	}

	u, err := url.Parse(target)
	if err != nil {
		return nil, errors.ErrInvalidHost
	}

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
			return types.DefaultHTTPSPort
		}
		return types.DefaultHTTPPort
	}

	var portNum int
	fmt.Sscanf(port, "%d", &portNum)
	return portNum
}

// Close 关闭资源
func (p *HTTPPinger) Close() error {
	// HTTP Pinger 不需要关闭资源
	return nil
}
