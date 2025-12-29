// Package trace 提供 ICMP Traceroute 功能
//
// ICMP Traceroute 使用 ICMP Echo Request 和逐步增加的 TTL 来追踪路由路径
// 需要 root 权限或 CAP_NET_RAW 能力
//
// 作者: Catsayer
package trace

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
	"github.com/catsayer/ntx/pkg/netutil"
	"github.com/catsayer/ntx/pkg/types"
)

const (
	// ProtocolICMP ICMP 协议号
	ProtocolICMP = 1
	// ProtocolIPv6ICMP ICMPv6 协议号
	ProtocolIPv6ICMP = 58
)

// ICMPTracer ICMP Traceroute 实现
type ICMPTracer struct {
	conn4 *icmp.PacketConn
	conn6 *icmp.PacketConn
	id    int
}

// NewICMPTracer 创建 ICMP Tracer
func NewICMPTracer() (*ICMPTracer, error) {
	t := &ICMPTracer{
		id: os.Getpid() & 0xffff,
	}

	// 尝试打开 ICMPv4 连接
	network := "ip4:icmp"
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		network = "udp4"
	}

	conn4, err := icmp.ListenPacket(network, "0.0.0.0")
	if err != nil {
		return nil, errors.NewPermissionError("icmp traceroute", "raw socket",
			"需要 root 权限或 CAP_NET_RAW 能力")
	}
	t.conn4 = conn4

	// 尝试打开 ICMPv6 连接
	network6 := "ip6:ipv6-icmp"
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		network6 = "udp6"
	}

	conn6, err := icmp.ListenPacket(network6, "::")
	if err != nil {
		t.conn6 = nil
	} else {
		t.conn6 = conn6
	}

	return t, nil
}

// Trace 执行 ICMP Traceroute
func (t *ICMPTracer) Trace(target string, opts *types.TraceOptions) (*types.TraceResult, error) {
	// 验证参数
	if target == "" {
		return nil, errors.ErrInvalidHost
	}
	if opts == nil {
		opts = types.DefaultTraceOptions()
	}

	// 解析主机信息
	hostInfo, err := netutil.ResolveHost(target, opts.IPVersion)
	if err != nil {
		return nil, errors.NewNetworkError("resolve", target, err)
	}

	// 检查连接是否可用
	if hostInfo.IPVersion == types.IPv4 && t.conn4 == nil {
		return nil, errors.NewPermissionError("icmp traceroute", "ipv4", "connection not available")
	}
	if hostInfo.IPVersion == types.IPv6 && t.conn6 == nil {
		return nil, errors.NewPermissionError("icmp traceroute", "ipv6", "connection not available")
	}

	// 创建结果对象
	result := &types.TraceResult{
		Target: &types.Host{
			Hostname:  target,
			IP:        hostInfo.IP,
			IPVersion: hostInfo.IPVersion,
		},
		Protocol:           types.ProtocolICMP,
		Hops:               make([]*types.TraceHop, 0, opts.MaxHops),
		ReachedDestination: false,
		Context: &types.ExecutionContext{
			StartTime: time.Now(),
		},
		Status: types.StatusSuccess,
	}

	// 获取主机名
	hostname, _ := os.Hostname()
	result.Context.Hostname = hostname

	// 执行 Traceroute
	ctx := context.Background()
	for ttl := opts.FirstTTL; ttl <= opts.MaxHops; ttl++ {
		hop := t.traceHop(ctx, hostInfo.IP, ttl, opts)
		result.AddHop(hop)

		// 检查是否到达目标
		if hop.IsDestination {
			result.ReachedDestination = true
			break
		}

		// 如果所有探测都失败，继续但记录
		if hop.GetSuccessCount() == 0 {
			// 连续多跳失败可能表示路径阻塞
			if ttl > opts.FirstTTL+5 {
				lastFiveAllFailed := true
				for i := len(result.Hops) - 1; i >= 0 && i >= len(result.Hops)-5; i-- {
					if result.Hops[i].GetSuccessCount() > 0 {
						lastFiveAllFailed = false
						break
					}
				}
				if lastFiveAllFailed {
					break
				}
			}
		}
	}

	// 更新上下文
	result.Context.EndTime = time.Now()
	result.Context.Duration = result.Context.EndTime.Sub(result.Context.StartTime)

	// 判断整体状态
	if !result.ReachedDestination && result.HopCount == 0 {
		result.Status = types.StatusFailure
	} else if !result.ReachedDestination {
		result.Status = types.StatusTimeout
	}

	return result, nil
}

