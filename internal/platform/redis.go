package platform

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const availabilityChannelPrefix = "spotsync:zone:"

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(redisURL string) (*RedisClient, error) {
	redisURL = strings.TrimSpace(redisURL)
	if redisURL == "" {
		return nil, nil
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse REDIS_URL: %w", err)
	}
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &RedisClient{client: client}, nil
}

func (r *RedisClient) Close() error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Close()
}

func (r *RedisClient) Ping(ctx context.Context) error {
	if r == nil || r.client == nil {
		return fmt.Errorf("redis not configured")
	}
	return r.client.Ping(ctx).Err()
}

func AvailabilityChannel(zoneID uint) string {
	return availabilityChannelPrefix + strconv.FormatUint(uint64(zoneID), 10)
}

func (r *RedisClient) Publish(ctx context.Context, channel, eventType string, payload []byte) error {
	if r == nil || r.client == nil {
		return nil
	}
	msg := eventType + "|" + string(payload)
	return r.client.Publish(ctx, channel, msg).Err()
}

func (r *RedisClient) PublishRaw(ctx context.Context, channel string, payload []byte) error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Publish(ctx, channel, payload).Err()
}

func (r *RedisClient) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.PSubscribe(ctx, patterns...)
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	if r == nil || r.client == nil {
		return "", redis.Nil
	}
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisClient) DecrIfPositive(ctx context.Context, key string) (int64, error) {
	if r == nil || r.client == nil {
		return 0, fmt.Errorf("redis not configured")
	}
	script := redis.NewScript(`
local v = redis.call('GET', KEYS[1])
if not v then return -1 end
local n = tonumber(v)
if n <= 0 then return 0 end
return redis.call('DECR', KEYS[1])
`)
	return script.Run(ctx, r.client, []string{key}).Int64()
}

func (r *RedisClient) Incr(ctx context.Context, key string) error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Incr(ctx, key).Err()
}

func ZoneAvailKey(zoneID uint) string {
	return fmt.Sprintf("spotsync:avail:%d", zoneID)
}

func ZoneCapacityKey(zoneID uint) string {
	return fmt.Sprintf("spotsync:capacity:%d", zoneID)
}

func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	if r == nil || r.client == nil {
		return false, fmt.Errorf("redis not configured")
	}
	n, err := r.client.Exists(ctx, key).Result()
	return n > 0, err
}

func (r *RedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (int64, error) {
	if r == nil || r.client == nil {
		return 0, fmt.Errorf("redis not configured")
	}
	return redis.NewScript(script).Run(ctx, r.client, keys, args...).Int64()
}

func (r *RedisClient) SetNX(ctx context.Context, key string, value int64) error {
	if r == nil || r.client == nil {
		return fmt.Errorf("redis not configured")
	}
	ok, err := r.client.SetNX(ctx, key, strconv.FormatInt(value, 10), 0).Result()
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return nil
}
