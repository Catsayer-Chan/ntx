package netutil

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/catsayer/ntx/pkg/types"
)

const defaultDNSTTL = 5 * time.Minute

var defaultDNSCache = newDNSCache(defaultDNSTTL)

type cacheEntry struct {
	host      *types.Host
	expiresAt time.Time
}

type dnsCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
}

func newDNSCache(ttl time.Duration) *dnsCache {
	return &dnsCache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
	}
}

func (c *dnsCache) get(key string) (*types.Host, bool) {
	if c == nil || c.ttl <= 0 {
		return nil, false
	}

	now := time.Now()
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if now.After(entry.expiresAt) {
		c.mu.Lock()
		if cur, exists := c.entries[key]; exists && now.After(cur.expiresAt) {
			delete(c.entries, key)
		}
		c.mu.Unlock()
		return nil, false
	}

	return cloneHost(entry.host), true
}

func (c *dnsCache) set(key string, host *types.Host) {
	if c == nil || c.ttl <= 0 {
		return
	}

	c.mu.Lock()
	c.entries[key] = cacheEntry{
		host:      cloneHost(host),
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

func cloneHost(h *types.Host) *types.Host {
	if h == nil {
		return nil
	}
	cloned := *h
	return &cloned
}

func cacheKey(host string, version types.IPVersion) string {
	return strings.ToLower(host) + "|" + strconv.Itoa(int(version))
}
