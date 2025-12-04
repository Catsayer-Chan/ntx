# NTX 安装指南

本文档介绍如何在不同平台上安装和配置 NTX (Network Tools eXtended)。

## 系统要求

- **操作系统**: Linux, macOS, Windows
- **Go 版本**: 1.21 或更高
- **权限**: 部分功能需要 root 权限或 CAP_NET_RAW 能力

## 安装方法

### 方法 1: 从源码编译（推荐）

#### 1. 安装 Go

确保已安装 Go 1.21 或更高版本：

```bash
go version
```

如果未安装，请访问 [Go 官网](https://go.dev/dl/) 下载安装。

#### 2. 克隆仓库

```bash
git clone https://github.com/Catsayer-Chan/ntx.git
cd ntx
```

#### 3. 编译

使用 Makefile 编译（推荐）：

```bash
make build
```

或直接使用 go build：

```bash
go build -o bin/ntx ./cmd/ntx
```

#### 4. 验证安装

```bash
./bin/ntx --version
```

### 方法 2: 使用 Go Install

直接从 GitHub 安装：

```bash
go install github.com/Catsayer-Chan/ntx/cmd/ntx@latest
```

安装后，`ntx` 命令将在 `$GOPATH/bin` 目录下。

### 方法 3: 下载预编译二进制文件

从 [Releases 页面](https://github.com/Catsayer-Chan/ntx/releases) 下载适合您系统的预编译版本。

#### Linux/macOS

```bash
# 下载
wget https://github.com/Catsayer-Chan/ntx/releases/download/v0.1.0/ntx-linux-amd64

# 添加执行权限
chmod +x ntx-linux-amd64

# 移动到系统路径
sudo mv ntx-linux-amd64 /usr/local/bin/ntx
```

#### Windows

1. 从 Releases 页面下载 `ntx-windows-amd64.exe`
2. 将文件重命名为 `ntx.exe`
3. 添加到系统 PATH 环境变量

## 系统安装

### Linux

#### 方法 1: 使用 make install

```bash
cd ntx
make install
```

这会将 `ntx` 安装到 `$GOPATH/bin`。

#### 方法 2: 手动安装

```bash
# 编译
make build

# 复制到系统目录
sudo cp bin/ntx /usr/local/bin/

# 验证
ntx --version
```

#### 设置权限（可选）

为了使用 ICMP Ping 和 Traceroute 而不需要 root 权限：

```bash
# 设置 CAP_NET_RAW 能力
sudo setcap cap_net_raw+ep /usr/local/bin/ntx
```

### macOS

#### 使用 Homebrew（推荐）

```bash
# 添加 tap
brew tap Catsayer-Chan/ntx

# 安装
brew install ntx
```

#### 手动安装

```bash
# 编译
make build

# 复制到系统目录
sudo cp bin/ntx /usr/local/bin/

# 验证
ntx --version
```

### Windows

#### 使用 Chocolatey

```powershell
# 安装
choco install ntx
```

#### 手动安装

1. 编译或下载 `ntx.exe`
2. 将 `ntx.exe` 放到任意目录（如 `C:\Program Files\ntx\`）
3. 添加到 PATH 环境变量：
   - 右键"此电脑" → "属性" → "高级系统设置" → "环境变量"
   - 在"系统变量"中找到 Path，点击"编辑"
   - 添加 ntx.exe 所在目录

## 配置

### 配置文件

NTX 支持通过配置文件自定义默认行为。配置文件位置：

- **Linux/macOS**: `~/.ntx.yaml` 或 `~/.config/ntx/config.yaml`
- **Windows**: `%USERPROFILE%\.ntx.yaml`

### 配置文件示例

创建 `~/.ntx.yaml`:

```yaml
# 全局配置
verbose: false
output: text
no-color: false

# Ping 配置
ping:
  protocol: tcp
  count: 4
  timeout: 5
  interval: 1
  port: 443

# Traceroute 配置
traceroute:
  max-hops: 30
  timeout: 3
  queries: 3

# 日志配置
log:
  level: info
  file: ~/.ntx/logs/ntx.log
```

### 环境变量

NTX 支持通过环境变量配置：

```bash
# 设置输出格式
export NTX_OUTPUT=json

# 启用详细输出
export NTX_VERBOSE=true

# 禁用彩色输出
export NTX_NO_COLOR=true
```

## 权限说明

### Linux

某些功能需要特殊权限：

| 功能 | 所需权限 | 说明 |
|------|----------|------|
| ICMP Ping | root 或 CAP_NET_RAW | 需要创建原始套接字 |
| TCP Ping | 普通用户 | 无需特殊权限 |
| HTTP Ping | 普通用户 | 无需特殊权限 |
| ICMP Traceroute | root 或 CAP_NET_RAW | 需要创建原始套接字 |
| 端口扫描 | 普通用户 | 无需特殊权限 |

#### 授予 CAP_NET_RAW 权限

```bash
sudo setcap cap_net_raw+ep $(which ntx)
```

### macOS

macOS 对原始套接字有严格限制：

- ICMP Ping: 需要 sudo
- ICMP Traceroute: 需要 sudo
- 其他功能: 普通用户

使用 sudo 运行：

```bash
sudo ntx ping google.com
sudo ntx trace google.com
```

### Windows

Windows 需要管理员权限运行某些功能：

1. 右键点击命令提示符
2. 选择"以管理员身份运行"
3. 执行 ntx 命令

或者在 PowerShell 中：

```powershell
# 以管理员身份运行
Start-Process ntx -ArgumentList "ping google.com" -Verb RunAs
```

## 验证安装

### 基本测试

```bash
# 查看版本
ntx --version

# 查看帮助
ntx --help

# TCP Ping 测试（无需 root）
ntx ping baidu.com --protocol tcp --port 443 -c 3

# HTTP Ping 测试（无需 root）
ntx ping https://www.google.com --protocol http -c 3
```

### 完整功能测试

```bash
# ICMP Ping（需要 root）
sudo ntx ping google.com -c 4

# Traceroute（需要 root）
sudo ntx trace google.com -m 15

# JSON 输出测试
ntx ping baidu.com --protocol tcp -c 2 -o json

# 详细输出测试
ntx ping google.com --protocol tcp -v
```

## 故障排除

### 常见问题

#### 1. "permission denied" 错误

**问题**: 运行 ICMP Ping 或 Traceroute 时出现权限错误

**解决方案**:
```bash
# Linux: 使用 sudo
sudo ntx ping google.com

# 或者授予 CAP_NET_RAW 权限
sudo setcap cap_net_raw+ep $(which ntx)

# macOS: 使用 sudo
sudo ntx ping google.com
```

#### 2. "command not found" 错误

**问题**: 找不到 ntx 命令

**解决方案**:
```bash
# 检查 PATH
echo $PATH

# 添加到 PATH（临时）
export PATH=$PATH:$(pwd)/bin

# 添加到 PATH（永久，Linux/macOS）
echo 'export PATH=$PATH:~/ntx/bin' >> ~/.bashrc
source ~/.bashrc
```

#### 3. 编译错误

**问题**: 编译时出现依赖错误

**解决方案**:
```bash
# 清理并重新下载依赖
go clean -modcache
go mod download
go mod tidy

# 重新编译
make clean
make build
```

#### 4. ICMP Ping 自动降级到 TCP

**问题**: ICMP Ping 自动切换到 TCP Ping

**原因**: 缺少必要权限

**解决方案**: 使用 sudo 或授予 CAP_NET_RAW 权限（见上文）

## 更新

### 从源码更新

```bash
cd ntx
git pull origin main
make clean
make build
make install
```

### 使用 Go Install 更新

```bash
go install github.com/Catsayer-Chan/ntx/cmd/ntx@latest
```

### 使用包管理器更新

```bash
# Homebrew (macOS)
brew upgrade ntx

# Chocolatey (Windows)
choco upgrade ntx
```

## 卸载

### Linux/macOS

```bash
# 删除二进制文件
sudo rm /usr/local/bin/ntx

# 删除配置文件
rm -rf ~/.ntx.yaml ~/.config/ntx

# 如果使用 Go Install 安装
rm $GOPATH/bin/ntx
```

### Windows

1. 删除 `ntx.exe` 文件
2. 从 PATH 环境变量中移除相关目录
3. 删除配置文件 `%USERPROFILE%\.ntx.yaml`

## 下一步

- 阅读 [使用说明](usage.md) 了解详细功能
- 查看 [示例](examples.md) 学习常用场景
- 参考 [API 文档](api.md) 进行二次开发