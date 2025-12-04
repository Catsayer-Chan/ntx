// Package main 是 NTX 网络工具的入口点
//
// NTX (Network Tools eXtended) 是一个现代化的网络调试命令整合工具，
// 整合了 ping、traceroute、netstat、ss、ifconfig 等常用网络调试命令，
// 提供统一的命令行接口和一致的输出格式。
//
// 核心特性：
// - 多协议支持（ICMP/TCP/HTTP Ping）
// - 跨平台一致性（Linux/macOS/Windows）
// - 智能诊断功能
// - 批量任务处理
// - 插件化架构
//
// 使用示例：
//   ntx ping google.com -c 5
//   ntx trace google.com
//   ntx scan 192.168.1.1 -p 1-1024
//   ntx diag
//
// 作者: Catsayer
package main

import (
	"fmt"
	"os"

	"github.com/catsayer/ntx/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}