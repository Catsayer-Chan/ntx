// Package types 提供网络相关的基础常量
package types

const (
	// ICMPIDMask ICMP 标识字段的 16 位掩码
	ICMPIDMask = 0xffff
	// StandardMTU 以太网标准 MTU
	StandardMTU = 1500
	// TCPHandshakeBytes TCP SYN+ACK 估算字节数
	TCPHandshakeBytes = 40
)
