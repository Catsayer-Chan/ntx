// Package whois 提供 Whois 查询功能
//
// 本模块实现了 Whois 协议查询:
// - 域名 Whois 查询
// - IP Whois 查询
// - AS 号查询
// - 结果解析
//
// 依赖:
// - net: 标准网络库
// - pkg/types: 类型定义
//
// 使用示例:
//
//	client := whois.NewClient()
//	result, err := client.Query(ctx, "google.com", opts)
//
// 作者: Catsayer
package whois

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
)

// Client Whois 客户端
type Client struct {
	timeout time.Duration
}

// NewClient 创建新的 Whois 客户端
func NewClient() *Client {
	return &Client{
		timeout: 10 * time.Second,
	}
}

// Query 执行 Whois 查询
//
// 参数:
//
//	ctx: 上下文
//	query: 查询内容（域名、IP 或 AS 号）
//	opts: 查询选项
//
// 返回:
//
//	*types.WhoisResult: 查询结果
//	error: 错误信息
func (c *Client) Query(ctx context.Context, query string, opts types.WhoisOptions) (*types.WhoisResult, error) {
	startTime := time.Now()

	logger.Info("开始 Whois 查询", zap.String("query", query))

	// 检测查询类型
	queryType := detectQueryType(query)

	// 选择 Whois 服务器
	server := opts.Server
	if server == "" {
		server = selectWhoisServer(query, queryType)
	}

	// 执行查询
	response, err := c.queryServer(ctx, server, query, opts.Timeout)
	if err != nil {
		return nil, fmt.Errorf("查询 Whois 服务器失败: %w", err)
	}

	result := &types.WhoisResult{
		Query:       query,
		Type:        queryType,
		Server:      server,
		RawResponse: response,
		QueryTime:   time.Since(startTime),
		Timestamp:   time.Now(),
	}

	// 解析响应
	result.ParsedData = parseWhoisResponse(response, queryType)

	logger.Info("Whois 查询完成",
		zap.String("query", query),
		zap.String("server", server),
		zap.Duration("duration", result.QueryTime),
	)

	return result, nil
}

// QueryBatch 批量查询
func (c *Client) QueryBatch(ctx context.Context, queries []string, opts types.WhoisOptions) ([]*types.WhoisResult, error) {
	results := make([]*types.WhoisResult, 0, len(queries))

	for _, query := range queries {
		result, err := c.Query(ctx, query, opts)
		if err != nil {
			logger.Error("查询失败", zap.String("query", query), zap.Error(err))
			continue
		}
		results = append(results, result)

		// 避免频繁查询被限流
		time.Sleep(1 * time.Second)
	}

	return results, nil
}

// queryServer 向 Whois 服务器发送查询
func (c *Client) queryServer(ctx context.Context, server, query string, timeout time.Duration) (string, error) {
	// 确保服务器地址包含端口
	if !strings.Contains(server, ":") {
		server = server + ":43"
	}

	// 连接服务器
	d := net.Dialer{Timeout: timeout}
	conn, err := d.DialContext(ctx, "tcp", server)
	if err != nil {
		return "", fmt.Errorf("连接服务器失败: %w", err)
	}
	defer conn.Close()

	// 设置读写超时
	conn.SetDeadline(time.Now().Add(timeout))

	// 发送查询
	_, err = fmt.Fprintf(conn, "%s\r\n", query)
	if err != nil {
		return "", fmt.Errorf("发送查询失败: %w", err)
	}

	// 读取响应
	var response strings.Builder
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		response.WriteString(scanner.Text())
		response.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	return response.String(), nil
}

// detectQueryType 检测查询类型
func detectQueryType(query string) types.WhoisType {
	// 检查是否是 IP 地址
	if net.ParseIP(query) != nil {
		return types.WhoisIP
	}

	// 检查是否是 AS 号
	if strings.HasPrefix(strings.ToUpper(query), "AS") {
		return types.WhoisAS
	}

	// 默认为域名查询
	return types.WhoisDomain
}

