# NTX (Network Tools eXtended)

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/catsayer/ntx)

**现代化的网络调试命令整合工具**

[特性](#特性) • [安装](#安装) • [快速开始](#快速开始) • [文档](#文档) • [贡献](#贡献)

</div>

---

## 简介

NTX (Network Tools eXtended) 是一个现代化的网络调试命令整合工具，整合了 `ping`、`traceroute`、`netstat`、`ss`、`ifconfig` 等常用网络调试命令，提供统一的命令行接口和一致的输出格式。

### 为什么选择 NTX？

- **统一接口**：一个工具整合所有常用网络调试命令
- **跨平台**：支持 Linux、macOS、Windows，提供一致的使用体验
- **现代化**：基于 Go 语言开发，性能优异，易于部署
- **智能诊断**：内置智能诊断功能，自动分析网络问题
- **灵活输出**：支持多种输出格式（文本、JSON、YAML）
- **批量处理**：支持批量任务处理，提高工作效率

## 特性

### ✅ 已实现功能

- **多协议 Ping** ⭐
  - ICMP Ping（传统 ping，需要 root 权限）
  - TCP Ping（通过 TCP 连接测试，无需 root）
  - HTTP Ping（通过 HTTP 请求测试，无需 root）
  - 支持多目标并发 Ping
  - 实时监控模式（ASCII 图表）
  - 统计信息（最小/最大/平均/标准差延迟、丢包率）
  - 自动权限降级（ICMP → TCP）

- **路由追踪** ⭐
  - ICMP Traceroute
  - 支持最大跳数、超时时间等配置
  - 路径可视化

- **DNS 查询** ⭐
  - A/AAAA/CNAME/MX/NS/TXT/SOA/PTR/SRV 记录查询
  - 反向 DNS 查询
  - 批量查询
  - 自定义 DNS 服务器
  - 查询所有记录类型

- **网络接口信息** ⭐
  - 接口列表和详细信息
  - IPv4/IPv6 地址、MAC 地址
  - 网卡状态和标志
  - 流量统计（Linux 完整支持）
  - 路由表信息（Linux）

- **HTTP 客户端** ⭐
  - GET/POST/PUT/DELETE/PATCH/HEAD/OPTIONS 方法
  - 自定义请求头和请求体
  - 超时控制和重定向控制
  - JSON 美化输出
  - 性能测试模式（Benchmark）

- **连接状态监控** ⭐
  - TCP/UDP 连接列表（Linux 完整支持）
  - 监听端口查看
  - 连接状态过滤
  - 端口过滤
  - 连接统计

- **端口扫描**
  - TCP Connect 扫描
  - TCP SYN 扫描（需要 root 权限）
  - UDP 扫描
  - 服务识别和版本探测

- **WHOIS 查询**
  - 域名 WHOIS
  - IP WHOIS
  - AS 查询

- **智能诊断**
  - 自动诊断网络问题
  - 连通性测试
  - DNS 解析测试
  - 路由测试
  - 生成诊断报告

- **批量任务**
  - YAML/JSON 配置文件支持
  - 并发执行
  - 任务调度
  - 结果聚合

### 设计特性

- **插件化架构**：易于扩展新功能
- **模块化设计**：各功能模块独立，职责清晰
- **接口抽象**：面向接口编程，易于测试和替换实现
- **平台适配**：自动适配不同操作系统的特性
- **结构化日志**：基于 zap 的高性能日志系统
- **配置管理**：支持配置文件和环境变量

## 安装

### 从源码编译

**前置要求**：
- Go 1.21 或更高版本
- Make（可选，用于构建自动化）

```bash
# 克隆仓库
git clone https://github.com/catsayer/ntx.git
cd ntx

# 使用 Make 编译
make build

# 或者直接使用 go build
go build -o bin/ntx ./cmd/ntx

# 安装到系统
make install
# 或者
go install ./cmd/ntx
```

### 使用 Go Install

```bash
go install github.com/catsayer/ntx/cmd/ntx@latest
```

### 预编译二进制文件

从 [Releases](https://github.com/catsayer/ntx/releases) 页面下载适合你系统的预编译二进制文件。

## 快速开始

### 基本命令概览

```bash
# 查看所有可用命令
ntx --help

# 查看特定命令的帮助
ntx ping --help
ntx dns --help
```

### 核心功能使用指南

#### 1. Ping 测试 - 多协议连通性检测

```bash
# ICMP Ping (传统 ping,需要 root/管理员权限)
sudo ntx ping google.com -c 5

# TCP Ping (无需 root 权限,推荐)
ntx ping google.com --protocol tcp --port 443

# HTTP Ping (通过 HTTP 请求测试)
ntx ping https://www.google.com --protocol http

# 多目标并发 Ping
ntx ping google.com baidu.com github.com -c 3

# 实时监控模式 (带 ASCII 图表)
ntx ping google.com --mode monitor

# 流式输出模式
ntx ping google.com --mode stream -c 10

# 批量并发模式 (JSON 输出)
ntx ping google.com baidu.com github.com --mode batch -o json

# 自定义参数
ntx ping 8.8.8.8 -c 10 -i 0.5 -t 3 --size 128 --ttl 64
```

**参数说明**:
- `-c, --count`: Ping 次数 (默认: 4)
- `-i, --interval`: 间隔时间/秒 (默认: 1)
- `-t, --timeout`: 超时时间/秒 (默认: 5)
- `--size`: 数据包大小/字节 (默认: 64)
- `--ttl`: 生存时间 (默认: 64)
- `--protocol`: 协议类型 (icmp/tcp/http)
- `--mode`: 输出模式 (stream/monitor/batch)

---

#### 2. DNS 查询 - 域名解析

```bash
# 查询 A 记录
ntx dns google.com

# 查询特定类型记录
ntx dns google.com --type MX     # 邮件服务器
ntx dns google.com --type NS     # 名称服务器
ntx dns google.com --type TXT    # 文本记录
ntx dns google.com --type AAAA   # IPv6 地址

# 查询所有类型记录
ntx dns google.com --all

# 指定 DNS 服务器
ntx dns google.com --server 1.1.1.1:53

# 反向 DNS 查询 (IP → 域名)
ntx dns --reverse 8.8.8.8

# 批量查询多个域名
ntx dns --batch domains.txt

# JSON 输出
ntx dns google.com -o json
```

**支持的记录类型**:
`A`, `AAAA`, `CNAME`, `MX`, `NS`, `TXT`, `SOA`, `PTR`, `SRV`

---

#### 3. 路由追踪 - 网络路径分析

```bash
# 基本路由追踪
ntx trace google.com

# 指定最大跳数
ntx trace google.com --max-hops 20

# 使用 TCP 协议 (某些网络 ICMP 被阻断)
ntx trace google.com --protocol tcp --port 443

# 自定义查询参数
ntx trace google.com --queries 5 --timeout 3

# 详细输出
ntx trace google.com -v
```

**参数说明**:
- `--max-hops`: 最大跳数 (默认: 30)
- `--timeout`: 每跳超时时间/秒 (默认: 3)
- `--queries`: 每跳查询次数 (默认: 3)
- `--first-ttl`: 起始 TTL (默认: 1)

---

#### 4. 端口扫描 - 端口开放检测

```bash
# 扫描常用端口
ntx scan 192.168.1.1

# 扫描指定端口
ntx scan 192.168.1.1 -p 80,443,8080

# 扫描端口范围
ntx scan 192.168.1.1 -p 1-1024

# 混合端口列表
ntx scan 192.168.1.1 -p 80,443,8000-9000

# 服务识别
ntx scan 192.168.1.1 -p 1-1000 --service-detect

# 调整并发数和超时
ntx scan 192.168.1.1 -p 1-1000 --concurrency 200 --timeout 2

# JSON 输出
ntx scan 192.168.1.1 -p 1-1024 -o json
```

**参数说明**:
- `-p, --ports`: 端口范围 (默认: 常用端口)
- `--timeout`: 超时时间/秒 (默认: 3)
- `--concurrency`: 并发数 (默认: 100)
- `--service-detect`: 启用服务识别

---

#### 5. HTTP 客户端 - HTTP 请求测试

```bash
# GET 请求
ntx http https://api.github.com

# POST 请求
ntx http https://api.example.com --method POST --data '{"key":"value"}'

# 自定义请求头
ntx http https://api.github.com -H "Authorization: token xxx"

# 不跟随重定向
ntx http https://google.com --no-follow-redirect

# 调整超时和最大重定向次数
ntx http https://example.com --timeout 10 --max-redirects 5

# HTTP 性能测试 (Benchmark)
ntx http https://example.com --benchmark -n 100 --concurrency 10

# 详细输出 (显示请求头和响应头)
ntx http https://api.github.com -v
```

**参数说明**:
- `--method`: HTTP 方法 (GET/POST/PUT/DELETE/PATCH/HEAD/OPTIONS)
- `--data`: 请求体数据
- `-H, --header`: 自定义请求头
- `--benchmark`: 启用性能测试模式
- `-n`: 请求次数 (benchmark 模式)

---

#### 6. 网络连接查看 - 连接状态监控

```bash
# 查看所有连接
ntx conn

# 只看 TCP 连接
ntx conn --tcp

# 只看 UDP 连接
ntx conn --udp

# 只看监听端口
ntx conn --listen

# 过滤特定端口
ntx conn --port 8080

# 过滤特定状态
ntx conn --state ESTABLISHED

# 显示统计信息
ntx conn --stats

# JSON 输出
ntx conn -o json
```

**支持的连接状态**:
`ESTABLISHED`, `LISTEN`, `TIME_WAIT`, `CLOSE_WAIT`, `SYN_SENT`, `SYN_RECEIVED`

---

#### 7. 网络接口信息 - 接口详情查看

```bash
# 查看所有网络接口
ntx iface

# 查看接口详细信息
ntx iface --detail

# 查看流量统计 (Linux 完整支持)
ntx iface --stats

# JSON 输出
ntx iface -o json
```

---

#### 8. WHOIS 查询 - 域名/IP 信息查询

```bash
# 域名 WHOIS 查询
ntx whois google.com

# IP WHOIS 查询
ntx whois 8.8.8.8

# AS 号查询
ntx whois AS15169

# 显示原始响应
ntx whois google.com --raw

# JSON 输出
ntx whois google.com -o json
```

---

#### 9. 智能诊断 - 自动网络问题诊断

```bash
# 诊断网络连接
ntx diag

# 诊断特定目标
ntx diag --target google.com

# 详细诊断报告
ntx diag -v

# JSON 输出
ntx diag -o json
```

**诊断内容**:
- 网络接口状态
- DNS 解析测试
- 连通性测试 (多协议)
- 路由追踪
- 网络配置检查

---

#### 10. 批量任务 - YAML 配置批量执行

```bash
# 执行批量任务
ntx batch -f tasks.yaml

# 指定并发数
ntx batch -f tasks.yaml --concurrency 20

# JSON 输出
ntx batch -f tasks.yaml -o json
```

**tasks.yaml 示例**:
```yaml
tasks:
  - name: "Ping 测试"
    type: ping
    enabled: true
    targets:
      - google.com
      - baidu.com
    options:
      count: 5
      protocol: tcp

  - name: "DNS 查询"
    type: dns
    enabled: true
    targets:
      - google.com
      - github.com
    options:
      type: A

  - name: "端口扫描"
    type: scan
    enabled: true
    targets:
      - 192.168.1.1
    options:
      ports: "80,443,8080"
```

### 高级用法

```bash
# TCP Ping
ntx ping example.com --protocol tcp --port 443

# HTTP Ping
ntx ping example.com --protocol http

# 指定超时和间隔
ntx ping google.com -c 10 -t 5 -i 0.5

# 使用 JSON 输出
ntx ping google.com -c 3 -o json

# 详细输出
ntx trace google.com -v

# 指定最大跳数
ntx trace google.com --max-hops 20

# 端口扫描指定端口范围
ntx scan 192.168.1.1 -p 80,443,8000-9000

# 查询特定 DNS 记录
ntx dns example.com --type MX --server 8.8.8.8

# 查看所有 TCP 连接
ntx conn --tcp --all

# 监听指定端口
ntx conn --listen --port 8080

# 查看接口流量统计
ntx iface --stats

# 诊断特定主机
ntx diag --target google.com

# HTTP POST 请求
ntx http https://api.example.com --method POST --data '{"key":"value"}'

# 批量执行并发任务
ntx batch -f tasks.yaml --concurrency 10
```

## 配置

NTX 支持通过配置文件自定义行为。默认搜索顺序（从高到低优先级）：

1. `--config` 参数指定的文件
2. 当前目录 `.ntx.yaml`
3. `~/.ntx.yaml`
4. `~/.config/ntx/config.yaml`
5. `/etc/ntx/config.yaml`

可复制 `configs/default.yaml` 到上述任意位置并按需修改。常用环境变量可覆盖配置，例如：

- `NTX_VERBOSE=true`
- `NTX_OUTPUT=json`
- `NTX_DNS_SERVER=1.1.1.1:53`
- `NTX_HTTP_TIMEOUT=10s`

配置文件示例：

```yaml
# 全局配置
global:
  verbose: false
  output: text
  no_color: false
  log_level: info

# Ping 配置
ping:
  count: 4
  timeout: 5s
  interval: 1s
  size: 64
  ttl: 64
  protocol: icmp

# 扫描配置
scan:
  timeout: 3s
  concurrency: 100

# DNS 配置
dns:
  server: "8.8.8.8:53"
  timeout: 5s
  fallback_servers:
    - "8.8.4.4:53"
    - "1.1.1.1:53"
    - "223.5.5.5:53"      # AliDNS (中国大陆)
    - "119.29.29.29:53"   # DNSPod (中国大陆)

# HTTP 配置
http:
  timeout: 30s
  follow_redirect: true
  max_redirects: 10
  user_agent: "NTX/dev"

# Traceroute 配置
trace:
  max_hops: 30
  timeout: 3s
```

## 开发

### 项目结构

```
ntx/
├── cmd/ntx/              # 应用入口
├── internal/
│   ├── cmd/              # CLI 命令实现
│   ├── core/             # 核心业务逻辑
│   ├── logger/           # 日志系统
│   └── output/           # 输出格式化
├── pkg/
│   ├── types/            # 类型定义
│   └── errors/           # 错误处理
├── configs/              # 配置文件
├── docs/                 # 文档
│   └── 关于NTX的优化和重构建议.md  # 代码审查报告
└── Makefile
```

### 构建命令

```bash
# 查看所有可用命令
make help

# 编译
make build

# 运行测试
make test

# 生成测试覆盖率报告
make coverage

# 代码检查
make lint

# 格式化代码
make fmt

# 交叉编译
make cross-build

# 完整构建流程
make all
```

## 贡献

欢迎贡献代码、报告问题或提出建议！

### 贡献步骤

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 开发指南

- 遵循项目代码规范
- 编写测试用例 (单元测试 + 集成测试)
- 更新文档 (代码注释 + README)
- 运行 `make check` 确保所有检查通过

### 提交消息规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型 (type)**:
- `feat`: 新功能
- `fix`: Bug 修复
- `refactor`: 重构
- `test`: 测试
- `docs`: 文档
- `chore`: 构建/工具

**示例**:
```
feat(ping): add DNS caching for host resolution

Implement in-memory DNS cache with configurable TTL to reduce
repeated DNS queries for the same host.

Closes #123
```

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 致谢

- [Cobra](https://github.com/spf13/cobra) - CLI 框架
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Zap](https://github.com/uber-go/zap) - 日志库
- [Color](https://github.com/fatih/color) - 终端颜色输出

## 作者

**Catsayer**

## 支持

如果你觉得这个项目有帮助，请给它一个 ⭐️！

---

<div align="center">
Made with ❤️ by Catsayer
</div>
