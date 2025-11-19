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

package maputil

import "fmt"

// Keys returns all keys from map
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values returns all values from map
func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// ContainsKey checks if map contains key
func ContainsKey[K comparable, V any](m map[K]V, key K) bool {
	_, exists := m[key]
	return exists
}

// Get returns value and existence flag
func Get[K comparable, V any](m map[K]V, key K) (V, bool) {
	value, exists := m[key]
	return value, exists
}

// GetOrDefault returns value or default if key doesn't exist
func GetOrDefault[K comparable, V any](m map[K]V, key K, defaultValue V) V {
	if value, exists := m[key]; exists {
		return value
	}
	return defaultValue
}

// Filter filters map entries based on predicate
func Filter[K comparable, V any](m map[K]V, fn func(K, V) bool) map[K]V {
	result := make(map[K]V, len(m))
	for k, v := range m {
		if fn(k, v) {
			result[k] = v
		}
	}
	return result
}

// Map applies function to each entry and returns new map
func Map[K comparable, V any, R any](m map[K]V, fn func(K, V) R) map[K]R {
	result := make(map[K]R, len(m))
	for k, v := range m {
		result[k] = fn(k, v)
	}
	return result
}

// Merge merges multiple maps into one (later maps override earlier ones)
func Merge[K comparable, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// Invert inverts map (keys become values, values become keys)
func Invert[K comparable, V comparable](m map[K]V) map[V]K {
	result := make(map[V]K, len(m))
	for k, v := range m {
		if _, exists := result[v]; exists {
			panic(fmt.Sprintf("duplicate value %v when inverting map", v))
		}
		result[v] = k
	}
	return result
}

// GroupBy groups slice elements by key function
func GroupBy[T any, K comparable](slice []T, keyFn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, v := range slice {
		key := keyFn(v)
		result[key] = append(result[key], v)
	}
	return result
}

// IndexBy indexes slice by key function
func IndexBy[T any, K comparable](slice []T, keyFn func(T) K) map[K]T {
	result := make(map[K]T, len(slice))
	for _, v := range slice {
		key := keyFn(v)
		result[key] = v
	}
	return result
}

// Reduce reduces map to a single value
func Reduce[K comparable, V any, R any](m map[K]V, initial R, fn func(R, K, V) R) R {
	result := initial
	for k, v := range m {
		result = fn(result, k, v)
	}
	return result
}

// Any checks if any entry satisfies predicate
func Any[K comparable, V any](m map[K]V, fn func(K, V) bool) bool {
	for k, v := range m {
		if fn(k, v) {
			return true
		}
	}
	return false
}

// All checks if all entries satisfy predicate
func All[K comparable, V any](m map[K]V, fn func(K, V) bool) bool {
	for k, v := range m {
		if !fn(k, v) {
			return false
		}
	}
	return true
}

// Clone creates a shallow copy of map
func Clone[K comparable, V any](m map[K]V) map[K]V {
	result := make(map[K]V, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// Equal checks if two maps are equal
func Equal[K comparable, V comparable](m1, m2 map[K]V) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v1 := range m1 {
		if v2, exists := m2[k]; !exists || v1 != v2 {
			return false
		}
	}
	return true
}

// Pick picks selected keys from map
func Pick[K comparable, V any](m map[K]V, keys []K) map[K]V {
	result := make(map[K]V, len(keys))
	for _, k := range keys {
		if v, exists := m[k]; exists {
			result[k] = v
		}
	}
	return result
}

// Omit omits selected keys from map
func Omit[K comparable, V any](m map[K]V, keys []K) map[K]V {
	keySet := make(map[K]bool, len(keys))
	for _, k := range keys {
		keySet[k] = true
	}

	result := make(map[K]V, len(m))
	for k, v := range m {
		if !keySet[k] {
			result[k] = v
		}
	}
	return result
}

// FromPairs creates map from key-value pairs
func FromPairs[K comparable, V any](pairs []Pair[K, V]) map[K]V {
	result := make(map[K]V, len(pairs))
	for _, p := range pairs {
		result[p.Key] = p.Value
	}
	return result
}

// ToPairs converts map to key-value pairs
func ToPairs[K comparable, V any](m map[K]V) []Pair[K, V] {
	pairs := make([]Pair[K, V], 0, len(m))
	for k, v := range m {
		pairs = append(pairs, Pair[K, V]{Key: k, Value: v})
	}
	return pairs
}

// Pair represents a key-value pair
type Pair[K any, V any] struct {
	Key   K
	Value V
}
