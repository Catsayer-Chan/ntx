package netutil

import (
	"net"
	"testing"

	"github.com/catsayer/ntx/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestResolveHostLiteralIPs(t *testing.T) {
	host, err := ResolveHost("127.0.0.1", types.IPvAny)
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1", host.IP)
	require.Equal(t, types.IPv4, host.IPVersion)

	host, err = ResolveHost("::1", types.IPvAny)
	require.NoError(t, err)
	require.Equal(t, "::1", host.IP)
	require.Equal(t, types.IPv6, host.IPVersion)
}

func TestSelectIPByVersion(t *testing.T) {
	ipv4 := net.ParseIP("192.0.2.1")
	ipv6 := net.ParseIP("2001:db8::1")
	ips := []net.IP{ipv6, ipv4}

	require.Equal(t, ipv4, selectIPByVersion(ips, types.IPv4))
	require.Equal(t, ipv6, selectIPByVersion(ips, types.IPv6))
	require.Equal(t, ipv6, selectIPByVersion(ips, types.IPvAny))
	require.Nil(t, selectIPByVersion([]net.IP{ipv4}, types.IPv6))
}
