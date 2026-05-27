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

package slice

import (
	"testing"

	"github.com/go-trellis/common/utils/testutils"
)

func TestContains(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	testutils.Assert(t, Contains(slice, 3), "should contain 3")
	testutils.Assert(t, !Contains(slice, 6), "should not contain 6")
}

func TestIndex(t *testing.T) {
	slice := []string{"a", "b", "c"}
	testutils.Equals(t, 1, Index(slice, "b"), "should find index")
	testutils.Equals(t, -1, Index(slice, "d"), "should return -1 if not found")
}

func TestFilter(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	result := Filter(slice, func(x int) bool {
		return x%2 == 0
	})
	testutils.Equals(t, []int{2, 4}, result, "should filter even numbers")
}

func TestMap(t *testing.T) {
	slice := []int{1, 2, 3}
	result := Map(slice, func(x int) int {
		return x * 2
	})
	testutils.Equals(t, []int{2, 4, 6}, result, "should map correctly")
}

func TestReduce(t *testing.T) {
	slice := []int{1, 2, 3, 4}
	result := Reduce(slice, 0, func(acc, x int) int {
		return acc + x
	})
	testutils.Equals(t, 10, result, "should sum correctly")
}

func TestRemove(t *testing.T) {
	slice := []int{1, 2, 3, 2, 4}
	result := Remove(slice, 2)
	testutils.Equals(t, []int{1, 3, 2, 4}, result, "should remove first occurrence")
}

func TestRemoveAll(t *testing.T) {
	slice := []int{1, 2, 3, 2, 4}
	result := RemoveAll(slice, 2)
	testutils.Equals(t, []int{1, 3, 4}, result, "should remove all occurrences")
}

func TestUnique(t *testing.T) {
	slice := []int{1, 2, 2, 3, 3, 3}
	result := Unique(slice)
	testutils.Equals(t, []int{1, 2, 3}, result, "should remove duplicates")
}

func TestChunk(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6, 7}
	chunks := Chunk(slice, 3)
	testutils.Equals(t, 3, len(chunks), "should create 3 chunks")
	testutils.Equals(t, []int{1, 2, 3}, chunks[0], "first chunk")
	testutils.Equals(t, []int{4, 5, 6}, chunks[1], "second chunk")
	testutils.Equals(t, []int{7}, chunks[2], "third chunk")
}

func TestReverse(t *testing.T) {
	slice := []int{1, 2, 3, 4}
	result := Reverse(slice)
	testutils.Equals(t, []int{4, 3, 2, 1}, result, "should reverse")
	testutils.Equals(t, []int{1, 2, 3, 4}, slice, "original should not change")
}

func TestReverseInPlace(t *testing.T) {
	slice := []int{1, 2, 3, 4}
	ReverseInPlace(slice)
	testutils.Equals(t, []int{4, 3, 2, 1}, slice, "should reverse in place")
}

func TestIntersect(t *testing.T) {
	slice1 := []int{1, 2, 3, 4}
	slice2 := []int{3, 4, 5, 6}
	result := Intersect(slice1, slice2)
	testutils.Equals(t, []int{3, 4}, result, "should find intersection")
}

func TestUnion(t *testing.T) {
	slice1 := []int{1, 2, 3}
	slice2 := []int{3, 4, 5}
	result := Union(slice1, slice2)
	testutils.Equals(t, 5, len(result), "should have 5 unique elements")
}

func TestDifference(t *testing.T) {
	slice1 := []int{1, 2, 3, 4}
	slice2 := []int{3, 4}
	result := Difference(slice1, slice2)
	testutils.Equals(t, []int{1, 2}, result, "should find difference")
}

func TestAny(t *testing.T) {
	slice := []int{1, 2, 3, 4}
	testutils.Assert(t, Any(slice, func(x int) bool { return x > 3 }), "should find any")
	testutils.Assert(t, !Any(slice, func(x int) bool { return x > 5 }), "should not find any")
}

func TestAll(t *testing.T) {
	slice := []int{2, 4, 6}
	testutils.Assert(t, All(slice, func(x int) bool { return x%2 == 0 }), "all should be even")
	testutils.Assert(t, !All([]int{2, 3, 4}, func(x int) bool { return x%2 == 0 }), "not all should be even")
}

func TestFirst(t *testing.T) {
	slice := []int{1, 2, 3, 4}
	val, ok := First(slice, func(x int) bool { return x > 2 })
	testutils.Assert(t, ok, "should find")
	testutils.Equals(t, 3, val, "should return first match")
}

func TestCount(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	count := Count(slice, func(x int) bool { return x%2 == 0 })
	testutils.Equals(t, 2, count, "should count 2 even numbers")
}

func TestFlatten(t *testing.T) {
	slices := [][]int{{1, 2}, {3, 4}, {5}}
	result := Flatten(slices)
	testutils.Equals(t, []int{1, 2, 3, 4, 5}, result, "should flatten")
}

func TestPartition(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	evens, odds := Partition(slice, func(x int) bool { return x%2 == 0 })
	testutils.Equals(t, []int{2, 4}, evens, "evens")
	testutils.Equals(t, []int{1, 3, 5}, odds, "odds")
}

func TestTake(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	result := Take(slice, 3)
	testutils.Equals(t, []int{1, 2, 3}, result, "should take 3")
}

func TestDrop(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	result := Drop(slice, 2)
	testutils.Equals(t, []int{3, 4, 5}, result, "should drop 2")
}

func TestZip(t *testing.T) {
	slice1 := []int{1, 2, 3}
	slice2 := []string{"a", "b", "c"}
	pairs := Zip(slice1, slice2)
	testutils.Equals(t, 3, len(pairs), "should create 3 pairs")
	testutils.Equals(t, 1, pairs[0].First, "first pair")
	testutils.Equals(t, "a", pairs[0].Second, "first pair")
}

func TestUnzip(t *testing.T) {
	pairs := []Pair[int, string]{
		{First: 1, Second: "a"},
		{First: 2, Second: "b"},
	}
	firsts, seconds := Unzip(pairs)
	testutils.Equals(t, []int{1, 2}, firsts, "firsts")
	testutils.Equals(t, []string{"a", "b"}, seconds, "seconds")
}

func TestSum(t *testing.T) {
	slice := []int{1, 2, 3, 4}
	testutils.Equals(t, 10, Sum(slice), "should sum")
}

func TestMax(t *testing.T) {
	slice := []int{1, 5, 3, 2}
	max, err := Max(slice)
	testutils.Ok(t, err)
	testutils.Equals(t, 5, max, "should find max")
}

func TestMin(t *testing.T) {
	slice := []int{5, 1, 3, 2}
	min, err := Min(slice)
	testutils.Ok(t, err)
	testutils.Equals(t, 1, min, "should find min")
}

func TestAverage(t *testing.T) {
	slice := []int{1, 2, 3, 4}
	avg, err := Average(slice)
	testutils.Ok(t, err)
	testutils.Equals(t, 2.5, avg, "should calculate average")
}
