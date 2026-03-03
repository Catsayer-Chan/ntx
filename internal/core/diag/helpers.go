package diag

import (
	"fmt"
	"strings"

	"github.com/catsayer/ntx/pkg/types"
)

// getDefaultGateway 获取默认网关
func (s *Service) getDefaultGateway() (string, error) {
	if s == nil || s.ifReader == nil {
		return "", fmt.Errorf("网络接口读取器未初始化")
	}

	routes, err := s.ifReader.GetRoutes()
	if err != nil {
		return "", fmt.Errorf("读取路由表失败: %w", err)
	}

	gateway := findDefaultGateway(routes)
	if gateway == "" {
		return "", fmt.Errorf("未找到默认网关")
	}
	return gateway, nil
}

func findDefaultGateway(routes []*types.Route) string {
	for _, route := range routes {
		if route == nil {
			continue
		}
		dest := strings.ToLower(strings.TrimSpace(route.Destination))
		if dest == "default" ||
			dest == "0.0.0.0" ||
			dest == "0.0.0.0/0.0.0.0" ||
			dest == "::/0" ||
			dest == "::" {
			gateway := strings.TrimSpace(route.Gateway)
			if gateway != "" && gateway != "*" && !strings.HasPrefix(gateway, "link#") {
				return gateway
			}
		}
	}
	return ""
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
