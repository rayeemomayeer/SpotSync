package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/platform"
)

const defaultAvailTTL = 30 * time.Second

type AvailabilityCache struct {
	redis *platform.RedisClient
	ttl   time.Duration
}

func NewAvailabilityCache(redis *platform.RedisClient, ttl time.Duration) *AvailabilityCache {
	if ttl < time.Second {
		ttl = defaultAvailTTL
	}
	return &AvailabilityCache{redis: redis, ttl: ttl}
}

func (c *AvailabilityCache) Get(ctx context.Context, zoneID uint) (int, bool) {
	if c == nil || c.redis == nil {
		return 0, false
	}
	raw, err := c.redis.Get(ctx, platform.ZoneAvailKey(zoneID))
	if err != nil || raw == "" {
		return 0, false
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return n, true
}

func (c *AvailabilityCache) Set(ctx context.Context, zoneID uint, available int) error {
	if c == nil || c.redis == nil {
		return nil
	}
	return c.redis.Set(ctx, platform.ZoneAvailKey(zoneID), strconv.Itoa(available), c.ttl)
}

func (c *AvailabilityCache) Invalidate(ctx context.Context, zoneID uint) error {
	if c == nil || c.redis == nil {
		return nil
	}
	return c.redis.Del(ctx, platform.ZoneAvailKey(zoneID))
}

func (c *AvailabilityCache) Key(zoneID uint) string {
	return fmt.Sprintf("zone:%d:available", zoneID)
}
