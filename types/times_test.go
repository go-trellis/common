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
	"testing"
	"time"

	"trellis.tech/trellis/common.v0.1/testutils"
	"trellis.tech/trellis/common.v0.1/types"
)

func TestTimeFunctions(t *testing.T) {
	var (
		testTime  = time.Date(2016, 11, 30, 17, 33, 0, 0, &time.Location{})
		testZTime = time.Date(2016, 1, 1, 0, 0, 0, 999, &time.Location{})
		zeroTime  = time.Time{}
	)

	testutils.Equals(t, "2016-11-30", types.FormatDate(testTime))

	testutils.Equals(t, "2016年11月30日", types.FormatChineseZDate(testTime))
	testutils.Equals(t, "2016年1月1日", types.FormatChineseZDate(testZTime))

	testutils.Equals(t, "2016年11月30日", types.FormatChineseDate(testTime))
	testutils.Equals(t, "2016年01月01日", types.FormatChineseDate(testZTime))
	testutils.Equals(t, "2016年11月30日17时33分00秒", types.FormatChineseDateTime(testTime))
	testutils.Equals(t, "2016年1月1日00时0分0秒", types.FormatChineseZDateTime(testZTime))

	testutils.Equals(t, "2016-1-1", types.FormatZDate(testZTime))
	testutils.Equals(t, "17:33:00", types.FormatTime(testTime))
	testutils.Equals(t, "2016-11-30-17-33-00", types.FormatDashTime(testTime))
	testutils.Equals(t, "2016-01-01T00:00:00Z", types.FormatRFC3339(testZTime))
	testutils.Equals(t, "2016-01-01T00:00:00.000000999Z", types.FormatRFC3339Nano(testZTime))
	testutils.Equals(t, "Wed, 30 Nov 2016 17:33:00 GMT", types.FormatHTTPGMT(testTime))

	testutils.Equals(t, true, types.IsZero(zeroTime))
	testutils.Equals(t, false, types.IsZero(testTime))

	var year = 2019
	testutils.Equals(t, 31, types.GetMonthDays(year, 1))
	testutils.Equals(t, 31, types.GetMonthDays(year, 3))
	testutils.Equals(t, 31, types.GetMonthDays(year, 5))
	testutils.Equals(t, 31, types.GetMonthDays(year, 7))
	testutils.Equals(t, 31, types.GetMonthDays(year, 8))
	testutils.Equals(t, 31, types.GetMonthDays(year, 10))
	testutils.Equals(t, 31, types.GetMonthDays(year, 12))

	testutils.Equals(t, 30, types.GetMonthDays(year, 4))
	testutils.Equals(t, 30, types.GetMonthDays(year, 6))
	testutils.Equals(t, 30, types.GetMonthDays(year, 9))
	testutils.Equals(t, 30, types.GetMonthDays(year, 11))

	testutils.Equals(t, 28, types.GetMonthDays(year, 2))
	year = 2020
	testutils.Equals(t, 29, types.GetMonthDays(year, 2))
	testutils.Equals(t, 0, types.GetMonthDays(year, 13))

	testutils.Equals(t, 30, types.GetTimeMonthDays(testTime))
	testutils.Equals(t, 31, types.GetTimeMonthDays(testZTime))
}
