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
	"testing"

	"github.com/go-trellis/common.v3/utils/testutils"
)

func TestIsEmail(t *testing.T) {
	testutils.Assert(t, IsEmail("test@example.com"), "should be valid email")
	testutils.Assert(t, IsEmail("user.name@example.co.uk"), "should be valid email")
	testutils.Assert(t, !IsEmail("invalid"), "should be invalid")
	testutils.Assert(t, !IsEmail("@example.com"), "should be invalid")
}

func TestIsURL(t *testing.T) {
	testutils.Assert(t, IsURL("http://example.com"), "should be valid URL")
	testutils.Assert(t, IsURL("https://example.com/path"), "should be valid URL")
	testutils.Assert(t, !IsURL("not-a-url"), "should be invalid")
	testutils.Assert(t, !IsURL("example.com"), "should be invalid")
}

func TestIsEmpty(t *testing.T) {
	testutils.Assert(t, IsEmpty(""), "empty string")
	testutils.Assert(t, IsEmpty("   "), "whitespace only")
	testutils.Assert(t, !IsEmpty("text"), "non-empty")
}

func TestLength(t *testing.T) {
	testutils.Assert(t, Length("hello", 1, 10), "should be in range")
	testutils.Assert(t, !Length("hello", 10, 20), "should be out of range")
}

func TestIsNumeric(t *testing.T) {
	testutils.Assert(t, IsNumeric("123"), "should be numeric")
	testutils.Assert(t, !IsNumeric("abc"), "should not be numeric")
	testutils.Assert(t, !IsNumeric(""), "empty should not be numeric")
}

func TestIsAlpha(t *testing.T) {
	testutils.Assert(t, IsAlpha("abc"), "should be alpha")
	testutils.Assert(t, !IsAlpha("123"), "should not be alpha")
	testutils.Assert(t, !IsAlpha(""), "empty should not be alpha")
}

func TestIsAlphaNumeric(t *testing.T) {
	testutils.Assert(t, IsAlphaNumeric("abc123"), "should be alphanumeric")
	testutils.Assert(t, !IsAlphaNumeric("abc-123"), "should not be alphanumeric")
	testutils.Assert(t, !IsAlphaNumeric(""), "empty should not be alphanumeric")
}

func TestIn(t *testing.T) {
	testutils.Assert(t, In(1, 1, 2, 3), "should be in list")
	testutils.Assert(t, !In(4, 1, 2, 3), "should not be in list")
}

func TestBetween(t *testing.T) {
	testutils.Assert(t, Between(5, 1, 10), "should be between")
	testutils.Assert(t, !Between(15, 1, 10), "should not be between")
}

func TestValidate(t *testing.T) {
	err := Validate("", RequiredString)
	testutils.NotOk(t, err)

	err = Validate("test", RequiredString)
	testutils.Ok(t, err)
}
