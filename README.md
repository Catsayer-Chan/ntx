# NTX (Network Tools eXtended)

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/catsayer/ntx)

**现代化的网络调试命令整合工具**

[特性](#特性) • [安装](#安装) • [快速开始](#快速开始) • [文档](#文档) • [贡献](#贡献)

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

### 核心功能

- **多协议 Ping**
  - ICMP Ping（传统 ping）
  - TCP Ping（通过 TCP 连接测试）
  - HTTP Ping（通过 HTTP 请求测试）

- **路由追踪**
  - ICMP Traceroute
  - UDP Traceroute
  - TCP Traceroute
  - 支持最大跳数、超时时间等配置

- **端口扫描**
  - TCP 端口扫描
  - UDP 端口扫描
  - SYN 扫描（需要 root 权限）
  - 支持端口范围、并发控制

- **DNS 查询**
  - A/AAAA/CNAME/MX/NS/TXT 记录查询
  - 反向 DNS 查询
  - 自定义 DNS 服务器
  - DNS 追踪

- **WHOIS 查询**
  - 域名 WHOIS
  - IP WHOIS
  - 自动选择合适的 WHOIS 服务器

- **连接状态监控**
  - TCP/UDP 连接列表
  - 监听端口查看
  - 进程关联（需要 root 权限）
  - 连接统计

- **网络接口信息**
  - 接口列表
  - IP 地址、MAC 地址
  - 网络统计信息
  - MTU、速率等详细信息

- **智能诊断**
  - 自动诊断网络问题
  - 连通性测试
  - DNS 解析测试
  - 路由测试
  - 生成诊断报告

- **HTTP 客户端**
  - GET/POST/PUT/DELETE 等方法
  - 自定义请求头
  - 请求/响应详情显示
  - 性能测试

- **批量任务**
  - 支持 YAML/JSON 配置文件
  - 并发执行
  - 任务依赖
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

### 基本用法

```bash
# 查看帮助
ntx --help

# Ping 测试
ntx ping google.com -c 5

# 路由追踪
ntx trace google.com

# 端口扫描
ntx scan 192.168.1.1 -p 1-1024

# DNS 查询
ntx dns google.com --type A

# WHOIS 查询
ntx whois google.com

# 查看网络连接
ntx conn --listen

# 查看网络接口
ntx iface --detail

# 智能诊断
ntx diag

# HTTP 请求
ntx http https://api.github.com

# 批量任务
ntx batch -f tasks.yaml
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

NTX 支持通过配置文件自定义行为。配置文件位置：

- Linux/macOS: `~/.ntx.yaml`
- Windows: `%USERPROFILE%\.ntx.yaml`
- 或使用 `--config` 参数指定

配置文件示例：

```yaml
# 全局配置
verbose: false
output: text
no-color: false

# Ping 配置
ping:
  count: 4
  timeout: 5
  interval: 1

# 扫描配置
scan:
  timeout: 3
  concurrency: 100

# DNS 配置
dns:
  server: 8.8.8.8
  timeout: 5

# HTTP 配置
http:
  timeout: 30
  follow-redirects: true
  max-redirects: 10
```

## 开发

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
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 开发指南

- 遵循项目代码规范
- 编写测试用例
- 更新文档
- 运行 `make check` 确保所有检查通过

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

