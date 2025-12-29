// Package types 提供网络连接相关类型定义
//
// 作者: Catsayer
package types

// ConnectionState 连接状态
type ConnectionState string

const (
	StateEstablished ConnectionState = "ESTABLISHED"
	StateSynSent     ConnectionState = "SYN_SENT"
	StateSynRecv     ConnectionState = "SYN_RECV"
	StateFinWait1    ConnectionState = "FIN_WAIT1"
	StateFinWait2    ConnectionState = "FIN_WAIT2"
	StateTimeWait    ConnectionState = "TIME_WAIT"
	StateClose       ConnectionState = "CLOSE"
	StateCloseWait   ConnectionState = "CLOSE_WAIT"
	StateLastAck     ConnectionState = "LAST_ACK"
	StateListen      ConnectionState = "LISTEN"
	StateClosing     ConnectionState = "CLOSING"
	StateUnknown     ConnectionState = "UNKNOWN"
)

// Connection 网络连接信息
type Connection struct {
	// Protocol 协议类型 (tcp, tcp6, udp, udp6)
	Protocol string `json:"protocol" yaml:"protocol"`

	// LocalAddr 本地地址
	LocalAddr string `json:"local_addr" yaml:"local_addr"`

	// LocalPort 本地端口
	LocalPort int `json:"local_port" yaml:"local_port"`

	// RemoteAddr 远程地址
	RemoteAddr string `json:"remote_addr" yaml:"remote_addr"`

	// RemotePort 远程端口
	RemotePort int `json:"remote_port" yaml:"remote_port"`

	// State 连接状态
	State ConnectionState `json:"state" yaml:"state"`

	// PID 进程 ID
	PID int `json:"pid,omitempty" yaml:"pid,omitempty"`

	// ProcessName 进程名称
	ProcessName string `json:"process_name,omitempty" yaml:"process_name,omitempty"`
}

// Listener 监听端口信息
type Listener struct {
	// Protocol 协议类型 (tcp, tcp6, udp, udp6)
	Protocol string `json:"protocol" yaml:"protocol"`

	// Addr 监听地址
	Addr string `json:"addr" yaml:"addr"`

	// Port 监听端口
	Port int `json:"port" yaml:"port"`

	// PID 进程 ID
	PID int `json:"pid,omitempty" yaml:"pid,omitempty"`

	// ProcessName 进程名称
	ProcessName string `json:"process_name,omitempty" yaml:"process_name,omitempty"`
}

// NetStatistics 网络连接统计
type NetStatistics struct {
	// TCPEstablished TCP ESTABLISHED 连接数
	TCPEstablished int `json:"tcp_established" yaml:"tcp_established"`

	// TCPListen TCP LISTEN 状态数
	TCPListen int `json:"tcp_listen" yaml:"tcp_listen"`

	// TCPTimeWait TCP TIME_WAIT 状态数
	TCPTimeWait int `json:"tcp_time_wait" yaml:"tcp_time_wait"`

	// TCPCloseWait TCP CLOSE_WAIT 状态数
	TCPCloseWait int `json:"tcp_close_wait" yaml:"tcp_close_wait"`

	// TCPTotal TCP 总连接数
	TCPTotal int `json:"tcp_total" yaml:"tcp_total"`

	// UDPTotal UDP 总连接数
	UDPTotal int `json:"udp_total" yaml:"udp_total"`

	// TotalConnections 总连接数
	TotalConnections int `json:"total_connections" yaml:"total_connections"`
}

// NetStatOptions 查询选项
type NetStatOptions struct {
	// Protocol 协议过滤 (tcp, udp, all)
	Protocol string `json:"protocol" yaml:"protocol"`

	// State 状态过滤
	State []ConnectionState `json:"state,omitempty" yaml:"state,omitempty"`

	// LocalPort 本地端口过滤
	LocalPort int `json:"local_port,omitempty" yaml:"local_port,omitempty"`

	// RemotePort 远程端口过滤
	RemotePort int `json:"remote_port,omitempty" yaml:"remote_port,omitempty"`

	// ProcessName 进程名过滤
	ProcessName string `json:"process_name,omitempty" yaml:"process_name,omitempty"`

	// IncludeProcess 是否包含进程信息
	IncludeProcess bool `json:"include_process" yaml:"include_process"`

	// ListenOnly 仅显示监听端口
	ListenOnly bool `json:"listen_only" yaml:"listen_only"`
}
