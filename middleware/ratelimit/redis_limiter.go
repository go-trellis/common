/*
Copyright © 2025 Henry Huang <hhh@rutcode.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package ratelimit

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisOptions captures the minimal options needed
type RedisOptions struct {
	Addr        string
	Username    string
	Password    string
	DB          int
	KeyPrefix   string
	PoolSize    int
	DialTimeout time.Duration
}

// RedisLimiter implements a distributed token-bucket using Redis + Lua
type RedisLimiter struct {
	client     redis.UniversalClient
	keyPrefix  string
	rate       int64
	period     time.Duration
	burst      int64
	allowBatch int64
}

// NewRedisLimiter creates a Redis-backed limiter
func NewRedisLimiter(opts RedisOptions, cfg *Config) (*RedisLimiter, error) {
	if cfg == nil {
		cfg = NewConfig(100, time.Second)
	}
	if cfg.Rate <= 0 {
		cfg.Rate = 100
	}
	if cfg.Period <= 0 {
		cfg.Period = time.Second
	}
	if cfg.Burst <= 0 {
		cfg.Burst = cfg.Rate
	}
	if opts.KeyPrefix == "" {
		opts.KeyPrefix = "rl:"
	}
	rc := redis.NewClient(&redis.Options{
		Addr:        opts.Addr,
		Username:    opts.Username,
		Password:    opts.Password,
		DB:          opts.DB,
		PoolSize:    opts.PoolSize,
		DialTimeout: opts.DialTimeout,
	})
	return &RedisLimiter{
		client:     rc,
		keyPrefix:  opts.KeyPrefix,
		rate:       cfg.Rate,
		period:     cfg.Period,
		burst:      cfg.Burst,
		allowBatch: 1,
	}, nil
}

// Close closes the client if needed
func (r *RedisLimiter) Close() error {
	if c, ok := r.client.(*redis.Client); ok {
		return c.Close()
	}
	return nil
}

// Lua script implements token bucket with refill and TTL
var luaTokenBucket = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local period = tonumber(ARGV[3])
local burst = tonumber(ARGV[4])
local tokens = tonumber(ARGV[5])

local data = redis.call('HMGET', key, 't', 'ts')
local cur_tokens = tonumber(data[1])
local last_ts = tonumber(data[2])

if not cur_tokens or not last_ts then
  cur_tokens = burst
  last_ts = now
  redis.call('HMSET', key, 't', cur_tokens, 'ts', last_ts)
  redis.call('PEXPIRE', key, period*2)
end

local elapsed = now - last_ts
if elapsed > 0 then
  local refill = (elapsed / period) * rate
  cur_tokens = math.min(burst, cur_tokens + refill)
  last_ts = now
end

local allowed = 0
if cur_tokens >= tokens then
  cur_tokens = cur_tokens - tokens
  allowed = 1
end

redis.call('HMSET', key, 't', cur_tokens, 'ts', last_ts)
redis.call('PEXPIRE', key, period*2)
return tostring(allowed)
`)

func (r *RedisLimiter) redisKey(key string) string {
	return r.keyPrefix + key
}

// Allow checks allowance using Redis
func (r *RedisLimiter) Allow(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, errors.New("empty key")
	}
	nowMs := time.Now().UnixMilli()
	res, err := luaTokenBucket.Run(ctx, r.client, []string{r.redisKey(key)},
		strconv.FormatInt(nowMs, 10),
		strconv.FormatInt(r.rate, 10),
		strconv.FormatInt(r.period.Milliseconds(), 10),
		strconv.FormatInt(r.burst, 10),
		strconv.FormatInt(r.allowBatch, 10),
	).Result()
	if err != nil {
		return false, err
	}
	switch v := res.(type) {
	case string:
		return v == "1", nil
	case int64:
		return v == 1, nil
	}
	return false, nil
}

// Wait polls Allow until permitted or ctx cancelled
func (r *RedisLimiter) Wait(ctx context.Context, key string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			ok, err := r.Allow(ctx, key)
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
			sleep := r.period / 10
			if sleep <= 0 {
				sleep = 50 * time.Millisecond
			}
			timer := time.NewTimer(sleep)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
}
