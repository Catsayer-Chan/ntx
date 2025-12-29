// Package dns 提供 DNS 查询功能
//
// 本模块实现了 DNS 查询功能,包括:
// - A/AAAA/CNAME/MX/NS/TXT/SOA 等记录查询
// - 反向 DNS 查询
// - 自定义 DNS 服务器
// - 批量查询
//
// 依赖:
// - github.com/miekg/dns: DNS 客户端库
//
// 使用示例:
//
//	resolver := dns.NewResolver(&dns.Options{Server: types.DefaultDNSServer})
//	result, err := resolver.Query("google.com", dns.TypeA)
//
// 作者: Catsayer
package dns

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/catsayer/ntx/pkg/errors"
	"github.com/catsayer/ntx/pkg/types"
	"github.com/miekg/dns"
)

// Resolver DNS 解析器
type Resolver struct {
	options *types.DNSOptions
	client  *dns.Client
}

// NewResolver 创建新的 DNS 解析器
func NewResolver(opts *types.DNSOptions) *Resolver {
	if opts == nil {
		opts = &types.DNSOptions{
			Server:  types.DefaultDNSServer,
			Timeout: types.DefaultDNSTimeout,
		}
	}

	if opts.Server == "" {
		opts.Server = types.DefaultDNSServer
	}
	opts.Server = types.FormatDNSServer(opts.Server)
	if opts.Timeout == 0 {
		opts.Timeout = types.DefaultDNSTimeout
	}

	return &Resolver{
		options: opts,
		client: &dns.Client{
			Timeout: opts.Timeout,
		},
	}
}

// Query 执行 DNS 查询
func (r *Resolver) Query(ctx context.Context, domain string, recordType types.DNSRecordType) (*types.DNSResult, error) {
	startTime := time.Now()

	// 验证域名
	if domain == "" {
		return nil, errors.ErrInvalidDomain
	}

	// 确保域名以 . 结尾
	if !strings.HasSuffix(domain, ".") {
		domain = domain + "."
	}

	// 构建 DNS 查询消息
	msg := new(dns.Msg)
	msg.SetQuestion(domain, uint16(recordType))
	msg.RecursionDesired = true

	// 执行查询
	response, rtt, err := r.client.ExchangeContext(ctx, msg, r.options.Server)
	if err != nil {
		return nil, fmt.Errorf("DNS 查询失败: %w", err)
	}

	// 检查响应状态
	if response.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("DNS 查询失败: %s", dns.RcodeToString[response.Rcode])
	}

	// 解析响应
	result := &types.DNSResult{
		Domain:     strings.TrimSuffix(domain, "."),
		RecordType: recordType,
		Server:     r.options.Server,
		QueryTime:  rtt,
		StartTime:  startTime,
		EndTime:    time.Now(),
		Records:    make([]*types.DNSRecord, 0),
	}

	// 解析 Answer 部分
	for _, rr := range response.Answer {
		record := r.parseRecord(rr)
		if record != nil {
			result.Records = append(result.Records, record)
		}
	}

	// 解析 Authority 部分
	for _, rr := range response.Ns {
		record := r.parseRecord(rr)
		if record != nil {
			result.Authority = append(result.Authority, record)
		}
	}

	// 解析 Additional 部分
	for _, rr := range response.Extra {
		record := r.parseRecord(rr)
		if record != nil {
			result.Additional = append(result.Additional, record)
		}
	}

	return result, nil
}

// QueryAll 查询所有常见记录类型
func (r *Resolver) QueryAll(ctx context.Context, domain string) (map[types.DNSRecordType]*types.DNSResult, error) {
	recordTypes := []types.DNSRecordType{
		types.DNSTypeA,
		types.DNSTypeAAAA,
		types.DNSTypeCNAME,
		types.DNSTypeMX,
		types.DNSTypeNS,
		types.DNSTypeTXT,
		types.DNSTypeSOA,
	}

	results := make(map[types.DNSRecordType]*types.DNSResult)

	for _, recordType := range recordTypes {
		result, err := r.Query(ctx, domain, recordType)
		if err == nil && len(result.Records) > 0 {
			results[recordType] = result
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("未找到任何 DNS 记录")
	}

	return results, nil
}

// QueryBatch 批量查询多个域名
func (r *Resolver) QueryBatch(ctx context.Context, domains []string, recordType types.DNSRecordType) ([]*types.DNSResult, error) {
	results := make([]*types.DNSResult, 0, len(domains))

	for _, domain := range domains {
		result, err := r.Query(ctx, domain, recordType)
		if err != nil {
			// 失败的查询也记录
			result = &types.DNSResult{
				Domain:     domain,
				RecordType: recordType,
				Server:     r.options.Server,
				Error:      err,
			}
		}
		results = append(results, result)
	}

	return results, nil
}

// Reverse 反向 DNS 查询
func (r *Resolver) Reverse(ctx context.Context, ip string) (*types.DNSResult, error) {
	// 验证 IP 地址
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, errors.ErrInvalidIP
	}

	// 构建反向查询域名
	var reverseDomain string
	if parsedIP.To4() != nil {
		// IPv4
		parts := strings.Split(ip, ".")
		if len(parts) != 4 {
			return nil, errors.ErrInvalidIP
		}
		reverseDomain = fmt.Sprintf("%s.%s.%s.%s.in-addr.arpa", parts[3], parts[2], parts[1], parts[0])
	} else {
		// IPv6
		reverseDomain, _ = dns.ReverseAddr(ip)
	}

	// 执行 PTR 查询
	return r.Query(ctx, reverseDomain, types.DNSTypePTR)
}

// parseRecord 解析 DNS 记录
func (r *Resolver) parseRecord(rr dns.RR) *types.DNSRecord {
	header := rr.Header()
	record := &types.DNSRecord{
		Name: strings.TrimSuffix(header.Name, "."),
		Type: types.DNSRecordType(header.Rrtype),
		TTL:  int(header.Ttl),
	}

	switch v := rr.(type) {
	case *dns.A:
		record.Value = v.A.String()
	case *dns.AAAA:
		record.Value = v.AAAA.String()
	case *dns.CNAME:
		record.Value = strings.TrimSuffix(v.Target, ".")
	case *dns.MX:
		record.Value = fmt.Sprintf("%d %s", v.Preference, strings.TrimSuffix(v.Mx, "."))
	case *dns.NS:
		record.Value = strings.TrimSuffix(v.Ns, ".")
	case *dns.TXT:
		record.Value = strings.Join(v.Txt, " ")
	case *dns.SOA:
		record.Value = fmt.Sprintf("%s %s %d %d %d %d %d",
			strings.TrimSuffix(v.Ns, "."),
			strings.TrimSuffix(v.Mbox, "."),
			v.Serial, v.Refresh, v.Retry, v.Expire, v.Minttl)
	case *dns.PTR:
		record.Value = strings.TrimSuffix(v.Ptr, ".")
	case *dns.SRV:
		record.Value = fmt.Sprintf("%d %d %d %s",
			v.Priority, v.Weight, v.Port, strings.TrimSuffix(v.Target, "."))
	default:
		record.Value = strings.TrimSpace(strings.TrimPrefix(rr.String(), header.String()))
	}

	return record
}

// Close 关闭解析器 (当前无需实际操作)
func (r *Resolver) Close() error {
	return nil
}
