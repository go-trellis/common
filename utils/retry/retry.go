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

package retry

import (
	"context"
	"errors"
	"math"
	"time"
)

var (
	// ErrMaxRetriesReached indicates that max retries have been reached
	ErrMaxRetriesReached = errors.New("max retries reached")
)

// BackoffStrategy defines how to calculate backoff delay
type BackoffStrategy interface {
	Next(retry int) time.Duration
}

// Config configures retry behavior
type Config struct {
	// MaxRetries is the maximum number of retries (0 means no retry, -1 means infinite)
	MaxRetries int

	// InitialDelay is the initial delay before first retry
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration

	// BackoffStrategy defines how to calculate backoff delay
	// If nil, uses ExponentialBackoff by default
	BackoffStrategy BackoffStrategy

	// RetryableErrors is a function that determines if an error should be retried
	// If nil, all errors are retried
	RetryableErrors func(error) bool
}

// ExponentialBackoff implements exponential backoff strategy
type ExponentialBackoff struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
	Factor    float64
}

// Next calculates the next backoff delay
func (b *ExponentialBackoff) Next(retry int) time.Duration {
	delay := time.Duration(float64(b.BaseDelay) * math.Pow(b.Factor, float64(retry)))
	if b.MaxDelay > 0 && delay > b.MaxDelay {
		delay = b.MaxDelay
	}
	return delay
}

// LinearBackoff implements linear backoff strategy
type LinearBackoff struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
	Increment time.Duration
}

// Next calculates the next backoff delay
func (b *LinearBackoff) Next(retry int) time.Duration {
	delay := b.BaseDelay + time.Duration(retry)*b.Increment
	if b.MaxDelay > 0 && delay > b.MaxDelay {
		delay = b.MaxDelay
	}
	return delay
}

// FixedBackoff implements fixed backoff strategy
type FixedBackoff struct {
	Delay time.Duration
}

// Next returns the fixed delay
func (b *FixedBackoff) Next(retry int) time.Duration {
	return b.Delay
}

// Do executes a function with retry logic
func Do(ctx context.Context, cfg Config, fn func() error) error {
	if cfg.MaxRetries == 0 {
		return fn()
	}

	maxRetries := cfg.MaxRetries
	if maxRetries < 0 {
		maxRetries = math.MaxInt
	}

	backoff := cfg.BackoffStrategy
	if backoff == nil {
		backoff = &ExponentialBackoff{
			BaseDelay: cfg.InitialDelay,
			MaxDelay:  cfg.MaxDelay,
			Factor:    2.0,
		}
		if cfg.InitialDelay == 0 {
			backoff.(*ExponentialBackoff).BaseDelay = 100 * time.Millisecond
		}
		if cfg.MaxDelay == 0 {
			backoff.(*ExponentialBackoff).MaxDelay = 30 * time.Second
		}
	}

	retryable := cfg.RetryableErrors
	if retryable == nil {
		retryable = func(error) bool { return true }
	}

	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		if ctx != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if !retryable(err) {
			return err
		}

		if i < maxRetries {
			delay := backoff.Next(i)
			if delay > 0 {
				if ctx != nil {
					timer := time.NewTimer(delay)
					select {
					case <-ctx.Done():
						timer.Stop()
						return ctx.Err()
					case <-timer.C:
					}
				} else {
					time.Sleep(delay)
				}
			}
		}
	}

	if lastErr != nil {
		return lastErr
	}
	return ErrMaxRetriesReached
}

// DoWithResult executes a function with retry logic and returns result
func DoWithResult[T any](ctx context.Context, cfg Config, fn func() (T, error)) (T, error) {
	var zero T
	if cfg.MaxRetries == 0 {
		return fn()
	}

	maxRetries := cfg.MaxRetries
	if maxRetries < 0 {
		maxRetries = math.MaxInt
	}

	backoff := cfg.BackoffStrategy
	if backoff == nil {
		backoff = &ExponentialBackoff{
			BaseDelay: cfg.InitialDelay,
			MaxDelay:  cfg.MaxDelay,
			Factor:    2.0,
		}
		if cfg.InitialDelay == 0 {
			backoff.(*ExponentialBackoff).BaseDelay = 100 * time.Millisecond
		}
		if cfg.MaxDelay == 0 {
			backoff.(*ExponentialBackoff).MaxDelay = 30 * time.Second
		}
	}

	retryable := cfg.RetryableErrors
	if retryable == nil {
		retryable = func(error) bool { return true }
	}

	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		if ctx != nil {
			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			default:
			}
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		if !retryable(err) {
			return zero, err
		}

		if i < maxRetries {
			delay := backoff.Next(i)
			if delay > 0 {
				if ctx != nil {
					timer := time.NewTimer(delay)
					select {
					case <-ctx.Done():
						timer.Stop()
						return zero, ctx.Err()
					case <-timer.C:
					}
				} else {
					time.Sleep(delay)
				}
			}
		}
	}

	if lastErr != nil {
		return zero, lastErr
	}
	return zero, ErrMaxRetriesReached
}
