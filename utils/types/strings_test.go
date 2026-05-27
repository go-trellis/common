/*
Copyright © 2020 Henry Huang <hhh@rutcode.com>

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

package types

import (
	"testing"

	"github.com/go-trellis/common/utils/testutils"
)

func TestIntToString(t *testing.T) {
	testCases := []struct {
		input    any
		expected string
	}{
		{int8(42), "42"},
		{int16(42), "42"},
		{int(42), "42"},
		{int32(42), "42"},
		{int64(42), "42"},
		{uint(42), "42"},
		{uint8(42), "42"},
		{uint16(42), "42"},
		{uint32(42), "42"},
		{uint64(42), "42"},
		{nil, ""},
	}

	for _, tc := range testCases {
		result, err := IntToString(tc.input)
		if tc.input == nil {
			testutils.Ok(t, err)
			testutils.Equals(t, tc.expected, result, "IntToString(nil) should return empty string")
		} else {
			testutils.Ok(t, err)
			testutils.Equals(t, tc.expected, result, "IntToString should convert correctly")
		}
	}
}

func TestIntToString_InvalidType(t *testing.T) {
	_, err := IntToString("invalid")
	testutils.NotOk(t, err, "should return error for invalid type")
}

func TestHideString(t *testing.T) {
	testCases := []struct {
		origin   string
		start    int
		length   int
		expected string
	}{
		{"test123", 1, 4, "****123"},
		{"test123", 2, 2, "t**t123"},
		{"test", 1, 10, "****"},
		{"", 1, 5, ""},
		{"test", 0, 5, "test"},
		{"test", 1, 0, "test"},
	}

	for _, tc := range testCases {
		result := HideString(tc.origin, tc.start, tc.length)
		testutils.Equals(t, tc.expected, result, "HideString should hide correctly")
	}
}

func TestHideString_Short(t *testing.T) {
	result := HideString("test", 5, 2)
	testutils.Equals(t, "test", result, "HideString should return original if start > length")
}

func TestRemoveDuplicateStringByMap(t *testing.T) {
	input := []string{"a", "b", "c", "a", "d", "b"}
	result := RemoveDuplicateStringByMap(input)
	testutils.Equals(t, []string{"a", "b", "c", "d"}, result, "should remove duplicates")
}

func TestRemoveDuplicateStringByMap_Empty(t *testing.T) {
	result := RemoveDuplicateStringByMap([]string{})
	testutils.Assert(t, result == nil, "should return nil for empty slice")
}

func TestRemoveDuplicateStringByMap_Nil(t *testing.T) {
	result := RemoveDuplicateStringByMap(nil)
	testutils.Assert(t, result == nil, "should return nil for nil slice")
}

func TestStringInSlice(t *testing.T) {
	haystack := []string{"a", "b", "c"}
	testutils.Assert(t, StringInSlice("a", haystack), "should find string in slice")
	testutils.Assert(t, !StringInSlice("d", haystack), "should not find string not in slice")
	testutils.Assert(t, !StringInSlice("", haystack), "empty string not in slice should return false")
}

func TestStringInSlice_Empty(t *testing.T) {
	testutils.Assert(t, !StringInSlice("a", []string{}), "empty slice should return false")
}

func TestSuffixStringInSlice(t *testing.T) {
	haystack := []string{".log", ".txt", ".json"}
	testutils.Assert(t, SuffixStringInSlice("test.log", haystack), "should find suffix in slice")
	testutils.Assert(t, !SuffixStringInSlice("test.doc", haystack), "should not find suffix not in slice")
}

func TestStringContainedInSlice(t *testing.T) {
	haystack := []string{"test", "hello", "world"}
	testutils.Assert(t, StringContainedInSlice("test123", haystack), "should find contained string")
	testutils.Assert(t, !StringContainedInSlice("xyz", haystack), "should not find not contained string")
}
