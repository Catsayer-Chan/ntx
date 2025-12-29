package netutil

import (
	"net"

	"github.com/catsayer/ntx/pkg/errors"
	"github.com/catsayer/ntx/pkg/types"
)

// ResolveHost 解析主机名或 IP 文本，按照 IP 版本偏好返回匹配的地址
func ResolveHost(host string, ipVersion types.IPVersion) (*types.Host, error) {
	if ip := net.ParseIP(host); ip != nil {
		ver := types.IPv4
		if ip.To4() == nil {
			ver = types.IPv6
		}
		return &types.Host{
			Hostname:  host,
			IP:        ip.String(),
			IPVersion: ver,
		}, nil
	}

	if cached, ok := defaultDNSCache.get(cacheKey(host, ipVersion)); ok {
		return cached, nil
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, errors.ErrDNSResolution
	}
	if len(ips) == 0 {
		return nil, errors.ErrNoAddress
	}

	selectedIP := selectIPByVersion(ips, ipVersion)
	if selectedIP == nil {
		return nil, errors.ErrNoAddress
	}

	ver := types.IPv4
	if selectedIP.To4() == nil {
		ver = types.IPv6
	}

	resolved := &types.Host{
		Hostname:  host,
		IP:        selectedIP.String(),
		IPVersion: ver,
	}

	defaultDNSCache.set(cacheKey(host, ipVersion), resolved)
	return resolved, nil
}

func selectIPByVersion(ips []net.IP, version types.IPVersion) net.IP {
	for _, ip := range ips {
		switch version {
		case types.IPv4:
			if ip.To4() != nil {
				return ip
			}
		case types.IPv6:
			if ip.To4() == nil {
				return ip
			}
		default:
			return ip
		}
	}
	return nil
}