// selectWhoisServer 根据查询类型选择 Whois 服务器
func selectWhoisServer(query string, queryType types.WhoisType) string {
	switch queryType {
	case types.WhoisIP:
		// IP 查询使用 ARIN
		return "whois.arin.net"
	case types.WhoisAS:
		// AS 查询使用 RADB
		return "whois.radb.net"
	case types.WhoisDomain:
		// 根据域名后缀选择服务器
		return selectDomainWhoisServer(query)
	default:
		return "whois.iana.org"
	}
}

// selectDomainWhoisServer 根据域名后缀选择 Whois 服务器
func selectDomainWhoisServer(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return "whois.iana.org"
	}

	tld := parts[len(parts)-1]

	serverMap := map[string]string{
		"com":  "whois.verisign-grs.com",
		"net":  "whois.verisign-grs.com",
		"org":  "whois.pir.org",
		"info": "whois.afilias.net",
		"biz":  "whois.biz",
		"cn":   "whois.cnnic.cn",
		"uk":   "whois.nic.uk",
		"de":   "whois.denic.de",
		"fr":   "whois.afnic.fr",
		"jp":   "whois.jprs.jp",
	}

	if server, ok := serverMap[tld]; ok {
		return server
	}

	return "whois.iana.org"
}

// parseWhoisResponse 解析 Whois 响应
func parseWhoisResponse(response string, queryType types.WhoisType) *types.WhoisData {
	data := &types.WhoisData{}

	switch queryType {
	case types.WhoisDomain:
		parseDomainWhois(response, data)
	case types.WhoisIP:
		parseIPWhois(response, data)
	case types.WhoisAS:
		parseASWhois(response, data)
	}

	return data
}

// parseDomainWhois 解析域名 Whois 响应
func parseDomainWhois(response string, data *types.WhoisData) {
	lines := strings.Split(response, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "%") || strings.HasPrefix(line, "#") {
			continue
		}

		// 分割键值对
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "domain name", "domain":
			data.Domain = value
		case "registrar":
			data.Registrar = value
		case "registrant name":
			data.RegistrantName = value
		case "registrant organization", "registrant org":
			data.RegistrantOrg = value
		case "registrant email":
			data.RegistrantEmail = value
		case "creation date", "created":
			data.CreationDate = parseDate(value)
		case "expiration date", "registry expiry date", "expires":
			data.ExpirationDate = parseDate(value)
		case "updated date", "last updated", "modified":
			data.UpdatedDate = parseDate(value)
		case "name server", "nserver":
			data.NameServers = append(data.NameServers, value)
		case "domain status", "status":
			data.Status = append(data.Status, value)
		}
	}
}

// parseIPWhois 解析 IP Whois 响应
func parseIPWhois(response string, data *types.WhoisData) {
	lines := strings.Split(response, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "%") || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "inetnum", "netrange":
			data.IPRange = value
		case "organization", "orgname", "org-name":
			data.Organization = value
		case "country":
			data.Country = value
		case "city":
			data.City = value
		case "address":
			data.Address = append(data.Address, value)
		}
	}
}

// parseASWhois 解析 AS Whois 响应
func parseASWhois(response string, data *types.WhoisData) {
	lines := strings.Split(response, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "%") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "as-name", "asname":
			data.ASName = value
		case "org-name", "organization":
			data.Organization = value
		}
	}
}

// parseDate 解析日期字符串
func parseDate(dateStr string) time.Time {
	// 常见日期格式
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02-Jan-2006",
		"2006/01/02",
	}

	// 清理日期字符串
	dateStr = strings.TrimSpace(dateStr)
	// 移除时区信息
	re := regexp.MustCompile(`\s+\([^)]+\)$`)
	dateStr = re.ReplaceAllString(dateStr, "")

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	return time.Time{}
}
