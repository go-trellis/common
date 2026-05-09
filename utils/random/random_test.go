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
	"testing"

	"github.com/go-trellis/common.v3/utils/testutils"
)

func TestString(t *testing.T) {
	s := String(10, "")
	testutils.Equals(t, 10, len(s), "should generate 10 chars")

	custom := String(5, "abc")
	testutils.Equals(t, 5, len(custom), "should generate 5 chars")
	for _, r := range custom {
		testutils.Assert(t, r == 'a' || r == 'b' || r == 'c', "should only contain a, b, c")
	}
}

func TestAlphaNumeric(t *testing.T) {
	s := AlphaNumeric(20)
	testutils.Equals(t, 20, len(s), "should generate 20 chars")
}

func TestNumeric(t *testing.T) {
	s := Numeric(10)
	testutils.Equals(t, 10, len(s), "should generate 10 digits")
	for _, r := range s {
		testutils.Assert(t, r >= '0' && r <= '9', "should only contain digits")
	}
}

func TestAlpha(t *testing.T) {
	s := Alpha(15)
	testutils.Equals(t, 15, len(s), "should generate 15 chars")
}

func TestInt(t *testing.T) {
	for i := 0; i < 100; i++ {
		n := Int(1, 10)
		testutils.Assert(t, n >= 1 && n <= 10, "should be in range")
	}
}

func TestInt64(t *testing.T) {
	for i := 0; i < 100; i++ {
		n := Int64(100, 200)
		testutils.Assert(t, n >= 100 && n <= 200, "should be in range")
	}
}

func TestFloat64(t *testing.T) {
	for i := 0; i < 100; i++ {
		f := Float64(0.0, 1.0)
		testutils.Assert(t, f >= 0.0 && f < 1.0, "should be in range")
	}
}

func TestBytes(t *testing.T) {
	b := Bytes(20)
	testutils.Equals(t, 20, len(b), "should generate 20 bytes")
}

func TestChoice(t *testing.T) {
	slice := []string{"a", "b", "c"}
	choice, err := Choice(slice)
	testutils.Ok(t, err)
	testutils.Assert(t, len(choice) > 0, "should choose an element")
}

func TestChoice_Empty(t *testing.T) {
	_, err := Choice([]string{})
	testutils.NotOk(t, err, "should return error for empty slice")
}

func TestChoices(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	result := Choices(slice, 3)
	testutils.Equals(t, 3, len(result), "should choose 3 elements")
}

func TestShuffle(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	result := Shuffle(slice)
	testutils.Equals(t, len(slice), len(result), "should have same length")
}

func TestWeightedChoice(t *testing.T) {
	items := []string{"a", "b", "c"}
	weights := []float64{0.1, 0.3, 0.6}
	choice, err := WeightedChoice(items, weights)
	testutils.Ok(t, err)
	testutils.Assert(t, len(choice) > 0, "should choose an item")
}

func TestWeightedChoice_Invalid(t *testing.T) {
	items := []string{"a", "b"}
	weights := []float64{0.5}
	_, err := WeightedChoice(items, weights)
	testutils.NotOk(t, err, "should return error for mismatched lengths")
}
