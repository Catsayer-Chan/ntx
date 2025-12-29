package batch

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func loadTaskConfig(path string) ([]Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config TaskConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return config.Tasks, nil
}

// GenerateSampleConfig 生成示例配置文件
func GenerateSampleConfig() string {
	return `# NTX 批量任务配置示例

tasks:
  # Ping 监控任务
  - name: "monitor-servers"
    type: "ping"
    enabled: true
    targets:
      - "google.com"
      - "baidu.com"
      - "github.com"
    options:
      count: 5
      timeout: 3
    concurrency: 3

  # DNS 查询任务
  - name: "dns-check"
    type: "dns"
    enabled: true
    targets:
      - "google.com"
      - "example.com"
    options:
      type: "A"
      server: "8.8.8.8"

  # 端口扫描任务
  - name: "scan-servers"
    type: "scan"
    enabled: false
    targets:
      - "192.168.1.1"
    options:
      ports: [22, 80, 443, 3306, 8080]
      timeout: 2
      concurrency: 50
`
}