// traceHop 追踪单个跳
func (t *ICMPTracer) traceHop(ctx context.Context, targetIP string, ttl int, opts *types.TraceOptions) *types.TraceHop {
	hop := &types.TraceHop{
		TTL:    ttl,
		Probes: make([]*types.TraceProbe, 0, opts.Queries),
	}

	// 执行多次探测
	for i := 0; i < opts.Queries; i++ {
		probe := t.probeOnce(ctx, targetIP, ttl, i+1, opts)
		hop.Probes = append(hop.Probes, probe)

		// 记录 IP 和主机名（使用第一个成功的响应）
		if probe.Status == types.StatusSuccess && hop.IP == "" {
			hop.IP = probe.IP

			// 尝试反向 DNS 解析
			if names, err := net.LookupAddr(probe.IP); err == nil && len(names) > 0 {
				hop.Hostname = names[0]
			} else {
				hop.Hostname = probe.IP
			}

			// 检查是否为目标
			if probe.IP == targetIP {
				hop.IsDestination = true
			}
		}
	}

	return hop
}

// probeOnce 执行单次探测
func (t *ICMPTracer) probeOnce(ctx context.Context, targetIP string, ttl, seq int, opts *types.TraceOptions) *types.TraceProbe {
	probe := &types.TraceProbe{
		Seq:    seq,
		Status: types.StatusSuccess,
	}

	// 解析目标 IP
	dst, err := net.ResolveIPAddr("ip", targetIP)
	if err != nil {
		probe.Status = types.StatusFailure
		probe.Error = err.Error()
		return probe
	}

	// 选择连接和消息类型
	var conn *icmp.PacketConn
	var msgType icmp.Type
	var ipVersion int

	if dst.IP.To4() != nil {
		conn = t.conn4
		msgType = ipv4.ICMPTypeEcho
		ipVersion = 4
	} else {
		conn = t.conn6
		msgType = ipv6.ICMPTypeEchoRequest
		ipVersion = 6
	}

	// 创建 ICMP 消息
	msg := &icmp.Message{
		Type: msgType,
		Code: 0,
		Body: &icmp.Echo{
			ID:   t.id,
			Seq:  seq,
			Data: make([]byte, opts.PacketSize),
		},
	}

	// 填充随机数据
	for i := range msg.Body.(*icmp.Echo).Data {
		msg.Body.(*icmp.Echo).Data[i] = byte(rand.Intn(256))
	}

	// 序列化消息
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		probe.Status = types.StatusFailure
		probe.Error = err.Error()
		return probe
	}

	// 注意：在 UDP 模式下（Darwin/Linux），无法直接设置 TTL
	// TTL 的设置需要通过特定的系统调用，这里暂时跳过
	// 在生产环境中，可以使用原始 socket 来实现完整的 TTL 控制

	// 设置超时
	deadline := time.Now().Add(opts.Timeout)
	conn.SetReadDeadline(deadline)
	conn.SetWriteDeadline(deadline)

	// 记录开始时间
	start := time.Now()

	// 发送 ICMP 请求
	_, err = conn.WriteTo(msgBytes, dst)
	if err != nil {
		probe.Status = types.StatusFailure
		probe.Error = err.Error()
		return probe
	}

	// 接收响应
	recvBuf := make([]byte, 1500)
	for {
		n, peer, err := conn.ReadFrom(recvBuf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				probe.Status = types.StatusTimeout
			} else {
				probe.Status = types.StatusFailure
				probe.Error = err.Error()
			}
			return probe
		}

		// 记录往返时间
		rtt := time.Since(start)

		// 解析响应
		var proto int
		if ipVersion == 4 {
			proto = ProtocolICMP
		} else {
			proto = ProtocolIPv6ICMP
		}

		rm, err := icmp.ParseMessage(proto, recvBuf[:n])
		if err != nil {
			continue
		}

		// 检查响应类型
		switch rm.Type {
		case ipv4.ICMPTypeEchoReply, ipv6.ICMPTypeEchoReply:
			// 到达目标
			if echo, ok := rm.Body.(*icmp.Echo); ok {
				if echo.ID == t.id && echo.Seq == seq {
					probe.RTT = rtt
					probe.IP = peer.String()
					return probe
				}
			}
		case ipv4.ICMPTypeTimeExceeded, ipv6.ICMPTypeTimeExceeded:
			// 中间路由器响应
			probe.RTT = rtt
			probe.IP = peer.String()
			return probe
		case ipv4.ICMPTypeDestinationUnreachable, ipv6.ICMPTypeDestinationUnreachable:
			// 目标不可达
			probe.Status = types.StatusFailure
			probe.Error = "destination unreachable"
			probe.IP = peer.String()
			return probe
		}

		// 检查是否超时
		if time.Now().After(deadline) {
			probe.Status = types.StatusTimeout
			return probe
		}
	}
}

// Close 关闭资源
func (t *ICMPTracer) Close() error {
	var err error
	if t.conn4 != nil {
		if e := t.conn4.Close(); e != nil {
			err = e
		}
	}
	if t.conn6 != nil {
		if e := t.conn6.Close(); e != nil {
			err = e
		}
	}
	return err
}
