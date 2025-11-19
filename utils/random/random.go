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

package random

import (
	cryptorand "crypto/rand"
	"math/big"
	mathrand "math/rand/v2"
	"time"
)

var defaultRand = cryptorand.Reader
var defaultMathRand = mathrand.New(mathrand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano())))

// String generates a random string of specified length using provided charset
func String(length int, charset string) string {
	if length <= 0 {
		return ""
	}
	if charset == "" {
		charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}

	result := make([]byte, length)
	for i := range result {
		n, err := cryptorand.Int(defaultRand, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback to pseudo-random if crypto/rand fails
			result[i] = charset[defaultMathRand.IntN(len(charset))]
			continue
		}
		result[i] = charset[n.Int64()]
	}
	return string(result)
}

// AlphaNumeric generates a random alphanumeric string
func AlphaNumeric(length int) string {
	return String(length, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
}

// Numeric generates a random numeric string
func Numeric(length int) string {
	return String(length, "0123456789")
}

// Alpha generates a random alphabetic string
func Alpha(length int) string {
	return String(length, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
}

// Int generates a random integer in range [min, max]
func Int(min, max int) int {
	if min >= max {
		return min
	}
	n, err := cryptorand.Int(defaultRand, big.NewInt(int64(max-min+1)))
	if err != nil {
		// Fallback to pseudo-random
		return min + defaultMathRand.IntN(max-min+1)
	}
	return min + int(n.Int64())
}

// Int64 generates a random int64 in range [min, max]
func Int64(min, max int64) int64 {
	if min >= max {
		return min
	}
	n, err := cryptorand.Int(defaultRand, big.NewInt(max-min+1))
	if err != nil {
		// Fallback to pseudo-random
		return min + defaultMathRand.Int64N(max-min+1)
	}
	return min + n.Int64()
}

// Float64 generates a random float64 in range [min, max)
func Float64(min, max float64) float64 {
	if min >= max {
		return min
	}
	return min + (max-min)*defaultMathRand.Float64()
}

// Bytes generates random bytes of specified length
func Bytes(length int) []byte {
	if length <= 0 {
		return nil
	}
	b := make([]byte, length)
	_, err := cryptorand.Read(b)
	if err != nil {
		// Fallback to pseudo-random
		for i := range b {
			b[i] = byte(defaultMathRand.IntN(256))
		}
	}
	return b
}

// Choice randomly selects one element from slice
func Choice[T any](slice []T) (T, error) {
	var zero T
	if len(slice) == 0 {
		return zero, ErrEmptySlice
	}
	index := Int(0, len(slice)-1)
	return slice[index], nil
}

// Choices randomly selects n elements from slice (with replacement)
func Choices[T any](slice []T, n int) []T {
	if len(slice) == 0 || n <= 0 {
		return nil
	}
	result := make([]T, n)
	for i := range result {
		index := Int(0, len(slice)-1)
		result[i] = slice[index]
	}
	return result
}

// Shuffle randomly shuffles slice
func Shuffle[T any](slice []T) []T {
	if len(slice) <= 1 {
		return slice
	}
	result := make([]T, len(slice))
	copy(result, slice)

	r := mathrand.New(mathrand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano())))
	for i := len(result) - 1; i > 0; i-- {
		j := r.IntN(i + 1)
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// WeightedChoice randomly selects based on weights
func WeightedChoice[T any](items []T, weights []float64) (T, error) {
	var zero T
	if len(items) != len(weights) {
		return zero, ErrWeightsMismatch
	}
	if len(items) == 0 {
		return zero, ErrEmptySlice
	}

	totalWeight := 0.0
	for _, w := range weights {
		if w < 0 {
			return zero, ErrNegativeWeight
		}
		totalWeight += w
	}

	if totalWeight == 0 {
		return zero, ErrZeroTotalWeight
	}

	r := Float64(0, totalWeight)
	currentWeight := 0.0

	for i, w := range weights {
		currentWeight += w
		if r <= currentWeight {
			return items[i], nil
		}
	}

	return items[len(items)-1], nil
}

var (
	// ErrEmptySlice is returned when slice is empty
	ErrEmptySlice = &RandomError{Message: "slice is empty"}
	// ErrWeightsMismatch is returned when items and weights length mismatch
	ErrWeightsMismatch = &RandomError{Message: "items and weights length mismatch"}
	// ErrNegativeWeight is returned when weight is negative
	ErrNegativeWeight = &RandomError{Message: "weight cannot be negative"}
	// ErrZeroTotalWeight is returned when total weight is zero
	ErrZeroTotalWeight = &RandomError{Message: "total weight is zero"}
)

// RandomError represents a random generation error
type RandomError struct {
	Message string
}

func (e *RandomError) Error() string {
	return e.Message
}
