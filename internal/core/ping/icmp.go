// Package ping 提供 ICMP Ping 功能
//
// ICMP Ping 使用 ICMP Echo Request/Reply 来测试网络连通性
// 需要 root 权限或 CAP_NET_RAW 能力
//
// 作者: Catsayer
package ping

import (
	"context"
	"math/rand"
	"net"
	"os"
	"runtime"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	"github.com/catsayer/ntx/pkg/errors"
	"github.com/catsayer/ntx/pkg/types"
)

const (
	// ProtocolICMP ICMP 协议号
	ProtocolICMP = 1
	// ProtocolIPv6ICMP ICMPv6 协议号
	ProtocolIPv6ICMP = 58
)

// ICMPPinger ICMP Ping 实现
type ICMPPinger struct {
	conn4 *icmp.PacketConn
	conn6 *icmp.PacketConn
	id    int
}

// NewICMPPinger 创建 ICMP Pinger
func NewICMPPinger() (*ICMPPinger, error) {
	p := &ICMPPinger{
		id: os.Getpid() & 0xffff,
	}

	// 尝试打开 ICMPv4 连接
	network := "udp4"
	if runtime.GOOS == "windows" {
		network = "ip4:icmp"
	}

	conn4, err := icmp.ListenPacket(network, "0.0.0.0")
	if err != nil {
		return nil, errors.NewPermissionError("icmp ping", "raw socket",
			"需要 root 权限或 CAP_NET_RAW 能力")
	}
	p.conn4 = conn4

	// 尝试打开 ICMPv6 连接
	network6 := "udp6"
	if runtime.GOOS == "windows" {
		network6 = "ip6:ipv6-icmp"
	}

	conn6, err := icmp.ListenPacket(network6, "::")
	if err != nil {
		// ICMPv6 可能不可用，不视为错误
		p.conn6 = nil
	} else {
		p.conn6 = conn6
	}

	return p, nil
}

// Ping 执行 ICMP Ping
func (p *ICMPPinger) Ping(target string, opts *types.PingOptions) (*types.PingResult, error) {
	// 验证参数
	if target == "" {
		return nil, errors.ErrInvalidHost
	}
	if opts == nil {
		opts = types.DefaultPingOptions()
	}

	// 解析主机信息
	hostInfo, err := p.resolveHost(target, opts.IPVersion)
	if err != nil {
		return nil, errors.NewNetworkError("resolve", target, err)
	}

	// 检查连接是否可用
	if hostInfo.IPVersion == types.IPv4 && p.conn4 == nil {
		return nil, errors.NewPermissionError("icmp ping", "ipv4", "connection not available")
	}
	if hostInfo.IPVersion == types.IPv6 && p.conn6 == nil {
		return nil, errors.NewPermissionError("icmp ping", "ipv6", "connection not available")
	}

	// 创建结果对象
	result := &types.PingResult{
		Target: &types.Host{
			Hostname:  target,
			IP:        hostInfo.IP,
			IPVersion: hostInfo.IPVersion,
		},
		Protocol:   types.ProtocolICMP,
		Replies:    make([]*types.PingReply, 0, opts.Count),
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
		reply := p.pingOnce(ctx, hostInfo.IP, i+1, opts)
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

// pingOnce 执行一次 ICMP Ping
func (p *ICMPPinger) pingOnce(ctx context.Context, ip string, seq int, opts *types.PingOptions) *types.PingReply {
	reply := &types.PingReply{
		Seq:    seq,
		From:   ip,
		Bytes:  opts.Size,
		TTL:    opts.TTL,
		Time:   time.Now(),
		Status: types.StatusSuccess,
	}

	// 解析 IP 地址
	dst, err := net.ResolveIPAddr("ip", ip)
	if err != nil {
		reply.Status = types.StatusFailure
		reply.Error = err.Error()
		return reply
	}

	// 选择连接和消息类型
	var conn *icmp.PacketConn
	var msgType icmp.Type

	if dst.IP.To4() != nil {
		conn = p.conn4
		msgType = ipv4.ICMPTypeEcho
	} else {
		conn = p.conn6
		msgType = ipv6.ICMPTypeEchoRequest
	}

	// 创建 ICMP 消息
	msg := &icmp.Message{
		Type: msgType,
		Code: 0,
		Body: &icmp.Echo{
			ID:   p.id,
			Seq:  seq,
			Data: make([]byte, opts.Size),
		},
	}

	// 填充随机数据
	for i := range msg.Body.(*icmp.Echo).Data {
		msg.Body.(*icmp.Echo).Data[i] = byte(rand.Intn(256))
	}

	// 序列化消息
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		reply.Status = types.StatusFailure
		reply.Error = err.Error()
		return reply
	}

	// 设置超时
	deadline := time.Now().Add(opts.Timeout)
	conn.SetReadDeadline(deadline)
	conn.SetWriteDeadline(deadline)

	// 记录开始时间
	start := time.Now()

	// 发送 ICMP 请求
	_, err = conn.WriteTo(msgBytes, dst)
	if err != nil {
		reply.Status = types.StatusFailure
		reply.Error = err.Error()
		return reply
	}

	// 接收 ICMP 响应
	recvBuf := make([]byte, 1500)
	for {
		n, peer, err := conn.ReadFrom(recvBuf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				reply.Status = types.StatusTimeout
			} else {
				reply.Status = types.StatusFailure
				reply.Error = err.Error()
			}
			return reply
		}

		// 记录往返时间
		rtt := time.Since(start)

		// 解析响应
		var proto int
		if dst.IP.To4() != nil {
			proto = ProtocolICMP
		} else {
			proto = ProtocolIPv6ICMP
		}

		rm, err := icmp.ParseMessage(proto, recvBuf[:n])
		if err != nil {
			continue
		}

		// 检查是否是我们的响应
		switch rm.Type {
		case ipv4.ICMPTypeEchoReply, ipv6.ICMPTypeEchoReply:
			if echo, ok := rm.Body.(*icmp.Echo); ok {
				if echo.ID == p.id && echo.Seq == seq {
					reply.RTT = rtt
					reply.From = peer.String()
					reply.Bytes = len(msgBytes)

					// 尝试获取 TTL
					if dst.IP.To4() != nil {
						reply.TTL = opts.TTL // IPv4 暂时使用配置的 TTL
					} else {
						reply.TTL = opts.TTL // IPv6 暂时使用配置的 TTL
					}

					return reply
				}
			}
		case ipv4.ICMPTypeDestinationUnreachable, ipv6.ICMPTypeDestinationUnreachable:
			reply.Status = types.StatusFailure
			reply.Error = "destination unreachable"
			return reply
		case ipv4.ICMPTypeTimeExceeded, ipv6.ICMPTypeTimeExceeded:
			reply.Status = types.StatusFailure
			reply.Error = "time exceeded"
			return reply
		}

		// 检查是否超时
		if time.Now().After(deadline) {
			reply.Status = types.StatusTimeout
			return reply
		}
	}
}

// resolveHost 解析主机名
func (p *ICMPPinger) resolveHost(host string, ipVersion types.IPVersion) (*types.Host, error) {
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
func (p *ICMPPinger) Close() error {
	var err error
	if p.conn4 != nil {
		if e := p.conn4.Close(); e != nil {
			err = e
		}
	}
	if p.conn6 != nil {
		if e := p.conn6.Close(); e != nil {
			err = e
		}
	}
	return err
}
