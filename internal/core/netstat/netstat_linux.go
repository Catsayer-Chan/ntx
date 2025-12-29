//go:build linux
// +build linux

// Package netstat 提供网络连接状态的 Linux 实现
package netstat

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/catsayer/ntx/pkg/types"
)

const (
	tcpStateEstablishedHex = 0x01
	tcpStateSynSentHex     = 0x02
	tcpStateSynRecvHex     = 0x03
	tcpStateFinWait1Hex    = 0x04
	tcpStateFinWait2Hex    = 0x05
	tcpStateTimeWaitHex    = 0x06
	tcpStateCloseHex       = 0x07
	tcpStateCloseWaitHex   = 0x08
	tcpStateLastAckHex     = 0x09
	tcpStateListenHex      = 0x0A
	tcpStateClosingHex     = 0x0B
)

type linuxReader struct{}

func newPlatformReader() platformReader {
	return &linuxReader{}
}

// getConnections 从 /proc/net 读取连接信息
func (r *linuxReader) getConnections(opts *types.NetStatOptions) ([]*types.Connection, error) {
	var connections []*types.Connection

	// 读取 TCP 连接
	if opts.Protocol == "all" || opts.Protocol == "tcp" {
		tcpConns, err := r.readTCPConnections(types.ProcNetTCP)
		if err == nil {
			connections = append(connections, tcpConns...)
		}

		tcp6Conns, err := r.readTCPConnections(types.ProcNetTCP6)
		if err == nil {
			connections = append(connections, tcp6Conns...)
		}
	}

	// 读取 UDP 连接
	if opts.Protocol == "all" || opts.Protocol == "udp" {
		udpConns, err := r.readUDPConnections(types.ProcNetUDP)
		if err == nil {
			connections = append(connections, udpConns...)
		}

		udp6Conns, err := r.readUDPConnections(types.ProcNetUDP6)
		if err == nil {
			connections = append(connections, udp6Conns...)
		}
	}

	return connections, nil
}

// getListeners 获取监听端口
func (r *linuxReader) getListeners(opts *types.NetStatOptions) ([]*types.Listener, error) {
	// 获取所有连接
	connections, err := r.getConnections(opts)
	if err != nil {
		return nil, err
	}

	// 过滤出 LISTEN 状态的连接
	listeners := make([]*types.Listener, 0)
	for _, conn := range connections {
		if conn.State == types.StateListen {
			listener := &types.Listener{
				Protocol:    conn.Protocol,
				Addr:        conn.LocalAddr,
				Port:        conn.LocalPort,
				PID:         conn.PID,
				ProcessName: conn.ProcessName,
			}
			listeners = append(listeners, listener)
		}
	}

	return listeners, nil
}

// getStatistics 获取统计信息
func (r *linuxReader) getStatistics() (*types.NetStatistics, error) {
	connections, err := r.getConnections(&types.NetStatOptions{Protocol: "all"})
	if err != nil {
		return nil, err
	}

	stats := &types.NetStatistics{}

	for _, conn := range connections {
		if strings.HasPrefix(conn.Protocol, "tcp") {
			stats.TCPTotal++
			switch conn.State {
			case types.StateEstablished:
				stats.TCPEstablished++
			case types.StateListen:
				stats.TCPListen++
			case types.StateTimeWait:
				stats.TCPTimeWait++
			case types.StateCloseWait:
				stats.TCPCloseWait++
			}
		} else if strings.HasPrefix(conn.Protocol, "udp") {
			stats.UDPTotal++
		}
		stats.TotalConnections++
	}

	return stats, nil
}

