// Package netstat 提供网络连接状态查询功能
//
// 本模块实现了网络连接状态查询功能,包括:
// - TCP/UDP 连接列表
// - 监听端口查看
// - 进程关联
// - 连接统计
//
// 依赖:
// - 平台特定实现 (Linux/macOS/Windows)
//
// 使用示例:
//
//	reader := netstat.NewNetStatReader()
//	connections, err := reader.GetConnections(&netstat.Options{Protocol: "tcp"})
//
// 作者: Catsayer
package netstat

import (
	"fmt"

	"github.com/catsayer/ntx/pkg/types"
)

// NetStatReader 网络连接状态读取器
type NetStatReader struct {
	impl platformReader
}

// platformReader 平台特定实现接口
type platformReader interface {
	getConnections(opts *types.NetStatOptions) ([]*types.Connection, error)
	getListeners(opts *types.NetStatOptions) ([]*types.Listener, error)
	getStatistics() (*types.NetStatistics, error)
}

// NewNetStatReader 创建新的网络连接状态读取器
func NewNetStatReader() *NetStatReader {
	return &NetStatReader{
		impl: newPlatformReader(),
	}
}

// GetConnections 获取网络连接列表
func (r *NetStatReader) GetConnections(opts *types.NetStatOptions) ([]*types.Connection, error) {
	if opts == nil {
		opts = &types.NetStatOptions{
			Protocol: "all",
		}
	}

	connections, err := r.impl.getConnections(opts)
	if err != nil {
		return nil, fmt.Errorf("获取连接列表失败: %w", err)
	}

	// 应用过滤器
	filtered := make([]*types.Connection, 0)
	for _, conn := range connections {
		if r.matchesFilter(conn, opts) {
			filtered = append(filtered, conn)
		}
	}

	return filtered, nil
}

// GetListeners 获取监听端口列表
func (r *NetStatReader) GetListeners(opts *types.NetStatOptions) ([]*types.Listener, error) {
	if opts == nil {
		opts = &types.NetStatOptions{
			Protocol: "all",
		}
	}

	listeners, err := r.impl.getListeners(opts)
	if err != nil {
		return nil, fmt.Errorf("获取监听端口失败: %w", err)
	}

	return listeners, nil
}

// GetStatistics 获取连接统计信息
func (r *NetStatReader) GetStatistics() (*types.NetStatistics, error) {
	stats, err := r.impl.getStatistics()
	if err != nil {
		return nil, fmt.Errorf("获取统计信息失败: %w", err)
	}

	return stats, nil
}

// matchesFilter 检查连接是否匹配过滤条件
func (r *NetStatReader) matchesFilter(conn *types.Connection, opts *types.NetStatOptions) bool {
	// 协议过滤
	if opts.Protocol != "all" {
		if opts.Protocol == "tcp" && conn.Protocol != "tcp" && conn.Protocol != "tcp6" {
			return false
		}
		if opts.Protocol == "udp" && conn.Protocol != "udp" && conn.Protocol != "udp6" {
			return false
		}
	}

	// 状态过滤
	if len(opts.State) > 0 {
		matched := false
		for _, state := range opts.State {
			if conn.State == state {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// 本地端口过滤
	if opts.LocalPort > 0 && conn.LocalPort != opts.LocalPort {
		return false
	}

	// 远程端口过滤
	if opts.RemotePort > 0 && conn.RemotePort != opts.RemotePort {
		return false
	}

	// 进程名过滤
	if opts.ProcessName != "" && conn.ProcessName != opts.ProcessName {
		return false
	}

	// 仅监听端口
	if opts.ListenOnly && conn.State != types.StateListen {
		return false
	}

	return true
}

// Close 关闭读取器 (当前无需实际操作)
func (r *NetStatReader) Close() error {
	return nil
}
