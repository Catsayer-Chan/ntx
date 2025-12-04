# NTX API 文档

本文档介绍如何在 Go 程序中使用 NTX 作为库。

## 目录

- [安装](#安装)
- [快速开始](#快速开始)
- [核心包](#核心包)
- [类型定义](#类型定义)
- [示例代码](#示例代码)

## 安装

```bash
go get github.com/Catsayer-Chan/ntx
```

## 快速开始

### 基本 Ping 示例

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Catsayer-Chan/ntx/internal/core/ping"
	"github.com/Catsayer-Chan/ntx/pkg/types"
)

func main() {
	// 创建 TCP Pinger
	pinger := ping.NewTCPPinger()
	defer pinger.Close()

	// 配置选项
	opts := &types.PingOptions{
		Protocol: types.ProtocolTCP,
		Count:    5,
		Interval: time.Second,
		Timeout:  5 * time.Second,
		Port:     443,
	}

	// 执行 Ping
	result, err := pinger.Ping("google.com", opts)
	if err != nil {
		log.Fatal(err)
	}

	// 输出结果
	fmt.Printf("Host: %s (%s)\n", result.Target.Hostname, result.Target.IP)
	fmt.Printf("Sent: %d, Received: %d\n",
		result.Statistics.Sent,
		result.Statistics.Received)
	fmt.Printf("Loss Rate: %.1f%%\n", result.Statistics.LossRate)
	fmt.Printf("Avg RTT: %v\n", result.Statistics.AvgRTT)
}
```

## 核心包

### pkg/types

定义了所有公共类型和接口。

#### 常用类型

```go
// 协议类型
type Protocol string
const (
    ProtocolICMP  Protocol = "icmp"
    ProtocolTCP   Protocol = "tcp"
    ProtocolUDP   Protocol = "udp"
    ProtocolHTTP  Protocol = "http"
    ProtocolHTTPS Protocol = "https"
)

// 输出格式
type OutputFormat string
const (
    OutputText  OutputFormat = "text"
    OutputJSON  OutputFormat = "json"
    OutputYAML  OutputFormat = "yaml"
    OutputTable OutputFormat = "table"
)

// 状态类型
type Status string
const (
    StatusSuccess Status = "success"
    StatusFailure Status = "failure"
    StatusTimeout Status = "timeout"
    StatusUnknown Status = "unknown"
)
```

#### Ping 相关类型

```go
// PingOptions Ping 配置选项
type PingOptions struct {
    Protocol     Protocol
    Count        int
    Interval     time.Duration
    Timeout      time.Duration
    Size         int
    TTL          int
    Port         int
    IPVersion    IPVersion
    HTTPMethod   string
    HTTPPath     string
}

// PingReply 单次 Ping 响应
type PingReply struct {
    Seq    int
    From   string
    Bytes  int
    TTL    int
    RTT    time.Duration
    Time   time.Time
    Status Status
    Error  string
}

// PingResult Ping 结果
type PingResult struct {
    Target     *Host
    Protocol   Protocol
    Replies    []*PingReply
    Statistics *Statistics
    Context    *ExecutionContext
    Status     Status
    Error      error
}

// Pinger 接口
type Pinger interface {
    Ping(target string, opts *PingOptions) (*PingResult, error)
    Close() error
}
```

#### Traceroute 相关类型

```go
// TraceOptions Traceroute 配置选项
type TraceOptions struct {
    Protocol     Protocol
    MaxHops      int
    Timeout      time.Duration
    Queries      int
    Port         int
    PacketSize   int
    IPVersion    IPVersion
    FirstTTL     int
}

// TraceHop 单跳信息
type TraceHop struct {
    TTL           int
    Probes        []*TraceProbe
    IP            string
    Hostname      string
    IsDestination bool
}

// TraceProbe 单次探测
type TraceProbe struct {
    Seq    int
    IP     string
    RTT    time.Duration
    Status Status
    Error  string
}

// TraceResult Traceroute 结果
type TraceResult struct {
    Target             *Host
    Protocol           Protocol
    Hops               []*TraceHop
    ReachedDestination bool
    HopCount           int
    Context            *ExecutionContext
    Status             Status
    Error              error
}

// Tracer 接口
type Tracer interface {
    Trace(target string, opts *TraceOptions) (*TraceResult, error)
    Close() error
}
```

### internal/core/ping

提供 Ping 功能的实现。

#### TCP Ping

```go
import "github.com/Catsayer-Chan/ntx/internal/core/ping"

// 创建 TCP Pinger
pinger := ping.NewTCPPinger()
defer pinger.Close()

// 执行 Ping
result, err := pinger.Ping("example.com", opts)
```

#### HTTP Ping

```go
// 创建 HTTP Pinger
pinger := ping.NewHTTPPinger()
defer pinger.Close()

// 执行 Ping
result, err := pinger.Ping("https://example.com", opts)
```

#### ICMP Ping

```go
// 创建 ICMP Pinger（需要权限）
pinger, err := ping.NewICMPPinger()
if err != nil {
    log.Fatal(err)
}
defer pinger.Close()

// 执行 Ping
result, err := pinger.Ping("example.com", opts)
```

### internal/core/trace

提供 Traceroute 功能的实现。

#### ICMP Traceroute

```go
import "github.com/Catsayer-Chan/ntx/internal/core/trace"

// 创建 ICMP Tracer（需要权限）
tracer, err := trace.NewICMPTracer()
if err != nil {
    log.Fatal(err)
}
defer tracer.Close()

// 执行 Traceroute
result, err := tracer.Trace("example.com", opts)
```

### internal/output/formatter

提供输出格式化功能。

```go
import (
    "github.com/Catsayer-Chan/ntx/internal/output/formatter"
    "github.com/Catsayer-Chan/ntx/pkg/types"
)

// 创建格式化器
f := formatter.NewFormatter(types.OutputJSON, false)

// 格式化结果
output, err := f.Format(result)
if err != nil {
    log.Fatal(err)
}

fmt.Println(output)
```

### pkg/errors

提供错误类型和错误处理工具。

```go
import "github.com/Catsayer-Chan/ntx/pkg/errors"

// 检查错误类型
if errors.IsTimeout(err) {
    fmt.Println("Operation timed out")
}

if errors.IsPermissionDenied(err) {
    fmt.Println("Permission denied")
}

// 创建自定义错误
err := errors.NewNetworkError("ping", "example.com", innerErr)
```

## 类型定义

### Host

```go
type Host struct {
    Hostname  string    // 主机名
    IP        string    // IP 地址
    IPVersion IPVersion // IP 版本 (4 或 6)
    Port      int       // 端口号（可选）
}
```

### Statistics

```go
type Statistics struct {
    Sent      int           // 发送数量
    Received  int           // 接收数量
    Loss      int           // 丢失数量
    LossRate  float64       // 丢包率 (0-100)
    MinRTT    time.Duration // 最小 RTT
    MaxRTT    time.Duration // 最大 RTT
    AvgRTT    time.Duration // 平均 RTT
    StdDevRTT time.Duration // RTT 标准差
    TotalTime time.Duration // 总时间
}
```

### ExecutionContext

```go
type ExecutionContext struct {
    StartTime   time.Time     // 开始时间
    EndTime     time.Time     // 结束时间
    Duration    time.Duration // 持续时间
    User        string        // 执行用户
    Hostname    string        // 执行主机
    CommandLine string        // 命令行
}
```

## 示例代码

### 完整的 TCP Ping 示例

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/Catsayer-Chan/ntx/internal/core/ping"
    "github.com/Catsayer-Chan/ntx/internal/output/formatter"
    "github.com/Catsayer-Chan/ntx/pkg/types"
)

func main() {
    // 创建 TCP Pinger
    pinger := ping.NewTCPPinger()
    defer pinger.Close()

    // 配置选项
    opts := &types.PingOptions{
        Protocol: types.ProtocolTCP,
        Count:    10,
        Interval: 500 * time.Millisecond,
        Timeout:  5 * time.Second,
        Port:     443,
        IPVersion: types.IPv4,
    }

    // 执行 Ping
    result, err := pinger.Ping("google.com", opts)
    if err != nil {
        log.Fatal(err)
    }

    // 文本输出
    f := formatter.NewFormatter(types.OutputText, false)
    output, _ := f.Format(result)
    fmt.Println(output)

    // 访问结构化数据
    fmt.Printf("\n=== Detailed Statistics ===\n")
    fmt.Printf("Target: %s (%s)\n", result.Target.Hostname, result.Target.IP)
    fmt.Printf("Sent: %d, Received: %d, Loss: %.1f%%\n",
        result.Statistics.Sent,
        result.Statistics.Received,
        result.Statistics.LossRate)

    fmt.Printf("RTT: min=%.2fms avg=%.2fms max=%.2fms\n",
        float64(result.Statistics.MinRTT.Microseconds())/1000,
        float64(result.Statistics.AvgRTT.Microseconds())/1000,
        float64(result.Statistics.MaxRTT.Microseconds())/1000)

    // 遍历每个响应
    fmt.Println("\n=== Individual Replies ===")
    for _, reply := range result.Replies {
        if reply.Status == types.StatusSuccess {
            fmt.Printf("Seq %d: %s - %.2fms\n",
                reply.Seq,
                reply.From,
                float64(reply.RTT.Microseconds())/1000)
        } else {
            fmt.Printf("Seq %d: %s\n", reply.Seq, reply.Status)
        }
    }
}
```

### HTTP Ping 示例

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/Catsayer-Chan/ntx/internal/core/ping"
    "github.com/Catsayer-Chan/ntx/pkg/types"
)

func main() {
    pinger := ping.NewHTTPPinger()
    defer pinger.Close()

    opts := &types.PingOptions{
        Protocol:   types.ProtocolHTTP,
        Count:      5,
        Interval:   time.Second,
        Timeout:    10 * time.Second,
        HTTPMethod: "GET",
        HTTPPath:   "/",
    }

    result, err := pinger.Ping("https://api.github.com", opts)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("HTTP Ping Results:\n")
    fmt.Printf("URL: %s\n", result.Target.Hostname)
    fmt.Printf("Status: %s\n", result.Status)
    fmt.Printf("Success Rate: %.1f%%\n", 100-result.Statistics.LossRate)
    fmt.Printf("Avg Response Time: %v\n", result.Statistics.AvgRTT)
}
```

### Traceroute 示例

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/Catsayer-Chan/ntx/internal/core/trace"
    "github.com/Catsayer-Chan/ntx/pkg/types"
)

func main() {
    // 创建 Tracer（需要 root 权限）
    tracer, err := trace.NewICMPTracer()
    if err != nil {
        log.Fatal(err)
    }
    defer tracer.Close()

    // 配置选项
    opts := &types.TraceOptions{
        Protocol: types.ProtocolICMP,
        MaxHops:  20,
        Timeout:  3 * time.Second,
        Queries:  3,
    }

    // 执行 Traceroute
    result, err := tracer.Trace("google.com", opts)
    if err != nil {
        log.Fatal(err)
    }

    // 输出结果
    fmt.Printf("Traceroute to %s (%s)\n", result.Target.Hostname, result.Target.IP)
    fmt.Printf("Max Hops: %d\n\n", opts.MaxHops)

    for _, hop := range result.Hops {
        fmt.Printf("%2d  ", hop.TTL)

        if hop.IP != "" {
            fmt.Printf("%-40s", hop.Hostname)

            for _, probe := range hop.Probes {
                if probe.Status == types.StatusSuccess {
                    fmt.Printf("  %.2fms",
                        float64(probe.RTT.Microseconds())/1000)
                } else {
                    fmt.Printf("  *")
                }
            }

            if hop.IsDestination {
                fmt.Printf("  [DEST]")
            }
        } else {
            fmt.Printf("* * *")
        }

        fmt.Println()
    }

    fmt.Printf("\nReached: %v\n", result.ReachedDestination)
    fmt.Printf("Total Hops: %d\n", result.HopCount)
}
```

### 并发 Ping 示例

```go
package main

import (
    "fmt"
    "sync"
    "time"

    "github.com/Catsayer-Chan/ntx/internal/core/ping"
    "github.com/Catsayer-Chan/ntx/pkg/types"
)

func main() {
    hosts := []string{"google.com", "github.com", "baidu.com"}
    var wg sync.WaitGroup

    results := make(map[string]*types.PingResult)
    var mu sync.Mutex

    for _, host := range hosts {
        wg.Add(1)
        go func(h string) {
            defer wg.Done()

            pinger := ping.NewTCPPinger()
            defer pinger.Close()

            opts := &types.PingOptions{
                Protocol: types.ProtocolTCP,
                Count:    5,
                Timeout:  3 * time.Second,
                Port:     443,
            }

            result, err := pinger.Ping(h, opts)
            if err != nil {
                fmt.Printf("Error pinging %s: %v\n", h, err)
                return
            }

            mu.Lock()
            results[h] = result
            mu.Unlock()
        }(host)
    }

    wg.Wait()

    // 输出结果
    for host, result := range results {
        fmt.Printf("%s: %.2fms (%.1f%% loss)\n",
            host,
            float64(result.Statistics.AvgRTT.Microseconds())/1000,
            result.Statistics.LossRate)
    }
}
```

### 自定义格式化器

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/Catsayer-Chan/ntx/internal/core/ping"
    "github.com/Catsayer-Chan/ntx/pkg/types"
)

// 自定义输出结构
type CustomOutput struct {
    Host      string  `json:"host"`
    IP        string  `json:"ip"`
    AvgRTT    float64 `json:"avg_rtt_ms"`
    LossRate  float64 `json:"loss_rate"`
    Timestamp string  `json:"timestamp"`
}

func main() {
    pinger := ping.NewTCPPinger()
    defer pinger.Close()

    result, _ := pinger.Ping("google.com", types.DefaultPingOptions())

    // 转换为自定义格式
    custom := CustomOutput{
        Host:      result.Target.Hostname,
        IP:        result.Target.IP,
        AvgRTT:    float64(result.Statistics.AvgRTT.Microseconds()) / 1000,
        LossRate:  result.Statistics.LossRate,
        Timestamp: result.Context.StartTime.Format("2006-01-02 15:04:05"),
    }

    // 输出 JSON
    output, _ := json.MarshalIndent(custom, "", "  ")
    fmt.Println(string(output))
}
```

## 错误处理

### 检查错误类型

```go
result, err := pinger.Ping("example.com", opts)
if err != nil {
    if errors.IsTimeout(err) {
        fmt.Println("Request timed out")
    } else if errors.IsPermissionDenied(err) {
        fmt.Println("Need root permission")
    } else if errors.IsNetworkError(err) {
        fmt.Println("Network error occurred")
    } else {
        fmt.Printf("Unknown error: %v\n", err)
    }
    return
}
```

### 创建自定义错误

```go
import "github.com/Catsayer-Chan/ntx/pkg/errors"

// 网络错误
err := errors.NewNetworkError("connect", "example.com", innerErr)

// 超时错误
err := errors.NewTimeoutError("ping", "5s")

// 权限错误
err := errors.NewPermissionError("icmp ping", "raw socket", "need CAP_NET_RAW")

// 验证错误
err := errors.NewValidationError("port", 70000, "port must be between 1-65535")
```

## 最佳实践

1. **始终关闭资源**
   ```go
   pinger := ping.NewTCPPinger()
   defer pinger.Close()
   ```

2. **处理错误**
   ```go
   result, err := pinger.Ping(target, opts)
   if err != nil {
       // 处理错误
       return
   }
   ```

3. **使用合理的超时**
   ```go
   opts.Timeout = 5 * time.Second
   ```

4. **并发时注意协程安全**
   ```go
   // 每个协程使用独立的 Pinger 实例
   pinger := ping.NewTCPPinger()
   defer pinger.Close()
   ```

5. **选择合适的协议**
   - 需要权限: ICMP
   - 测试端口: TCP
   - 测试服务: HTTP

## 更多资源

- [使用说明](usage.md)
- [示例代码](examples.md)
- [GitHub 仓库](https://github.com/Catsayer-Chan/ntx)
- [问题追踪](https://github.com/Catsayer-Chan/ntx/issues)