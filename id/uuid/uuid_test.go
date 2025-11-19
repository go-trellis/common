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
	"testing"

	"trellis.tech/trellis/common.v3/utils/testutils"
)

func TestNew(t *testing.T) {
	id := New()
	testutils.Assert(t, len(id) > 0, "should generate UUID")
	testutils.Assert(t, IsValid(id), "should be valid UUID")
}

func TestNewUUID(t *testing.T) {
	id := NewUUID()
	testutils.Assert(t, id.String() != "", "should generate UUID")
}

func TestParse(t *testing.T) {
	id := New()
	parsed, err := Parse(id)
	testutils.Ok(t, err)
	testutils.Equals(t, id, parsed.String(), "should parse correctly")
}

func TestParse_Invalid(t *testing.T) {
	_, err := Parse("invalid-uuid")
	testutils.NotOk(t, err, "should return error")
}

func TestMustParse(t *testing.T) {
	id := New()
	parsed := MustParse(id)
	testutils.Equals(t, id, parsed.String(), "should parse correctly")
}

func TestIsValid(t *testing.T) {
	testutils.Assert(t, IsValid(New()), "should be valid")
	testutils.Assert(t, IsValid("550e8400-e29b-41d4-a716-446655440000"), "should be valid")
	testutils.Assert(t, !IsValid("invalid"), "should be invalid")
	testutils.Assert(t, !IsValid(""), "should be invalid")
}

func TestNewString(t *testing.T) {
	id := NewString()
	testutils.Assert(t, IsValid(id), "should generate valid UUID")
}
