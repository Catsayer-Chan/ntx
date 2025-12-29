// Package ping 提供 ICMP Ping 功能
//
// ICMP Ping 使用 ICMP Echo Request/Reply 来测试网络连通性
// 需要 root 权限或 CAP_NET_RAW 能力
//
// 作者: Catsayer
package ping

import (
	"context"
	stdErrors "errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/pkg/errors"
	"github.com/catsayer/ntx/pkg/netutil"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
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
func NewICMPPinger(opts *types.PingOptions) (*ICMPPinger, error) {
	p := &ICMPPinger{
		id: os.Getpid() & types.ICMPIDMask,
	}

	// 打开 ICMPv4 连接
	conn4, err := icmp.ListenPacket("ip4:icmp", "")
	if err != nil {
		return nil, errors.NewPermissionError("icmp ping", "raw socket", getPermissionHint())
	}
	p.conn4 = conn4

	// 尝试打开 ICMPv6 连接（可选）
	conn6, err := icmp.ListenPacket("ip6:ipv6-icmp", "")
	if err != nil {
		p.conn6 = nil
	} else {
		p.conn6 = conn6
	}

	// 设置 TOS (仅 IPv4 支持)
	if opts.TOS > 0 && p.conn4 != nil {
		if err := p.conn4.IPv4PacketConn().SetTOS(opts.TOS); err != nil {
			logger.Warn("无法为 IPv4 设置 TOS", zap.Error(err))
		}
	}

	return p, nil
}

// getPermissionHint 根据操作系统返回权限提示
func getPermissionHint() string {
	switch runtime.GOOS {
	case "darwin":
		return "需要 root 权限，请使用: sudo ntx ping <target>"
	case "linux":
		return "需要 root 权限或 CAP_NET_RAW 能力，请使用: sudo ntx ping <target>"
	case "windows":
		return "需要管理员权限，请以管理员身份运行"
	default:
		return "需要特权"
	}
}

// Ping 执行 ICMP Ping
func (p *ICMPPinger) Ping(ctx context.Context, target string, opts *types.PingOptions) (*types.PingResult, error) {
	if target == "" {
		return nil, errors.ErrInvalidHost
	}
	if opts == nil {
		opts = types.DefaultPingOptions()
	}

	hostInfo, err := netutil.ResolveHost(target, opts.IPVersion)
	if err != nil {
		return nil, errors.NewNetworkError("resolve", target, err)
	}

	if hostInfo.IPVersion == types.IPv4 && p.conn4 == nil {
		return nil, errors.NewPermissionError("icmp ping", "ipv4", "connection not available")
	}
	if hostInfo.IPVersion == types.IPv6 && p.conn6 == nil {
		return nil, errors.NewPermissionError("icmp ping", "ipv6", "connection not available")
	}

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

		reply := p.pingOnce(ctx, hostInfo.IP, i+1, opts)
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

// PingStream 执行实时 ICMP Ping
func (p *ICMPPinger) PingStream(ctx context.Context, target string, opts *types.PingOptions) (<-chan *types.PingReply, error) {
	if target == "" {
		return nil, errors.ErrInvalidHost
	}
	if opts == nil {
		opts = types.DefaultPingOptions()
	}

	hostInfo, err := netutil.ResolveHost(target, opts.IPVersion)
	if err != nil {
		return nil, errors.NewNetworkError("resolve", target, err)
	}

	if hostInfo.IPVersion == types.IPv4 && p.conn4 == nil {
		return nil, errors.NewPermissionError("icmp ping", "ipv4", "connection not available")
	}
	if hostInfo.IPVersion == types.IPv6 && p.conn6 == nil {
		return nil, errors.NewPermissionError("icmp ping", "ipv6", "connection not available")
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

			reply := p.pingOnce(ctx, hostInfo.IP, i+1, opts)
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

	dst, err := net.ResolveIPAddr("ip", ip)
	if err != nil {
		reply.Status = types.StatusFailure
		reply.Error = err.Error()
		return reply
	}

	var conn *icmp.PacketConn
	var msgType icmp.Type

	if dst.IP.To4() != nil {
		conn = p.conn4
		msgType = ipv4.ICMPTypeEcho
	} else {
		conn = p.conn6
		msgType = ipv6.ICMPTypeEchoRequest
	}

	msg := &icmp.Message{
		Type: msgType,
		Code: 0,
		Body: &icmp.Echo{
			ID:   p.id,
			Seq:  seq,
			Data: make([]byte, opts.Size),
		},
	}

	for i := range msg.Body.(*icmp.Echo).Data {
		msg.Body.(*icmp.Echo).Data[i] = byte(rand.Intn(256))
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		reply.Status = types.StatusFailure
		reply.Error = err.Error()
		return reply
	}

	deadline := time.Now().Add(opts.Timeout)
	conn.SetReadDeadline(deadline)
	conn.SetWriteDeadline(deadline)

	start := time.Now()

	if err := ctx.Err(); err != nil {
		reply.Status = types.StatusFailure
		reply.Error = err.Error()
		return reply
	}

	_, err = conn.WriteTo(msgBytes, dst)
	if err != nil {
		reply.Status = types.StatusFailure
		reply.Error = err.Error()
		return reply
	}

	recvBuf := make([]byte, types.StandardMTU)

	cancelRead := make(chan struct{})
	defer close(cancelRead)
	go func() {
		select {
		case <-ctx.Done():
			// 提前唤醒阻塞的 ReadFrom
			_ = conn.SetReadDeadline(time.Now())
		case <-cancelRead:
		}
	}()

	for {
		if err := ctx.Err(); err != nil {
			reply.Status = types.StatusFailure
			reply.Error = err.Error()
			return reply
		}

		n, peer, err := conn.ReadFrom(recvBuf)
		if err != nil {
			if ctx.Err() != nil {
				reply.Status = types.StatusFailure
				reply.Error = ctx.Err().Error()
				return reply
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				reply.Status = types.StatusTimeout
			} else {
				reply.Status = types.StatusFailure
				reply.Error = err.Error()
			}
			return reply
		}

		rtt := time.Since(start)

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

		switch rm.Type {
		case ipv4.ICMPTypeEchoReply, ipv6.ICMPTypeEchoReply:
			if echo, ok := rm.Body.(*icmp.Echo); ok {
				if echo.ID == p.id && echo.Seq == seq {
					reply.RTT = rtt
					reply.From = peer.String()
					reply.Bytes = len(msgBytes)
					// TTL 从 IP 包头获取,默认设置为配置的 TTL
					reply.TTL = opts.TTL

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

		if time.Now().After(deadline) {
			reply.Status = types.StatusTimeout
			return reply
		}
	}
}

// Close 关闭资源
func (p *ICMPPinger) Close() error {
	var errs []error
	if p.conn4 != nil {
		if e := p.conn4.Close(); e != nil {
			errs = append(errs, fmt.Errorf("close IPv4 conn: %w", e))
		}
	}
	if p.conn6 != nil {
		if e := p.conn6.Close(); e != nil {
			errs = append(errs, fmt.Errorf("close IPv6 conn: %w", e))
		}
	}
	return stdErrors.Join(errs...)
}
