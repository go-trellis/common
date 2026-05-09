/*
Copyright © 2022 Henry Huang <hhh@rutcode.com>

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
	"flag"
	"sort"
	"testing"

	"github.com/go-trellis/common.v3/utils/testutils"
)

func TestStrings_Len(t *testing.T) {
	ss := Strings{"a", "b", "c"}
	testutils.Equals(t, 3, ss.Len(), "Len should return correct length")
}

func TestStrings_Empty(t *testing.T) {
	ss := Strings{}
	testutils.Equals(t, 0, ss.Len(), "Empty Strings should have length 0")
}

func TestStrings_Less(t *testing.T) {
	ss := Strings{"a", "b", "c"}
	testutils.Assert(t, ss.Less(0, 1), "a should be less than b")
	testutils.Assert(t, !ss.Less(1, 0), "b should not be less than a")
}

func TestStrings_Swap(t *testing.T) {
	ss := Strings{"a", "b", "c"}
	ss.Swap(0, 2)
	testutils.Equals(t, Strings{"c", "b", "a"}, ss, "Swap should swap elements")
}

func TestStrings_String(t *testing.T) {
	ss := Strings{"a", "b", "c"}
	str := ss.String()
	testutils.Assert(t, len(str) > 0, "String should not be empty")
}

func TestStrings_Set(t *testing.T) {
	var ss Strings
	err := ss.Set("a")
	testutils.Ok(t, err)
	testutils.Assert(t, len(ss) == 1, "Set should add element")
	testutils.Equals(t, "a", ss[0], "Set should add correct element")

	err = ss.Set("b")
	testutils.Ok(t, err)
	testutils.Assert(t, len(ss) == 2, "Set should append element")
}

func TestStrings_Sort(t *testing.T) {
	ss := Strings{"c", "a", "b"}
	sort.Sort(ss)
	testutils.Equals(t, Strings{"a", "b", "c"}, ss, "should sort correctly")
}

func TestStrings_FlagValue(t *testing.T) {
	var ss Strings
	testutils.Assert(t, flag.Value(&ss) != nil, "Strings should implement flag.Value")

	ss.Set("flag1")
	ss.Set("flag2")
	testutils.Assert(t, len(ss) == 2, "should have 2 elements")
}
