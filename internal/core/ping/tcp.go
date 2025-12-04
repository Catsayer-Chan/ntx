// Package ping 提供 TCP Ping 功能
//
// TCP Ping 通过建立 TCP 连接来测试目标主机的可达性和响应时间
//
// 作者: Catsayer
package ping

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/catsayer/ntx/pkg/errors"
	"github.com/catsayer/ntx/pkg/types"
)

// TCPPinger TCP Ping 实现
type TCPPinger struct {
	dialer *net.Dialer
}

// NewTCPPinger 创建 TCP Pinger
func NewTCPPinger() *TCPPinger {
	return &TCPPinger{
		dialer: &net.Dialer{},
	}
}

// Ping 执行 TCP Ping
func (p *TCPPinger) Ping(target string, opts *types.PingOptions) (*types.PingResult, error) {
	// 验证参数
	if target == "" {
		return nil, errors.ErrInvalidHost
	}
	if opts == nil {
		opts = types.DefaultPingOptions()
	}

	// 解析目标地址
	host, port, err := p.parseTarget(target, opts.Port)
	if err != nil {
		return nil, err
	}

	// 解析主机名
	hostInfo, err := p.resolveHost(host, opts.IPVersion)
	if err != nil {
		return nil, errors.NewNetworkError("resolve", target, err)
	}

	// 创建结果对象
	result := &types.PingResult{
		Target: &types.Host{
			Hostname:  host,
			IP:        hostInfo.IP,
			IPVersion: hostInfo.IPVersion,
			Port:      port,
		},
		Protocol: types.ProtocolTCP,
		Replies:  make([]*types.PingReply, 0, opts.Count),
		Statistics: &types.Statistics{},
		Context: &types.ExecutionContext{
			StartTime: time.Now(),
		},
		Status: types.StatusSuccess,
	}

	// 获取主机名
	hostname, _ := os.Hostname()
	result.Context.Hostname = hostname

	// 执行 Ping
	ctx := context.Background()
	for i := 0; i < opts.Count; i++ {
		reply := p.pingOnce(ctx, hostInfo.IP, port, i+1, opts)
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

// pingOnce 执行一次 TCP Ping
func (p *TCPPinger) pingOnce(ctx context.Context, ip string, port, seq int, opts *types.PingOptions) *types.PingReply {
	reply := &types.PingReply{
		Seq:    seq,
		From:   ip,
		Bytes:  0,
		TTL:    0,
		Time:   time.Now(),
		Status: types.StatusSuccess,
	}

	// 创建带超时的上下文
	dialCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	// 记录开始时间
	start := time.Now()

	// 连接目标
	addr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := p.dialer.DialContext(dialCtx, "tcp", addr)

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

	// 关闭连接
	conn.Close()

	// TCP 连接成功，设置响应信息
	reply.Bytes = 40 // TCP SYN+ACK 大约 40 字节
	reply.TTL = opts.TTL

	return reply
}

// parseTarget 解析目标地址和端口
func (p *TCPPinger) parseTarget(target string, defaultPort int) (string, int, error) {
	// 尝试分离主机和端口
	host, portStr, err := net.SplitHostPort(target)
	if err != nil {
		// 没有端口，使用默认端口
		if defaultPort <= 0 {
			defaultPort = 80 // TCP Ping 默认端口
		}
		return target, defaultPort, nil
	}

	// 解析端口
	port := defaultPort
	if portStr != "" {
		var portNum int
		_, err = fmt.Sscanf(portStr, "%d", &portNum)
		if err != nil || portNum <= 0 || portNum > 65535 {
			return "", 0, errors.ErrInvalidPort
		}
		port = portNum
	}

	return host, port, nil
}

// resolveHost 解析主机名
func (p *TCPPinger) resolveHost(host string, ipVersion types.IPVersion) (*types.Host, error) {
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
func (p *TCPPinger) Close() error {
	// TCP Pinger 不需要关闭资源
	return nil
}