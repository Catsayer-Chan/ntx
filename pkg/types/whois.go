// Package types 定义 NTX 工具的公共类型
//
// 本文件定义 Whois 查询相关的类型和接口
//
// 作者: Catsayer
package types

import (
	"net"
	"time"
)

// WhoisType Whois 查询类型
type WhoisType int

const (
	// WhoisDomain 域名 Whois
	WhoisDomain WhoisType = iota
	// WhoisIP IP Whois
	WhoisIP
	// WhoisAS AS 号查询
	WhoisAS
)

// WhoisOptions Whois 查询选项
type WhoisOptions struct {
	// Server Whois 服务器地址（可选）
	Server string
	// Timeout 查询超时时间
	Timeout time.Duration
	// FollowReferrals 是否跟随 referral
	FollowReferrals bool
}

// DefaultWhoisOptions 返回默认 Whois 选项
func DefaultWhoisOptions() WhoisOptions {
	return WhoisOptions{
		Server:          "",
		Timeout:         10 * time.Second,
		FollowReferrals: true,
	}
}

// WhoisResult Whois 查询结果
type WhoisResult struct {
	// Query 查询内容
	Query string
	// Type 查询类型
	Type WhoisType
	// Server 使用的 Whois 服务器
	Server string
	// RawResponse 原始响应数据
	RawResponse string
	// ParsedData 解析后的数据
	ParsedData *WhoisData
	// QueryTime 查询耗时
	QueryTime time.Duration
	// Timestamp 查询时间戳
	Timestamp time.Time
}

// WhoisData 解析后的 Whois 数据
type WhoisData struct {
	// 域名信息
	Domain          string    `json:"domain,omitempty"`
	Registrar       string    `json:"registrar,omitempty"`
	RegistrantName  string    `json:"registrant_name,omitempty"`
	RegistrantOrg   string    `json:"registrant_org,omitempty"`
	RegistrantEmail string    `json:"registrant_email,omitempty"`
	CreationDate    time.Time `json:"creation_date,omitempty"`
	ExpirationDate  time.Time `json:"expiration_date,omitempty"`
	UpdatedDate     time.Time `json:"updated_date,omitempty"`
	NameServers     []string  `json:"name_servers,omitempty"`
	Status          []string  `json:"status,omitempty"`

	// IP 信息
	IP           net.IP   `json:"ip,omitempty"`
	IPRange      string   `json:"ip_range,omitempty"`
	Organization string   `json:"organization,omitempty"`
	Country      string   `json:"country,omitempty"`
	City         string   `json:"city,omitempty"`
	Address      []string `json:"address,omitempty"`

	// AS 信息
	ASN    int    `json:"asn,omitempty"`
	ASName string `json:"as_name,omitempty"`

	// 其他字段
	AdminContact string `json:"admin_contact,omitempty"`
	TechContact  string `json:"tech_contact,omitempty"`
}
