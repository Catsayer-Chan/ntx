# NTX 使用示例

本文档提供 NTX 在各种实际场景中的使用示例。

## 目录

- [网络诊断](#网络诊断)
- [服务监控](#服务监控)
- [性能测试](#性能测试)
- [自动化脚本](#自动化脚本)
- [CI/CD 集成](#cicd-集成)

## 网络诊断

### 检查网络连通性

```bash
# 快速检查主机是否在线
ntx ping google.com --protocol tcp -c 1

# 检查多个主机
for host in google.com baidu.com github.com; do
    echo -n "Checking $host... "
    if ntx ping "$host" --protocol tcp -c 1 -o json 2>/dev/null | jq -e '.statistics.received > 0' >/dev/null; then
        echo "✓ Online"
    else
        echo "✗ Offline"
    fi
done
```

### 诊断网络延迟

```bash
# 测试到不同地区服务器的延迟
declare -A servers=(
    ["US West"]="us-west.example.com"
    ["US East"]="us-east.example.com"
    ["Europe"]="eu.example.com"
    ["Asia"]="asia.example.com"
)

for region in "${!servers[@]}"; do
    host="${servers[$region]}"
    result=$(ntx ping "$host" --protocol tcp -c 10 -o json 2>/dev/null)
    avg_rtt=$(echo "$result" | jq '.statistics.avg_rtt / 1000000')
    echo "$region: ${avg_rtt}ms"
done
```

### 追踪网络路径

```bash
# 追踪到目标的路由路径
sudo ntx trace google.com -m 20

# 保存路由信息
sudo ntx trace google.com -o json > route_$(date +%Y%m%d).json

# 比较不同时间的路由变化
sudo ntx trace example.com -o json | jq '.hops[].ip' > route_now.txt
diff route_before.txt route_now.txt
```

### 诊断丢包问题

```bash
# 长时间测试丢包率
ntx ping problem-host.com --protocol tcp -c 100 -i 0.5 -o json | \
    jq '.statistics | "Loss: \(.loss_rate)%, Avg RTT: \(.avg_rtt/1000000)ms"'

# 监控丢包率
watch -n 10 'ntx ping google.com --protocol tcp -c 20 -o json | jq .statistics'
```

## 服务监控

### HTTP 服务健康检查

```bash
#!/bin/bash

check_http_service() {
    local url=$1
    local expected_status=${2:-200}

    result=$(ntx ping "$url" --protocol http -c 1 -o json 2>/dev/null)

    if [ $? -eq 0 ]; then
        received=$(echo "$result" | jq '.statistics.received')
        if [ "$received" -gt 0 ]; then
            echo "✓ $url is healthy"
            return 0
        fi
    fi

    echo "✗ $url is down"
    return 1
}

# 检查多个服务
services=(
    "https://api.example.com"
    "https://www.example.com"
    "https://cdn.example.com"
)

for service in "${services[@]}"; do
    check_http_service "$service"
done
```

### 端口可用性监控

```bash
#!/bin/bash

# 监控关键端口
PORTS=(80 443 22 3306 6379)
HOST="production-server.com"

for port in "${PORTS[@]}"; do
    echo -n "Port $port: "
    result=$(ntx ping "$HOST" --protocol tcp --port "$port" -c 1 -o json 2>/dev/null)

    if echo "$result" | jq -e '.statistics.received > 0' >/dev/null 2>&1; then
        rtt=$(echo "$result" | jq '.replies[0].rtt / 1000000')
        echo "✓ Open (${rtt}ms)"
    else
        echo "✗ Closed or filtered"
    fi
done
```

### 服务响应时间监控

```bash
#!/bin/bash

# 持续监控服务响应时间
LOG_FILE="service_monitor.log"
ALERT_THRESHOLD=500  # 毫秒

monitor_service() {
    local url=$1
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    result=$(ntx ping "$url" --protocol http -c 5 -o json 2>/dev/null)

    if [ $? -eq 0 ]; then
        avg_rtt=$(echo "$result" | jq '.statistics.avg_rtt / 1000000')
        loss_rate=$(echo "$result" | jq '.statistics.loss_rate')

        echo "$timestamp,$url,$avg_rtt,$loss_rate" >> "$LOG_FILE"

        # 检查是否超过阈值
        if (( $(echo "$avg_rtt > $ALERT_THRESHOLD" | bc -l) )); then
            echo "⚠ ALERT: $url response time ${avg_rtt}ms exceeds threshold"
            # 发送告警通知...
        fi
    fi
}

# 持续监控
while true; do
    monitor_service "https://api.example.com"
    sleep 60
done
```

## 性能测试

### 网络带宽估算

```bash
#!/bin/bash

# 通过 HTTP Ping 测试下载速度
test_bandwidth() {
    local url=$1
    local count=10

    echo "Testing $url..."
    result=$(ntx ping "$url" --protocol http -c "$count" -o json 2>/dev/null)

    total_bytes=$(echo "$result" | jq '[.replies[].bytes] | add')
    total_time=$(echo "$result" | jq '.statistics.total_time / 1000000000')
    avg_rtt=$(echo "$result" | jq '.statistics.avg_rtt / 1000000')

    if [ "$total_time" != "0" ]; then
        bandwidth=$(echo "scale=2; $total_bytes / $total_time / 1024" | bc)
        echo "Average RTT: ${avg_rtt}ms"
        echo "Approximate bandwidth: ${bandwidth} KB/s"
    fi
}

test_bandwidth "https://www.google.com"
```

### 负载测试

```bash
#!/bin/bash

# 并发负载测试
concurrent_test() {
    local url=$1
    local concurrency=${2:-10}
    local iterations=${3:-5}

    echo "Running concurrent test: $concurrency connections, $iterations iterations each"

    for i in $(seq 1 "$concurrency"); do
        (
            ntx ping "$url" --protocol http -c "$iterations" -o json > "result_$i.json" 2>&1
        ) &
    done

    wait

    # 汇总结果
    echo "Results:"
    for i in $(seq 1 "$concurrency"); do
        if [ -f "result_$i.json" ]; then
            avg_rtt=$(jq '.statistics.avg_rtt / 1000000' "result_$i.json" 2>/dev/null || echo "N/A")
            loss=$(jq '.statistics.loss_rate' "result_$i.json" 2>/dev/null || echo "N/A")
            echo "Thread $i: Avg RTT=${avg_rtt}ms, Loss=${loss}%"
        fi
    done

    # 清理
    rm -f result_*.json
}

concurrent_test "https://api.example.com" 20 10
```

### 性能基准测试

```bash
#!/bin/bash

# 对比不同协议性能
benchmark() {
    local host=$1

    echo "=== Benchmarking $host ==="

    # TCP Ping
    echo -n "TCP Ping: "
    tcp_result=$(ntx ping "$host" --protocol tcp --port 443 -c 50 -o json 2>/dev/null)
    tcp_avg=$(echo "$tcp_result" | jq '.statistics.avg_rtt / 1000000')
    tcp_std=$(echo "$tcp_result" | jq '.statistics.stddev_rtt / 1000000')
    echo "Avg=${tcp_avg}ms, StdDev=${tcp_std}ms"

    # HTTP Ping
    echo -n "HTTP Ping: "
    http_result=$(ntx ping "https://$host" --protocol http -c 50 -o json 2>/dev/null)
    http_avg=$(echo "$http_result" | jq '.statistics.avg_rtt / 1000000')
    http_std=$(echo "$http_result" | jq '.statistics.stddev_rtt / 1000000')
    echo "Avg=${http_avg}ms, StdDev=${http_std}ms"

    # ICMP Ping (需要 sudo)
    if [ "$EUID" -eq 0 ]; then
        echo -n "ICMP Ping: "
        icmp_result=$(ntx ping "$host" --protocol icmp -c 50 -o json 2>/dev/null)
        icmp_avg=$(echo "$icmp_result" | jq '.statistics.avg_rtt / 1000000')
        icmp_std=$(echo "$icmp_result" | jq '.statistics.stddev_rtt / 1000000')
        echo "Avg=${icmp_avg}ms, StdDev=${icmp_std}ms"
    fi
}

benchmark "google.com"
benchmark "github.com"
```

## 自动化脚本

### 自动故障检测和恢复

```bash
#!/bin/bash

# 服务故障检测和自动重启
SERVICE_NAME="my-service"
CHECK_URL="http://localhost:8080/health"
MAX_FAILURES=3
failures=0

while true; do
    result=$(ntx ping "$CHECK_URL" --protocol http -c 1 -o json 2>/dev/null)

    if echo "$result" | jq -e '.statistics.received == 0' >/dev/null 2>&1; then
        ((failures++))
        echo "$(date): Service check failed ($failures/$MAX_FAILURES)"

        if [ $failures -ge $MAX_FAILURES ]; then
            echo "$(date): Restarting $SERVICE_NAME..."
            systemctl restart "$SERVICE_NAME"
            failures=0
            sleep 30
        fi
    else
        if [ $failures -gt 0 ]; then
            echo "$(date): Service recovered"
        fi
        failures=0
    fi

    sleep 10
done
```

### 网络质量报告生成

```bash
#!/bin/bash

# 生成每日网络质量报告
REPORT_DATE=$(date +%Y-%m-%d)
REPORT_FILE="network_report_$REPORT_DATE.html"

generate_report() {
    cat > "$REPORT_FILE" <<EOF
<!DOCTYPE html>
<html>
<head>
    <title>Network Quality Report - $REPORT_DATE</title>
    <style>
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #4CAF50; color: white; }
        .good { color: green; }
        .warning { color: orange; }
        .bad { color: red; }
    </style>
</head>
<body>
    <h1>Network Quality Report - $REPORT_DATE</h1>
    <table>
        <tr>
            <th>Host</th>
            <th>Avg RTT</th>
            <th>Loss Rate</th>
            <th>Status</th>
        </tr>
EOF

    hosts=("google.com" "github.com" "baidu.com")

    for host in "${hosts[@]}"; do
        result=$(ntx ping "$host" --protocol tcp -c 20 -o json 2>/dev/null)

        avg_rtt=$(echo "$result" | jq '.statistics.avg_rtt / 1000000')
        loss_rate=$(echo "$result" | jq '.statistics.loss_rate')

        # 判断状态
        if (( $(echo "$loss_rate < 1 && $avg_rtt < 100" | bc -l) )); then
            status="<span class='good'>Good</span>"
        elif (( $(echo "$loss_rate < 5 && $avg_rtt < 200" | bc -l) )); then
            status="<span class='warning'>Warning</span>"
        else
            status="<span class='bad'>Bad</span>"
        fi

        cat >> "$REPORT_FILE" <<EOF
        <tr>
            <td>$host</td>
            <td>${avg_rtt}ms</td>
            <td>${loss_rate}%</td>
            <td>$status</td>
        </tr>
EOF
    done

    cat >> "$REPORT_FILE" <<EOF
    </table>
    <p>Generated at: $(date)</p>
</body>
</html>
EOF

    echo "Report generated: $REPORT_FILE"
}

generate_report
```

### 多区域延迟监控

```bash
#!/bin/bash

# 监控多个地理区域的网络延迟
declare -A regions=(
    ["US-West"]="us-west.example.com"
    ["US-East"]="us-east.example.com"
    ["Europe"]="eu.example.com"
    ["Asia"]="asia.example.com"
)

OUTPUT_FILE="regional_latency.csv"
echo "timestamp,region,host,avg_rtt,min_rtt,max_rtt,loss_rate" > "$OUTPUT_FILE"

while true; do
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    for region in "${!regions[@]}"; do
        host="${regions[$region]}"
        result=$(ntx ping "$host" --protocol tcp -c 10 -o json 2>/dev/null)

        if [ $? -eq 0 ]; then
            avg_rtt=$(echo "$result" | jq '.statistics.avg_rtt / 1000000')
            min_rtt=$(echo "$result" | jq '.statistics.min_rtt / 1000000')
            max_rtt=$(echo "$result" | jq '.statistics.max_rtt / 1000000')
            loss_rate=$(echo "$result" | jq '.statistics.loss_rate')

            echo "$timestamp,$region,$host,$avg_rtt,$min_rtt,$max_rtt,$loss_rate" >> "$OUTPUT_FILE"
        fi
    done

    sleep 300  # 每 5 分钟检测一次
done
```

## CI/CD 集成

### GitHub Actions

```yaml
name: Network Health Check

on:
  schedule:
    - cron: '*/30 * * * *'  # 每 30 分钟
  workflow_dispatch:

jobs:
  health-check:
    runs-on: ubuntu-latest
    steps:
      - name: Install NTX
        run: |
          wget https://github.com/Catsayer-Chan/ntx/releases/latest/download/ntx-linux-amd64
          chmod +x ntx-linux-amd64
          sudo mv ntx-linux-amd64 /usr/local/bin/ntx

      - name: Check Production Services
        run: |
          services=(
            "https://api.example.com"
            "https://www.example.com"
            "https://cdn.example.com"
          )

          for service in "${services[@]}"; do
            echo "Checking $service..."
            ntx ping "$service" --protocol http -c 5 -o json > result.json

            loss_rate=$(jq '.statistics.loss_rate' result.json)
            avg_rtt=$(jq '.statistics.avg_rtt / 1000000' result.json)

            echo "$service: Loss=$loss_rate%, Avg RTT=${avg_rtt}ms"

            if (( $(echo "$loss_rate > 10" | bc -l) )); then
              echo "::error::High packet loss detected for $service"
              exit 1
            fi
          done

      - name: Upload Results
        uses: actions/upload-artifact@v2
        with:
          name: health-check-results
          path: result.json
```

### GitLab CI

```yaml
network_health_check:
  stage: test
  image: ubuntu:latest
  script:
    - apt-get update && apt-get install -y wget jq bc
    - wget https://github.com/Catsayer-Chan/ntx/releases/latest/download/ntx-linux-amd64
    - chmod +x ntx-linux-amd64 && mv ntx-linux-amd64 /usr/local/bin/ntx
    - |
      for service in api.example.com www.example.com; do
        echo "Testing $service..."
        ntx ping "$service" --protocol tcp --port 443 -c 5 -o json > result.json
        loss_rate=$(jq '.statistics.loss_rate' result.json)
        if (( $(echo "$loss_rate > 5" | bc -l) )); then
          echo "ERROR: High packet loss: $loss_rate%"
          exit 1
        fi
      done
  artifacts:
    paths:
      - result.json
    expire_in: 1 week
  only:
    - schedules
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any

    triggers {
        cron('H/30 * * * *')  // 每 30 分钟
    }

    stages {
        stage('Install NTX') {
            steps {
                sh '''
                    if [ ! -f /usr/local/bin/ntx ]; then
                        wget https://github.com/Catsayer-Chan/ntx/releases/latest/download/ntx-linux-amd64
                        chmod +x ntx-linux-amd64
                        sudo mv ntx-linux-amd64 /usr/local/bin/ntx
                    fi
                '''
            }
        }

        stage('Health Check') {
            steps {
                script {
                    def services = [
                        'api.example.com',
                        'www.example.com',
                        'cdn.example.com'
                    ]

                    services.each { service ->
                        sh """
                            echo "Checking ${service}..."
                            ntx ping "${service}" --protocol tcp --port 443 -c 5 -o json > ${service}.json
                        """

                        def result = readJSON file: "${service}.json"
                        def lossRate = result.statistics.loss_rate

                        if (lossRate > 5) {
                            error("High packet loss for ${service}: ${lossRate}%")
                        }
                    }
                }
            }
        }
    }

    post {
        always {
            archiveArtifacts artifacts: '*.json', fingerprint: true
        }
        failure {
            emailext (
                subject: "Network Health Check Failed",
                body: "One or more services failed the health check.",
                to: "ops@example.com"
            )
        }
    }
}
```

## Python 集成示例

```python
#!/usr/bin/env python3
import subprocess
import json
from typing import Dict, Optional

class NTXClient:
    """NTX Python 客户端封装"""

    def __init__(self, ntx_path: str = 'ntx'):
        self.ntx_path = ntx_path

    def ping(self, target: str, protocol: str = 'tcp',
             count: int = 3, **kwargs) -> Optional[Dict]:
        """执行 Ping 操作"""
        cmd = [
            self.ntx_path, 'ping', target,
            '--protocol', protocol,
            '-c', str(count),
            '-o', 'json'
        ]

        # 添加额外参数
        for key, value in kwargs.items():
            cmd.extend([f'--{key}', str(value)])

        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                check=True
            )
            return json.loads(result.stdout)
        except subprocess.CalledProcessError as e:
            print(f"Error: {e}")
            return None
        except json.JSONDecodeError as e:
            print(f"JSON decode error: {e}")
            return None

    def trace(self, target: str, max_hops: int = 30) -> Optional[Dict]:
        """执行 Traceroute 操作"""
        cmd = [
            self.ntx_path, 'trace', target,
            '-m', str(max_hops),
            '-o', 'json'
        ]

        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                check=True
            )
            return json.loads(result.stdout)
        except (subprocess.CalledProcessError, json.JSONDecodeError) as e:
            print(f"Error: {e}")
            return None

# 使用示例
if __name__ == '__main__':
    client = NTXClient()

    # Ping 测试
    result = client.ping('google.com', protocol='tcp', count=5)
    if result:
        stats = result['statistics']
        print(f"Loss Rate: {stats['loss_rate']}%")
        print(f"Avg RTT: {stats['avg_rtt']/1000000:.2f}ms")

    # Traceroute 测试
    result = client.trace('google.com', max_hops=15)
    if result:
        print(f"Hops: {result['hop_count']}")
        print(f"Reached: {result['reached_destination']}")
```

## 总结

这些示例展示了 NTX 在各种实际场景中的应用。您可以根据具体需求修改和组合这些示例。

更多信息请参考：
- [使用说明](usage.md)
- [安装指南](installation.md)
- [API 文档](api.md)