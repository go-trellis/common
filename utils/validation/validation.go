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

package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// EmailRegex is the regex pattern for email validation
var EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// IsEmail validates email address
func IsEmail(email string) bool {
	return EmailRegex.MatchString(email)
}

// IsURL validates URL
func IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// IsEmpty checks if string is empty (trimmed)
func IsEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}

// IsNotEmpty checks if string is not empty
func IsNotEmpty(str string) bool {
	return !IsEmpty(str)
}

// Length checks if string length is within range
func Length(str string, min, max int) bool {
	length := len([]rune(str))
	return length >= min && length <= max
}

// MinLength checks if string has minimum length
func MinLength(str string, min int) bool {
	return len([]rune(str)) >= min
}

// MaxLength checks if string has maximum length
func MaxLength(str string, max int) bool {
	return len([]rune(str)) <= max
}

// Matches checks if string matches regex pattern
func Matches(str, pattern string) bool {
	matched, err := regexp.MatchString(pattern, str)
	return err == nil && matched
}

// IsNumeric checks if string contains only digits
func IsNumeric(str string) bool {
	if str == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^\d+$`, str)
	return matched
}

// IsAlpha checks if string contains only letters
func IsAlpha(str string) bool {
	if str == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z]+$`, str)
	return matched
}

// IsAlphaNumeric checks if string contains only letters and digits
func IsAlphaNumeric(str string) bool {
	if str == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, str)
	return matched
}

// IsInt checks if string is a valid integer
func IsInt(str string) bool {
	if str == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^-?\d+$`, str)
	return matched
}

// IsFloat checks if string is a valid float
func IsFloat(str string) bool {
	if str == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^-?\d+(\.\d+)?$`, str)
	return matched
}

// In checks if value is in the provided list
func In[T comparable](value T, list ...T) bool {
	for _, v := range list {
		if value == v {
			return true
		}
	}
	return false
}

// NotIn checks if value is not in the provided list
func NotIn[T comparable](value T, list ...T) bool {
	return !In(value, list...)
}

// Between checks if numeric value is between min and max (inclusive)
func Between[T Number](value, min, max T) bool {
	return value >= min && value <= max
}

// Number is a type constraint for numeric types
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

// Validator defines validation function
type Validator[T any] func(T) error

// Validate validates value using validator function
func Validate[T any](value T, validators ...Validator[T]) error {
	for _, v := range validators {
		if err := v(value); err != nil {
			return err
		}
	}
	return nil
}

// Required validates that value is not zero/nil/empty
func Required[T comparable](value T, zero T) error {
	if value == zero {
		return fmt.Errorf("value is required")
	}
	return nil
}

// RequiredString validates that string is not empty
func RequiredString(str string) error {
	if IsEmpty(str) {
		return fmt.Errorf("string is required")
	}
	return nil
}

// ValidatorFunc is a helper to create validator from boolean function
func ValidatorFunc[T any](fn func(T) bool, message string) Validator[T] {
	return func(value T) error {
		if !fn(value) {
			return fmt.Errorf("%s", message)
		}
		return nil
	}
}
