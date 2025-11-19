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
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// TokenBucketLimiter implements Limiter using token bucket algorithm
type TokenBucketLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	config   *Config
}

// NewTokenBucketLimiter creates a new token bucket limiter
func NewTokenBucketLimiter(config *Config) *TokenBucketLimiter {
	if config == nil {
		config = NewConfig(100, time.Second)
	}
	if config.Rate <= 0 {
		config.Rate = 100
	}
	if config.Period <= 0 {
		config.Period = time.Second
	}
	if config.Burst <= 0 {
		config.Burst = config.Rate
	}
	return &TokenBucketLimiter{
		limiters: make(map[string]*rate.Limiter),
		config:   config,
	}
}

// getLimiter gets or creates a limiter for the given key
func (t *TokenBucketLimiter) getLimiter(key string) *rate.Limiter {
	// Fast path: read lock for existing limiters
	t.mu.RLock()
	limiter, exists := t.limiters[key]
	t.mu.RUnlock()

	if exists {
		return limiter
	}

	// Slow path: write lock for creating new limiter
	t.mu.Lock()
	defer t.mu.Unlock()

	// Double check after acquiring write lock (another goroutine might have created it)
	if limiter, exists := t.limiters[key]; exists {
		return limiter
	}

	// Calculate rate limit per second
	periodSeconds := t.config.Period.Seconds()
	if periodSeconds <= 0 {
		periodSeconds = 1.0
	}
	rateLimit := rate.Limit(float64(t.config.Rate) / periodSeconds)

	// Create new limiter
	limiter = rate.NewLimiter(rateLimit, int(t.config.Burst))
	t.limiters[key] = limiter
	return limiter
}

// Allow checks if a request is allowed
func (t *TokenBucketLimiter) Allow(ctx context.Context, key string) (bool, error) {
	limiter := t.getLimiter(key)
	return limiter.Allow(), nil
}

// Wait waits until a request is allowed
func (t *TokenBucketLimiter) Wait(ctx context.Context, key string) error {
	limiter := t.getLimiter(key)
	return limiter.Wait(ctx)
}
