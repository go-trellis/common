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

import "context"

// MultiLimiter composes multiple limiters; only allows if ALL allow
type MultiLimiter struct {
	limiters []Limiter
}

// NewMultiLimiter creates a multi-limiter; nils are ignored
func NewMultiLimiter(limiters ...Limiter) *MultiLimiter {
	ls := make([]Limiter, 0, len(limiters))
	for _, l := range limiters {
		if l != nil {
			ls = append(ls, l)
		}
	}
	return &MultiLimiter{limiters: ls}
}

// Allow returns true only if every underlying limiter allows
func (m *MultiLimiter) Allow(ctx context.Context, key string) (bool, error) {
	for _, l := range m.limiters {
		ok, err := l.Allow(ctx, key)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// Wait waits on each limiter sequentially
func (m *MultiLimiter) Wait(ctx context.Context, key string) error {
	for _, l := range m.limiters {
		if err := l.Wait(ctx, key); err != nil {
			return err
		}
	}
	return nil
}
