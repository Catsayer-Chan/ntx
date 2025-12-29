//go:build windows
// +build windows

package netstat

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/catsayer/ntx/pkg/types"
)

type windowsReader struct{}

func newPlatformReader() platformReader {
	return &windowsReader{}
}

func (r *windowsReader) getConnections(opts *types.NetStatOptions) ([]*types.Connection, error) {
	protocols := parseProtocolRequest(opts)
	connections := make([]*types.Connection, 0)

	for _, proto := range protocols {
		cmd := exec.Command("netstat", "-ano", "-p", proto)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("执行 netstat -ano -p %s 失败: %w (%s)", proto, err, strings.TrimSpace(string(output)))
		}
		connections = append(connections, parseWindowsConnections(string(output))...)
	}

	return connections, nil
}

func (r *windowsReader) getListeners(opts *types.NetStatOptions) ([]*types.Listener, error) {
	connections, err := r.getConnections(opts)
	if err != nil {
		return nil, err
	}

	listeners := make([]*types.Listener, 0)
	for _, conn := range connections {
		if conn.State == types.StateListen {
			listeners = append(listeners, &types.Listener{
				Protocol:    conn.Protocol,
				Addr:        conn.LocalAddr,
				Port:        conn.LocalPort,
				PID:         conn.PID,
				ProcessName: conn.ProcessName,
			})
		}
	}

	return listeners, nil
}

func (r *windowsReader) getStatistics() (*types.NetStatistics, error) {
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

func parseWindowsConnections(raw string) []*types.Connection {
	scanner := bufio.NewScanner(strings.NewReader(raw))
	connections := make([]*types.Connection, 0)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "Active") || strings.HasPrefix(line, "Proto") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		protocol := strings.ToLower(fields[0])
		localIdx, remoteIdx, stateIdx, pidIdx := 1, 2, -1, len(fields)-1

		state := types.StateUnknown
		if strings.HasPrefix(protocol, "tcp") && len(fields) >= 5 {
			stateIdx = len(fields) - 2
			pidIdx = len(fields) - 1
			state = mapConnState(fields[stateIdx])
		}

		localAddr, localPort := splitWindowsAddress(fields[localIdx])
		remoteAddr, remotePort := splitWindowsAddress(fields[remoteIdx])
		pid, _ := strconv.Atoi(fields[pidIdx])

		conn := &types.Connection{
			Protocol:   protocol,
			LocalAddr:  localAddr,
			LocalPort:  localPort,
			RemoteAddr: remoteAddr,
			RemotePort: remotePort,
			State:      state,
			PID:        pid,
		}

		connections = append(connections, conn)
	}

	return connections
}

func splitWindowsAddress(addr string) (string, int) {
	if addr == "*:*" {
		return "*", 0
	}

	// Windows IPv6 地址使用 [addr]:port 格式
	if strings.HasPrefix(addr, "[") {
		if idx := strings.LastIndex(addr, "]:"); idx != -1 {
			host := addr[1:idx]
			portStr := addr[idx+2:]
			port, _ := strconv.Atoi(portStr)
			return host, port
		}
	}

	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		host := addr[:idx]
		portStr := addr[idx+1:]
		port, _ := strconv.Atoi(portStr)
		return host, port
	}
	return addr, 0
}

func parseProtocolRequest(opts *types.NetStatOptions) []string {
	switch strings.ToLower(opts.Protocol) {
	case "tcp":
		return []string{"tcp"}
	case "udp":
		return []string{"udp"}
	default:
		return []string{"tcp", "udp"}
	}
}

func mapConnState(state string) types.ConnectionState {
	switch strings.ToUpper(state) {
	case "ESTABLISHED":
		return types.StateEstablished
	case "SYN_SENT":
		return types.StateSynSent
	case "SYN_RCVD", "SYN_RECV":
		return types.StateSynRecv
	case "FIN_WAIT_1", "FIN_WAIT1":
		return types.StateFinWait1
	case "FIN_WAIT_2", "FIN_WAIT2":
		return types.StateFinWait2
	case "TIME_WAIT":
		return types.StateTimeWait
	case "CLOSE_WAIT":
		return types.StateCloseWait
	case "CLOSED", "CLOSE":
		return types.StateClose
	case "LAST_ACK":
		return types.StateLastAck
	case "LISTENING", "LISTEN":
		return types.StateListen
	case "CLOSING":
		return types.StateClosing
	default:
		return types.StateUnknown
	}
}
