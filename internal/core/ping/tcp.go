// Package ping 提供 TCP Ping 功能
//
// # TCP Ping 通过建立 TCP 连接来测试目标主机的可达性和响应时间
//
// 作者: Catsayer
package ping

import (
	"context"
	"fmt"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/catsayer/ntx/pkg/errors"
	"github.com/catsayer/ntx/pkg/netutil"
	"github.com/catsayer/ntx/pkg/types"
)

// TCPPinger TCP Ping 实现
type TCPPinger struct {
	dialer *net.Dialer
}

// NewTCPPinger 创建 TCP Pinger
func NewTCPPinger(opts ...*types.PingOptions) *TCPPinger {
	var cfg *types.PingOptions
	if len(opts) > 0 {
		cfg = opts[0]
	}
	if cfg == nil {
		cfg = types.DefaultPingOptions()
	}

	dialer := &net.Dialer{}
	if cfg.TOS > 0 {
		// 设置 Dialer 的 Control 函数来设置 TOS
		dialer.Control = func(network, address string, c syscall.RawConn) error {
			var controlErr error
			err := c.Control(func(fd uintptr) {
				// 在 Linux/macOS 上设置 IPv4 TOS
				controlErr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_TOS, cfg.TOS)
			})
			if err != nil {
				return err
			}
			return controlErr
		}
	}

	return &TCPPinger{
		dialer: dialer,
	}
}

// Ping 执行 TCP Ping
func (p *TCPPinger) Ping(ctx context.Context, target string, opts *types.PingOptions) (*types.PingResult, error) {
	if target == "" {
		return nil, errors.ErrInvalidHost
	}
	if opts == nil {
		opts = types.DefaultPingOptions()
	}

	host, port, err := p.parseTarget(target, opts.Port)
	if err != nil {
		return nil, err
	}
	if host == "" {
		return nil, errors.ErrInvalidHost
	}

	hostInfo, err := netutil.ResolveHost(host, opts.IPVersion)
	if err != nil {
		return nil, errors.NewNetworkError("resolve", target, err)
	}

	result := &types.PingResult{
		Target: &types.Host{
			Hostname:  host,
			IP:        hostInfo.IP,
			IPVersion: hostInfo.IPVersion,
			Port:      port,
		},
		Protocol:   types.ProtocolTCP,
		Replies:    make([]*types.PingReply, 0, opts.Count),
		Statistics: &types.Statistics{},
		Context: &types.ExecutionContext{
			StartTime: time.Now(),
		},
		Status: types.StatusSuccess,
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

		reply := p.pingOnce(ctx, hostInfo.IP, port, i+1, opts)
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

// PingStream 执行实时 TCP Ping
func (p *TCPPinger) PingStream(ctx context.Context, target string, opts *types.PingOptions) (<-chan *types.PingReply, error) {
	if target == "" {
		return nil, errors.ErrInvalidHost
	}
	if opts == nil {
		opts = types.DefaultPingOptions()
	}

	host, port, err := p.parseTarget(target, opts.Port)
	if err != nil {
		return nil, err
	}
	if host == "" {
		return nil, errors.ErrInvalidHost
	}

	hostInfo, err := netutil.ResolveHost(host, opts.IPVersion)
	if err != nil {
		return nil, errors.NewNetworkError("resolve", target, err)
	}

	replyChan := make(chan *types.PingReply)

	go func() {
		defer close(replyChan)

		for i := 0; opts.Count <= 0 || i < opts.Count; i++ {
			select {
			case <-ctx.Done():
				return
			default:
			}

			reply := p.pingOnce(ctx, hostInfo.IP, port, i+1, opts)
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

// pingOnce 执行一次 TCP Ping
func (p *TCPPinger) pingOnce(ctx context.Context, ip string, port, seq int, opts *types.PingOptions) *types.PingReply {
	reply := &types.PingReply{
		Seq:    seq,
		From:   ip,
		Time:   time.Now(),
		Status: types.StatusSuccess,
	}

	dialCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	start := time.Now()
	addr := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
	conn, err := p.dialer.DialContext(dialCtx, "tcp", addr)
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
	conn.Close()

	reply.Bytes = types.TCPHandshakeBytes
	return reply
}

// parseTarget 解析目标地址和端口
func (p *TCPPinger) parseTarget(target string, defaultPort int) (string, int, error) {
	host, portStr, err := net.SplitHostPort(target)
	if err != nil {
		if defaultPort <= 0 {
			defaultPort = types.DefaultTCPPort
		}
		return target, defaultPort, nil
	}

	port := defaultPort
	if portStr != "" {
		var portNum int
		_, err = fmt.Sscanf(portStr, "%d", &portNum)
		if err != nil || portNum < types.MinPort || portNum > types.MaxPort {
			return "", 0, errors.ErrInvalidPort
		}
		port = portNum
	}
	return host, port, nil
}

// Close 关闭资源
func (p *TCPPinger) Close() error {
	// TCP Pinger 不需要关闭资源
	return nil
}
