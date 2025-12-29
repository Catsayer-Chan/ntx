package diag

// getDefaultGateway 获取默认网关
func getDefaultGateway() (string, error) {
	// TODO: 解析路由表获取真实网关
	return "192.168.1.1", nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
