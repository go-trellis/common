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

package context

import (
	"context"
	"time"
)

// WithTimeout creates context with timeout, returns context and cancel function
func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(parent, timeout)
}

// WithDeadline creates context with deadline
func WithDeadline(parent context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithDeadline(parent, deadline)
}

// WithValue wraps context.WithValue with type safety
func WithValue[K comparable, V any](parent context.Context, key K, val V) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithValue(parent, key, val)
}

// Value safely retrieves value from context with type assertion
func Value[K comparable, V any](ctx context.Context, key K) (V, bool) {
	if ctx == nil {
		var zero V
		return zero, false
	}
	val := ctx.Value(key)
	if val == nil {
		var zero V
		return zero, false
	}
	v, ok := val.(V)
	return v, ok
}

// ValueOrDefault retrieves value from context or returns default
func ValueOrDefault[K comparable, V any](ctx context.Context, key K, defaultValue V) V {
	if val, ok := Value[K, V](ctx, key); ok {
		return val
	}
	return defaultValue
}

// IsDone checks if context is done/cancelled
func IsDone(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// Err returns context error if context is done, nil otherwise
func Err(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// WithCancel creates context with cancel function
func WithCancel(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithCancel(parent)
}

// Background returns background context
func Background() context.Context {
	return context.Background()
}

// TODO returns TODO context
func TODO() context.Context {
	return context.TODO()
}

// Merge merges multiple contexts, returns first cancelled context or nil
func Merge(ctxs ...context.Context) context.Context {
	if len(ctxs) == 0 {
		return context.Background()
	}
	if len(ctxs) == 1 {
		if ctxs[0] == nil {
			return context.Background()
		}
		return ctxs[0]
	}

	ctx, cancel := context.WithCancel(context.Background())

	for _, c := range ctxs {
		if c == nil {
			continue
		}
		go func(c context.Context) {
			select {
			case <-c.Done():
				cancel()
			case <-ctx.Done():
			}
		}(c)
	}

	return ctx
}

// WithValues merges multiple key-value pairs into context
func WithValues(parent context.Context, values map[interface{}]interface{}) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	ctx := parent
	for k, v := range values {
		ctx = context.WithValue(ctx, k, v)
	}
	return ctx
}