// readTCPConnections 读取 TCP 连接
func (r *linuxReader) readTCPConnections(path string) ([]*types.Connection, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var connections []*types.Connection
	scanner := bufio.NewScanner(file)

	// 跳过表头
	scanner.Scan()

	protocol := "tcp"
	if strings.Contains(path, "tcp6") {
		protocol = "tcp6"
	}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		// 解析本地地址
		localAddr, localPort := parseAddress(fields[1])
		// 解析远程地址
		remoteAddr, remotePort := parseAddress(fields[2])
		// 解析状态
		state := parseTCPState(fields[3])

		conn := &types.Connection{
			Protocol:   protocol,
			LocalAddr:  localAddr,
			LocalPort:  localPort,
			RemoteAddr: remoteAddr,
			RemotePort: remotePort,
			State:      state,
		}

		connections = append(connections, conn)
	}

	return connections, scanner.Err()
}

// readUDPConnections 读取 UDP 连接
func (r *linuxReader) readUDPConnections(path string) ([]*types.Connection, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var connections []*types.Connection
	scanner := bufio.NewScanner(file)

	// 跳过表头
	scanner.Scan()

	protocol := "udp"
	if strings.Contains(path, "udp6") {
		protocol = "udp6"
	}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		// 解析本地地址
		localAddr, localPort := parseAddress(fields[1])
		// 解析远程地址
		remoteAddr, remotePort := parseAddress(fields[2])

		conn := &types.Connection{
			Protocol:   protocol,
			LocalAddr:  localAddr,
			LocalPort:  localPort,
			RemoteAddr: remoteAddr,
			RemotePort: remotePort,
			State:      types.StateUnknown, // UDP 没有状态
		}

		connections = append(connections, conn)
	}

	return connections, scanner.Err()
}

// parseAddress 解析地址字符串 (格式: 0100007F:1F90)
func parseAddress(addrStr string) (string, int) {
	parts := strings.Split(addrStr, ":")
	if len(parts) != 2 {
		return "0.0.0.0", 0
	}

	// 解析 IP (十六进制小端序)
	ipHex := parts[0]
	var ip string
	if len(ipHex) == 8 {
		// IPv4
		ip = parseIPv4(ipHex)
	} else {
		// IPv6
		ip = parseIPv6(ipHex)
	}

	// 解析端口
	portHex := parts[1]
	port, _ := strconv.ParseInt(portHex, 16, 32)

	return ip, int(port)
}

// parseIPv4 解析 IPv4 地址
func parseIPv4(hexIP string) string {
	if len(hexIP) != 8 {
		return "0.0.0.0"
	}

	var ip [4]byte
	for i := 0; i < 4; i++ {
		val, _ := strconv.ParseUint(hexIP[i*2:(i+1)*2], 16, 8)
		ip[3-i] = byte(val)
	}

	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

// parseIPv6 解析 IPv6 地址
func parseIPv6(hexIP string) string {
	if len(hexIP) != 32 {
		return "::"
	}

	// 简化版实现,实际应该处理压缩表示
	var parts []string
	for i := 0; i < 8; i++ {
		start := i * 4
		end := start + 4
		parts = append(parts, hexIP[start:end])
	}

	return strings.Join(parts, ":")
}

// parseTCPState 解析 TCP 状态
func parseTCPState(stateHex string) types.ConnectionState {
	state, _ := strconv.ParseInt(stateHex, 16, 32)

	switch state {
	case tcpStateEstablishedHex:
		return types.StateEstablished
	case tcpStateSynSentHex:
		return types.StateSynSent
	case tcpStateSynRecvHex:
		return types.StateSynRecv
	case tcpStateFinWait1Hex:
		return types.StateFinWait1
	case tcpStateFinWait2Hex:
		return types.StateFinWait2
	case tcpStateTimeWaitHex:
		return types.StateTimeWait
	case tcpStateCloseHex:
		return types.StateClose
	case tcpStateCloseWaitHex:
		return types.StateCloseWait
	case tcpStateLastAckHex:
		return types.StateLastAck
	case tcpStateListenHex:
		return types.StateListen
	case tcpStateClosingHex:
		return types.StateClosing
	default:
		return types.StateUnknown
	}
}
