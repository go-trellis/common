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

package uuid

import (
	"github.com/google/uuid"
)

// New generates a new random UUID
func New() string {
	return uuid.New().String()
}

// NewUUID generates a new UUID and returns uuid.UUID
func NewUUID() uuid.UUID {
	return uuid.New()
}

// Parse parses UUID string
func Parse(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// MustParse parses UUID string or panics
func MustParse(s string) uuid.UUID {
	return uuid.MustParse(s)
}

// IsValid checks if string is a valid UUID
func IsValid(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// NewString returns a new UUID as string (alias for New)
func NewString() string {
	return New()
}
