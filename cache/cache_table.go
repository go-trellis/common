/*
Copyright © 2016 Henry Huang <hhh@rutcode.com>

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

package cache

import (
	"time"
)

// NoExpire Timers
const (
	NoExpire time.Duration = 0
)

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback func(key any, value any)

// TableCache table manager for k-vs functions
type TableCache interface {
	// Insert the object or all of the objects in list.
	Insert(key, values any) bool
	// InsertExpire insert the object or all of the objects with expired time in list.
	InsertExpire(key, value any, expire time.Duration) bool
	// DeleteObject Deletes all objects with key: Key.
	DeleteObject(key any) bool
	// DeleteObjects Delete all objects in the table Tab. Remain table in cache.
	DeleteObjects()
	// Member Returns true if one or more elements in the table has key: Key, otherwise false.
	Member(key any) bool
	// Members Returns all keys in the table Tab.
	Members() ([]any, bool)
	// Lookup Look up values with key: Key.
	Lookup(key any) ([]any, bool)
	// LookupAll Look up all values in the Tab.
	LookupAll() (map[any][]any, bool)
	// SetExpire Set Key Expire time
	SetExpire(key any, expire time.Duration) bool
}

// OptionFunc 参数处理函数
type OptionFunc func(*Options)

// Options configure
type Options struct {
	ValueMode ValueMode

	Size int

	Evict EvictCallback
}

// OptionValueMode set the values' model
func OptionValueMode(mode ValueMode) OptionFunc {
	return func(t *Options) {
		t.ValueMode = mode
	}
}

// OptionKeySize set the size of keys
func OptionKeySize(size int) OptionFunc {
	return func(t *Options) {
		t.Size = size
	}
}

// OptionEvict set the evict ballback
func OptionEvict(evict EvictCallback) OptionFunc {
	return func(t *Options) {
		t.Evict = evict
	}
}
