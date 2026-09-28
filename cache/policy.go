package cache

import (
	"strings"
	"time"
)

// CachePolicyMode decides when a cached tile may be served and when the
// origin sources are consulted again.
//
//	cache_first  (default) A cached tile is served forever; sources are only
//	                 consulted on a cache miss. Freshness is managed by
//	                 seeding/cleaning the cache explicitly.
//	ttl              A cached tile is served until it is older than TTL,
//	                 then the next request refreshes it from the sources.
//	revalidate       Like ttl, but an expired tile is still served from the
//	                 cache while a background refresh runs (the response is
//	                 never delayed by the origin).
//	bypass           The cache is neither read nor written; every request
//	                 goes to the sources.
type CachePolicyMode string

const (
	PolicyCacheFirst CachePolicyMode = "cache_first"
	PolicyTTL        CachePolicyMode = "ttl"
	PolicyRevalidate CachePolicyMode = "revalidate"
	PolicyBypass     CachePolicyMode = "bypass"
)

// DefaultPolicy matches the historical behaviour: serve whatever is cached.
func DefaultPolicy() CachePolicy {
	return CachePolicy{Mode: PolicyCacheFirst}
}

// CachePolicy is the parsed form of the per-cache "cache_policy" setting.
type CachePolicy struct {
	Mode CachePolicyMode
	TTL  time.Duration // effective for ttl and revalidate; 0 = never expires
}

// ParseCachePolicy validates and normalizes a mode string and ttl hours.
func ParseCachePolicy(mode string, ttlHours int) (CachePolicy, error) {
	normalized := CachePolicyMode(strings.ToLower(strings.TrimSpace(mode)))
	if normalized == "" {
		normalized = PolicyCacheFirst
	}
	switch normalized {
	case PolicyCacheFirst:
		return DefaultPolicy(), nil
	case PolicyBypass:
		return CachePolicy{Mode: PolicyBypass}, nil
	case PolicyTTL, PolicyRevalidate:
		if ttlHours <= 0 {
			return CachePolicy{}, ErrInvalidCachePolicy{Reason: "ttl_hours must be > 0 for ttl/revalidate mode"}
		}
		return CachePolicy{Mode: normalized, TTL: time.Duration(ttlHours) * time.Hour}, nil
	}
	return CachePolicy{}, ErrInvalidCachePolicy{Reason: "unknown mode " + mode}
}

// ErrInvalidCachePolicy is returned for malformed policy settings.
type ErrInvalidCachePolicy struct {
	Reason string
}

func (e ErrInvalidCachePolicy) Error() string {
	return "invalid cache policy: " + e.Reason
}

// expired reports whether a tile with the given modification time may no
// longer be served under this policy.
func (p CachePolicy) expired(modTime time.Time, now time.Time) bool {
	switch p.Mode {
	case PolicyTTL, PolicyRevalidate:
		if p.TTL <= 0 {
			return false
		}
		return modTime.Add(p.TTL).Before(now)
	}
	return false
}

// shouldStoreFor reports whether the manager persists freshly created tiles
// (bypass policy never does). Uses a type assertion so mock managers in
// tests are unaffected.
func shouldStoreFor(m Manager) bool {
	if tm, ok := m.(*TileManager); ok {
		return tm.ShouldStoreToCache()
	}
	return true
}
