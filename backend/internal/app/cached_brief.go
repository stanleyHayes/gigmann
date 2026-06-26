package app

import (
	"context"
	"sync"
	"time"

	"github.com/xcreativs/gigmann/internal/core/brief"
	"github.com/xcreativs/gigmann/internal/ports"
)

// CachedBrief wraps a BriefGenerator with a TTL cache. The Daily Brief is an
// expensive LLM call, so it is generated at most once per TTL: the first (cold)
// request generates synchronously, later requests serve the cache instantly and
// trigger a background refresh once stale. A failed refresh keeps the last good
// brief (serve-stale-on-error), so a transient model outage never blanks the brief.
type CachedBrief struct {
	inner ports.BriefGenerator
	ttl   time.Duration
	now   func() time.Time

	mu         sync.Mutex
	cached     brief.Brief
	at         time.Time
	has        bool
	refreshing bool
}

var _ ports.BriefGenerator = (*CachedBrief)(nil)

// NewCachedBrief wraps inner with a TTL cache.
func NewCachedBrief(inner ports.BriefGenerator, ttl time.Duration) *CachedBrief {
	return &CachedBrief{inner: inner, ttl: ttl, now: time.Now}
}

// Generate serves the cached brief when present (refreshing in the background
// when stale); the first call generates synchronously.
func (c *CachedBrief) Generate(ctx context.Context) (brief.Brief, error) {
	c.mu.Lock()
	if c.has {
		cached := c.cached
		if c.now().Sub(c.at) >= c.ttl && !c.refreshing {
			c.refreshing = true
			go c.refresh() //nolint:contextcheck // detached background refresh, not request-scoped
		}
		c.mu.Unlock()
		return cached, nil
	}
	c.mu.Unlock()

	b, err := c.inner.Generate(ctx)
	if err != nil {
		return brief.Brief{}, err
	}
	c.store(b)
	return b, nil
}

func (c *CachedBrief) refresh() {
	defer func() {
		c.mu.Lock()
		c.refreshing = false
		c.mu.Unlock()
	}()
	if b, err := c.inner.Generate(context.Background()); err == nil {
		c.store(b)
	}
}

func (c *CachedBrief) store(b brief.Brief) {
	c.mu.Lock()
	c.cached, c.at, c.has = b, c.now(), true
	c.mu.Unlock()
}
