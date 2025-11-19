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
	"testing"
	"time"
)

func TestTokenBucketLimiter(t *testing.T) {
	config := NewConfig(10, time.Second)
	limiter := NewTokenBucketLimiter(config)

	ctx := context.Background()
	key := "test_key"

	// First 10 requests should be allowed
	for i := 0; i < 10; i++ {
		allowed, err := limiter.Allow(ctx, key)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !allowed {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// 11th request should be rate limited
	allowed, err := limiter.Allow(ctx, key)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("request 11 should be rate limited")
	}
}

func TestTokenBucketLimiter_DifferentKeys(t *testing.T) {
	config := NewConfig(5, time.Second)
	limiter := NewTokenBucketLimiter(config)

	ctx := context.Background()

	// Different keys should have separate rate limits
	key1 := "key1"
	key2 := "key2"

	// Use all tokens for key1
	for i := 0; i < 5; i++ {
		allowed, _ := limiter.Allow(ctx, key1)
		if !allowed {
			t.Errorf("request %d for key1 should be allowed", i+1)
		}
	}

	// Key2 should still have tokens
	allowed, _ := limiter.Allow(ctx, key2)
	if !allowed {
		t.Error("key2 should still have tokens available")
	}
}

func TestNewConfig(t *testing.T) {
	config := NewConfig(100, time.Second)
	if config.Rate != 100 {
		t.Errorf("expected rate 100, got %d", config.Rate)
	}
	if config.Period != time.Second {
		t.Errorf("expected period 1s, got %v", config.Period)
	}
	if config.Burst != 100 {
		t.Errorf("expected burst 100, got %d", config.Burst)
	}
}
