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

package types_test

import (
	"encoding/json"
	"testing"

	"trellis.tech/trellis/common.v2/testutils"
	"trellis.tech/trellis/common.v2/types"
)

func Test_ToFloat64(t *testing.T) {
	f32 := float32(1.0)
	f, err := types.ToFloat64(f32)
	testutils.Equals(t, 1.0, f)
	testutils.Ok(t, err)
	f64 := float64(2.0)
	f, err = types.ToFloat64(f64)
	testutils.Equals(t, 2.0, f)
	testutils.Ok(t, err)

	var i = 10
	f, err = types.ToFloat64(i)
	testutils.Ok(t, err)
	testutils.Equals(t, 10.0, f)

	var s = "10.0"
	f, err = types.ToFloat64(s)
	testutils.Ok(t, err)
	testutils.Equals(t, 10.0, f)

	var js json.Number = "20.0"
	f, err = types.ToFloat64(js)
	testutils.Equals(t, 20.0, f)
	testutils.Ok(t, err)

	failed := map[string]string{}
	_, err = types.ToFloat64(failed)
	testutils.NotOk(t, err)
}

func Test_RoundFund(t *testing.T) {
	i := types.RoundFund(1.1)
	testutils.Equals(t, int64(1), i)
	i = types.RoundFund(1.5)
	testutils.Equals(t, int64(2), i)
}
