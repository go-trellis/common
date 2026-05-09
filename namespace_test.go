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

package common

import (
	"strings"
	"testing"

	"github.com/go-trellis/common.v3/utils/testutils"
)

func TestFormatNamespaceString(t *testing.T) {
	result := FormatNamespaceString("test")
	testutils.Assert(t, strings.Contains(result, "T:C"), "result should contain namespace")
	testutils.Assert(t, strings.Contains(result, "test"), "result should contain input data")
	testutils.Assert(t, strings.Contains(result, "=>"), "result should contain separator")
}

func TestFormatNamespaceString_Empty(t *testing.T) {
	result := FormatNamespaceString("")
	testutils.Assert(t, strings.Contains(result, "T:C"), "result should contain namespace")
	testutils.Assert(t, strings.Contains(result, "[]"), "result should contain empty brackets")
}

func TestFormatNamespaceString_SpecialChars(t *testing.T) {
	result := FormatNamespaceString("test@example.com")
	testutils.Assert(t, strings.Contains(result, "T:C"), "result should contain namespace")
	testutils.Assert(t, strings.Contains(result, "test@example.com"), "result should contain input data")
}

func TestFormatNamespaceString_Unicode(t *testing.T) {
	result := FormatNamespaceString("测试")
	testutils.Assert(t, strings.Contains(result, "T:C"), "result should contain namespace")
	testutils.Assert(t, strings.Contains(result, "测试"), "result should contain unicode data")
}

func TestFormatNamespaceString_Format(t *testing.T) {
	result := FormatNamespaceString("test")
	expected := "T:C=>[test]"
	testutils.Equals(t, expected, result, "format should match expected")
}
