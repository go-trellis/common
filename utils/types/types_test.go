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

	"github.com/go-trellis/common.v3/utils/testutils"
)

func TestReduplicate_Ints(t *testing.T) {
	input := []int{1, 2, 3, 2, 4, 1, 5}
	result := Reduplicate(input)
	resultSlice := result.([]int)
	testutils.Assert(t, len(resultSlice) == 5, "should have 5 unique elements")
	testutils.Equals(t, []int{1, 2, 3, 4, 5}, resultSlice, "should remove duplicates")
}

func TestReduplicate_Strings(t *testing.T) {
	input := []string{"a", "b", "c", "a", "d", "b"}
	result := Reduplicate(input)
	resultSlice := result.([]string)
	testutils.Assert(t, len(resultSlice) == 4, "should have 4 unique elements")
	testutils.Equals(t, []string{"a", "b", "c", "d"}, resultSlice, "should remove duplicates")
}

func TestReduplicate_Empty(t *testing.T) {
	input := []int{}
	result := Reduplicate(input)
	resultSlice := result.([]int)
	testutils.Assert(t, len(resultSlice) == 0, "should return empty slice")
}

func TestReduplicate_NotSlice(t *testing.T) {
	input := "not a slice"
	result := Reduplicate(input)
	testutils.Equals(t, "not a slice", result, "should return original value for non-slice")
}

func TestReduplicate_Nil(t *testing.T) {
	var input []int = nil
	result := Reduplicate(input)
	// Reduplicate may return empty slice or nil for nil input
	resultSlice, ok := result.([]int)
	testutils.Assert(t, ok, "should return []int type")
	_ = resultSlice // Accept both nil and empty slice
}
