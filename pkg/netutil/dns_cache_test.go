package netutil

import (
	"testing"
	"time"

	"github.com/catsayer/ntx/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestDNSCacheStoresAndClonesHost(t *testing.T) {
	cache := newDNSCache(5 * time.Minute)
	key := cacheKey("Example.COM", types.IPv4)

	original := &types.Host{
		Hostname:  "example.com",
		IP:        "1.1.1.1",
		IPVersion: types.IPv4,
	}

	cache.set(key, original)

	cached, ok := cache.get(key)
	require.True(t, ok)
	require.NotNil(t, cached)
	require.Equal(t, "example.com", cached.Hostname)
	require.Equal(t, "1.1.1.1", cached.IP)

	cached.IP = "9.9.9.9"
	cachedAgain, ok := cache.get(key)
	require.True(t, ok)
	require.Equal(t, "1.1.1.1", cachedAgain.IP)
}
