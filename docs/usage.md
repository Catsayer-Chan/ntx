# NTX 使用说明

本文档详细介绍 NTX (Network Tools eXtended) 的所有功能和使用方法。

## 目录

- [基本概念](#基本概念)
- [全局选项](#全局选项)
- [Ping 命令](#ping-命令)
- [Traceroute 命令](#traceroute-命令)
- [输出格式](#输出格式)
- [配置文件](#配置文件)
- [高级用法](#高级用法)

## 基本概念

NTX 是一个现代化的网络调试工具集，提供统一的命令行接口来执行常见的网络诊断任务。

### 设计理念

- **统一接口**: 所有网络工具使用一致的命令行参数
- **多协议支持**: 支持 ICMP、TCP、HTTP 等多种协议
- **跨平台**: 在 Linux、macOS、Windows 上提供一致的体验
- **灵活输出**: 支持文本、JSON、YAML 等多种输出格式
- **智能降级**: 权限不足时自动切换到替代方案

## 全局选项

所有 NTX 命令都支持以下全局选项：

```bash
ntx [command] [flags]
```

### 通用标志

| 标志 | 简写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--config` | | string | ~/.ntx.yaml | 配置文件路径 |
| `--verbose` | `-v` | bool | false | 启用详细输出 |
| `--output` | `-o` | string | text | 输出格式 (text/json/yaml/table) |
| `--no-color` | | bool | false | 禁用彩色输出 |
| `--help` | `-h` | bool | false | 显示帮助信息 |
| `--version` | | bool | false | 显示版本信息 |

### 示例

```bash
# 查看版本
ntx --version

# 查看帮助
ntx --help

# 启用详细输出
ntx ping google.com -v

# 使用 JSON 输出
ntx ping google.com -o json

# 禁用彩色输出
ntx ping google.com --no-color

# 使用自定义配置文件
ntx ping google.com --config /path/to/config.yaml
```

## Ping 命令

Ping 用于测试网络连通性和延迟。NTX 支持三种 Ping 协议。

### 基本用法

```bash
ntx ping <target> [flags]
```

### 协议类型

#### 1. ICMP Ping（默认）

使用 ICMP Echo Request/Reply 协议，需要 root 权限。

```bash
# 基本 ICMP Ping
sudo ntx ping google.com

# 指定次数
sudo ntx ping google.com -c 10

# 设置间隔和超时
sudo ntx ping google.com -c 5 -i 0.5 -t 3
```

#### 2. TCP Ping

通过建立 TCP 连接测试端口可达性，无需特殊权限。

```bash
# TCP Ping 默认端口 80
ntx ping google.com --protocol tcp

# 指定端口
ntx ping google.com --protocol tcp --port 443

# 测试多个端口
ntx ping example.com -p tcp --port 80
ntx ping example.com -p tcp --port 443
ntx ping example.com -p tcp --port 8080
```

#### 3. HTTP Ping

通过发送 HTTP 请求测试 Web 服务，无需特殊权限。

```bash
# HTTP Ping
ntx ping http://example.com --protocol http

# HTTPS Ping
ntx ping https://example.com --protocol http

# 指定路径
ntx ping https://api.github.com --protocol http
```

### 参数说明

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--protocol` | `-p` | string | icmp | 协议类型: icmp, tcp, http |
| `--count` | `-c` | int | 4 | 发送次数，0 表示无限次 |
| `--interval` | `-i` | float | 1.0 | 发送间隔（秒） |
| `--timeout` | `-t` | float | 5.0 | 超时时间（秒） |
| `--size` | `-s` | int | 64 | 数据包大小（字节） |
| `--ttl` | | int | 64 | Time To Live |
| `--port` | | int | 0 | 端口号（TCP/HTTP） |
| `--ipv4` | `-4` | bool | false | 强制使用 IPv4 |
| `--ipv6` | `-6` | bool | false | 强制使用 IPv6 |

### 使用示例

#### 基本 Ping

```bash
# 简单 Ping（自动选择 TCP 如果没有 root 权限）
ntx ping google.com

# 指定次数
ntx ping google.com -c 10

# 快速 Ping（短间隔）
ntx ping google.com -c 20 -i 0.2
```

#### 端口测试

```bash
# 测试 HTTPS 端口
ntx ping google.com --protocol tcp --port 443

# 测试 SSH 端口
ntx ping example.com --protocol tcp --port 22

# 测试自定义端口
ntx ping myserver.com --protocol tcp --port 8080
```

#### HTTP 服务测试

```bash
# 测试网站可达性
ntx ping https://www.google.com --protocol http

# 测试 API 端点
ntx ping https://api.github.com --protocol http -c 5

# 测试响应时间
ntx ping https://example.com --protocol http -c 10 -i 0.5
```

#### IPv6 测试

```bash
# 强制使用 IPv6
ntx ping google.com -6

# IPv6 地址
ntx ping 2001:4860:4860::8888
```

### 输出解释

```bash
$ ntx ping baidu.com --protocol tcp --port 443 -c 3

PING baidu.com (111.63.65.103) tcp protocol
------------------------------------------------------------
Reply from 111.63.65.103: bytes=40 time=22.069ms ttl=64 seq=1
Reply from 111.63.65.103: bytes=40 time=32.946ms ttl=64 seq=2
Reply from 111.63.65.103: bytes=40 time=32.235ms ttl=64 seq=3

------------------------------------------------------------
--- baidu.com ping statistics ---
3 packets transmitted, 3 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 22.069ms/29.083ms/32.946ms/4.968ms
time 2.090s
```

字段说明：
- **from**: 响应来源 IP
- **bytes**: 响应字节数
- **time**: 往返时间（RTT）
- **ttl**: Time To Live
- **seq**: 序列号
- **transmitted**: 发送的数据包数
- **received**: 接收的数据包数
- **packet loss**: 丢包率
- **min/avg/max/stddev**: 最小/平均/最大/标准差 RTT

## Traceroute 命令

Traceroute 用于追踪数据包到达目标主机所经过的路由路径。

### 基本用法

```bash
ntx trace <target> [flags]

# 别名
ntx traceroute <target> [flags]
ntx tr <target> [flags]
```

### 参数说明

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--max-hops` | `-m` | int | 30 | 最大跳数 |
| `--timeout` | `-t` | float | 3.0 | 每跳超时时间（秒） |
| `--queries` | `-q` | int | 3 | 每跳查询次数 |
| `--port` | `-p` | int | 33434 | 起始端口号（UDP） |
| `--first-ttl` | | int | 1 | 起始 TTL 值 |
| `--ipv4` | `-4` | bool | false | 强制使用 IPv4 |
| `--ipv6` | `-6` | bool | false | 强制使用 IPv6 |

### 使用示例

#### 基本 Traceroute

```bash
# 简单追踪（需要 sudo）
sudo ntx trace google.com

# 限制最大跳数
sudo ntx trace google.com -m 15

# 每跳查询 5 次
sudo ntx trace google.com -q 5
```

#### 高级选项

```bash
# 从第 5 跳开始
sudo ntx trace google.com --first-ttl 5

# 设置超时时间
sudo ntx trace google.com -t 5

# 指定端口
sudo ntx trace google.com -p 33434
```

#### 不同输出格式

```bash
# 文本输出（默认）
sudo ntx trace google.com

# 表格输出
sudo ntx trace google.com -o table

# JSON 输出（便于解析）
sudo ntx trace google.com -o json

# YAML 输出
sudo ntx trace google.com -o yaml
```

### 输出解释

```bash
$ sudo ntx trace baidu.com -m 10 -q 2

traceroute to baidu.com (111.63.65.103), 10 hops max, icmp protocol
----------------------------------------------------------------------
 1  192.168.1.1                               1.234ms  1.123ms
 2  10.0.0.1                                  5.678ms  5.543ms
 3  * * *
 4  202.97.33.1                               15.234ms 15.123ms
 5  202.97.50.166                             20.456ms 20.334ms
 6  111.63.65.103                             25.789ms 25.678ms  [DEST]
----------------------------------------------------------------------
Trace complete: reached baidu.com in 6 hops
Time: 18.234s
```

字段说明：
- **第一列**: 跳数（TTL）
- **第二列**: 路由器主机名或 IP
- **后续列**: 每次探测的 RTT
- **`*`**: 超时或未响应
- **`[DEST]`**: 到达目标主机

## 输出格式

NTX 支持多种输出格式，适用于不同场景。

### Text 格式（默认）

人类可读的文本格式，包含彩色输出。

```bash
ntx ping google.com -c 3
```

### JSON 格式

适合程序解析和自动化处理。

```bash
ntx ping google.com -c 3 -o json
```

输出示例：
```json
{
  "target": {
    "hostname": "google.com",
    "ip": "142.250.185.78",
    "ip_version": 4,
    "port": 80
  },
  "protocol": "tcp",
  "replies": [...],
  "statistics": {
    "sent": 3,
    "received": 3,
    "loss": 0,
    "loss_rate": 0,
    "min_rtt": 20123456,
    "max_rtt": 25678901,
    "avg_rtt": 22901178
  }
}
```

### YAML 格式

适合配置文件和人类阅读。

```bash
ntx ping google.com -c 3 -o yaml
```

输出示例：
```yaml
target:
  hostname: google.com
  ip: 142.250.185.78
  ip_version: 4
protocol: tcp
statistics:
  sent: 3
  received: 3
  loss_rate: 0
```

### Table 格式

表格化展示，适合数据对比。

```bash
ntx ping google.com -c 5 -o table
ntx trace google.com -o table
```

## 配置文件

### 配置文件位置

默认搜索顺序（从高到低优先级）：

1. `--config` 参数指定的文件
2. 当前目录 `.ntx.yaml`
3. `~/.ntx.yaml`
4. `~/.config/ntx/config.yaml`
5. `/etc/ntx/config.yaml`

项目根目录 `configs/default.yaml` 可作为模板复制到上述任意位置。

### 配置文件示例

```yaml
global:
  verbose: false
  output: text
  no_color: false
  log_level: info

ping:
  protocol: icmp
  count: 4
  timeout: 5s
  interval: 1s
  port: 0
  ttl: 64

dns:
  server: "1.1.1.1:53"
  timeout: 3s

http:
  timeout: 15s
  follow_redirect: true
  max_redirects: 5

scan:
  timeout: 3s
  concurrency: 200
  service_detect: true

trace:
  max_hops: 25
  timeout: 3s
  queries: 3
  port: 33434
  first_ttl: 1
```

### 使用自定义配置

```bash
# 使用指定配置文件
ntx ping google.com --config /path/to/config.yaml

# 环境变量配置
export NTX_OUTPUT=json
export NTX_VERBOSE=true
export NTX_DNS_SERVER=1.1.1.1:53
ntx ping google.com
```

## 高级用法

### 批量测试

```bash
# 测试多个主机
for host in google.com baidu.com github.com; do
  echo "Testing $host..."
  ntx ping $host --protocol tcp -c 5
done

# 测试多个端口
for port in 80 443 8080; do
  echo "Testing port $port..."
  ntx ping example.com --protocol tcp --port $port -c 3
done
```

### 脚本集成

#### Bash 脚本

```bash
#!/bin/bash

# 检查主机是否在线
check_host() {
    local host=$1
    result=$(ntx ping "$host" --protocol tcp -c 1 -o json 2>/dev/null)

    if echo "$result" | jq -e '.statistics.received > 0' > /dev/null 2>&1; then
        echo "$host is online"
        return 0
    else
        echo "$host is offline"
        return 1
    fi
}

# 测试主机列表
hosts=("google.com" "baidu.com" "github.com")
for host in "${hosts[@]}"; do
    check_host "$host"
done
```

#### Python 脚本

```python
import subprocess
import json

def ping_host(host, count=3):
    """Ping 主机并返回结果"""
    cmd = ['ntx', 'ping', host, '--protocol', 'tcp', '-c', str(count), '-o', 'json']
    result = subprocess.run(cmd, capture_output=True, text=True)

    if result.returncode == 0:
        return json.loads(result.stdout)
    else:
        return None

# 测试主机
result = ping_host('google.com')
if result:
    stats = result['statistics']
    print(f"Loss Rate: {stats['loss_rate']}%")
    print(f"Avg RTT: {stats['avg_rtt']/1000000:.2f}ms")
```

### 监控和告警

```bash
#!/bin/bash

# 监控脚本示例
THRESHOLD=100  # 延迟阈值（毫秒）
LOSS_THRESHOLD=10  # 丢包率阈值（百分比）

monitor_host() {
    local host=$1
    result=$(ntx ping "$host" --protocol tcp -c 10 -o json 2>/dev/null)

    loss_rate=$(echo "$result" | jq '.statistics.loss_rate')
    avg_rtt=$(echo "$result" | jq '.statistics.avg_rtt / 1000000')

    if (( $(echo "$loss_rate > $LOSS_THRESHOLD" | bc -l) )); then
        echo "WARNING: High packet loss to $host: ${loss_rate}%"
        # 发送告警...
    fi

    if (( $(echo "$avg_rtt > $THRESHOLD" | bc -l) )); then
        echo "WARNING: High latency to $host: ${avg_rtt}ms"
        # 发送告警...
    fi
}

# 持续监控
while true; do
    monitor_host "google.com"
    sleep 60
done
```

### 性能测试

```bash
# 测试不同时段的网络性能
#!/bin/bash

OUTPUT_FILE="network_performance.csv"
echo "timestamp,host,protocol,loss_rate,avg_rtt" > "$OUTPUT_FILE"

while true; do
    timestamp=$(date +"%Y-%m-%d %H:%M:%S")

    for host in google.com baidu.com; do
        for protocol in tcp http; do
            result=$(ntx ping "$host" --protocol "$protocol" -c 5 -o json 2>/dev/null)

            loss_rate=$(echo "$result" | jq '.statistics.loss_rate')
            avg_rtt=$(echo "$result" | jq '.statistics.avg_rtt / 1000000')

            echo "$timestamp,$host,$protocol,$loss_rate,$avg_rtt" >> "$OUTPUT_FILE"
        done
    done

    sleep 300  # 每 5 分钟测试一次
done
```

## 故障排除

### 常见问题

1. **权限错误**
   ```bash
   # 使用 sudo
   sudo ntx ping google.com

   # 或授予能力
   sudo setcap cap_net_raw+ep $(which ntx)
   ```

2. **自动降级**
   ```bash
   # ICMP 失败时自动使用 TCP
   ntx ping google.com  # 自动切换到 TCP
   ```

3. **超时问题**
   ```bash
   # 增加超时时间
   ntx ping slow-host.com -t 10
   ```

4. **防火墙阻断**
   ```bash
   # 尝试不同协议和端口
   ntx ping example.com --protocol tcp --port 443
   ntx ping example.com --protocol http
   ```

## 最佳实践

1. **选择合适的协议**
   - 诊断网络: 使用 ICMP
   - 测试服务: 使用 TCP 或 HTTP
   - 无 root 权限: 使用 TCP 或 HTTP

2. **合理设置参数**
   - 间隔不要太短（避免被限流）
   - 超时时间根据网络状况调整
   - 次数根据需要平衡速度和准确性

3. **输出格式选择**
   - 人类查看: 使用 text 或 table
   - 自动化处理: 使用 json
   - 配置文件: 使用 yaml

4. **日志和监控**
   - 启用详细日志用于调试
   - 使用 JSON 输出便于日志分析
   - 定期监控关键服务

## 下一步

- 查看 [安装指南](installation.md) 了解安装方法
- 查看 [示例](examples.md) 学习实际应用场景
- 查看 [API 文档](api.md) 进行二次开发
