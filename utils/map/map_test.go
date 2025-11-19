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

package maputil

import (
	"testing"

	"trellis.tech/trellis/common.v3/utils/testutils"
)

func TestKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	keys := Keys(m)
	testutils.Equals(t, 3, len(keys), "should have 3 keys")
}

func TestValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	values := Values(m)
	testutils.Equals(t, 3, len(values), "should have 3 values")
}

func TestContainsKey(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	testutils.Assert(t, ContainsKey(m, "a"), "should contain key")
	testutils.Assert(t, !ContainsKey(m, "c"), "should not contain key")
}

func TestGet(t *testing.T) {
	m := map[string]int{"a": 1}
	val, ok := Get(m, "a")
	testutils.Assert(t, ok, "should exist")
	testutils.Equals(t, 1, val, "should return value")

	_, ok = Get(m, "b")
	testutils.Assert(t, !ok, "should not exist")
}

func TestGetOrDefault(t *testing.T) {
	m := map[string]int{"a": 1}
	testutils.Equals(t, 1, GetOrDefault(m, "a", 0), "should return value")
	testutils.Equals(t, 10, GetOrDefault(m, "b", 10), "should return default")
}

func TestFilter(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	result := Filter(m, func(k string, v int) bool {
		return v > 1
	})
	testutils.Equals(t, 2, len(result), "should filter correctly")
}

func TestMap(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	result := Map(m, func(k string, v int) int {
		return v * 2
	})
	testutils.Equals(t, 2, result["a"], "should map values")
}

func TestMerge(t *testing.T) {
	m1 := map[string]int{"a": 1}
	m2 := map[string]int{"b": 2}
	m3 := map[string]int{"a": 3}
	result := Merge(m1, m2, m3)
	testutils.Equals(t, 3, result["a"], "later should override")
	testutils.Equals(t, 2, result["b"], "should merge")
}

func TestGroupBy(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	result := GroupBy(slice, func(x int) string {
		if x%2 == 0 {
			return "even"
		}
		return "odd"
	})
	testutils.Equals(t, 2, len(result), "should have 2 groups")
	testutils.Equals(t, 3, len(result["odd"]), "should have 3 odd numbers")
}

func TestIndexBy(t *testing.T) {
	type Item struct {
		ID   int
		Name string
	}
	items := []Item{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}}
	result := IndexBy(items, func(item Item) int {
		return item.ID
	})
	testutils.Equals(t, "a", result[1].Name, "should index correctly")
}

func TestClone(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	clone := Clone(m)
	testutils.Equals(t, m["a"], clone["a"], "should clone")

	m["a"] = 10
	testutils.Equals(t, 1, clone["a"], "clone should be independent")
}

func TestEqual(t *testing.T) {
	m1 := map[string]int{"a": 1, "b": 2}
	m2 := map[string]int{"a": 1, "b": 2}
	m3 := map[string]int{"a": 1, "b": 3}
	testutils.Assert(t, Equal(m1, m2), "should be equal")
	testutils.Assert(t, !Equal(m1, m3), "should not be equal")
}

func TestPick(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	result := Pick(m, []string{"a", "c"})
	testutils.Equals(t, 2, len(result), "should pick 2 keys")
	testutils.Equals(t, 1, result["a"], "should pick a")
}

func TestOmit(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	result := Omit(m, []string{"a", "c"})
	testutils.Equals(t, 1, len(result), "should omit 2 keys")
	testutils.Equals(t, 2, result["b"], "should keep b")
}
