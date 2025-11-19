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
	"time"
)

// Limiter defines the interface for rate limiting
type Limiter interface {
	// Allow checks if a request is allowed
	Allow(ctx context.Context, key string) (bool, error)
	// Wait waits until a request is allowed
	Wait(ctx context.Context, key string) error
}

// KeyExtractor extracts a key from context for rate limiting
type KeyExtractor func(ctx context.Context) string

// Config defines rate limiter configuration
type Config struct {
	// Rate is the number of requests allowed per period
	Rate int64 `json:"rate" yaml:"rate"`
	// Period is the time period for the rate limit
	Period time.Duration `json:"period" yaml:"period"`
	// Burst is the maximum burst size (for token bucket)
	Burst int64 `json:"burst" yaml:"burst"`
}

// NewConfig creates a new rate limiter config with default values
func NewConfig(rate int64, period time.Duration) *Config {
	return &Config{
		Rate:   rate,
		Period: period,
		Burst:  rate,
	}
}
